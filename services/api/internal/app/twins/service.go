// Package twins implements the twin lifecycle use cases: create (which
// kicks off an async build), start/stop/destroy, and the event feed the
// transport layer streams to the browser over SSE
// (docs/MASTER_PLAN.md §5.3 "app" layer, §5.7).
package twins

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type Service struct {
	repo     ports.TwinRepository
	provider ports.LabProvider
	audit    ports.AuditRepository

	mu   sync.Mutex
	subs map[uuid.UUID]map[chan twin.Event]struct{} // twinID -> live SSE subscribers
}

func NewService(repo ports.TwinRepository, provider ports.LabProvider, audit ports.AuditRepository) *Service {
	return &Service{
		repo: repo, provider: provider, audit: audit,
		subs: map[uuid.UUID]map[chan twin.Event]struct{}{},
	}
}

// Create records a new twin and starts its build in the background,
// returning as soon as the Requested record exists — the caller (an HTTP
// handler) does not block on the build. snapshotID is not yet validated
// against a real snapshots repository: Assets/Snapshots upload is still
// unbuilt (docs/PROGRESS.md), so any UUID is accepted for now.
func (s *Service) Create(ctx context.Context, actorID, workspaceID, snapshotID uuid.UUID) (twin.Twin, error) {
	t, err := s.repo.Create(ctx, workspaceID, snapshotID, "fake")
	if err != nil {
		return twin.Twin{}, fmt.Errorf("twins: create: %w", err)
	}

	if _, err := s.audit.Record(ctx, workspaceID, &actorID, "twin.create", "twin", t.ID.String(), nil); err != nil {
		return twin.Twin{}, fmt.Errorf("twins: audit create: %w", err)
	}

	// The build must outlive this request: detach from the request's
	// context (which is cancelled the moment the HTTP handler returns)
	// but keep it otherwise, so a build started by one request isn't
	// silently killed when that response is written.
	go s.runBuild(context.WithoutCancel(ctx), t)

	return t, nil
}

func (s *Service) runBuild(ctx context.Context, t twin.Twin) {
	current := t.State
	step := func(ctx context.Context, to twin.State, message string) error {
		if !twin.CanTransition(current, to) {
			return fmt.Errorf("twins: provider requested invalid transition %s -> %s", current, to)
		}
		updated, err := s.repo.UpdateState(ctx, t.ID, to)
		if err != nil {
			return err
		}
		current = updated.State
		s.recordEvent(ctx, t.ID, to, message)
		return nil
	}

	s.provider.Build(ctx, t.ID, step)

	// A build that didn't reach Ready (cancelled, or the provider stopped
	// early on a step error) must not be left in a non-terminal,
	// non-"ready" limbo forever — mark it Failed so the console can show
	// that and a retry path (create a new twin) is the only way forward,
	// per the state machine's "no resuming a failed build" rule
	// (internal/domain/twin).
	if current != twin.StateReady && twin.CanTransition(current, twin.StateFailed) {
		if _, err := s.repo.UpdateState(ctx, t.ID, twin.StateFailed); err == nil {
			s.recordEvent(ctx, t.ID, twin.StateFailed, "build did not complete")
		}
	}
}

func (s *Service) recordEvent(ctx context.Context, twinID uuid.UUID, state twin.State, message string) {
	ev, err := s.repo.AppendEvent(ctx, twin.Event{TwinID: twinID, State: state, Message: message})
	if err != nil {
		return
	}
	s.publish(ev)
}

func (s *Service) Get(ctx context.Context, workspaceID, twinID uuid.UUID) (twin.Twin, error) {
	return s.repo.Get(ctx, workspaceID, twinID)
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]twin.Twin, error) {
	return s.repo.List(ctx, workspaceID)
}

// Events returns everything recorded for twinID after afterSequence, for a
// reconnecting SSE client to catch up before it starts live-streaming
// (docs/MASTER_PLAN.md §5.1: event-sourced runs/builds).
func (s *Service) Events(ctx context.Context, twinID uuid.UUID, afterSequence int) ([]twin.Event, error) {
	return s.repo.ListEvents(ctx, twinID, afterSequence)
}

func (s *Service) Start(ctx context.Context, actorID, workspaceID, twinID uuid.UUID) (twin.Twin, error) {
	return s.transition(ctx, actorID, workspaceID, twinID, twin.StateRunning, s.provider.Start, "twin.start")
}

func (s *Service) Stop(ctx context.Context, actorID, workspaceID, twinID uuid.UUID) (twin.Twin, error) {
	return s.transition(ctx, actorID, workspaceID, twinID, twin.StateStopped, s.provider.Stop, "twin.stop")
}

func (s *Service) Destroy(ctx context.Context, actorID, workspaceID, twinID uuid.UUID) (twin.Twin, error) {
	return s.transition(ctx, actorID, workspaceID, twinID, twin.StateDestroyed, s.provider.Destroy, "twin.destroy")
}

var ErrInvalidTransition = fmt.Errorf("twins: invalid state transition")

func (s *Service) transition(ctx context.Context, actorID, workspaceID, twinID uuid.UUID, to twin.State, op func(context.Context, uuid.UUID) error, action string) (twin.Twin, error) {
	current, err := s.repo.Get(ctx, workspaceID, twinID)
	if err != nil {
		return twin.Twin{}, err
	}
	if !twin.CanTransition(current.State, to) {
		return twin.Twin{}, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current.State, to)
	}
	if err := op(ctx, twinID); err != nil {
		return twin.Twin{}, err
	}
	updated, err := s.repo.UpdateState(ctx, twinID, to)
	if err != nil {
		return twin.Twin{}, err
	}
	if _, err := s.audit.Record(ctx, workspaceID, &actorID, action, "twin", twinID.String(), nil); err != nil {
		return twin.Twin{}, err
	}
	s.recordEvent(ctx, twinID, to, action)
	return updated, nil
}

// Subscribe registers ch to receive every future event for twinID until
// cancel is called. The channel is buffered and publish never blocks on a
// slow/absent reader (a dropped event is recoverable — the client can
// always call Events with its last-seen sequence to catch up — a stalled
// build is not).
func (s *Service) Subscribe(twinID uuid.UUID) (ch <-chan twin.Event, cancel func()) {
	c := make(chan twin.Event, 16)
	s.mu.Lock()
	if s.subs[twinID] == nil {
		s.subs[twinID] = map[chan twin.Event]struct{}{}
	}
	s.subs[twinID][c] = struct{}{}
	s.mu.Unlock()

	return c, func() {
		s.mu.Lock()
		delete(s.subs[twinID], c)
		s.mu.Unlock()
	}
}

func (s *Service) publish(e twin.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for c := range s.subs[e.TwinID] {
		select {
		case c <- e:
		default:
		}
	}
}

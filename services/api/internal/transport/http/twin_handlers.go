package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	apptwins "github.com/sshreyx1/cyberditto/services/api/internal/app/twins"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

type twinHandlers struct {
	svc *apptwins.Service
}

func twinResponse(t twin.Twin) map[string]any {
	return map[string]any{
		"id":             t.ID,
		"workspace_id":   t.WorkspaceID,
		"snapshot_id":    t.SnapshotID,
		"provider":       t.Provider,
		"state":          t.State,
		"fidelity_score": t.FidelityScore,
		"created_at":     t.CreatedAt,
		"updated_at":     t.UpdatedAt,
	}
}

type createTwinRequest struct {
	// SnapshotID is optional in Phase 1: Assets/Snapshots upload doesn't
	// exist yet (docs/PROGRESS.md), so a caller that omits it gets a
	// synthetic one rather than being forced to fabricate a real-looking
	// UUID client-side.
	SnapshotID string `json:"snapshot_id,omitempty"`
}

func (h *twinHandlers) create(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())

	var req createTwinRequest
	if r.ContentLength != 0 {
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			writeProblem(w, http.StatusBadRequest, "invalid request body", err.Error())
			return
		}
	}

	snapshotID := uuid.New()
	if req.SnapshotID != "" {
		id, err := uuid.Parse(req.SnapshotID)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "invalid snapshot_id", "must be a UUID")
			return
		}
		snapshotID = id
	}

	t, err := h.svc.Create(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, snapshotID)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "create twin failed", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(twinResponse(t))
}

func (h *twinHandlers) list(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())

	ts, err := h.svc.List(r.Context(), ac.Membership.WorkspaceID)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "list twins failed", "")
		return
	}

	out := make([]map[string]any, 0, len(ts))
	for _, t := range ts {
		out = append(out, twinResponse(t))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"twins": out})
}

func (h *twinHandlers) get(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid twin id", "")
		return
	}

	t, err := h.svc.Get(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		writeProblem(w, http.StatusNotFound, "twin not found", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(twinResponse(t))
}

type twinOp func(ctx context.Context, actorID, workspaceID, twinID uuid.UUID) (twin.Twin, error)

func (h *twinHandlers) start(w http.ResponseWriter, r *http.Request) {
	h.doTransition(w, r, h.svc.Start)
}
func (h *twinHandlers) stop(w http.ResponseWriter, r *http.Request) { h.doTransition(w, r, h.svc.Stop) }
func (h *twinHandlers) destroy(w http.ResponseWriter, r *http.Request) {
	h.doTransition(w, r, h.svc.Destroy)
}

func (h *twinHandlers) doTransition(w http.ResponseWriter, r *http.Request, op twinOp) {
	ac := authFromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid twin id", "")
		return
	}

	t, err := op(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, id)
	if err != nil {
		if errors.Is(err, apptwins.ErrInvalidTransition) {
			writeProblem(w, http.StatusConflict, "invalid state transition", err.Error())
			return
		}
		writeProblem(w, http.StatusNotFound, "twin not found", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(twinResponse(t))
}

// events streams a twin's build/lifecycle progress to the browser over SSE
// (docs/MASTER_PLAN.md §5.1: "event-sourced where it matters... UI streams
// them via SSE"). It first replays anything already recorded, then streams
// new events live, deduplicating by sequence number so an event published
// in the gap between subscribing and the catch-up read is never sent twice.
func (h *twinHandlers) events(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid twin id", "")
		return
	}
	if _, err := h.svc.Get(r.Context(), ac.Membership.WorkspaceID, id); err != nil {
		writeProblem(w, http.StatusNotFound, "twin not found", "")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeProblem(w, http.StatusInternalServerError, "streaming unsupported", "")
		return
	}

	// Subscribe before the catch-up read, not after: subscribing second
	// would let an event published in between be missed entirely, rather
	// than merely duplicated (which lastSeq below already guards against).
	live, cancel := h.svc.Subscribe(id)
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	lastSeq := 0
	if catchUp, err := h.svc.Events(r.Context(), id, 0); err == nil {
		for _, ev := range catchUp {
			writeSSEEvent(w, ev)
			lastSeq = ev.Sequence
		}
		flusher.Flush()
	}

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-live:
			if !ok {
				return
			}
			if ev.Sequence <= lastSeq {
				continue
			}
			writeSSEEvent(w, ev)
			lastSeq = ev.Sequence
			flusher.Flush()
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, ev twin.Event) {
	data, err := json.Marshal(map[string]any{
		"sequence":   ev.Sequence,
		"state":      ev.State,
		"message":    ev.Message,
		"created_at": ev.CreatedAt,
	})
	if err != nil {
		return
	}
	fmt.Fprintf(w, "id: %d\ndata: %s\n\n", ev.Sequence, data)
}

package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	apptwins "github.com/sshreyx1/cyberditto/services/api/internal/app/twins"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
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

func writeTwinError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apptwins.ErrInvalidTransition):
		writeProblem(w, http.StatusConflict, "invalid state transition", err.Error())
	case errors.Is(err, apptwins.ErrSnapshotNotFound):
		writeProblem(w, http.StatusUnprocessableEntity, "snapshot not found", "snapshot_id does not reference a snapshot in this workspace")
	case errors.Is(err, ports.ErrNotFound):
		writeProblem(w, http.StatusNotFound, "twin not found", "")
	default:
		writeProblem(w, http.StatusInternalServerError, "request failed", "")
	}
}

type createTwinRequest struct {
	SnapshotID string `json:"snapshot_id"`
}

func (h *twinHandlers) create(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())

	var req createTwinRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	snapshotID, err := uuid.Parse(req.SnapshotID)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid snapshot_id", "snapshot_id is required and must be a UUID")
		return
	}

	t, err := h.svc.Create(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, snapshotID)
	if err != nil {
		writeTwinError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, twinResponse(t))
}

func (h *twinHandlers) list(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	ts, err := h.svc.List(r.Context(), ac.Membership.WorkspaceID)
	if err != nil {
		writeTwinError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(ts))
	for _, t := range ts {
		out = append(out, twinResponse(t))
	}
	writeJSON(w, http.StatusOK, map[string]any{"twins": out})
}

func (h *twinHandlers) get(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	t, err := h.svc.Get(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		writeTwinError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, twinResponse(t))
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
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	t, err := op(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, id)
	if err != nil {
		writeTwinError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, twinResponse(t))
}

// events streams a twin's lifecycle events over SSE: it replays what's
// recorded after the client's Last-Event-ID (EventSource sends it
// automatically on reconnect), then streams live events, deduplicating by
// sequence so an event published between subscribing and the catch-up read
// is never sent twice.
func (h *twinHandlers) events(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	workspaceID := ac.Membership.WorkspaceID
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Get(r.Context(), workspaceID, id); err != nil {
		writeTwinError(w, err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeProblem(w, http.StatusInternalServerError, "streaming unsupported", "")
		return
	}

	lastSeq := 0
	if v := r.Header.Get("Last-Event-ID"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			lastSeq = n
		}
	}

	// Subscribe before the catch-up read so nothing published in between
	// is missed (duplicates are filtered by lastSeq below).
	live, cancel := h.svc.Subscribe(id)
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	if catchUp, err := h.svc.Events(r.Context(), workspaceID, id, lastSeq); err == nil {
		for _, ev := range catchUp {
			writeSSEEvent(w, ev)
			lastSeq = ev.Sequence
		}
	}
	flusher.Flush()

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

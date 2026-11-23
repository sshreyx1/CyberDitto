package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	appassets "github.com/sshreyx1/cyberditto/services/api/internal/app/assets"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

// snapshotDigestHeader carries the SHA-256 the collector printed for the
// file being uploaded, so the server can detect a corrupted/truncated
// upload. Optional.
const snapshotDigestHeader = "X-Snapshot-SHA256"

type assetHandlers struct {
	svc *appassets.Service
}

func assetResponse(a asset.Asset) map[string]any {
	return map[string]any{
		"id":           a.ID,
		"workspace_id": a.WorkspaceID,
		"name":         a.Name,
		"created_at":   a.CreatedAt,
		"updated_at":   a.UpdatedAt,
	}
}

func snapshotResponse(s asset.Snapshot) map[string]any {
	out := map[string]any{
		"id":         s.ID,
		"asset_id":   s.AssetID,
		"version":    s.Version,
		"sha256":     s.SHA256,
		"created_at": s.CreatedAt,
	}
	if s.Data != nil {
		out["data"] = s.Data
	}
	return out
}

// writeAssetError maps app/domain errors to HTTP status codes in one place.
func writeAssetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ports.ErrNotFound):
		writeProblem(w, http.StatusNotFound, "not found", "")
	case errors.Is(err, asset.ErrInvalidName):
		writeProblem(w, http.StatusBadRequest, "invalid asset name", err.Error())
	case errors.Is(err, asset.ErrInvalidDocument):
		writeProblem(w, http.StatusBadRequest, "invalid snapshot", err.Error())
	case errors.Is(err, appassets.ErrDigestMismatch):
		writeProblem(w, http.StatusUnprocessableEntity, "snapshot digest mismatch", err.Error())
	default:
		writeProblem(w, http.StatusInternalServerError, "request failed", "")
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func pathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid "+name, "must be a UUID")
		return uuid.Nil, false
	}
	return id, true
}

type createAssetRequest struct {
	Name string `json:"name"`
}

func (h *assetHandlers) create(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	var req createAssetRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	a, err := h.svc.Create(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, req.Name)
	if err != nil {
		writeAssetError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, assetResponse(a))
}

func (h *assetHandlers) list(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	as, err := h.svc.List(r.Context(), ac.Membership.WorkspaceID)
	if err != nil {
		writeAssetError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(as))
	for _, a := range as {
		out = append(out, assetResponse(a))
	}
	writeJSON(w, http.StatusOK, map[string]any{"assets": out})
}

func (h *assetHandlers) get(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	a, err := h.svc.Get(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		writeAssetError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, assetResponse(a))
}

// uploadSnapshot takes the collector's JSON output as the raw request
// body, exactly as written to disk — no wrapping envelope — so the digest
// the collector printed can be checked against the received bytes.
func (h *assetHandlers) uploadSnapshot(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeProblem(w, http.StatusRequestEntityTooLarge, "snapshot too large", "")
			return
		}
		writeProblem(w, http.StatusBadRequest, "could not read request body", "")
		return
	}
	snap, err := h.svc.UploadSnapshot(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, id, body, r.Header.Get(snapshotDigestHeader))
	if err != nil {
		writeAssetError(w, err)
		return
	}
	snap.Data = nil
	writeJSON(w, http.StatusCreated, snapshotResponse(snap))
}

func (h *assetHandlers) listSnapshots(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	snaps, err := h.svc.ListSnapshots(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		writeAssetError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(snaps))
	for _, s := range snaps {
		out = append(out, snapshotResponse(s))
	}
	writeJSON(w, http.StatusOK, map[string]any{"snapshots": out})
}

func (h *assetHandlers) getSnapshot(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	s, err := h.svc.GetSnapshot(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		writeAssetError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshotResponse(s))
}

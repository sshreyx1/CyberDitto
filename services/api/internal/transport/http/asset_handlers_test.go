package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

func TestAssets_CreateUploadListGet(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	assetID, snapshotID := createAssetWithSnapshot(t, router, cookies)

	sum := sha256.Sum256([]byte(testSnapshotDoc))
	digest := hex.EncodeToString(sum[:])
	rec := doRaw(t, router, http.MethodPost, "/api/v1/assets/"+assetID+"/snapshots", []byte(testSnapshotDoc),
		map[string]string{snapshotDigestHeader: digest}, cookies)
	if rec.Code != http.StatusCreated {
		t.Fatalf("second upload status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var second struct {
		Version int    `json:"version"`
		SHA256  string `json:"sha256"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&second)
	if second.Version != 2 || second.SHA256 != digest {
		t.Errorf("second snapshot = %+v, want version 2 with sha256 %s", second, digest)
	}

	rec = doJSON(t, router, http.MethodGet, "/api/v1/assets", nil, cookies)
	var list struct {
		Assets []map[string]any `json:"assets"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&list)
	if len(list.Assets) != 1 {
		t.Fatalf("assets = %d, want 1", len(list.Assets))
	}

	rec = doJSON(t, router, http.MethodGet, "/api/v1/assets/"+assetID+"/snapshots", nil, cookies)
	var snaps struct {
		Snapshots []map[string]any `json:"snapshots"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&snaps)
	if len(snaps.Snapshots) != 2 {
		t.Fatalf("snapshots = %d, want 2", len(snaps.Snapshots))
	}
	if _, hasData := snaps.Snapshots[0]["data"]; hasData {
		t.Error("snapshot list must not include document bodies")
	}

	rec = doJSON(t, router, http.MethodGet, "/api/v1/snapshots/"+snapshotID, nil, cookies)
	var full struct {
		Data json.RawMessage `json:"data"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&full)
	if _, err := asset.ParseDocument(full.Data); err != nil {
		t.Errorf("GET snapshot data does not round-trip as a valid document: %v", err)
	}
}

func TestAssets_UploadRejections(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	assetID, _ := createAssetWithSnapshot(t, router, cookies)
	path := "/api/v1/assets/" + assetID + "/snapshots"

	cases := []struct {
		name    string
		body    []byte
		headers map[string]string
		want    int
	}{
		{"invalid document", []byte(`{"schema_version":"9"}`), nil, http.StatusBadRequest},
		{"digest mismatch", []byte(testSnapshotDoc), map[string]string{snapshotDigestHeader: "00"}, http.StatusUnprocessableEntity},
		{"too large", bytes.Repeat([]byte(" "), asset.MaxDocumentBytes+1), nil, http.StatusRequestEntityTooLarge},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if rec := doRaw(t, router, http.MethodPost, path, c.body, c.headers, cookies); rec.Code != c.want {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, c.want, rec.Body.String())
			}
		})
	}

	if rec := doJSON(t, router, http.MethodPost, "/api/v1/assets", createAssetRequest{Name: ""}, cookies); rec.Code != http.StatusBadRequest {
		t.Errorf("blank asset name status = %d, want 400", rec.Code)
	}
}

// A JSON body over the default 1 MiB cap is still rejected on ordinary
// routes; only snapshot uploads get the larger limit.
func TestBodyLimit_DefaultStillAppliesElsewhere(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	big := append([]byte(`{"name":"`), bytes.Repeat([]byte("a"), maxRequestBodyBytes+1)...)
	big = append(big, []byte(`"}`)...)
	if rec := doRaw(t, router, http.MethodPost, "/api/v1/assets", big, nil, cookies); rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized JSON on /assets: status = %d, want 400", rec.Code)
	}
}

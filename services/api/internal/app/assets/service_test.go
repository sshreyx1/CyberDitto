package assets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/fake"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

const doc = `{"schema_version":"1","hostname":"HOST-1","collected_at":"2026-11-23T06:00:00Z","fields":{"os.version":{"source":"os.version","collected_at":"2026-11-23T06:00:00Z","value":{"product_name":"Windows 11 Pro"}}}}`

func newService() (*Service, *fake.AuditRepo) {
	audit := fake.NewAuditRepo()
	return NewService(fake.NewAssetRepo(), audit), audit
}

func TestUploadSnapshot_VersionsAndDigest(t *testing.T) {
	svc, audit := newService()
	ctx := context.Background()
	actor, ws := uuid.New(), uuid.New()

	a, err := svc.Create(ctx, actor, ws, "  laptop  ")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if a.Name != "laptop" {
		t.Errorf("name = %q, want trimmed", a.Name)
	}

	sum := sha256.Sum256([]byte(doc))
	digest := hex.EncodeToString(sum[:])

	first, err := svc.UploadSnapshot(ctx, actor, ws, a.ID, []byte(doc), digest)
	if err != nil {
		t.Fatalf("UploadSnapshot: %v", err)
	}
	second, err := svc.UploadSnapshot(ctx, actor, ws, a.ID, []byte(doc), "")
	if err != nil {
		t.Fatalf("UploadSnapshot (no declared digest): %v", err)
	}
	if first.Version != 1 || second.Version != 2 {
		t.Errorf("versions = %d, %d; want 1, 2", first.Version, second.Version)
	}
	if first.SHA256 != digest {
		t.Errorf("stored sha256 = %s, want %s", first.SHA256, digest)
	}

	list, err := svc.ListSnapshots(ctx, ws, a.ID)
	if err != nil || len(list) != 2 || list[0].Data != nil {
		t.Fatalf("ListSnapshots = %+v, %v; want 2 summaries without data", list, err)
	}

	events, _ := audit.List(ctx, ws, 10, 0)
	if len(events) != 3 { // asset.create + 2 × snapshot.upload
		t.Errorf("audit events = %d, want 3", len(events))
	}
}

func TestUploadSnapshot_Rejections(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()
	actor, ws := uuid.New(), uuid.New()
	a, _ := svc.Create(ctx, actor, ws, "laptop")

	if _, err := svc.UploadSnapshot(ctx, actor, ws, a.ID, []byte(`{"nope":1}`), ""); !errors.Is(err, asset.ErrInvalidDocument) {
		t.Errorf("invalid document error = %v", err)
	}
	if _, err := svc.UploadSnapshot(ctx, actor, ws, a.ID, []byte(doc), "deadbeef"); !errors.Is(err, ErrDigestMismatch) {
		t.Errorf("digest mismatch error = %v", err)
	}
	if _, err := svc.UploadSnapshot(ctx, actor, uuid.New(), a.ID, []byte(doc), ""); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("other-workspace upload error = %v, want ErrNotFound", err)
	}
	if _, err := svc.Create(ctx, actor, ws, ""); !errors.Is(err, asset.ErrInvalidName) {
		t.Errorf("blank name error = %v", err)
	}
}

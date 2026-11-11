package audit

import (
	"testing"

	"github.com/google/uuid"
)

func TestComputeHash_Deterministic(t *testing.T) {
	ws := uuid.New()
	actor := uuid.New()
	h1, err := ComputeHash(ws, &actor, "session.login", "user", actor.String(), map[string]any{"ip": "redacted"}, GenesisHash)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := ComputeHash(ws, &actor, "session.login", "user", actor.String(), map[string]any{"ip": "redacted"}, GenesisHash)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("ComputeHash is not deterministic: %q != %q", h1, h2)
	}
}

func TestComputeHash_DifferentPrevHashDiffers(t *testing.T) {
	ws := uuid.New()
	h1, _ := ComputeHash(ws, nil, "workspace.setup", "workspace", ws.String(), nil, GenesisHash)
	h2, _ := ComputeHash(ws, nil, "workspace.setup", "workspace", ws.String(), nil, "some-other-prev-hash")
	if h1 == h2 {
		t.Fatal("expected different prev_hash to produce a different hash (this is what makes the chain tamper-evident)")
	}
}

func TestVerify_DetectsTampering(t *testing.T) {
	ws := uuid.New()
	hash, err := ComputeHash(ws, nil, "workspace.setup", "workspace", ws.String(), nil, GenesisHash)
	if err != nil {
		t.Fatal(err)
	}
	e := Event{
		WorkspaceID: ws,
		Action:      "workspace.setup",
		TargetType:  "workspace",
		TargetID:    ws.String(),
		PrevHash:    GenesisHash,
		Hash:        hash,
	}
	ok, err := Verify(e)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Verify should accept an untampered event")
	}

	tampered := e
	tampered.TargetID = uuid.New().String() // simulate someone rewriting a historical row
	ok, err = Verify(tampered)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Verify should reject a tampered event")
	}
}

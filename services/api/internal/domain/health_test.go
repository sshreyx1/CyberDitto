package domain

import "testing"

func TestNewHealth(t *testing.T) {
	h := NewHealth()
	if h.Status != "ok" {
		t.Fatalf("Status = %q, want %q", h.Status, "ok")
	}
	if h.Version == "" {
		t.Fatal("Version must not be empty")
	}
}

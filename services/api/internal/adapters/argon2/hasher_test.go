package argon2

import "testing"

func TestHashAndVerify_RoundTrip(t *testing.T) {
	h := New()
	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := h.Verify("correct horse battery staple", hash)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("Verify(correct password) = false, want true")
	}
}

func TestVerify_WrongPasswordRejected(t *testing.T) {
	h := New()
	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := h.Verify("wrong password", hash)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if ok {
		t.Fatal("Verify(wrong password) = true, want false")
	}
}

func TestHash_ProducesDifferentSaltsEachTime(t *testing.T) {
	h := New()
	a, err := h.Hash("same password")
	if err != nil {
		t.Fatal(err)
	}
	b, err := h.Hash("same password")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("two hashes of the same password were identical; salts must be random per call")
	}
}

func TestVerify_RejectsMalformedHash(t *testing.T) {
	h := New()
	if _, err := h.Verify("anything", "not-an-argon2-hash"); err != ErrInvalidHash {
		t.Fatalf("Verify(malformed hash) error = %v, want ErrInvalidHash", err)
	}
}

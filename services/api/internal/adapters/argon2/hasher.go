// Package argon2 implements ports.PasswordHasher with Argon2id
// (docs/MASTER_PLAN.md §6.2).
package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Parameters follow the OWASP Password Storage Cheat Sheet's Argon2id
// baseline recommendation (as of this writing) for an interactive login
// path: 19 MiB memory, 2 iterations, 1 degree of parallelism, 16-byte salt,
// 32-byte key. Tuned further once real load-testing data exists.
const (
	memoryKiB   = 19 * 1024
	iterations  = 2
	parallelism = 1
	saltLen     = 16
	keyLen      = 32
)

var ErrInvalidHash = errors.New("argon2: invalid encoded hash")

type Hasher struct{}

func New() *Hasher { return &Hasher{} }

func (*Hasher) Hash(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("argon2: generate salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, iterations, memoryKiB, parallelism, keyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memoryKiB, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func (*Hasher) Verify(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, ErrInvalidHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, ErrInvalidHash
	}
	var m, t, p int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false, ErrInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrInvalidHash
	}

	got := argon2.IDKey([]byte(password), salt, uint32(t), uint32(m), uint8(p), uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

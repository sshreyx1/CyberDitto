// Package token generates cryptographically random session tokens
// (docs/MASTER_PLAN.md §6.2: "random 256-bit IDs").
package token

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const byteLen = 32 // 256 bits

type Generator struct{}

func New() *Generator { return &Generator{} }

func (*Generator) Generate() (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("token: generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

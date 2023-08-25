package utils

import (
	"crypto/sha256"
	"fmt"
)

func ComputeSha256(val string) string {
	h := sha256.New()
	h.Write([]byte(val))
	// Calculate and print the hash
	return fmt.Sprintf("%x", h.Sum(nil))
}

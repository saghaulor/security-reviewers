// Package uuidgen generates RFC 4122 version 4 (random) UUIDs using
// crypto/rand. It has no external dependencies.
package uuidgen

import (
	"crypto/rand"
	"fmt"
)

// New returns a new random RFC 4122 v4 UUID in canonical lowercase form:
//
//	xxxxxxxx-xxxx-4xxx-[89ab]xxx-xxxxxxxxxxxx
//
// It panics only if crypto/rand is unavailable, which cannot happen on any
// supported platform (Linux, macOS, Windows).
func New() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("uuidgen: crypto/rand unavailable: " + err.Error())
	}
	// Set version 4 (random): top 4 bits of byte 6 = 0100.
	b[6] = (b[6] & 0x0f) | 0x40
	// Set RFC 4122 variant: top 2 bits of byte 8 = 10.
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16],
	)
}

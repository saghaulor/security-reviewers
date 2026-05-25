package uuidgen_test

import (
	"regexp"
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/uuidgen"
)

// rfc4122v4 matches a canonical lowercase RFC 4122 v4 UUID:
// xxxxxxxx-xxxx-4xxx-[89ab]xxx-xxxxxxxxxxxx
var rfc4122v4 = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

func TestNew_Format(t *testing.T) {
	id := uuidgen.New()
	if !rfc4122v4.MatchString(id) {
		t.Errorf("New() = %q; want RFC 4122 v4 format", id)
	}
}

func TestNew_Unique(t *testing.T) {
	const n = 100
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id := uuidgen.New()
		if _, dup := seen[id]; dup {
			t.Fatalf("New() returned duplicate UUID after %d calls: %s", i+1, id)
		}
		seen[id] = struct{}{}
	}
}

func TestNew_VersionBit(t *testing.T) {
	for i := 0; i < 20; i++ {
		id := uuidgen.New()
		// The 15th character (index 14) must be '4' for v4.
		if id[14] != '4' {
			t.Errorf("New() = %q; version nibble at index 14 = %c, want '4'", id, id[14])
		}
	}
}

func TestNew_VariantBits(t *testing.T) {
	for i := 0; i < 20; i++ {
		id := uuidgen.New()
		// The 20th character (index 19) must be one of 8,9,a,b for RFC 4122 variant.
		c := id[19]
		if c != '8' && c != '9' && c != 'a' && c != 'b' {
			t.Errorf("New() = %q; variant nibble at index 19 = %c, want one of [89ab]", id, c)
		}
	}
}

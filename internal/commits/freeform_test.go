package commits

import (
	"testing"

	"github.com/0x1306e6d/release-cli/internal/version"
)

func TestFreeform_Parse(t *testing.T) {
	f := Freeform()

	cases := []struct {
		subject string
		body    string
	}{
		{"Add user export feature", ""},
		{"cleanup", ""},
		{"feat: this looks conventional but is still freeform", "BREAKING CHANGE: ignored"},
	}

	for _, tc := range cases {
		p := f.Parse(tc.subject, tc.body)
		if p == nil {
			t.Fatalf("expected non-nil result for %q", tc.subject)
		}
		if p.Bump != version.BumpPatch {
			t.Errorf("bump = %v, want patch for %q", p.Bump, tc.subject)
		}
		if p.Type != "other" {
			t.Errorf("type = %q, want %q for %q", p.Type, "other", tc.subject)
		}
		if p.Subject != tc.subject {
			t.Errorf("subject = %q, want %q", p.Subject, tc.subject)
		}
		if p.Breaking {
			t.Errorf("breaking = true, want false for %q", tc.subject)
		}
	}
}

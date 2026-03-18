package commits

import (
	"testing"

	"github.com/0x1306e6d/release-cli/internal/version"
)

func TestFreeformCommits_Parse(t *testing.T) {
	f := &FreeformCommits{}

	tests := []struct {
		name    string
		subject string
		body    string
	}{
		{"plain English", "Add user export feature", ""},
		{"single word", "cleanup", ""},
		{"special characters", "fix: handle 100% CPU & memory issues (#42)", ""},
		{"empty subject", "", ""},
		{"with body", "Update docs", "Some detailed description"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Parse(tt.subject, tt.body)
			if result == nil {
				t.Fatal("expected non-nil result, freeform should accept all commits")
			}
			if result.Type != "other" {
				t.Errorf("Type = %q, want %q", result.Type, "other")
			}
			if result.Subject != tt.subject {
				t.Errorf("Subject = %q, want %q", result.Subject, tt.subject)
			}
			if result.Body != tt.body {
				t.Errorf("Body = %q, want %q", result.Body, tt.body)
			}
			if result.Breaking {
				t.Error("Breaking = true, want false")
			}
			if result.Bump != version.BumpPatch {
				t.Errorf("Bump = %v, want BumpPatch", result.Bump)
			}
		})
	}
}

func TestFreeformCommits_IgnoresBreakingChangeFooter(t *testing.T) {
	f := &FreeformCommits{}

	result := f.Parse("some change", "BREAKING CHANGE: removed old API")
	if result.Breaking {
		t.Error("freeform should not detect breaking changes from footer")
	}
	if result.Bump != version.BumpPatch {
		t.Errorf("Bump = %v, want BumpPatch even with BREAKING CHANGE footer", result.Bump)
	}
}

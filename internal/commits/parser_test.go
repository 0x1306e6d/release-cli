package commits

import (
	"testing"

	"github.com/0x1306e6d/release-cli/internal/version"
)

func TestConventional_Feat(t *testing.T) {
	conv := &ConventionalCommits{}
	p := conv.Parse("feat: add user export", "")
	if p == nil {
		t.Fatal("expected parsed commit")
	}
	if p.Type != "feat" {
		t.Errorf("type = %q, want %q", p.Type, "feat")
	}
	if p.Bump != version.BumpMinor {
		t.Errorf("bump = %v, want minor", p.Bump)
	}
}

func TestConventional_Fix(t *testing.T) {
	conv := &ConventionalCommits{}
	p := conv.Parse("fix: correct null pointer", "")
	if p == nil {
		t.Fatal("expected parsed commit")
	}
	if p.Bump != version.BumpPatch {
		t.Errorf("bump = %v, want patch", p.Bump)
	}
}

func TestConventional_BreakingBang(t *testing.T) {
	conv := &ConventionalCommits{}
	p := conv.Parse("feat!: redesign API", "")
	if p == nil {
		t.Fatal("expected parsed commit")
	}
	if !p.Breaking {
		t.Error("expected breaking")
	}
	if p.Bump != version.BumpMajor {
		t.Errorf("bump = %v, want major", p.Bump)
	}
}

func TestConventional_BreakingFooter(t *testing.T) {
	conv := &ConventionalCommits{}
	p := conv.Parse("feat: change auth", "BREAKING CHANGE: removed legacy endpoint")
	if p == nil {
		t.Fatal("expected parsed commit")
	}
	if !p.Breaking {
		t.Error("expected breaking")
	}
	if p.Bump != version.BumpMajor {
		t.Errorf("bump = %v, want major", p.Bump)
	}
}

func TestConventional_WithScope(t *testing.T) {
	conv := &ConventionalCommits{}
	p := conv.Parse("feat(auth): add oauth support", "")
	if p == nil {
		t.Fatal("expected parsed commit")
	}
	if p.Scope != "auth" {
		t.Errorf("scope = %q, want %q", p.Scope, "auth")
	}
}

func TestConventional_NonConforming(t *testing.T) {
	conv := &ConventionalCommits{}
	p := conv.Parse("updated readme", "")
	if p != nil {
		t.Error("expected nil for non-conforming commit")
	}
}

func TestAnalyze_HighestBumpWins(t *testing.T) {
	conv := &ConventionalCommits{}
	raw := []RawCommit{
		{Subject: "fix: typo"},
		{Subject: "feat: new feature"},
	}
	_, bump := Analyze(raw, conv)
	if bump == nil {
		t.Fatal("expected non-nil bump")
	}
	if *bump != version.BumpMinor {
		t.Errorf("bump = %v, want minor", *bump)
	}
}

func TestAnalyze_NoReleasable(t *testing.T) {
	conv := &ConventionalCommits{}
	raw := []RawCommit{
		{Subject: "update docs"},
		{Subject: "cleanup"},
	}
	_, bump := Analyze(raw, conv)
	if bump != nil {
		t.Errorf("expected nil bump, got %v", *bump)
	}
}

func TestCustom_Convention(t *testing.T) {
	conv := NewCustomCommits(nil, []string{"feature", "enhancement"}, []string{"bugfix"})
	p := conv.Parse("feature: add dark mode", "")
	if p == nil {
		t.Fatal("expected parsed commit")
	}
	if p.Bump != version.BumpMinor {
		t.Errorf("bump = %v, want minor", p.Bump)
	}
}

func TestCustom_UnknownType(t *testing.T) {
	conv := NewCustomCommits(nil, []string{"feature"}, nil)
	p := conv.Parse("chore: update deps", "")
	if p != nil {
		t.Error("expected nil for unrecognized custom type")
	}
}

package version

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		input      string
		wantMajor  int
		wantMinor  int
		wantPatch  int
		wantPre    string
		wantString string
	}{
		{"1.2.3", 1, 2, 3, "", "1.2.3"},
		{"v1.2.3", 1, 2, 3, "", "1.2.3"},
		{"0.0.0", 0, 0, 0, "", "0.0.0"},
		{"1.4.0-SNAPSHOT", 1, 4, 0, "SNAPSHOT", "1.4.0-SNAPSHOT"},
		{"v2.0.0-rc.1", 2, 0, 0, "rc.1", "2.0.0-rc.1"},
		{"1.4.0.dev0", 1, 4, 0, "dev0", "1.4.0-dev0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v.Major != tt.wantMajor || v.Minor != tt.wantMinor || v.Patch != tt.wantPatch {
				t.Errorf("got %d.%d.%d, want %d.%d.%d", v.Major, v.Minor, v.Patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
			}
			if v.PreRelease != tt.wantPre {
				t.Errorf("pre-release = %q, want %q", v.PreRelease, tt.wantPre)
			}
			if v.String() != tt.wantString {
				t.Errorf("String() = %q, want %q", v.String(), tt.wantString)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	invalids := []string{"", "v", "abc", "1.2", "1"}
	for _, s := range invalids {
		t.Run(s, func(t *testing.T) {
			_, err := Parse(s)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestBump(t *testing.T) {
	v := Semver{Major: 1, Minor: 2, Patch: 3}

	tests := []struct {
		bump BumpType
		want string
	}{
		{BumpPatch, "1.2.4"},
		{BumpMinor, "1.3.0"},
		{BumpMajor, "2.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.bump.String(), func(t *testing.T) {
			got := v.Bump(tt.bump)
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestBump_StripsPreRelease(t *testing.T) {
	v := Semver{Major: 1, Minor: 4, Patch: 0, PreRelease: "SNAPSHOT"}
	got := v.Bump(BumpMinor)
	if got.PreRelease != "" {
		t.Errorf("expected no pre-release, got %q", got.PreRelease)
	}
	if got.String() != "1.5.0" {
		t.Errorf("got %q, want %q", got.String(), "1.5.0")
	}
}

func TestTagString(t *testing.T) {
	v := Semver{Major: 1, Minor: 4, Patch: 0}
	if got := v.TagString(); got != "v1.4.0" {
		t.Errorf("got %q, want %q", got, "v1.4.0")
	}
}

func TestIsPreRelease(t *testing.T) {
	v1 := Semver{Major: 1, Minor: 0, Patch: 0}
	if v1.IsPreRelease() {
		t.Error("expected false for release version")
	}

	v2 := Semver{Major: 1, Minor: 0, Patch: 0, PreRelease: "SNAPSHOT"}
	if !v2.IsPreRelease() {
		t.Error("expected true for pre-release version")
	}
}

func TestStripPreRelease(t *testing.T) {
	v := Semver{Major: 1, Minor: 4, Patch: 0, PreRelease: "SNAPSHOT"}
	got := v.StripPreRelease()
	if got.String() != "1.4.0" {
		t.Errorf("got %q, want %q", got.String(), "1.4.0")
	}
}

func TestNextSnapshot(t *testing.T) {
	released := Semver{Major: 1, Minor: 4, Patch: 0}
	got := NextSnapshot(released, "SNAPSHOT")
	if got.String() != "1.5.0-SNAPSHOT" {
		t.Errorf("got %q, want %q", got.String(), "1.5.0-SNAPSHOT")
	}
}

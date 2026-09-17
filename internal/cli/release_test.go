package cli

import "testing"

func TestValidateVersionFlags(t *testing.T) {
	for _, tt := range []struct {
		name, release, next, bump string
		wantErr                   bool
	}{
		{"release", "1.2.0", "", "", false},
		{"snapshot", "1.2.0", "1.3.0-SNAPSHOT", "", false},
		{"bump conflict", "1.2.0", "", "patch", true},
		{"orphan next", "", "1.3.0-SNAPSHOT", "", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVersionFlags(tt.release, tt.next, tt.bump)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateVersionFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

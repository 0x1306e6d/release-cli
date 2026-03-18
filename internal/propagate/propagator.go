package propagate

import (
	"fmt"

	"github.com/0x1306e6d/release-cli/internal/config"
)

// Propagate updates all configured propagation targets with the new version.
func Propagate(dir string, targets []config.PropagateTarget, newVersion string) error {
	for _, target := range targets {
		if err := propagateOne(dir, target, newVersion); err != nil {
			return fmt.Errorf("propagating to %s: %w", target.File, err)
		}
	}
	return nil
}

func propagateOne(dir string, target config.PropagateTarget, newVersion string) error {
	switch {
	case target.Type != "":
		return propagateBuiltIn(dir, target, newVersion)
	case target.Field != "":
		return propagateStructured(dir, target.File, target.Field, newVersion)
	case target.Pattern != "":
		return propagatePattern(dir, target.File, target.Pattern, newVersion)
	default:
		return fmt.Errorf("no propagation strategy specified")
	}
}

// builtInTypes maps type names to pattern templates.
var builtInTypes = map[string]string{
	"docker-label": `LABEL version="{{.Version}}"`,
}

func propagateBuiltIn(dir string, target config.PropagateTarget, newVersion string) error {
	pattern, ok := builtInTypes[target.Type]
	if !ok {
		return fmt.Errorf("unknown built-in propagation type %q", target.Type)
	}
	return propagatePattern(dir, target.File, pattern, newVersion)
}

package monorepo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/0x1306e6d/release-cli/internal/config"
)

// PackageNode represents a package in the monorepo tree.
type PackageNode struct {
	Name      string         // Tag prefix: root uses config Name, children use relative path
	Path      string         // Relative path from repo root (e.g., "cli", "workflow/sub")
	Config    *config.Config // Parsed .release.yaml for this package
	Children  []*PackageNode
}

// LoadTree reads .release.yaml from rootDir and recursively loads all declared
// child modules, building the full package tree. Returns the root node.
func LoadTree(rootDir string) (*PackageNode, error) {
	visited := make(map[string]bool)
	return loadNode(rootDir, "", visited)
}

func loadNode(rootDir, relPath string, visited map[string]bool) (*PackageNode, error) {
	absPath := rootDir
	if relPath != "" {
		absPath = filepath.Join(rootDir, relPath)
	}

	absClean := filepath.Clean(absPath)
	if visited[absClean] {
		return nil, fmt.Errorf("circular module reference detected at %q", relPath)
	}
	visited[absClean] = true

	cfg, warnings, err := config.Load(absClean)
	if err != nil {
		if relPath == "" {
			return nil, err
		}
		return nil, fmt.Errorf("module %q: %w", relPath, err)
	}
	_ = warnings

	// Determine the node name (used as tag prefix).
	var name string
	if relPath == "" {
		// Root node: use the Name field from config.
		name = cfg.Name
	} else {
		// Child node: use relative path from repo root.
		name = relPath
	}

	node := &PackageNode{
		Name:   name,
		Path:   relPath,
		Config: cfg,
	}

	// Recursively load children.
	for _, mod := range cfg.Modules {
		childRel := mod
		if relPath != "" {
			childRel = filepath.Join(relPath, mod)
		}
		// Normalize to forward slashes for consistency.
		childRel = filepath.ToSlash(childRel)

		child, err := loadNode(rootDir, childRel, visited)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	return node, nil
}

// Flatten returns the node and all its descendants in a flat slice.
func (n *PackageNode) Flatten() []*PackageNode {
	var result []*PackageNode
	n.flatten(&result)
	return result
}

func (n *PackageNode) flatten(acc *[]*PackageNode) {
	*acc = append(*acc, n)
	for _, child := range n.Children {
		child.flatten(acc)
	}
}

// Find searches the tree for a node matching the given name.
// Returns nil if not found.
func (n *PackageNode) Find(name string) *PackageNode {
	if n.Name == name {
		return n
	}
	for _, child := range n.Children {
		if found := child.Find(name); found != nil {
			return found
		}
	}
	return nil
}

// Resolve finds a package by name and returns it plus all its descendants.
// Returns an error if the package is not found.
func (n *PackageNode) Resolve(name string) ([]*PackageNode, error) {
	node := n.Find(name)
	if node == nil {
		return nil, fmt.Errorf("package %q not found in config", name)
	}
	return node.Flatten(), nil
}

// TagPrefix returns the tag prefix for this node.
// For the root node (empty path), it uses the Name field.
// For child nodes, it uses the relative path.
func (n *PackageNode) TagPrefix() string {
	return n.Name
}

// ValidateTree checks the tree for overlapping sibling paths.
func ValidateTree(root *PackageNode) error {
	return validateSiblings(root)
}

func validateSiblings(node *PackageNode) error {
	for i := 0; i < len(node.Children); i++ {
		for j := i + 1; j < len(node.Children); j++ {
			a, b := node.Children[i], node.Children[j]
			aPath := a.Path + "/"
			bPath := b.Path + "/"
			if strings.HasPrefix(bPath, aPath) {
				return fmt.Errorf("overlapping module paths: %q contains %q", a.Path, b.Path)
			}
			if strings.HasPrefix(aPath, bPath) {
				return fmt.Errorf("overlapping module paths: %q contains %q", b.Path, a.Path)
			}
		}
	}
	// Recurse into children.
	for _, child := range node.Children {
		if err := validateSiblings(child); err != nil {
			return err
		}
	}
	return nil
}

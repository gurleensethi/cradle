package types

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CradleProject represents a project managed by cradle.
type CradleProject struct {
	// Path is the absolute path to the project directory.
	Path string `yaml:"path"`
	// CreatedAt is the timestamp when the project was registered.
	CreatedAt time.Time `yaml:"created_at"`
	// Temporary indicates whether the project is temporary.
	Temporary bool `yaml:"temporary"`
	// UniqueNameFromPath is a human-readable name derived from the project path.
	// This field is not serialized to YAML (used for display and lookup only).
	UniqueNameFromPath string `yaml:"-"`
	// CreatedBy identifies the tool or user that registered this project.
	CreatedBy string `yaml:"created_by"`
}

// MatchPathOrName reports whether the project's path or unique name exactly matches the query.
func (p CradleProject) MatchPathOrName(query string) bool {
	return p.Path == query || p.UniqueNameFromPath == query
}

func (p CradleProject) GetPathWithTruncatedHome() string {
	homeDir, err := os.UserHomeDir()
	if err != nil || !strings.HasPrefix(p.Path, homeDir) {
		return p.Path
	}

	if p.Path == homeDir {
		return "~"
	}

	prefix := homeDir + string(filepath.Separator)

	if strings.HasPrefix(p.Path, prefix) {
		relativePath := p.Path[len(prefix):]

		return "~" + string(filepath.Separator) + relativePath
	}

	return p.Path
}

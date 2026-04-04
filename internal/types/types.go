package types

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CradleProject represents a project managed by cradle.
type CradleProject struct {
	Path               string    `yaml:"path"`
	CreatedAt          time.Time `yaml:"created_at"`
	Temporary          bool      `yaml:"temporary"`
	UniqueNameFromPath string    `yaml:"-"`
	CreatedBy          string    `yaml:"created_by"`
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

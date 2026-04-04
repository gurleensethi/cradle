package types

import "time"

// CradleProject represents a project managed by cradle.
type CradleProject struct {
	Path               string    `yaml:"path"`
	CreatedAt          time.Time `yaml:"created_at"`
	Temporary          bool      `yaml:"temporary"`
	UniqueNameFromPath string    `yaml:"-"`
	CreatedBy          string    `yaml:"created_by"`
}

// MatchPathOrName checks if the project matches the given query by path or name.
func (p CradleProject) MatchPathOrName(query string) bool {
	return p.Path == query || p.UniqueNameFromPath == query
}

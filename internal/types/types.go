package types

import "time"

// CradleProject represents a project managed by cradle.
type CradleProject struct {
	Path               string    `toml:"path"`
	CreatedAt          time.Time `toml:"created_at"`
	Temporary          bool      `toml:"temporary"`
	UniqueNameFromPath string    `toml:"-"`
	CreatedBy          string    `toml:"created_by"`
}

// MatchPathOrName checks if the project matches the given query by path or name.
func (p CradleProject) MatchPathOrName(query string) bool {
	return p.Path == query || p.UniqueNameFromPath == query
}

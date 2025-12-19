package main

import "time"

type CradleProject struct {
	Path               string    `toml:"path"`
	CreatedAt          time.Time `toml:"created_at"`
	Temporary          bool      `toml:"temporary"`
	UniqueNameFromPath string    `toml:"-"`
	CreatedBy          string    `toml:"created_by"`
}

func (p CradleProject) MatchPathOrName(query string) bool {
	return p.Path == query || p.UniqueNameFromPath == query
}

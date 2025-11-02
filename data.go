package main

import "time"

type CradleProject struct {
	Path        string
	CreatedAt   time.Time
	IsTemporary bool
}

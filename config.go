package main

import (
	"github.com/gurleensethi/cradle/internal/config"
)

// InitConfig initializes the cradle configuration.
// Deprecated: Use config.Init() instead.
func InitConfig() error {
	return config.Init()
}

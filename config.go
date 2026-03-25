package main

import (
	"github.com/gurleensethi/cradle/internal/config"
)

// InitConfig initializes the cradle configuration.
// Deprecated: Use config.Init() instead.
func InitConfig() error {
	return config.Init()
}

// UpdateCradleConfigFile writes the current configuration to the config file.
// Deprecated: Use config.UpdateConfigFile() instead.
func UpdateCradleConfigFile() error {
	return config.UpdateConfigFile()
}

// CradleConfig represents the cradle.toml file.
// Deprecated: Use config.CradleConfig instead.
type CradleConfig = config.CradleConfig

// Config holds the cradle configuration.
// Deprecated: Use config.Config instead.
type Config = config.Config

// getConfig returns the global configuration instance.
// Deprecated: Use config.Get() instead.
func getConfig() *config.Config {
	return config.Get()
}

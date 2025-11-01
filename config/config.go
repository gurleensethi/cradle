package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	EnvCradleHome        = "CRADLE_HOME"
	CradleConfigFileName = "cradle.toml"
)

type Config struct {
	CradleHomeDirPath string
}

var cfg *Config

func Init() error {
	cfg = &Config{}

	cradleHomePath, err := getCradleHomeDir()
	if err != nil {
		return err
	}

	// Make sure home directory exists.
	err = ensureCradleHomeDir(cradleHomePath)
	if err != nil {
		return err
	}

	cfg.CradleHomeDirPath = cradleHomePath

	return nil
}

func getCradleHomeDir() (string, error) {
	cradleHomePath := strings.TrimSpace(os.Getenv(EnvCradleHome))
	if cradleHomePath == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		cradleHomePath = path.Join(userHomeDir, "cradle")
	}

	return cradleHomePath, nil
}

func ensureCradleHomeDir(dirPath string) error {
	dirStat, err := os.Stat(dirPath)
	if err != nil {
		// Directory doesn't exist, try to create it (including parents).
		if errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				return err
			}

			return nil
		}

		return err
	}

	// Directory already exists, we are good.
	if dirStat.IsDir() {
		return nil
	}

	return fmt.Errorf("%s is a file, either delete the file or change cradle home path by setting `CRADLE_HOME` env to something else", dirPath)
}

// func ensureCradleConfigFile(dirPath string) error {
// }

func Get() *Config {
	if cfg == nil {
		panic("call config.Init() before using config")
	}

	return cfg
}

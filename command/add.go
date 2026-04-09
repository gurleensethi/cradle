package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gurleensethi/cradle/internal/config"
	"github.com/gurleensethi/cradle/internal/types"
	"github.com/urfave/cli/v3"
)

// Add returns the add command for adding existing projects to cradle.
func Add() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add an existing project into cradle",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "path",
				UsageText: "path to the project directory",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			projectPath := c.StringArg("path")
			if projectPath == "" {
				return fmt.Errorf("provide a project path")
			}

			absProjectDirPath, err := addProject(projectPath)
			if err != nil {
				return err
			}

			fmt.Println("Project added: ", absProjectDirPath)

			return nil
		},
	}
}

// addProject validates a directory and registers it as a cradle project. Returns the absolute path.
func addProject(projectDirPath string) (string, error) {
	projectDirPath, err := filepath.Abs(projectDirPath)
	if err != nil {
		return "", err
	}

	dirStat, err := os.Stat(projectDirPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("project directory with specified path doesn't exist")
		}

		return "", err
	}

	if !dirStat.IsDir() {
		return "", fmt.Errorf("specified path is not a directory")
	}

	// Make sure there is no existing project with same name
	for _, project := range config.Projects() {
		if project.Path == projectDirPath {
			return "", fmt.Errorf("project already exists")
		}
	}

	cradleProject := types.CradleProject{
		Path:      projectDirPath,
		Temporary: false,
		CreatedAt: time.Now(),
	}

	return projectDirPath, config.AddProject(cradleProject)
}

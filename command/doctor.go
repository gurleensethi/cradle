package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gurleensethi/cradle/internal/config"
	"github.com/gurleensethi/cradle/internal/types"
	"github.com/urfave/cli/v3"
)

// Doctor returns the doctor command for checking the health of cradle.
func Doctor() *cli.Command {
	return &cli.Command{
		Name:  "doctor",
		Usage: "Check health of cradle, find broken projects and fix problems",
		Action: func(ctx context.Context, c *cli.Command) error {
			return doctor()
		},
	}
}

// doctor checks all registered project paths for existence and validity.
// It reports any issues found without modifying the configuration.
func doctor() error {
	var issues []string

	config.ForEachProject(func(project types.CradleProject) bool {
		stat, err := os.Stat(project.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				issues = append(issues, fmt.Sprintf("- %s: project path does not exist", project.Path))
			} else {
				// Collect all errors instead of returning early for generic os.Stat errors
				issues = append(issues, fmt.Sprintf("- %s: failed to access project path: %v", project.Path, err))
			}
			return true // continue
		}

		if !stat.IsDir() {
			issues = append(issues, fmt.Sprintf("- %s: project path is a file, not a directory", project.Path))
		}

		return true // continue
	})

	if len(issues) == 0 {
		fmt.Println("No issues found ✓")
	} else {
		fmt.Printf("Found %d issues:\n", len(issues))
		fmt.Println(strings.Join(issues, "\n"))
	}

	return nil
}

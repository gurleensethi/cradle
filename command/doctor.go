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
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "fix",
				Usage: "remove broken project entries (paths that do not exist) from the config",
				Value: false,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return doctor(c.Bool("fix"))
		},
	}
}

// doctor checks all registered project paths and reports any issues found.
// Detection and fixing are kept separate: issues/broken entries are collected
// first, and only when fix is true are the broken entries removed.
func doctor(fix bool) error {
	// Detection phase: collect all issues and broken entries.
	var issues []string
	var brokenProjectPaths []types.CradleProject

	for _, project := range config.Projects() {
		stat, err := os.Stat(project.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				issues = append(issues, fmt.Sprintf("- %s: project path does not exist", project.Path))
				brokenProjectPaths = append(brokenProjectPaths, project)
			} else {
				// Collect all errors instead of returning early for generic os.Stat errors
				issues = append(issues, fmt.Sprintf("- %s: failed to access project path: %v", project.Path, err))
			}
			continue
		}

		if !stat.IsDir() {
			issues = append(issues, fmt.Sprintf("- %s: project path is a file, not a directory", project.Path))
		}
	}

	if len(issues) == 0 {
		fmt.Println("No issues found ✓")
		return nil
	}

	fmt.Printf("Found %d issues:\n", len(issues))
	fmt.Println(strings.Join(issues, "\n"))

	// Fix phase: only runs when requested.
	if fix {
		if len(brokenProjectPaths) == 0 {
			fmt.Println("Nothing to fix.")
			return nil
		}

		for _, project := range brokenProjectPaths {
			if err := config.RemoveProjectByName(project.Path); err != nil {
				return err
			}

			fmt.Printf("Fixed: removed broken project entry %s\n", project.Path)
		}
	} else {
		fmt.Println("\nRun `cradle doctor --fix` to fix these issues.")
	}

	return nil
}

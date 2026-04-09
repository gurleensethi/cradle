package command

import (
	"context"
	"fmt"
	"os"

	"github.com/gurleensethi/cradle/internal/config"
	"github.com/urfave/cli/v3"
)

// Cleanup returns the cleanup command for removing temporary projects.
func Cleanup() *cli.Command {
	return &cli.Command{
		Name:  "cleanup",
		Usage: "Remove all temporary projects that were created by cradle. ",
		Action: func(ctx context.Context, c *cli.Command) error {
			_, err := cleanupTemporaryProjects()
			return err
		},
	}
}

// cleanupTemporaryProjects removes all temporary projects after user confirmation. Returns the count of remaining permanent projects.
func cleanupTemporaryProjects() (int, error) {
	tempProjects := config.TemporaryProjects()
	count := len(tempProjects)

	if count == 0 {
		fmt.Println("No temporary projects to clean up.")
		return 0, nil
	}

	var confirmation string
	fmt.Printf("Are you sure you want to delete %d temporary projects (Y/N):", count)
	fmt.Scanln(&confirmation)

	if confirmation != "Y" && confirmation != "y" {
		fmt.Println("Cleanup aborted.")
		return 0, nil
	}

	for _, project := range tempProjects {
		// Remove the project directory
		err := os.RemoveAll(project.Path)
		if err != nil {
			return 0, err
		}
	}

	// Update config with only permanent projects
	permanentProjects := config.PermanentProjects()
	if err := config.UpdateProjects(permanentProjects); err != nil {
		return 0, err
	}

	fmt.Printf("Removed %d temporary projects.\n", count)

	return len(permanentProjects), nil
}

package command

import (
	"context"
	"fmt"

	"github.com/gurleensethi/cradle/internal/config"
	"github.com/urfave/cli/v3"
)

// Remove returns the remove command for removing projects from cradle.
func Remove() *cli.Command {
	return &cli.Command{
		Name:    "remove",
		Usage:   "Remove a project from cradle's management (this does not delete the project files)",
		Aliases: []string{"rm"},
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "name",
				UsageText: "name of the project to remove",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			name := c.StringArg("name")
			if name == "" {
				return fmt.Errorf("provide a project name")
			}

			if err := config.RemoveProjectByName(name); err != nil {
				return err
			}

			fmt.Println("Project removed from cradle")

			return nil
		},
	}
}

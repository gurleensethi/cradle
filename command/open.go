package command

import (
	"context"
	"fmt"
	"os"

	"github.com/gurleensethi/cradle/internal/config"
	"github.com/gurleensethi/cradle/internal/types"
	"github.com/urfave/cli/v3"
)

// Open returns the open command for opening a project.
func Open() *cli.Command {
	return &cli.Command{
		Name:  "open",
		Usage: "Open a project",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "name",
				UsageText: "name of the project to open",
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

			project, found := config.FindProject(name)
			if !found {
				return fmt.Errorf("%s project not found", name)
			}

			if config.Get().CradleCommandOut {
				fmt.Fprintf(os.Stderr, "eval cd %s", project.Path)
			}

			return nil
		},
	}
}

// openProject looks up a project by name or path and returns it if found.
func openProject(query string) (types.CradleProject, error) {
	project, found := config.FindProject(query)
	if !found {
		return types.CradleProject{}, fmt.Errorf("%s project not found", query)
	}
	return project, nil
}

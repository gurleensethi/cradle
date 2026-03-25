package command

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/gurleensethi/cradle/internal/config"
	"github.com/gurleensethi/cradle/internal/types"
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

			project, err := openProject(name)
			if err != nil {
				return err
			}

			if config.Get().CradleCommandOut {
				fmt.Fprintf(os.Stderr, "eval cd %s", project.Path)
			}

			return nil
		},
	}
}

func openProject(query string) (types.CradleProject, error) {
	for _, project := range config.Get().CradleConfig.Projects {
		if project.MatchPathOrName(query) {
			return project, nil
		}
	}

	return types.CradleProject{}, fmt.Errorf("%s project not found", query)
}

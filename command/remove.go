package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"github.com/gurleensethi/cradle/internal/config"
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

			err := removeProject(name)
			if err != nil {
				return err
			}

			fmt.Println("Project removed from cradle")

			return nil
		},
	}
}

func removeProject(name string) error {
	for i, project := range config.Get().CradleConfig.Projects {
		if project.MatchPathOrName(name) {
			config.Get().CradleConfig.Projects = append(config.Get().CradleConfig.Projects[:i], config.Get().CradleConfig.Projects[i+1:]...)
			return config.UpdateConfigFile()
		}
	}

	return fmt.Errorf("project not found")
}

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:        "cradle",
		Description: "CLI to manage local projects",
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			return ctx, InitConfig()
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() > 0 {
				fmt.Printf("Unknown command: `%s`\n\n", c.Args().Get(0))
				return cli.ShowRootCommandHelp(c.Root())
			}

			program := tea.NewProgram(NewCradleUIModel(), tea.WithAltScreen())
			model, err := program.Run()
			if model, ok := model.(CradleUIModel); ok {
				if model.SelectedProjectPath != "" && config.CradleCommandOut {
					fmt.Fprintf(os.Stderr, "eval cd %s", model.SelectedProjectPath)
				}
			}
			return err
		},
		Commands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List all projects managed by cradle",
				Aliases: []string{"ls"},
				Action: func(ctx context.Context, c *cli.Command) error {
					err := ListProjects()
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:        "create",
				Usage:       "Create a new project",
				Description: "create new project",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "temp",
						DefaultText: "create a temporary project",
						Value:       false,
						Usage:       "--temp",
					},
					&cli.StringFlag{
						Name:     "template",
						Usage:    "specify a template to use for project creation",
						Required: false,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					if c.Args().Len() == 0 {
						return fmt.Errorf("provide a project name")
					}

					newProjectPath, err := CreateProject(CreateProjectParams{
						Name:     strings.Join(c.Args().Slice(), "-"),
						Temp:     c.Bool("temp"),
						Template: c.String("template"),
					})
					if err != nil {
						return err
					}

					if config.CradleCommandOut {
						fmt.Fprintf(os.Stderr, "eval cd %s", newProjectPath)
					}

					return nil
				},
			},
			{
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

					absProjectDirPath, err := AddProject(projectPath)
					if err != nil {
						return err
					}

					fmt.Println("Project added: ", absProjectDirPath)

					return nil
				},
			},
			{
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

					err := RemoveProject(name)
					if err != nil {
						return err
					}

					fmt.Println("Project removed from cradle")

					return nil
				},
			},
			{
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

					project, err := OpenProject(name)
					if err != nil {
						return err
					}

					if config.CradleCommandOut {
						fmt.Fprintf(os.Stderr, "eval cd %s", project.Path)
					}

					return nil
				},
			},
			{
				Name:  "cleanup",
				Usage: "Remove all temporary projects that were created by cradle. ",
				Action: func(ctx context.Context, c *cli.Command) error {
					_, err := CleanupTemporaryProjects()
					return err
				},
			},
			{
				Name:  "doctor",
				Usage: "Check health of cradle, find broken projects and fix problems",
				Action: func(ctx context.Context, c *cli.Command) error {
					return Doctor()
				},
			},
		},
	}

	err := cmd.Run(context.TODO(), os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

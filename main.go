package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gurleensethi/cradle/command"
	"github.com/gurleensethi/cradle/internal/config"
	"github.com/gurleensethi/cradle/internal/template"
	"github.com/urfave/cli/v3"
)

//go:embed templates/*.yaml
var templateFS embed.FS

func main() {
	// Set the template filesystem for the template package
	template.SetTemplateFS(templateFS)

	cmd := &cli.Command{
		Name:        "cradle",
		Description: "CLI to manage local projects",
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			return ctx, config.Init()
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() > 0 {
				fmt.Printf("Unknown command: `%s`\n\n", c.Args().Get(0))
				return cli.ShowRootCommandHelp(c.Root())
			}

			program := tea.NewProgram(NewCradleUIModel(), tea.WithAltScreen())
			model, err := program.Run()
			if model, ok := model.(CradleUIModel); ok {
				if model.SelectedProjectPath != "" && config.Get().CradleCommandOut {
					fmt.Fprintf(os.Stderr, "eval cd %s", model.SelectedProjectPath)
				}
			}
			return err
		},
		Commands: []*cli.Command{
			command.List(),
			command.Create(),
			command.Add(),
			command.Remove(),
			command.Open(),
			command.Cleanup(),
			command.Doctor(),
		},
	}

	err := cmd.Run(context.TODO(), os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

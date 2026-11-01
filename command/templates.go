package command

import (
	"context"
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/gurleensethi/cradle/internal/template"
	"github.com/urfave/cli/v3"
)

// Templates returns the templates command for managing project templates.
func Templates() *cli.Command {
	return &cli.Command{
		Name:  "templates",
		Usage: "Manage project templates",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List available templates",
				Action: func(ctx context.Context, c *cli.Command) error {
					return listTemplates(ctx, c)
				},
			},
		},
	}
}

// listTemplates displays all available project templates in a table.
func listTemplates(_ context.Context, _ *cli.Command) error {
	templates, err := template.ListTemplates()
	if err != nil {
		return err
	}

	rows := [][]string{}
	for _, tmpl := range templates {
		rows = append(rows, []string{tmpl.FileName, tmpl.Description})
	}

	rowStyle := lipgloss.NewStyle().Padding(0, 1)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			return rowStyle
		}).
		Headers("Template", "Description").
		Rows(rows...)

	fmt.Println(t)
	return nil
}

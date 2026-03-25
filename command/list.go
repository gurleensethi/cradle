package command

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gurleensethi/cradle/internal/config"
	"github.com/urfave/cli/v3"
)

// List returns the list command for displaying all managed projects.
func List() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Usage:   "List all projects managed by cradle",
		Aliases: []string{"ls"},
		Action: func(ctx context.Context, c *cli.Command) error {
			return listProjects(ctx, c)
		},
	}
}

func listProjects(ctx context.Context, c *cli.Command) error {
	projects := config.Projects()
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	rows := [][]string{}
	for _, project := range projects {
		var temp string
		if project.Temporary {
			temp = "Yes"
		} else {
			temp = "No"
		}

		rows = append(rows, []string{
			project.UniqueNameFromPath,
			project.Path,
			temp,
			project.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	rowStyle := lipgloss.NewStyle().Padding(0, 1)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			return rowStyle
		}).
		Headers("Name", "Path", "Temporary", "Time").
		Rows(rows...)

	fmt.Println(t)

	return nil
}

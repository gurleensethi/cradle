package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CradleUIModel struct {
	table               table.Model
	SelectedProjectPath string
}

func NewCradleUIModel() CradleUIModel {
	t := table.New(
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return CradleUIModel{
		table: t,
	}
}

func (c CradleUIModel) Init() tea.Cmd {
	return nil
}

func (c CradleUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		width := msg.Width - 8

		columns := []table.Column{
			{Title: "Name", Width: ((20 * width) / 100)},
			{Title: "Path", Width: ((40 * width) / 100)},
			{Title: "Temporary", Width: (10 * width) / 100},
			{Title: "Time", Width: (30 * width) / 100},
		}

		var rows []table.Row
		for _, project := range config.CradleConfig.Projects {
			rows = append(rows, table.Row{
				project.UniqueNameFromPath,
				project.Path,
				fmt.Sprintf("%v", project.Temporary),
				project.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		c.table.SetColumns(columns)
		c.table.SetRows(rows)

		c.table.SetWidth(width)
		c.table.SetHeight(msg.Height - 2)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return c, tea.Quit
		case "enter":
			c.SelectedProjectPath = c.table.SelectedRow()[1]
			return c, tea.Quit
		}
	default:
		_ = msg
	}

	c.table, cmd = c.table.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c CradleUIModel) View() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Render(c.table.View())
}

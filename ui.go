package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CradleUIModel struct {
	SelectedProjectPath string
	List                list.Model
	Width               int
	Height              int
}

type ProjectListItem struct {
	Project CradleProject
}

func (p ProjectListItem) Title() string       { return p.Project.UniqueNameFromPath }
func (p ProjectListItem) Description() string { return p.Project.Path }
func (p ProjectListItem) FilterValue() string {
	return p.Project.UniqueNameFromPath + " " + p.Project.Path
}

type ProjectListDeletegate struct{}

func (p ProjectListDeletegate) Height() int  { return 3 }
func (p ProjectListDeletegate) Spacing() int { return 0 }

func (p ProjectListDeletegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	projectItem, ok := item.(ProjectListItem)
	if !ok {
		return
	}

	// Style for the title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Width(m.Width()).
		Foreground(lipgloss.AdaptiveColor{
			Light: "0",
			Dark:  "209",
		})

	// Base style for each item
	style := lipgloss.NewStyle().
		Width(m.Width()).
		Margin(0, 2, 1, 2).
		PaddingLeft(1).
		PaddingRight(1)

	// Style for temporary project indicator
	tempStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#FFFF00",
			Dark:  "#FFFF00",
		})

	if index == m.Index() {
		style = style.
			Background(lipgloss.AdaptiveColor{
				Light: "#D3D3D3",
				Dark:  "#484848ff",
			}).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: "209",
				Dark:  "209",
			})

		titleStyle = titleStyle.Bold(true)
	} else {
		style = style.
			Foreground(lipgloss.AdaptiveColor{
				Light: "240",
				Dark:  "250",
			}).
			PaddingLeft(2)
	}

	tempState := ""
	if projectItem.Project.Temporary {
		tempState = tempStyle.Render("(temporary)")
	}
	title := titleStyle.Render(projectItem.Project.UniqueNameFromPath + " " + tempState)

	str := lipgloss.JoinVertical(lipgloss.Left,
		title,
		projectItem.Project.Path,
	)

	fmt.Fprint(w, style.Render(str))
}

func (p ProjectListDeletegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func NewCradleUIModel() CradleUIModel {
	var listItems []list.Item
	for _, project := range config.CradleConfig.Projects {
		listItems = append(listItems, ProjectListItem{Project: project})
	}

	projectList := list.New(listItems, ProjectListDeletegate{}, 0, 0)
	projectList.Title = "Select a project"
	projectList.Styles.Title = lipgloss.NewStyle().Bold(true)

	return CradleUIModel{
		List: projectList,
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
		c.Height = msg.Height
		c.Width = msg.Width
		width := msg.Width - 8
		c.List.SetHeight(msg.Height - 7)
		c.List.SetWidth(MinInt(width, 100))
	case tea.KeyMsg:
		if c.List.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return c, tea.Quit
		case "enter":
			selectedItem, ok := c.List.SelectedItem().(ProjectListItem)
			if ok {
				c.SelectedProjectPath = selectedItem.Project.Path
				return c, tea.Quit
			}
		}
	default:
		_ = msg
	}

	c.List, cmd = c.List.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c CradleUIModel) Title() string {
	return lipgloss.NewStyle().
		Width(c.Width).
		Height(4).
		MarginBottom(1).
		// Foreground(lipgloss.Color("255")).
		// Background(lipgloss.Color("100")).
		Bold(true).
		Align(lipgloss.Center).
		Render(` ▗▄▄▖▗▄▄▖  ▗▄▖ ▗▄▄▄  ▗▖   ▗▄▄▄▖
▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌  █ ▐▌   ▐▌   
▐▌   ▐▛▀▚▖▐▛▀▜▌▐▌  █ ▐▌   ▐▛▀▀▘
▝▚▄▄▖▐▌ ▐▌▐▌ ▐▌▐▙▄▄▀ ▐▙▄▄▖▐▙▄▄▖`)
}

func (c CradleUIModel) View() string {
	return lipgloss.NewStyle().
		Width(c.Width).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				c.Title(),
				lipgloss.NewStyle().
					Render(
						lipgloss.NewStyle().
							Border(lipgloss.RoundedBorder()).
							BorderForeground(lipgloss.Color("#7a7a7aff")).
							Render(c.List.View()),
					),
			),
		)
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gurleensethi/cradle/internal/config"
	"github.com/gurleensethi/cradle/internal/types"
)

// CradleUIModel is the main TUI model.
type CradleUIModel struct {
	SelectedProjectPath string
	ProjectList         list.Model
	Width               int
	Height              int
}

type ProjectListItem struct {
	Project types.CradleProject
}

func (p ProjectListItem) Title() string { return p.Project.UniqueNameFromPath }

func (p ProjectListItem) Description() string { return p.Project.Path }

func (p ProjectListItem) FilterValue() string {
	return p.Project.UniqueNameFromPath + " " + p.Project.Path
}

type ProjectListDelegate struct{}

func (p ProjectListDelegate) Height() int { return 3 }

func (p ProjectListDelegate) Spacing() int { return 0 }

func (p ProjectListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	projectItem, ok := item.(ProjectListItem)
	if !ok {
		return
	}

	isSelectedItem := index == m.Index()

	// ========== Styles ==========
	nonSelectedTitle := lipgloss.NewStyle().
		Bold(true).
		Width(m.Width()).
		Foreground(lipgloss.AdaptiveColor{
			Light: "0",
			Dark:  "209",
		})
	selectedTitle := nonSelectedTitle.Bold(true)
	// ============================

	// Style for the title
	titleStyle := nonSelectedTitle

	// Base style for each item
	style := lipgloss.NewStyle().
		Width(m.Width()-3).
		Margin(0, 1, 0, 1).
		PaddingLeft(1).
		PaddingRight(1)

	// Style for temporary project indicator
	tempStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#FFFF00",
			Dark:  "#FFFF00",
		})

	if isSelectedItem {
		style = style.
			Background(lipgloss.AdaptiveColor{
				Light: "#D3D3D3",
				Dark:  "#484848",
			}).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: "209",
				Dark:  "209",
			})

		titleStyle = selectedTitle
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
		projectItem.Project.GetPathWithTruncatedHome(),
	)

	fmt.Fprint(w, style.Render(str))
}

// Update performs no custom updates.
func (p ProjectListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// NewCradleUIModel returns a new TUI model populated with projects.
func NewCradleUIModel() CradleUIModel {
	var listItems []list.Item
	for _, project := range config.Projects() {
		listItems = append(listItems, ProjectListItem{Project: project})
	}

	projectList := list.New(listItems, ProjectListDelegate{}, 0, 0)
	projectList.SetShowTitle(false)
	projectList.FilterInput.Prompt = "Search: "
	projectList.FilterInput.PromptStyle = lipgloss.NewStyle()

	return CradleUIModel{
		ProjectList: projectList,
	}
}

func (c CradleUIModel) Init() tea.Cmd {
	return nil
}

// Update handles TUI events and returns the updated model.
func (c CradleUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.Height = msg.Height
		c.Width = msg.Width
		c.ProjectList.SetHeight(msg.Height - 3)
		c.ProjectList.SetWidth(msg.Width)
	case tea.KeyMsg:
		if c.ProjectList.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return c, tea.Quit
		case "enter":
			selectedItem, ok := c.ProjectList.SelectedItem().(ProjectListItem)
			if ok {
				c.SelectedProjectPath = selectedItem.Project.Path
				return c, tea.Quit
			}
		}
	default:
		_ = msg
	}

	c.ProjectList, cmd = c.ProjectList.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c CradleUIModel) Title() string {
	return lipgloss.NewStyle().
		Width(c.Width).
		MarginBottom(1).
		Bold(true).
		Align(lipgloss.Center).
		Background(lipgloss.Color("#ff7300")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render("cradle")
}

func (c CradleUIModel) View() string {
	return lipgloss.NewStyle().
		Width(c.Width).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				c.Title(),
				lipgloss.NewStyle().
					Render(
						c.ProjectList.View(),
					),
			),
		)
}

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

// CradleUIModel implements the tea.Model interface for the main TUI.
// It holds the project list state and the selected project path after user confirmation.
type CradleUIModel struct {
	// SelectedProjectPath stores the path of the project selected by the user.
	SelectedProjectPath string
	// ProjectList is the bubbletea list component displaying projects.
	ProjectList list.Model
	// Width is the terminal width used for rendering.
	Width int
	// Height is the terminal height used for rendering.
	Height int
}

// ProjectListItem implements list.Item for displaying a project in the TUI list.
type ProjectListItem struct {
	Project types.CradleProject
}

// Title returns the display title for the list item (the project's unique name).
func (p ProjectListItem) Title() string { return p.Project.UniqueNameFromPath }

// Description returns the display description for the list item (the project path).
func (p ProjectListItem) Description() string { return p.Project.Path }

// FilterValue returns the text used for filtering list items.
func (p ProjectListItem) FilterValue() string {
	return p.Project.UniqueNameFromPath + " " + p.Project.Path
}

// ProjectListDelegate implements list.ItemDelegate for rendering project list items.
type ProjectListDelegate struct{}

// Height returns the height of a single rendered list item.
func (p ProjectListDelegate) Height() int { return 3 }

// Spacing returns the spacing between list items.
func (p ProjectListDelegate) Spacing() int { return 0 }

// Render draws a single list item to the writer with appropriate styling.
func (p ProjectListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
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
		Margin(0, 1, 1, 1).
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
		projectItem.Project.GetPathWithTruncatedHome(),
	)

	fmt.Fprint(w, style.Render(str))
}

// Update handles tea.Msg for the project list delegate.
// It performs no custom updates and returns nil.
func (p ProjectListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// NewCradleUIModel creates and initializes a new CradleUIModel with projects from the config.
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

// Init implements tea.Model.Init. It performs no initialization and returns nil.
func (c CradleUIModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.Update. It handles window sizing, key events,
// and delegates list updates. Returns the updated model and a quit command
// when the user selects a project or exits.
func (c CradleUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.Height = msg.Height
		c.Width = msg.Width
		width := msg.Width - 2
		c.ProjectList.SetHeight(msg.Height - 3)
		c.ProjectList.SetWidth(min(width))
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

// Title returns the ASCII art banner for the TUI.
func (c CradleUIModel) Title() string {
	return lipgloss.NewStyle().
		Width(c.Width - 1).
		MarginBottom(1).
		Bold(true).
		Align(lipgloss.Center).
		Background(lipgloss.Color("#ff7300")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render(`cradle`)
}

// View implements tea.Model.View. It renders the full TUI layout with the
// title banner and the bordered project list.
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

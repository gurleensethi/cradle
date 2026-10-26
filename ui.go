package main

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
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

// Title returns the project's unique name for display in the list.
func (p ProjectListItem) Title() string { return p.Project.UniqueNameFromPath }

// Description returns the project's full path shown beneath its title.
func (p ProjectListItem) Description() string { return p.Project.Path }

// FilterValue returns the text matched against when filtering projects.
func (p ProjectListItem) FilterValue() string {
	return p.Project.UniqueNameFromPath
}

type ProjectListDelegate struct{}

// Height returns the number of terminal lines a single list item occupies.
func (p ProjectListDelegate) Height() int { return 2 }

// Spacing returns the number of blank lines between list items.
func (p ProjectListDelegate) Spacing() int { return 0 }

// Render draws a single project item in the list.
func (p ProjectListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	projectItem, ok := item.(ProjectListItem)
	if !ok {
		return
	}

	isSelectedItem := index == m.Index()

	// ========== Styles ==========
	nonSelectedTitle := lipgloss.NewStyle().
		Bold(false).
		Width(m.Width()).
		Foreground(compat.AdaptiveColor{
			Light: lipgloss.Color("0"),
			Dark:  lipgloss.Color("#FFFFFF"),
		})
	selectedTitle := nonSelectedTitle.Bold(true)
	subTitle := lipgloss.NewStyle().Faint(true)
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
		Foreground(compat.AdaptiveColor{
			Light: lipgloss.Color("#B8860B"),
			Dark:  lipgloss.Color("#FFD700"),
		})

	if isSelectedItem {
		style = style.
			Background(compat.AdaptiveColor{
				Light: lipgloss.Color("#E5E5E5"),
				Dark:  lipgloss.Color("#2A2A2A"),
			}).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(compat.AdaptiveColor{
				Light: lipgloss.Color("#808080"),
				Dark:  lipgloss.Color("#808080"),
			})

		titleStyle = selectedTitle
	} else {
		style = style.
			Foreground(compat.AdaptiveColor{
				Light: lipgloss.Color("0"),
				Dark:  lipgloss.Color("#FFFFFF"),
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
		subTitle.Render(projectItem.Project.GetPathWithTruncatedHome()),
	)

	fmt.Fprint(w, style.Render(str))
}

// Update performs no custom updates.
// Update handles item-level updates (none for this delegate).
func (p ProjectListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// NewCradleUIModel returns a new TUI model populated with projects.
// NewCradleUIModel builds the TUI model populated with configured projects.
func NewCradleUIModel() CradleUIModel {
	var listItems []list.Item
	for _, project := range config.Projects() {
		listItems = append(listItems, ProjectListItem{Project: project})
	}

	projectList := list.New(listItems, ProjectListDelegate{}, 0, 0)
	projectList.SetShowTitle(false)
	projectList.SetShowHelp(false)
	projectList.SetShowStatusBar(false)
	projectList.FilterInput.Prompt = "Search: "

	filterStyles := projectList.FilterInput.Styles()
	filterStyles.Focused.Prompt = lipgloss.NewStyle()
	filterStyles.Blurred.Prompt = lipgloss.NewStyle()
	projectList.FilterInput.SetStyles(filterStyles)

	var selectedProjectPath string
	if len(config.Projects()) > 0 {
		selectedProjectPath = config.Projects()[0].Path
	}

	return CradleUIModel{
		ProjectList:         projectList,
		SelectedProjectPath: selectedProjectPath,
	}
}

// Init returns the initial command (none).
func (c CradleUIModel) Init() tea.Cmd {
	return nil
}

// Update handles TUI events and returns the updated model.
// Update processes TUI messages and updates the model.
func (c CradleUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.Height = msg.Height
		c.Width = msg.Width
		c.ProjectList.SetSize(msg.Width, msg.Height-6)
	case tea.KeyPressMsg:
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

// Header renders the top title bar.
func (c CradleUIModel) Header() string {
	return lipgloss.NewStyle().
		Padding(1).
		MarginBottom(1).
		Bold(true).
		Width(c.Width).
		Align(lipgloss.Center).
		Background(lipgloss.Color("#2D2D2D")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render("cradle")
}

// Footer renders the bottom bar.
func (c CradleUIModel) Footer() string {
	textStyle := lipgloss.NewStyle().Faint(true)

	return lipgloss.NewStyle().
		MarginTop(1).
		Padding(0, 2).
		Bold(true).
		Width(c.Width).
		Align(lipgloss.Left).
		Background(lipgloss.Color("#2D2D2D")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render(
			textStyle.Render(fmt.Sprintf("%d projects", len(config.Projects()))),
		)
}

// View renders the full TUI.
func (c CradleUIModel) View() tea.View {
	view := tea.NewView(
		lipgloss.NewStyle().
			Width(c.Width).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					c.Header(),
					lipgloss.NewStyle().
						Render(
							c.ProjectList.View(),
						),
					c.Footer(),
				),
			),
	)

	view.AltScreen = true

	return view
}

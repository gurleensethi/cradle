package main

import (
	"fmt"
	"io"
	"strings"

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
		Faint(true).
		Foreground(compat.AdaptiveColor{
			Light: lipgloss.Color("0"),
			Dark:  lipgloss.Color("#ff7300"),
		})
	selectedTitle := nonSelectedTitle.Bold(true).
		Faint(false)
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
			Light: lipgloss.Color("#FFFF00"),
			Dark:  lipgloss.Color("#FFFF00"),
		})

	if isSelectedItem {
		style = style.
			Background(compat.AdaptiveColor{
				Light: lipgloss.Color("#D3D3D3"),
				Dark:  lipgloss.Color("#484848"),
			}).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(compat.AdaptiveColor{
				Light: lipgloss.Color("209"),
				Dark:  lipgloss.Color("209"),
			})

		titleStyle = selectedTitle
	} else {
		style = style.
			Foreground(compat.AdaptiveColor{
				Light: lipgloss.Color("240"),
				Dark:  lipgloss.Color("250"),
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
	projectList.SetShowHelp(false)
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
		c.ProjectList.SetSize(msg.Width, msg.Height-3)
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

func (c CradleUIModel) Title() string {
	return lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		MarginBottom(1).
		Bold(true).
		Align(lipgloss.Left).
		Background(lipgloss.Color("#ff7300")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render("cradle")
}

func (c CradleUIModel) ItemList() string {
	strBuilder := strings.Builder{}

	// Marker follows the list cursor so arrow keys visibly move the selection.
	cursorItem, _ := c.ProjectList.SelectedItem().(ProjectListItem)

	for _, project := range config.Projects() {
		prefix := " "
		if project.Path == cursorItem.Project.Path {
			prefix = ">"
		}

		strBuilder.WriteString(prefix + project.UniqueNameFromPath + " " + "\n")
	}

	return strBuilder.String()
}

func (c CradleUIModel) View() tea.View {
	view := tea.NewView(
		lipgloss.NewStyle().
			Width(c.Width).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					c.Title(),
					c.ItemList(),
					lipgloss.NewStyle().
						Render(
						// c.ProjectList.View(),
						),
				),
			),
	)

	view.AltScreen = true

	return view
}

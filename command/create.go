package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/gurleensethi/cradle/internal/config"
	cradleTemplate "github.com/gurleensethi/cradle/internal/template"
	"github.com/gurleensethi/cradle/internal/types"
	"github.com/urfave/cli/v3"
)

// Create returns the create command for creating new projects.
func Create() *cli.Command {
	return &cli.Command{
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
			&cli.BoolFlag{
				Name:     "i",
				Usage:    "create project interactively, launches a TUI",
				Required: false,
				Value:    false,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Bool("i") {
				newProjectPath, err := createProjectInteractive()
				if err != nil {
					return err
				}

				if config.Get().CradleCommandOut {
					fmt.Fprintf(os.Stderr, "eval cd %s", newProjectPath)
				}

				return nil
			}

			if c.Args().Len() == 0 {
				return fmt.Errorf("provide a project name")
			}

			newProjectPath, err := createProject(createProjectParams{
				Name:     strings.Join(c.Args().Slice(), "-"),
				Temp:     c.Bool("temp"),
				Template: c.String("template"),
			})
			if err != nil {
				return err
			}

			if config.Get().CradleCommandOut {
				fmt.Fprintf(os.Stderr, "eval cd %s", newProjectPath)
			}

			return nil
		},
	}
}

type createProjectParams struct {
	Name     string
	Temp     bool
	Template string
}

func createProjectInteractive() (string, error) {
	program := tea.NewProgram(newCreateProjectModel())

	returnModel, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("failed to launch in interactive mode: %w", err)
	}

	if createProjectModel, ok := returnModel.(createProjectModel); ok {
		fmt.Println(createProjectModel)
	}

	return "", nil
}

// createProject prepares the project files from the template (if any) and
// persists the project to disk and config. Returns the created project path.
func createProject(params createProjectParams) (string, error) {
	newProjectPath := path.Join(config.Get().CradleHomeDirPath, params.Name)

	// Make sure there is no existing project with same name
	for _, project := range config.Projects() {
		if project.Path == newProjectPath {
			return "", fmt.Errorf("project already exists")
		}
	}

	files := []projectFile{}

	// If a template is specified, use it to create the project
	if params.Template != "" {
		templateData, err := cradleTemplate.GetTemplate(params.Template)
		if err != nil {
			if errors.Is(err, cradleTemplate.ErrNotExists) {
				return "", fmt.Errorf("template %s does not exist", params.Template)
			}

			return "", err
		}

		userInputs, err := cradleTemplate.ReadUserInputs(templateData)
		if err != nil {
			return "", err
		}

		templateInput := map[string]any{
			"ProjectName": params.Name,
		}

		maps.Copy(templateInput, userInputs)

		for filePath, templateFile := range templateData.Files {
			buffer := bytes.NewBuffer([]byte{})

			t, err := template.New("").Parse(string(templateFile))
			if err != nil {
				// TODO: wrap the error with the failing file path and the
				// underlying parse reason for easier debugging.
				return "", err
			}

			err = t.Execute(buffer, templateInput)
			if err != nil {
				return "", err
			}

			files = append(files, projectFile{
				Path:    filePath,
				Content: buffer.String(),
			})
		}
	}

	err := createProjectOnDisk(newProjectPath, files)
	if err != nil {
		return "", err
	}

	cradleProject := types.CradleProject{
		Path:      newProjectPath,
		Temporary: params.Temp,
		CreatedAt: time.Now(),
		CreatedBy: "cradle",
	}

	return newProjectPath, config.AddProject(cradleProject)
}

// projectFile represents a file that is to be created inside a project.
type projectFile struct {
	Path    string
	Content string
}

// createProjectOnDisk creates the project folder and provided files on disk.
func createProjectOnDisk(projectPath string, files []projectFile) error {
	// Create the project directory
	err := os.Mkdir(projectPath, os.ModePerm)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("failed: '%s' directory already exists", projectPath)
		}
		return err
	}

	for _, file := range files {
		// Make sure the full directory path exists
		directory := path.Join(projectPath, filepath.Dir(file.Path))

		err = os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", directory, err)
		}

		err := os.WriteFile(path.Join(projectPath, file.Path), []byte(file.Content), 0o644)
		if err != nil {
			// TODO: include the project path in the error so callers can tell
			// which file failed to be written.
			return fmt.Errorf("failed to write to file '%s': %w", file.Path, err)
		}
	}

	return nil
}

type createProjectModel struct {
	UserInputProjectName     string
	UserInputProjectTemplate string

	Form      *huh.Form
	FormModel huh.Model

	WindowSize tea.WindowSizeMsg
}

// Init returns the initial command.
func (m createProjectModel) Init() tea.Cmd {
	return m.Form.Init()
}

// Update handles messages for the create project model.
func (m createProjectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowSize = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	_, cmd = m.Form.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the create project model.
func (m createProjectModel) View() tea.View {
	view := tea.View{}

	width := m.WindowSize.Width
	height := m.WindowSize.Height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// Header bar, consistent with the main Cradle TUI.
	header := lipgloss.NewStyle().
		Padding(1).
		MarginBottom(1).
		Bold(true).
		Width(width).
		Align(lipgloss.Center).
		Background(lipgloss.Color("#2D2D2D")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render("cradle  ·  create new project")

	// Footer bar, consistent with the main Cradle TUI.
	footer := lipgloss.NewStyle().
		MarginTop(1).
		Padding(0, 2).
		Bold(true).
		Width(width).
		Align(lipgloss.Left).
		Background(lipgloss.Color("#2D2D2D")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render("esc to cancel")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#808080")).
		Padding(2, 4)

	form := m.Form.WithWidth(width / 2).WithHeight(height / 2).View()

	// Body fills the remaining height so the footer sits at the bottom.
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	bodyHeight := height - headerHeight - footerHeight
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	body := lipgloss.NewStyle().
		Width(width).
		Height(bodyHeight).
		MaxHeight(bodyHeight).
		AlignVertical(lipgloss.Center).
		AlignHorizontal(lipgloss.Center).
		Render(boxStyle.Render(form))

	view.SetContent(
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			body,
			footer,
		),
	)

	view.AltScreen = true

	return view
}

func newCreateProjectModel() createProjectModel {
	templates, err := cradleTemplate.ListTemplates()
	if err != nil {
		// TODO: handle error properly.
		panic(err)
	}

	templateOptions := []huh.Option[string]{
		huh.NewOption("Empty", "empty"),
	}
	for _, template := range templates {
		templateOptions = append(templateOptions, huh.NewOption(template.Name, template.FileName))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("Name of the project to create.").
				Placeholder("my-project").
				Key("project_name"),
			huh.NewConfirm().
				Title("Temporary project").
				Description("Mark the project as temporary so it can be cleaned up later.").
				Key("temporary"),
			huh.NewSelect[string]().
				Title("Template").
				Description("Template used to scaffold the project.").
				Options(templateOptions...).
				Key("template"),
		),
	).WithShowHelp(true)

	return createProjectModel{
		Form: form,
	}
}

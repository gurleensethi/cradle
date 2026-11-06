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
		},
		Action: func(ctx context.Context, c *cli.Command) error {
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

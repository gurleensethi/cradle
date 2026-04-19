package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path"
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

// createProject creates a project directory and registers it. Returns the created path.
func createProject(params createProjectParams) (string, error) {
	newProjectPath := path.Join(config.Get().CradleHomeDirPath, params.Name)

	// Make sure there is no existing project with same name
	for _, project := range config.Projects() {
		if project.Path == newProjectPath {
			return "", fmt.Errorf("project already exists")
		}
	}

	files := make(map[string]string)

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

		templateInput := map[string]string{
			"ProjectName": params.Name,
		}

		maps.Copy(templateInput, userInputs)

		for filePath, templateFile := range templateData.Files {
			buffer := bytes.NewBuffer([]byte{})

			t, err := template.New("").Parse(string(templateFile))
			if err != nil {
				// TODO: add better error message, denoting which file failed
				// TODO: rendering and why.
				return "", err
			}

			err = t.Execute(buffer, templateInput)
			if err != nil {
				return "", err
			}

			files[filePath] = buffer.String()
		}
	}

	// Create the project directory
	err := os.Mkdir(newProjectPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	for filePath, fileContent := range files {
		err := os.WriteFile(path.Join(newProjectPath, filePath), []byte(fileContent), 0o644)
		if err != nil {
			// TODO: return better error message, denoting which file failed to be created.
			return "", err
		}
	}

	cradleProject := types.CradleProject{
		Path:      newProjectPath,
		Temporary: params.Temp,
		CreatedAt: time.Now(),
		CreatedBy: "cradle",
	}

	return newProjectPath, config.AddProject(cradleProject)
}

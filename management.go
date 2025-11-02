package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

func OpenProject(query string) (CradleProject, error) {
	for _, project := range config.CradleConfig.Projects {
		if path.Base(project.Path) == query {
			return project, nil
		}
	}

	return CradleProject{}, fmt.Errorf("`%s` project not found", query)
}

func ListProjects() error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Project Path", "Temporary", "Created Time"})

	for i, project := range config.CradleConfig.Projects {
		t.AppendRow(table.Row{i + 1, project.Path, project.IsTemporary, project.CreatedAt.Format("2006-01-02 15:04:05")})
	}

	t.Render()

	return nil
}

type CreateProjectParams struct {
	Name string
	Temp bool
}

func CreateProject(params CreateProjectParams) (string, error) {
	newProjectPath := path.Join(config.CradleHomeDirPath, params.Name)

	// Make sure there is no existing project with same name
	for _, project := range config.CradleConfig.Projects {
		if project.Path == newProjectPath {
			return "", fmt.Errorf("project already exists")
		}
	}

	// Create the project directory
	err := os.Mkdir(newProjectPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	cradleProject := CradleProject{
		Path:        newProjectPath,
		IsTemporary: params.Temp,
		CreatedAt:   time.Now(),
	}

	config.CradleConfig.Projects = append(config.CradleConfig.Projects, cradleProject)

	return newProjectPath, UpdateCradleConfigFile()
}

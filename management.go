package main

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func OpenProject(query string) (CradleProject, error) {
	for _, project := range config.CradleConfig.Projects {
		if project.MatchPathOrName(query) {
			return project, nil
		}
	}

	return CradleProject{}, fmt.Errorf("%s project not found", query)
}

func ListProjects() error {
	if len(config.CradleConfig.Projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	rows := [][]string{}
	for _, project := range config.CradleConfig.Projects {
		var temp string
		if project.Temporary {
			temp = "Yes"
		} else {
			temp = "No"
		}

		rows = append(rows, []string{
			project.UniqueNameFromPath,
			project.Path,
			temp,
			project.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	rowStyle := lipgloss.NewStyle().Padding(0, 1)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			return rowStyle
		}).
		Headers("Name", "Path", "Temporary", "Time").
		Rows(rows...)

	fmt.Println(t)

	return nil
}

type CreateProjectParams struct {
	Name     string
	Temp     bool
	Template string
}

func CreateProject(params CreateProjectParams) (string, error) {
	newProjectPath := path.Join(config.CradleHomeDirPath, params.Name)

	// Make sure there is no existing project with same name
	for _, project := range config.CradleConfig.Projects {
		if project.Path == newProjectPath {
			return "", fmt.Errorf("project already exists")
		}
	}

	files := make(map[string]string)

	// If a template is specified, use it to create the project
	if params.Template != "" {
		templateData, err := GetTemplate(params.Template)
		if err != nil {
			if errors.Is(err, ErrNotExists) {
				return "", fmt.Errorf("template %s does not exist", params.Template)
			}

			return "", err
		}

		userInputs, err := ReadUserInputs(templateData)
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
		err := os.WriteFile(path.Join(newProjectPath, filePath), []byte(fileContent), 0644)
		if err != nil {
			// TODO: return better error message, denoting which file failed to be created.
			return "", err
		}
	}

	cradleProject := CradleProject{
		Path:      newProjectPath,
		Temporary: params.Temp,
		CreatedAt: time.Now(),
		CreatedBy: "cradle",
	}

	config.CradleConfig.Projects = append(config.CradleConfig.Projects, cradleProject)

	return newProjectPath, UpdateCradleConfigFile()
}

func AddProject(projectDirPath string) (string, error) {
	projectDirPath, err := filepath.Abs(projectDirPath)
	if err != nil {
		return "", err
	}

	dirStat, err := os.Stat(projectDirPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("project directory with specified path doesn't exist")
		}

		return "", err
	}

	if !dirStat.IsDir() {
		return "", fmt.Errorf("specified path is not a directory")
	}

	// Make sure there is no existing project with same name
	for _, project := range config.CradleConfig.Projects {
		if project.Path == projectDirPath {
			return "", fmt.Errorf("project already exists")
		}
	}

	cradleProject := CradleProject{
		Path:      projectDirPath,
		Temporary: false,
		CreatedAt: time.Now(),
	}

	config.CradleConfig.Projects = append(config.CradleConfig.Projects, cradleProject)

	return projectDirPath, UpdateCradleConfigFile()
}

func RemoveProject(name string) error {
	for i, project := range config.CradleConfig.Projects {
		if project.MatchPathOrName(name) {
			config.CradleConfig.Projects = append(config.CradleConfig.Projects[:i], config.CradleConfig.Projects[i+1:]...)
			return UpdateCradleConfigFile()
		}
	}

	return fmt.Errorf("project not found")
}

func CleanupTemporaryProjects() (int, error) {
	var count int
	for _, project := range config.CradleConfig.Projects {
		if project.Temporary {
			count++
		}
	}

	if count == 0 {
		fmt.Println("No temporary projects to clean up.")
		return 0, nil
	}

	var confirmation string
	fmt.Printf("Are you sure you want to delete %d temporary projects (Y/N):", count)
	fmt.Scanln(&confirmation)

	if confirmation != "Y" && confirmation != "y" {
		fmt.Println("Cleanup aborted.")
		return 0, nil
	}

	updatedProjects := []CradleProject{}
	for _, project := range config.CradleConfig.Projects {
		if project.Temporary {
			// Remove the project directory
			err := os.RemoveAll(project.Path)
			if err != nil {
				return 0, err
			}
		} else {
			updatedProjects = append(updatedProjects, project)
		}
	}

	fmt.Printf("Removed %d temporary projects.\n", count)

	config.CradleConfig.Projects = updatedProjects

	return len(updatedProjects), UpdateCradleConfigFile()
}

func Doctor() error {
	var issues []string

	for _, project := range config.CradleConfig.Projects {
		stat, err := os.Stat(project.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				issues = append(issues, fmt.Sprintf("- %s: project path does not exist", project.Path))
			} else {
				// Collect all errors instead of returning early for generic os.Stat errors
				issues = append(issues, fmt.Sprintf("- %s: failed to access project path: %v", project.Path, err))
			}
			continue
		}

		if !stat.IsDir() {
			issues = append(issues, fmt.Sprintf("- %s: project path is a file, not a directory", project.Path))
		}
	}

	if len(issues) == 0 {
		fmt.Println("No issues found ✓")
	} else {
		fmt.Printf("Found %d issues:\n", len(issues))
		fmt.Println(strings.Join(issues, "\n"))
	}

	return nil
}

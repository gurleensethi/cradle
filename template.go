package main

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"

	"github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*.yaml
var templateFS embed.FS

var ErrNotExists = errors.New("template does not exist")

type TemplateInput struct {
	Name     string `yaml:"name"`
	Default  string `yaml:"default"`
	Required bool   `yaml:"required"`
	Help     string `yaml:"help"`
}

type TemplateFile string

type TemplateData struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Inputs      []TemplateInput         `yaml:"inputs"`
	Files       map[string]TemplateFile `yaml:"files"`
}

func GetTemplate(templateName string) (*TemplateData, error) {
	// Check if the template exists in the embedded filesystem
	data, err := templateFS.ReadFile("templates/" + templateName + ".yaml")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotExists
		}

		return nil, err
	}

	var template TemplateData
	buffer := bytes.NewBuffer(data)

	yamlDecoder := yaml.NewDecoder(buffer)

	err = yamlDecoder.Decode(&template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

func ReadUserInputs(td *TemplateData) (map[string]string, error) {
	userInputs := make(map[string]string)
	for _, input := range td.Inputs {
		userInputs[input.Name] = input.Default
	}

	var fields []huh.Field
	for _, input := range td.Inputs {
		field := huh.NewInput().
			Title(input.Name).
			Description(input.Help).
			Validate(func(s string) error {
				if input.Required && s == "" {
					return errors.New("this field is required")
				}
				return nil
			}).
			Key(input.Name)

		fields = append(fields, field)
	}

	form := huh.NewForm(
		huh.NewGroup(fields...),
	)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	return userInputs, nil
}

package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*.yaml
var templateFS embed.FS

var ErrNotExists = errors.New("template does not exist")

type FloatValidationRule struct {
	Message string   `yaml:"message"`
	Min     *float64 `yaml:"min,omitempty"`
	Max     *float64 `yaml:"max,omitempty"`
}

type StringValidationRule struct {
	Pattern string `yaml:"pattern"`
	Message string `yaml:"message"`
	MinLen  *int   `yaml:"min_len,omitempty"`
	MaxLen  *int   `yaml:"max_len,omitempty"`
}

type TemplateInputValidation struct {
	Float  *FloatValidationRule  `yaml:"float,omitempty"`
	String *StringValidationRule `yaml:"string,omitempty"`
}

func (tiv *TemplateInputValidation) Validate(input string) string {
	if tiv.Float != nil {
		numValue, err := strconv.ParseFloat(input, 64)
		if err != nil {
			if tiv.Float.Message != "" {
				return tiv.Float.Message
			}

			return "invalid number"
		}

		if tiv.Float.Min != nil && float64(numValue) < *tiv.Float.Min {
			return fmt.Sprintf("number must be greater than %v", *tiv.Float.Min)
		}

		if tiv.Float.Max != nil && float64(numValue) > *tiv.Float.Max {
			return fmt.Sprintf("number must be less than %v", *tiv.Float.Max)
		}
	}

	if tiv.String != nil {
		// Validate as string
		// (Implementation of string validation can be added here)
	}

	return ""
}

type TemplateInput struct {
	Name     string                  `yaml:"name"`
	Default  string                  `yaml:"default"`
	Required bool                    `yaml:"required"`
	Help     string                  `yaml:"help"`
	Pattern  string                  `yaml:"pattern"`
	Validate TemplateInputValidation `yaml:"validate,omitempty"`
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

	if len(td.Inputs) == 0 {
		return userInputs, nil
	}

	var fields []huh.Field
	for _, input := range td.Inputs {
		field := huh.NewInput().
			Title(input.Name).
			Description(input.Help).
			Validate(func(s string) error {
				if input.Required && s == "" {
					return errors.New(input.Name + " is required")
				}

				validationMessage := input.Validate.Validate(s)
				if validationMessage != "" {
					return errors.New(input.Name + " " + validationMessage)
				}

				userInputs[input.Name] = s

				return nil
			}).
			Key(input.Name)

		fields = append(fields, field)
	}

	form := huh.NewForm(
		huh.NewGroup(fields...),
	).WithProgramOptions(tea.WithOutput(os.Stdout))

	err := form.Run()
	if err != nil {
		return nil, err
	}

	return userInputs, nil
}

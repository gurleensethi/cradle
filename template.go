package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*.yaml
var templateFS embed.FS

var ErrNotExists = errors.New("template does not exist")

type TemplateInputValidation struct {
	Pattern string   `yaml:"pattern"`
	Message string   `yaml:"message"`
	MinLen  *int     `yaml:"min_len,omitempty"`
	MaxLen  *int     `yaml:"max_len,omitempty"`
	Min     *float64 `yaml:"min,omitempty"`
	Max     *float64 `yaml:"max,omitempty"`
}

func (tiv *TemplateInputValidation) Validate(inputType, input string) string {
	if inputType == "float" {
		numValue, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return "must be a valid float"
		}

		if tiv.Min != nil && float64(numValue) < *tiv.Min {
			return fmt.Sprintf("number must be greater than %v", *tiv.Min)
		}

		if tiv.Max != nil && float64(numValue) > *tiv.Max {
			return fmt.Sprintf("number must be less than %v", *tiv.Max)
		}
	}

	if inputType == "string" {
		// Validate as string
		// (Implementation of string validation can be added here)

		if tiv.Pattern != "" {
			regex, err := regexp.Compile(tiv.Pattern)
			if err != nil {
				return fmt.Sprintf("invalid regex: %v %v", err, tiv.Pattern)
			}

			if !regex.MatchString(input) {
				return fmt.Sprintf("should match pattern `%s`", tiv.Pattern)
			}
		}

		if tiv.MinLen != nil {
			if len(input) < *tiv.MinLen {
				return fmt.Sprintf("need to be at least %d characters", *tiv.MinLen)
			}
		}

		if tiv.MaxLen != nil {
			if len(input) > *tiv.MaxLen {
				return fmt.Sprintf("cannot to be longer than %d characters", *tiv.MaxLen)
			}
		}
	}

	return ""
}

type TemplateInput struct {
	Name        string                  `yaml:"name"`
	Default     string                  `yaml:"default"`
	Required    bool                    `yaml:"required"`
	Description string                  `yaml:"description"`
	Pattern     string                  `yaml:"pattern"`
	Type        string                  `yaml:"type"`
	Validate    TemplateInputValidation `yaml:"validate,omitempty"`
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
		title := input.Name
		if input.Required {
			title += "*"
		}

		description := input.Description
		if input.Default != "" {
			description += " (default: " + input.Default + ")"
		}

		field := huh.NewInput().
			Title(title).
			Description(description).
			PlaceholderFunc(func() string {
				if input.Default != "" {
					return "default: " + input.Default + ""
				}
				return ""
			}, nil).
			Validate(func(s string) error {
				if input.Required {
					if input.Default != "" {
						userInputs[input.Name] = input.Default
					} else {
						return errors.New(input.Name + " is required")
					}
				}

				// Only validate if value is provided
				if s != "" {
					validationMessage := input.Validate.Validate(input.Type, s)
					if validationMessage != "" {
						return errors.New(input.Name + " " + validationMessage)
					}
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

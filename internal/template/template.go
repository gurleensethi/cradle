package template

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
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// templateFS holds the embedded template files.
// It must be set via SetTemplateFS before using template functions.
var templateFS *embed.FS

// SetTemplateFS sets the embed.FS containing template files.
// This must be called before using GetTemplate or ReadUserInputs.
func SetTemplateFS(fs embed.FS) {
	templateFS = &fs
}

var ErrNotExists = errors.New("template does not exist")

// TemplateInputValidation defines validation rules for template inputs.
type TemplateInputValidation struct {
	Pattern string   `yaml:"pattern"`
	Message string   `yaml:"message"`
	MinLen  *int     `yaml:"min_len,omitempty"`
	MaxLen  *int     `yaml:"max_len,omitempty"`
	Min     *float64 `yaml:"min,omitempty"`
	Max     *float64 `yaml:"max,omitempty"`
}

// Validate validates an input value against the validation rules.
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

	if inputType == "int" {
		numValue, err := strconv.Atoi(input)
		if err != nil {
			return "must be a valid integer"
		}

		if tiv.Min != nil && numValue < int(*tiv.Min) {
			return fmt.Sprintf("number must be greater than %v", *tiv.Min)
		}

		if tiv.Max != nil && numValue > int(*tiv.Max) {
			return fmt.Sprintf("number must be less than %v", *tiv.Max)
		}
	}

	if inputType == "string" {
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

// TemplateInput defines a user input field for a template.
type TemplateInput struct {
	Name        string                  `yaml:"name"`
	Default     string                  `yaml:"default"`
	Required    bool                    `yaml:"required"`
	Description string                  `yaml:"description"`
	Pattern     string                  `yaml:"pattern"`
	Type        string                  `yaml:"type"`
	Validate    TemplateInputValidation `yaml:"validate,omitempty"`
}

// TemplateFile represents a file in a template.
type TemplateFile string

// TemplateData represents a project template.
type TemplateData struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Inputs      []TemplateInput         `yaml:"inputs"`
	Files       map[string]TemplateFile `yaml:"files"`
}

// Validate validates the template data.
func (td *TemplateData) Validate() error {
	if len(td.Name) == 0 {
		return errors.New("template name cannot be empty")
	}

	if len(td.Files) == 0 {
		return errors.New("template must define at least one file")
	}

	return nil
}

// GetTemplate loads a template by name from the embedded filesystem.
func GetTemplate(templateName string) (*TemplateData, error) {
	if templateFS == nil {
		panic("template.SetTemplateFS must be called before GetTemplate")
	}

	data, err := (*templateFS).ReadFile("templates/" + templateName + ".yaml")
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

// ReadUserInputs prompts the user for template input values.
func ReadUserInputs(td *TemplateData) (map[string]string, error) {
	userInputs := make(map[string]string)
	for _, input := range td.Inputs {
		userInputs[input.Name] = input.Default
	}

	if len(td.Inputs) == 0 {
		return userInputs, nil
	}

	fmt.Println(
		lipgloss.NewStyle().
			Margin(1).
			Underline(true).
			Render(td.Name + " - " + td.Description),
	)

	var fields []huh.Field
	for _, input := range td.Inputs {
		title := input.Name
		if input.Required {
			title += "*"
		}

		description := input.Description
		if input.Default != "" {
			description += "\n(default: " + input.Default + ")"
		}

		field := huh.NewInput().
			Title(title).
			Description(description).
			PlaceholderFunc(func() string {
				return input.Default
			}, nil).
			Validate(func(s string) error {
				if s == "" {
					if input.Default != "" {
						userInputs[input.Name] = input.Default
						return nil
					} else if input.Required {
						return errors.New(input.Name + " is required")
					}
				} else {
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

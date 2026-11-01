package template

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"gopkg.in/yaml.v3"
)

// templateFS is the embedded template filesystem.
var templateFS *embed.FS

// SetTemplateFS sets the embed.FS containing template files.
// This must be called before using GetTemplate or ReadUserInputs.
func SetTemplateFS(fs embed.FS) {
	templateFS = &fs
}

// ErrNotExists is returned when a requested template is not found.
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

// Validate checks the given input against the configured rules for inputType.
// It returns a human-readable error message and false when validation fails,
// or an empty string and true when the input is valid.
func (tiv *TemplateInputValidation) Validate(inputType, input string) (string, bool) {
	switch inputType {
	case "float":
		numValue, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return "must be a valid number (float)", false
		}
		if tiv.Min != nil && numValue < *tiv.Min {
			return fmt.Sprintf("number must be greater than %v", *tiv.Min), false
		}
		if tiv.Max != nil && numValue > *tiv.Max {
			return fmt.Sprintf("number must be less than %v", *tiv.Max), false
		}
	case "int":
		numValue, err := strconv.Atoi(input)
		if err != nil {
			return "must be a valid number (integer)", false
		}
		if tiv.Min != nil && numValue < int(*tiv.Min) {
			return fmt.Sprintf("number must be greater than %v", *tiv.Min), false
		}
		if tiv.Max != nil && numValue > int(*tiv.Max) {
			return fmt.Sprintf("number must be less than %v", *tiv.Max), false
		}
	case "string":
		if tiv.Pattern != "" {
			regex, err := regexp.Compile(tiv.Pattern)
			if err != nil {
				return fmt.Sprintf("invalid regex: %v %v", err, tiv.Pattern), false
			}
			if !regex.MatchString(input) {
				return fmt.Sprintf("should match pattern `%s`", tiv.Pattern), false
			}
		}
		if tiv.MinLen != nil && len(input) < *tiv.MinLen {
			return fmt.Sprintf("need to be at least %d characters", *tiv.MinLen), false
		}
		if tiv.MaxLen != nil && len(input) > *tiv.MaxLen {
			return fmt.Sprintf("cannot be longer than %d characters", *tiv.MaxLen), false
		}
	}

	return "", true
}

// TemplateInput defines a user input field for a template.
type TemplateInput struct {
	Name        string                  `yaml:"name"`
	Default     string                  `yaml:"default"`
	Required    bool                    `yaml:"required"`
	Description string                  `yaml:"description"`
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
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, err
	}

	return &template, nil
}

// castInput converts a string value to the type declared by the template input.
func castInput(inputType, value string) any {
	switch inputType {
	case "int":
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	case "float":
		if n, err := strconv.ParseFloat(value, 64); err == nil {
			return n
		}
	case "bool":
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}

	return value
}

// ReadUserInputs collects validated input values from the user.
// Values are cast to their declared type (int, float, or string) before being returned.
func ReadUserInputs(td *TemplateData) (map[string]any, error) {
	userInputs := make(map[string]any)
	for _, input := range td.Inputs {
		userInputs[input.Name] = nil
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

	// fields rendered in the form to collect variables
	// values from the user.
	var fields []huh.Field

	stringPtrs := make(map[string]*string)
	boolPtrs := make(map[string]*bool)

	for _, input := range td.Inputs {
		title := input.Name
		if input.Required {
			title += "*"
		}

		description := input.Description
		if input.Default != "" {
			description += "\n(default: " + input.Default + ")"
		}

		switch input.Type {
		case "bool":
			b := false
			if v, ok := castInput(input.Type, input.Default).(bool); ok {
				b = v
			}
			ptr := &b
			boolPtrs[input.Name] = ptr

			field := huh.NewConfirm().
				Title(title).
				Description(description).
				Value(ptr).
				Key(input.Name)

			fields = append(fields, field)
		default:
			value := ""
			ptr := &value
			stringPtrs[input.Name] = ptr

			field := huh.NewInput().
				Title(title).
				Description(description).
				PlaceholderFunc(func() string {
					return input.Default
				}, nil).
				Value(ptr).
				Validate(func(s string) error {
					if s == "" {
						if input.Default != "" {
							return nil
						}
						if input.Required {
							return errors.New(input.Name + " is required")
						}
						return nil
					}

					validationMessage, isValid := input.Validate.Validate(input.Type, s)
					if !isValid {
						if input.Validate.Message != "" {
							validationMessage = input.Validate.Message
						}
						return errors.New(input.Name + " " + validationMessage)
					}
					return nil
				}).
				Key(input.Name)

			fields = append(fields, field)
		}
	}

	form := huh.NewForm(
		huh.NewGroup(fields...),
	).WithProgramOptions(tea.WithOutput(os.Stdout))

	if err := form.Run(); err != nil {
		return nil, err
	}

	// Sync bound values back into the result map, applying defaults for empty inputs.
	for _, input := range td.Inputs {
		switch input.Type {
		case "bool":
			if ptr, ok := boolPtrs[input.Name]; ok && ptr != nil {
				userInputs[input.Name] = *ptr
			}
		default:
			if ptr, ok := stringPtrs[input.Name]; ok && ptr != nil {
				value := *ptr
				if value == "" && input.Default != "" {
					value = input.Default
				}
				userInputs[input.Name] = castInput(input.Type, value)
			}
		}
	}

	return userInputs, nil
}

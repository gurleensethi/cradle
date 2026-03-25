package main

import (
	"github.com/gurleensethi/cradle/internal/template"
)

// TemplateInputValidation defines validation rules for template inputs.
// Deprecated: Use template.TemplateInputValidation instead.
type TemplateInputValidation = template.TemplateInputValidation

// TemplateInput defines a user input field for a template.
// Deprecated: Use template.TemplateInput instead.
type TemplateInput = template.TemplateInput

// TemplateFile represents a file in a template.
// Deprecated: Use template.TemplateFile instead.
type TemplateFile = template.TemplateFile

// TemplateData represents a project template.
// Deprecated: Use template.TemplateData instead.
type TemplateData = template.TemplateData

// GetTemplate loads a template by name from the embedded filesystem.
// Deprecated: Use template.GetTemplate() instead.
func GetTemplate(templateName string) (*template.TemplateData, error) {
	return template.GetTemplate(templateName)
}

// ReadUserInputs prompts the user for template input values.
// Deprecated: Use template.ReadUserInputs() instead.
func ReadUserInputs(td *template.TemplateData) (map[string]string, error) {
	return template.ReadUserInputs(td)
}

// ErrNotExists indicates that a template does not exist.
// Deprecated: Use template.ErrNotExists instead.
var ErrNotExists = template.ErrNotExists

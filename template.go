package main

import (
	"github.com/gurleensethi/cradle/internal/template"
)

// Deprecated: Use template.TemplateInputValidation instead.
type TemplateInputValidation = template.TemplateInputValidation

// Deprecated: Use template.TemplateInput instead.
type TemplateInput = template.TemplateInput

// Deprecated: Use template.TemplateFile instead.
type TemplateFile = template.TemplateFile

// Deprecated: Use template.TemplateData instead.
type TemplateData = template.TemplateData

// Deprecated: Use template.GetTemplate instead.
func GetTemplate(templateName string) (*template.TemplateData, error) {
	return template.GetTemplate(templateName)
}

// Deprecated: Use template.ReadUserInputs instead.
func ReadUserInputs(td *template.TemplateData) (map[string]string, error) {
	return template.ReadUserInputs(td)
}

// Deprecated: Use template.ErrNotExists instead.
var ErrNotExists = template.ErrNotExists

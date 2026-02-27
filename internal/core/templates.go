package core

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

// renderTemplate executes a named template with the given data.
func renderTemplate(name string, data any) (string, error) {
	raw, err := templateFS.ReadFile(name)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// readTemplate returns the raw content of an embedded template file.
func readTemplate(name string) (string, error) {
	raw, err := templateFS.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

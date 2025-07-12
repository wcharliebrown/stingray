package tests

import (
	"strings"
	"testing"
	"stingray/templates"
)

func TestLoadTemplate(t *testing.T) {
	// Test loading a template that should exist
	content, err := templates.LoadTemplate("login_form")
	if err != nil {
		t.Fatalf("Failed to load login_form template: %v", err)
	}
	if content == "" {
		t.Error("Template content is empty")
	}
}

func TestProcessEmbeddedTemplates(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No embedded templates",
			input:    "<div>Hello World</div>",
			expected: "<div>Hello World</div>",
		},
		{
			name:     "With embedded template",
			input:    "<div>{{template_login_form}}</div>",
			expected: "<div>", // Should contain the login form content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := templates.ProcessEmbeddedTemplates(tt.input)
			if err != nil {
				t.Fatalf("Error processing templates: %v", err)
			}
			
			if tt.name == "No embedded templates" {
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			} else {
				// For embedded templates, just check that something changed
				if result == tt.input {
					t.Error("Template was not processed")
				}
			}
		})
	}
}

func TestProcessEmbeddedTemplatesRecursive(t *testing.T) {
	content := "<div>{{template_login_form}}<p>{{template_login_form}}</p></div>"
	processed, err := templates.ProcessEmbeddedTemplates(content)
	if err != nil {
		t.Fatalf("Error processing recursive templates: %v", err)
	}
	
	// Should not contain the original template references
	if strings.Contains(processed, "{{template_login_form}}") {
		t.Error("Template references were not fully processed")
	}
} 
package templates

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"stingray/models"
)

// LoadTemplate reads a template from file
func LoadTemplate(name string) (string, error) {
	// Try current directory first
	paths := []string{"templates/" + name, "../templates/" + name}
	var lastErr error
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content), nil
		}
		lastErr = err
	}
	return "", fmt.Errorf("template %s not found: %v", name, lastErr)
}

// ProcessEmbeddedTemplates processes embedded template references in content
func ProcessEmbeddedTemplates(content string) (string, error) {
	// Find all template references like {{template_name}}
	processed := content

	for {
		// Find the next template reference
		if !strings.Contains(processed, "{{template_") {
			break
		}

		start := strings.Index(processed, "{{template_")
		if start == -1 {
			break
		}

		// Find the closing }} for this specific template reference
		// Start searching from the position after the opening {{
		searchStart := start + 2
		end := strings.Index(processed[searchStart:], "}}")
		if end == -1 {
			// No closing braces found, break to avoid infinite loop
			log.Printf("Warning: Template reference without closing braces found at position %d", start)
			break
		}
		end = searchStart + end + 2 // Adjust for the slice and add 2 for the }}

		// Extract the template reference
		templateRef := processed[start+2 : end-2] // Remove {{ and }}
		templateName := strings.TrimPrefix(templateRef, "template_")

		templateContent, err := LoadTemplate(templateName)
		if err != nil {
			log.Printf("Warning: Template %s not found, removing reference", templateName)
			processed = processed[:start] + processed[end:]
		} else {
			processed = processed[:start] + templateContent + processed[end:]
		}
	}

	return processed, nil
}

// RenderPage renders a page using its template
func RenderPage(page *models.Page) (string, error) {
	// Load the main template
	templateContent, err := LoadTemplate(page.Template)
	if err != nil {
		return "", err
	}

	// Process embedded templates in the template itself
	templateContent, err = ProcessEmbeddedTemplates(templateContent)
	if err != nil {
		return "", err
	}

	// Process embedded templates in content
	processedContent, err := ProcessEmbeddedTemplates(page.MainContent)
	if err != nil {
		return "", err
	}

	// Create template data
	data := map[string]interface{}{
		"Title":          page.Title,
		"MetaDescription": page.MetaDescription,
		"Header":         template.HTML(page.Header),
		"Navigation":     template.HTML(page.Navigation),
		"MainContent":    template.HTML(processedContent),
		"Sidebar":        template.HTML(page.Sidebar),
		"Footer":         template.HTML(page.Footer),
		"CSSClass":       page.CSSClass,
		"Scripts":        template.HTML(page.Scripts),
	}

	// Parse and execute template
	tmpl, err := template.New("page").Parse(templateContent)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// RenderMetadataPage renders a metadata page using the metadata template
func RenderMetadataPage(data map[string]interface{}) (string, error) {
	// Load the metadata template
	templateContent, err := LoadTemplate("metadata")
	if err != nil {
		return "", err
	}

	// Process embedded templates in the template itself
	templateContent, err = ProcessEmbeddedTemplates(templateContent)
	if err != nil {
		return "", err
	}

	// Parse and execute template
	tmpl, err := template.New("metadata").Parse(templateContent)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		return "", err
	}

	return result.String(), nil
} 
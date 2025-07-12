package tests

import (
	"testing"
	"stingray/config"
	"stingray/database"
	"stingray/templates"
)

func TestDatabaseConnection(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
}

func TestGetPage(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	page, err := db.GetPage("home")
	if err != nil {
		t.Fatalf("Failed to get home page: %v", err)
	}
	if page.Slug != "home" {
		t.Errorf("Expected slug 'home', got '%s'", page.Slug)
	}
}

func TestProcessEmbeddedTemplatesInDatabase(t *testing.T) {
	content := "<div>{{template_login_form}}</div>"
	processed, err := templates.ProcessEmbeddedTemplates(content)
	if err != nil {
		t.Fatalf("Error processing embedded templates: %v", err)
	}
	if processed == content {
		t.Errorf("Template was not processed: %s", processed)
	}
} 
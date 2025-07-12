package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Page struct {
	ID             int
	Slug           string
	Title          string
	MetaDescription string
	Header         string
	Navigation     string
	MainContent    string
	Sidebar        string
	Footer         string
	CSSClass       string
	Scripts        string
	Template       string
}

type Database struct {
	db *sql.DB
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.initDatabase(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) initDatabase() error {
	// Create database if it doesn't exist
	createDBQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", "stingray")
	_, err := d.db.Exec(createDBQuery)
	if err != nil {
		return err
	}

	// Create pages table
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS pages (
		id INT AUTO_INCREMENT PRIMARY KEY,
		slug VARCHAR(255) UNIQUE NOT NULL,
		title VARCHAR(255) NOT NULL,
		meta_description TEXT,
		header TEXT,
		navigation TEXT,
		main_content TEXT,
		sidebar TEXT,
		footer TEXT,
		css_class VARCHAR(255),
		scripts TEXT,
		template VARCHAR(100) DEFAULT 'default'
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.db.Exec(createTableQuery)
	if err != nil {
		return err
	}

	// Initialize with default pages
	return d.initializePages()
}

func (d *Database) initializePages() error {
	pages := []Page{
		{
			Slug:           "home",
			Title:          "Welcome to Sting Ray",
			MetaDescription: "A modern content management system built with Go",
			Header:         "Welcome to Sting Ray",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Welcome to Sting Ray</h2><p>This is a modern content management system built with Go and MySQL. Features include dynamic page serving, template system, and RESTful API endpoints.</p>`,
			Sidebar:        `<h3>Quick Links</h3><ul><li><a href="/page/about">About</a></li><li><a href="/page/demo">Demo</a></li><li><a href="/user/login">Login</a></li></ul>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "about",
			Title:          "About Sting Ray",
			MetaDescription: "Learn about the Sting Ray content management system",
			Header:         "About Sting Ray",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>About Sting Ray</h2><p>Sting Ray is a modern content management system built with Go and MySQL.</p><h3>Features:</h3><ul><li>Dynamic page serving</li><li>Template system with embedded templates</li><li>RESTful API endpoints</li><li>MySQL database backend</li><li>Responsive design</li></ul>`,
			Sidebar:        `<h3>Technology Stack</h3><ul><li>Go 1.24.4</li><li>MySQL 5.7+</li><li>HTML Templates</li><li>RESTful APIs</li></ul>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "login",
			Title:          "Login - Sting Ray",
			MetaDescription: "User login page",
			Header:         "User Login",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Login</h2><p>Please enter your credentials to access the system.</p>{{template_login_form}}`,
			Sidebar:        `<h3>Need Help?</h3><p>Contact the administrator for login credentials.</p>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "shutdown",
			Title:          "Server Shutdown - Sting Ray",
			MetaDescription: "Server shutdown confirmation",
			Header:         "Server Shutdown",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Server Shutdown</h2><p>The server is shutting down gracefully. Please wait...</p><div id="countdown">30</div><script>let count = 30; const timer = setInterval(() => { count--; document.getElementById('countdown').textContent = count; if (count <= 0) { clearInterval(timer); window.location.href = '/'; } }, 1000);</script>`,
			Sidebar:        `<h3>Shutdown Progress</h3><p>Server will be unavailable during shutdown.</p>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "demo",
			Title:          "Template Demo - Sting Ray",
			MetaDescription: "Demonstration of embedded templates",
			Header:         "Template Demo",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Embedded Templates Demo</h2><p>This page demonstrates the embedded template system.</p><h3>Login Form Embedded:</h3>{{template_login_form}}<h3>Custom Content:</h3><p>This content is rendered with the modern template system.</p>`,
			Sidebar:        `<h3>Template Features</h3><ul><li>Embedded templates</li><li>Recursive processing</li><li>Multiple template support</li><li>Responsive design</li></ul>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
	}

	for _, page := range pages {
		if err := d.createPageIfNotExists(page); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) createPageIfNotExists(page Page) error {
	// Check if page exists
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM pages WHERE slug = ?", page.Slug).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = d.db.Exec(`
			INSERT INTO pages (slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			page.Slug, page.Title, page.MetaDescription, page.Header, page.Navigation,
			page.MainContent, page.Sidebar, page.Footer, page.CSSClass, page.Scripts, page.Template)
		return err
	}

	return nil
}

func (d *Database) GetPage(slug string) (*Page, error) {
	var page Page
	err := d.db.QueryRow(`
		SELECT id, slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template
		FROM pages WHERE slug = ?`, slug).Scan(
		&page.ID, &page.Slug, &page.Title, &page.MetaDescription, &page.Header,
		&page.Navigation, &page.MainContent, &page.Sidebar, &page.Footer,
		&page.CSSClass, &page.Scripts, &page.Template)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (d *Database) GetAllPages() ([]Page, error) {
	rows, err := d.db.Query(`
		SELECT id, slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template
		FROM pages ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []Page
	for rows.Next() {
		var page Page
		err := rows.Scan(
			&page.ID, &page.Slug, &page.Title, &page.MetaDescription, &page.Header,
			&page.Navigation, &page.MainContent, &page.Sidebar, &page.Footer,
			&page.CSSClass, &page.Scripts, &page.Template)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// Template processing functions
func loadTemplate(name string) (string, error) {
	// Read template from file
	content, err := os.ReadFile("templates/" + name)
	if err != nil {
		return "", fmt.Errorf("template %s not found: %v", name, err)
	}
	return string(content), nil
}

func processEmbeddedTemplates(content string) (string, error) {
	// Find all template references like {{template_name}}
	processed := content

	for {
		// Simple template replacement
		if strings.Contains(processed, "{{template_") {
			start := strings.Index(processed, "{{template_")
			end := strings.Index(processed, "}}")
			if start == -1 || end == -1 {
				break
			}

			templateRef := processed[start+2 : end] // Remove {{ and }}
			templateName := strings.TrimPrefix(templateRef, "template_")

			templateContent, err := loadTemplate(templateName)
			if err != nil {
				log.Printf("Warning: Template %s not found, removing reference", templateName)
				processed = strings.Replace(processed, processed[start:end+2], "", 1)
			} else {
				processed = strings.Replace(processed, processed[start:end+2], templateContent, 1)
			}
		} else {
			break
		}
	}

	return processed, nil
}

func renderPage(page *Page) (string, error) {
	// Load the main template
	templateContent, err := loadTemplate(page.Template)
	if err != nil {
		return "", err
	}

	// Process embedded templates in content
	processedContent, err := processEmbeddedTemplates(page.MainContent)
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
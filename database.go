package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// Page represents a page in the database
type Page struct {
	ID             int                `json:"id"`
	Slug           string             `json:"slug"`
	Title          string             `json:"title"`
	MetaDescription string            `json:"meta_description"`
	Header         htmltemplate.HTML  `json:"header"`
	Navigation     htmltemplate.HTML  `json:"navigation"`
	MainContent    htmltemplate.HTML  `json:"main_content"`
	Sidebar        htmltemplate.HTML  `json:"sidebar"`
	Footer         htmltemplate.HTML  `json:"footer"`
	CSSClass       string             `json:"css_class"`
	Scripts        htmltemplate.HTML  `json:"scripts"`
	Template       string             `json:"template"`
}

// InitDatabase initializes the MySQL database
func InitDatabase() error {
	var err error
	
	// Get database configuration
	config := GetDatabaseConfig()
	
	// Open database connection
	db, err = sql.Open("mysql", config.GetDSN())
	if err != nil {
		return err
	}
	
	// Test the connection
	if err = db.Ping(); err != nil {
		return err
	}
	
	// Check if the database and table exist, create if they don't
	if err = ensureDatabaseExists(); err != nil {
		return err
	}
	
	return nil
}

// ensureDatabaseExists creates the database and table if they don't exist
func ensureDatabaseExists() error {
	config := GetDatabaseConfig()
	
	// First, connect without specifying a database to create it if needed
	tempDB, err := sql.Open("mysql", config.GetDSNWithoutDB())
	if err != nil {
		return err
	}
	defer tempDB.Close()
	
	// Create database if it doesn't exist
	_, err = tempDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", config.Database))
	if err != nil {
		return err
	}
	
	// Now connect to the specific database
	db, err = sql.Open("mysql", config.GetDSN())
	if err != nil {
		return err
	}
	
	// Test the connection
	if err = db.Ping(); err != nil {
		return err
	}
	
	// Check if table exists
	var tableExists int
	err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = 'pages'", config.Database)).Scan(&tableExists)
	if err != nil {
		return err
	}
	
	// Create table and insert initial data only if it doesn't exist
	if tableExists == 0 {
		if err = createPageTable(); err != nil {
			return err
		}
		if err = insertInitialData(); err != nil {
			return err
		}
	}
	
	return nil
}

// createPageTable creates the pages table
func createPageTable() error {
	query := `
	CREATE TABLE pages (
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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
	
	_, err := db.Exec(query)
	return err
}

// insertInitialData inserts the initial page data
func insertInitialData() error {
	pages := []Page{
		{
			Slug:           "home",
			Title:          "Welcome to Sting Ray",
			MetaDescription: "A modern web application built with Go",
			Header:         htmltemplate.HTML("<h1>Welcome to Sting Ray</h1>"),
			Navigation:     htmltemplate.HTML(`<ul><li><a href="/">Home</a></li><li><a href="/page/about">About</a></li><li><a href="/user/login">Login</a></li></ul>`),
			MainContent:    htmltemplate.HTML("<p>This is the main content area of the home page. Welcome to our application!</p>"),
			Sidebar:        htmltemplate.HTML("<div class='sidebar'><h3>Quick Links</h3><ul><li><a href='/page/about'>About</a></li><li><a href='/user/login'>Login</a></li></ul></div>"),
			Footer:         htmltemplate.HTML("<footer>&copy; 2024 Sting Ray. All rights reserved.</footer>"),
			CSSClass:       "home-page",
			Scripts:        htmltemplate.HTML("<script>console.log('Home page loaded');</script>"),
			Template:       "modern",
		},
		{
			Slug:           "about",
			Title:          "About Sting Ray",
			MetaDescription: "Learn more about the Sting Ray application",
			Header:         htmltemplate.HTML("<h1>About Sting Ray</h1>"),
			Navigation:     htmltemplate.HTML(`<ul><li><a href="/">Home</a></li><li><a href="/page/about">About</a></li><li><a href="/user/login">Login</a></li></ul>`),
			MainContent:    htmltemplate.HTML("<p>Sting Ray is a modern web application built with Go. It provides a simple and efficient way to serve web content.</p><p>Features include:</p><ul><li>Fast Go backend</li><li>JSON API endpoints</li><li>Static page serving</li><li>User authentication</li></ul>"),
			Sidebar:        htmltemplate.HTML("<div class='sidebar'><h3>Contact</h3><p>Get in touch with us for more information.</p></div>"),
			Footer:         htmltemplate.HTML("<footer>&copy; 2024 Sting Ray. All rights reserved.</footer>"),
			CSSClass:       "about-page",
			Scripts:        htmltemplate.HTML("<script>console.log('About page loaded');</script>"),
			Template:       "modern",
		},
		{
			Slug:           "login",
			Title:          "User Login",
			MetaDescription: "Login to your Sting Ray account",
			Header:         htmltemplate.HTML("<h1>User Login</h1>"),
			Navigation:     htmltemplate.HTML(`<ul><li><a href="/">Home</a></li><li><a href="/page/about">About</a></li><li><a href="/user/login">Login</a></li></ul>`),
			MainContent:    htmltemplate.HTML("<p>Please enter your credentials to access your account.</p>{{template_login_form}}"),
			Sidebar:        htmltemplate.HTML("<div class='sidebar'><h3>Need Help?</h3><p>Contact support if you're having trouble logging in.</p></div>"),
			Footer:         htmltemplate.HTML("<footer>&copy; 2024 Sting Ray. All rights reserved.</footer>"),
			CSSClass:       "login-page",
			Scripts:        htmltemplate.HTML("<script>console.log('Login page loaded');</script>"),
			Template:       "modern",
		},
		{
			Slug:           "shutdown",
			Title:          "Shutdown",
			MetaDescription: "Server shutdown page",
			Header:         htmltemplate.HTML("<h1>Shutdown Initiated</h1>"),
			Navigation:     htmltemplate.HTML(`<ul><li><a href="/">Home</a></li></ul>`),
			MainContent:    htmltemplate.HTML("<p>The server is shutting down gracefully.</p>"),
			Sidebar:        htmltemplate.HTML("<div class='sidebar'><h3>Info</h3><p>This page will close shortly.</p></div>"),
			Footer:         htmltemplate.HTML("<footer>&copy; 2024 Sting Ray</footer>"),
			CSSClass:       "shutdown-page",
			Scripts:        htmltemplate.HTML("<script>console.log('Shutdown page loaded');</script>"),
			Template:       "modern",
		},
		{
			Slug:           "demo",
			Title:          "Embedded Templates Demo",
			MetaDescription: "Demonstration of embedded templates functionality",
			Header:         htmltemplate.HTML("<h1>Embedded Templates Demo</h1>"),
			Navigation:     htmltemplate.HTML(`<ul><li><a href="/">Home</a></li><li><a href="/page/demo">Demo</a></li></ul>`),
			MainContent:    htmltemplate.HTML("<p>This page demonstrates the embedded templates functionality. The header and footer are embedded templates.</p><p>You can also embed forms and other components:</p>{{template_login_form}}"),
			Sidebar:        htmltemplate.HTML("<div class='sidebar'><h3>Available Templates</h3><ul><li>modern_header</li><li>modern_footer</li><li>login_form</li></ul></div>"),
			Footer:         htmltemplate.HTML("<footer>&copy; 2024 Sting Ray</footer>"),
			CSSClass:       "demo-page",
			Scripts:        htmltemplate.HTML("<script>console.log('Demo page loaded');</script>"),
			Template:       "modern",
		},
	}
	
	for _, page := range pages {
		if err := insertPage(&page); err != nil {
			return err
		}
	}
	
	return nil
}

// insertPage inserts a single page into the database
func insertPage(page *Page) error {
	query := `
	INSERT INTO pages (slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := db.Exec(query, 
		page.Slug, 
		page.Title, 
		page.MetaDescription, 
		string(page.Header), 
		string(page.Navigation), 
		string(page.MainContent), 
		string(page.Sidebar), 
		string(page.Footer), 
		page.CSSClass, 
		string(page.Scripts), 
		page.Template)
	
	return err
}

// GetPageBySlug retrieves a page by its slug/identifier
func GetPageBySlug(slug string) (*Page, bool) {
	query := `SELECT id, slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template 
			  FROM pages WHERE slug = ?`
	
	var page Page
	err := db.QueryRow(query, slug).Scan(
		&page.ID,
		&page.Slug,
		&page.Title,
		&page.MetaDescription,
		&page.Header,
		&page.Navigation,
		&page.MainContent,
		&page.Sidebar,
		&page.Footer,
		&page.CSSClass,
		&page.Scripts,
		&page.Template,
	)
	
	if err != nil {
		return nil, false
	}
	
	return &page, true
}

// GetAllPages returns all available pages
func GetAllPages() map[string]*Page {
	query := `SELECT id, slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template 
			  FROM pages ORDER BY slug`
	
	rows, err := db.Query(query)
	if err != nil {
		return make(map[string]*Page)
	}
	defer rows.Close()
	
	pages := make(map[string]*Page)
	for rows.Next() {
		var page Page
		err := rows.Scan(
			&page.ID,
			&page.Slug,
			&page.Title,
			&page.MetaDescription,
			&page.Header,
			&page.Navigation,
			&page.MainContent,
			&page.Sidebar,
			&page.Footer,
			&page.CSSClass,
			&page.Scripts,
			&page.Template,
		)
		if err == nil {
			pages[page.Slug] = &page
		}
	}
	
	return pages
}

// processEmbeddedTemplates processes a template string and replaces embedded template references
// like {{template_anything}} with the actual template content
func processEmbeddedTemplates(templateContent string) (string, error) {
	// Regular expression to find embedded template references
	// Matches {{template_anything}} pattern
	re := regexp.MustCompile(`\{\{template_([^}]+)\}\}`)
	
	// Find all matches
	matches := re.FindAllStringSubmatch(templateContent, -1)
	
	// If no embedded templates found, return original content
	if len(matches) == 0 {
		return templateContent, nil
	}
	
	// Process each embedded template
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		
		embeddedTemplateName := match[1]
		fullMatch := match[0]
		
		// Load the embedded template
		embeddedContent, err := loadTemplateFromFile(embeddedTemplateName)
		if err != nil {
			// If template not found, replace with a comment indicating the missing template
			embeddedContent = fmt.Sprintf("<!-- Missing embedded template: %s -->", embeddedTemplateName)
		}
		
		// Replace the embedded template reference with its content
		templateContent = strings.ReplaceAll(templateContent, fullMatch, embeddedContent)
	}
	
	return templateContent, nil
}

// processEmbeddedTemplatesInContent processes embedded templates in page content
// This is similar to processEmbeddedTemplates but doesn't return an error
func processEmbeddedTemplatesInContent(content string) string {
	// Regular expression to find embedded template references
	// Matches {{template_anything}} pattern
	re := regexp.MustCompile(`\{\{template_([^}]+)\}\}`)
	
	// Find all matches
	matches := re.FindAllStringSubmatch(content, -1)
	
	// If no embedded templates found, return original content
	if len(matches) == 0 {
		return content
	}
	
	// Process each embedded template
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		
		embeddedTemplateName := match[1]
		fullMatch := match[0]
		
		// Load the embedded template
		embeddedContent, err := loadTemplateFromFile(embeddedTemplateName)
		if err != nil {
			// If template not found, replace with a comment indicating the missing template
			embeddedContent = fmt.Sprintf("<!-- Missing embedded template: %s -->", embeddedTemplateName)
		}
		
		// Replace the embedded template reference with its content
		content = strings.ReplaceAll(content, fullMatch, embeddedContent)
	}
	
	return content
}

// loadTemplateFromFile loads a template from a file in the templates directory
func loadTemplateFromFile(name string) (string, error) {
	filename := filepath.Join("templates", name)
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	
	templateContent := string(content)
	
	// Process embedded templates recursively
	processedContent, err := processEmbeddedTemplates(templateContent)
	if err != nil {
		return "", err
	}
	
	return processedContent, nil
}

// templateExists checks if a template file exists
func templateExists(name string) bool {
	filename := filepath.Join("templates", name)
	_, err := os.Stat(filename)
	return err == nil
}

// getAvailableTemplates returns a list of available template names
func getAvailableTemplates() []string {
	templates := []string{}
	
	// Check for common template files
	commonTemplates := []string{"default", "simple", "modern"}
	
	for _, name := range commonTemplates {
		if templateExists(name) {
			templates = append(templates, name)
		}
	}
	
	return templates
}

// getResponseFormat parses the response_format parameter from the request
// Defaults to "html" if not specified, only "html" and "json" are valid
func getResponseFormat(r *http.Request) string {
	format := r.URL.Query().Get("response_format")
	if format == "" {
		return "html"
	}
	if format == "json" {
		return "json"
	}
	return "html" // Default to html for any invalid values
}

// renderHTML renders a page as HTML using a template from the database
func renderHTML(w http.ResponseWriter, page *Page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Use the template specified in the page, or default to "default"
	templateName := page.Template
	if templateName == "" {
		templateName = "default"
	}
	
	// Load template from file
	html, err := loadTemplateFromFile(templateName)
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	
	// Process embedded templates in the page content
	processedPage := *page
	processedPage.MainContent = htmltemplate.HTML(processEmbeddedTemplatesInContent(string(page.MainContent)))
	processedPage.Header = htmltemplate.HTML(processEmbeddedTemplatesInContent(string(page.Header)))
	processedPage.Navigation = htmltemplate.HTML(processEmbeddedTemplatesInContent(string(page.Navigation)))
	processedPage.Sidebar = htmltemplate.HTML(processEmbeddedTemplatesInContent(string(page.Sidebar)))
	processedPage.Footer = htmltemplate.HTML(processEmbeddedTemplatesInContent(string(page.Footer)))
	processedPage.Scripts = htmltemplate.HTML(processEmbeddedTemplatesInContent(string(page.Scripts)))
	
	tmpl, err := htmltemplate.New("page").Parse(html)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	
	err = tmpl.Execute(w, &processedPage)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// renderHTMLWithTemplate renders any data using a template from the database
func renderHTMLWithTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Load template from file
	html, err := loadTemplateFromFile(templateName)
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	
	tmpl, err := htmltemplate.New("page").Parse(html)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// HandlePageRequest handles page requests and returns response based on response_format parameter
func HandlePageRequest(w http.ResponseWriter, r *http.Request, slug string) {
	page, exists := GetPageBySlug(slug)
	if !exists {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	responseFormat := getResponseFormat(r)
	if responseFormat == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	} else {
		renderHTML(w, page)
	}
}
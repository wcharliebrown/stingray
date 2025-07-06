package main

import (
	"encoding/json"
	htmltemplate "html/template"
	"net/http"
	"os"
	"path/filepath"
)

// Page represents a page in the database
type Page struct {
	Title          string `json:"title"`
	MetaDescription string `json:"meta_description"`
	Header         htmltemplate.HTML `json:"header"`
	Navigation     htmltemplate.HTML `json:"navigation"`
	MainContent    htmltemplate.HTML `json:"main_content"`
	Sidebar        htmltemplate.HTML `json:"sidebar"`
	Footer         htmltemplate.HTML `json:"footer"`
	CSSClass       string `json:"css_class"`
	Scripts        htmltemplate.HTML `json:"scripts"`
	Template       string `json:"template"`
}

// Template represents a template in the database
type Template struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	HTML string `json:"html"`
}

// GetPageBySlug retrieves a page by its slug/identifier
func GetPageBySlug(slug string) (*Page, bool) {
	// Static database content
	pages := map[string]*Page{
		"home": {
			Title:          "Welcome to Sting Ray",
			MetaDescription: "A modern web application built with Go",
			Header:         htmltemplate.HTML("<h1>Welcome to Sting Ray</h1>"),
			Navigation:     htmltemplate.HTML(`<nav><a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a></nav>`),
			MainContent:    htmltemplate.HTML("<p>This is the main content area of the home page. Welcome to our application!</p>"),
			Sidebar:        htmltemplate.HTML("<div class='sidebar'><h3>Quick Links</h3><ul><li><a href='/page/about'>About</a></li><li><a href='/user/login'>Login</a></li></ul></div>"),
			Footer:         htmltemplate.HTML("<footer>&copy; 2024 Sting Ray. All rights reserved.</footer>"),
			CSSClass:       "home-page",
			Scripts:        htmltemplate.HTML("<script>console.log('Home page loaded');</script>"),
			Template:       "default",
		},
		"about": {
			Title:          "About Sting Ray",
			MetaDescription: "Learn more about the Sting Ray application",
			Header:         htmltemplate.HTML("<h1>About Sting Ray</h1>"),
			Navigation:     htmltemplate.HTML(`<nav><a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a></nav>`),
			MainContent:    htmltemplate.HTML("<p>Sting Ray is a modern web application built with Go. It provides a simple and efficient way to serve web content.</p><p>Features include:</p><ul><li>Fast Go backend</li><li>JSON API endpoints</li><li>Static page serving</li><li>User authentication</li></ul>"),
			Sidebar:        htmltemplate.HTML("<div class='sidebar'><h3>Contact</h3><p>Get in touch with us for more information.</p></div>"),
			Footer:         htmltemplate.HTML("<footer>&copy; 2024 Sting Ray. All rights reserved.</footer>"),
			CSSClass:       "about-page",
			Scripts:        htmltemplate.HTML("<script>console.log('About page loaded');</script>"),
			Template:       "default",
		},
		"login": {
			Title:          "User Login",
			MetaDescription: "Login to your Sting Ray account",
			Header:         htmltemplate.HTML("<h1>User Login</h1>"),
			Navigation:     htmltemplate.HTML(`<nav><a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a></nav>`),
			MainContent:    htmltemplate.HTML("<p>Please enter your credentials to access your account.</p>"),
			Sidebar:        htmltemplate.HTML("<div class='sidebar'><h3>Need Help?</h3><p>Contact support if you're having trouble logging in.</p></div>"),
			Footer:         htmltemplate.HTML("<footer>&copy; 2024 Sting Ray. All rights reserved.</footer>"),
			CSSClass:       "login-page",
			Scripts:        htmltemplate.HTML("<script>console.log('Login page loaded');</script>"),
			Template:       "default",
		},
	}

	page, exists := pages[slug]
	return page, exists
}

// GetAllPages returns all available pages
func GetAllPages() map[string]*Page {
	home, _ := GetPageBySlug("home")
	about, _ := GetPageBySlug("about")
	login, _ := GetPageBySlug("login")
	
	return map[string]*Page{
		"home":  home,
		"about": about,
		"login": login,
	}
}

// loadTemplateFromFile loads a template from a file in the templates directory
func loadTemplateFromFile(name string) (string, error) {
	filename := filepath.Join("templates", name)
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ReloadTemplates reloads all templates from files
// This is useful during development for hot reloading
func ReloadTemplates() error {
	// This function can be called to refresh templates from disk
	// In a production environment, you might want to add file watching
	// or call this function periodically
	return nil
}

// GetTemplateByID retrieves a template by its ID
func GetTemplateByID(id int) (*Template, bool) {
	// Map IDs to template names
	idToName := map[int]string{
		1: "default",
		2: "simple",
	}
	
	name, exists := idToName[id]
	if !exists {
		return nil, false
	}
	
	html, err := loadTemplateFromFile(name)
	if err != nil {
		return nil, false
	}
	
	return &Template{
		ID:   id,
		Name: name,
		HTML: html,
	}, true
}

// GetTemplateByName retrieves a template by its name
func GetTemplateByName(name string) (*Template, bool) {
	html, err := loadTemplateFromFile(name)
	if err != nil {
		return nil, false
	}
	
	// Map names to IDs
	nameToID := map[string]int{
		"default": 1,
		"simple":  2,
	}
	
	id, exists := nameToID[name]
	if !exists {
		return nil, false
	}
	
	return &Template{
		ID:   id,
		Name: name,
		HTML: html,
	}, true
}

// GetAllTemplates returns all available templates
func GetAllTemplates() map[int]*Template {
	templates := make(map[int]*Template)
	
	// Try to load each template
	for id, name := range map[int]string{
		1: "default",
		2: "simple",
	} {
		if template, exists := GetTemplateByName(name); exists {
			templates[id] = template
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
	
	// Get the default template from the database
	template, exists := GetTemplateByName("default")
	if !exists {
		http.Error(w, "Default template not found", http.StatusInternalServerError)
		return
	}
	
	tmpl, err := htmltemplate.New("page").Parse(template.HTML)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	
	err = tmpl.Execute(w, page)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// renderHTMLWithTemplate renders any data using a template from the database
func renderHTMLWithTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Get the template from the database
	template, exists := GetTemplateByName(templateName)
	if !exists {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	
	tmpl, err := htmltemplate.New("page").Parse(template.HTML)
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
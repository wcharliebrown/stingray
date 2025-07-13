package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"stingray/database"
	"stingray/templates"
)

// PageHandler handles page-related requests
type PageHandler struct {
	db *database.Database
	sm *SessionMiddleware
}

// NewPageHandler creates a new page handler
func NewPageHandler(db *database.Database) *PageHandler {
	return &PageHandler{
		db: db,
		sm: NewSessionMiddleware(db),
	}
}

// HandleHome handles the home page request
func (h *PageHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		RenderMessage(w, "404 Not Found", "Page Not Found", "error", "The page you requested does not exist.", "/", "Go Home", http.StatusNotFound)
		return
	}

	page, err := h.db.GetPage("home")
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	// Modify navigation based on authentication status
	if h.sm.IsAuthenticated(r) {
		username := r.Header.Get("X-Username")
		if username == "" {
			username = "User"
		}
		page.Navigation = `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/profile">Profile</a> | <a href="/user/logout">Logout</a>`
		page.Sidebar = `<h3>Welcome, ` + username + `!</h3><ul><li><a href="/page/about">About</a></li><li><a href="/page/demo">Demo</a></li><li><a href="/user/profile">Profile</a></li><li><a href="/user/logout">Logout</a></li></ul>`
	} else {
		page.Navigation = `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`
		page.Sidebar = `<h3>Quick Links</h3><ul><li><a href="/page/about">About</a></li><li><a href="/page/demo">Demo</a></li><li><a href="/user/login">Login</a></li></ul>`
	}

	html, err := templates.RenderPage(page)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// HandlePage handles individual page requests
func (h *PageHandler) HandlePage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/page/")
	if path == "" {
		RenderMessage(w, "404 Not Found", "Page Not Found", "error", "The page you requested does not exist.", "/", "Go Home", http.StatusNotFound)
		return
	}

	page, err := h.db.GetPage(path)
	if err != nil {
		database.LogSQLError(err)
		RenderMessage(w, "404 Not Found", "Page Not Found", "error", "The page you requested does not exist.", "/", "Go Home", http.StatusNotFound)
		return
	}

	// Modify navigation based on authentication status for login page
	if path == "login" {
		if h.sm.IsAuthenticated(r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	} else {
		// Modify navigation for other pages based on authentication status
		if h.sm.IsAuthenticated(r) {
			username := r.Header.Get("X-Username")
			if username == "" {
				username = "User"
			}
			page.Navigation = `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/profile">Profile</a> | <a href="/user/logout">Logout</a>`
		} else {
			page.Navigation = `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`
		}
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
		return
	}

	html, err := templates.RenderPage(page)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// HandlePages handles the pages listing request
func (h *PageHandler) HandlePages(w http.ResponseWriter, r *http.Request) {
	pages, err := h.db.GetAllPages()
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Error fetching pages", http.StatusInternalServerError)
		return
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pages)
		return
	}

	// HTML response
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>All Pages - Sting Ray</title>
		<style>
			body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; margin: 0; padding: 2rem; }
			.container { max-width: 800px; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
			h1 { color: #2c3e50; margin-bottom: 1rem; }
			.page-list { list-style: none; padding: 0; }
			.page-item { padding: 1rem; border-bottom: 1px solid #e9ecef; }
			.page-item:last-child { border-bottom: none; }
			.page-item a { color: #667eea; text-decoration: none; font-weight: 500; }
			.page-item a:hover { text-decoration: underline; }
			.page-meta { color: #6c757d; font-size: 0.9rem; margin-top: 0.5rem; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>All Pages</h1>
			<ul class="page-list">
				{{range .}}
				<li class="page-item">
					<a href="/page/{{.Slug}}">{{.Title}}</a>
					<div class="page-meta">Slug: {{.Slug}} | Template: {{.Template}}</div>
				</li>
				{{end}}
			</ul>
		</div>
	</body>
	</html>`

	t, err := template.New("pages").Parse(tmpl)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, pages)
}

// HandleTemplates handles the templates listing request
func (h *PageHandler) HandleTemplates(w http.ResponseWriter, r *http.Request) {
	templates := []string{"default", "simple", "modern", "modern_header", "modern_footer", "login_form"}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(templates)
		return
	}

	// HTML response
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Available Templates - Sting Ray</title>
		<style>
			body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; margin: 0; padding: 2rem; }
			.container { max-width: 800px; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
			h1 { color: #2c3e50; margin-bottom: 1rem; }
			.template-list { list-style: none; padding: 0; }
			.template-item { padding: 1rem; border-bottom: 1px solid #e9ecef; }
			.template-item:last-child { border-bottom: none; }
			.template-item a { color: #667eea; text-decoration: none; font-weight: 500; }
			.template-item a:hover { text-decoration: underline; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Available Templates</h1>
			<ul class="template-list">
				{{range .}}
				<li class="template-item">
					<a href="/template/{{.}}?response_format=json">{{.}}</a>
				</li>
				{{end}}
			</ul>
		</div>
	</body>
	</html>`

	t, err := template.New("templates").Parse(tmpl)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, templates)
}

// HandleTemplate handles individual template requests
func (h *PageHandler) HandleTemplate(w http.ResponseWriter, r *http.Request) {
	templateName := strings.TrimPrefix(r.URL.Path, "/template/")
	if templateName == "" {
		RenderMessage(w, "404 Not Found", "Template Not Found", "error", "The requested template does not exist.", "/templates", "View Templates", http.StatusNotFound)
		return
	}

	templateContent, err := templates.LoadTemplate(templateName)
	if err != nil {
		database.LogSQLError(err)
		RenderMessage(w, "404 Not Found", "Template Not Found", "error", "The requested template does not exist.", "/templates", "View Templates", http.StatusNotFound)
		return
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		response := map[string]string{
			"name":    templateName,
			"content": templateContent,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// HTML response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(templateContent))
} 
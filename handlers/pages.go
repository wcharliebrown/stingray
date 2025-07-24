package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"stingray/database"
	"stingray/templates"
	"stingray/config"
	"io/ioutil"
)

// PageHandler handles page-related requests
type PageHandler struct {
	db   *database.Database
	sm   *SessionMiddleware
	cfg  *config.Config // Add config reference
}

// NewPageHandler creates a new page handler
func NewPageHandler(db *database.Database, cfg *config.Config) *PageHandler {
	return &PageHandler{
		db:  db,
		sm:  NewSessionMiddleware(db),
		cfg: cfg,
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
		
		// Check if user is admin or engineer
		session, err := h.sm.GetSessionFromRequest(r)
		var isAdmin, isEngineer bool
		if err == nil {
			isAdmin, _ = h.db.IsUserInGroup(session.UserID, "admin")
			isEngineer, _ = h.db.IsUserInGroup(session.UserID, "engineer")
		}
		
		// Build navigation with admin/engineer-specific links
		nav := `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/profile">Profile</a> | <a href="/user/logout">Logout</a> | <a href="/config">Config</a>`
		sidebar := `<h3>Welcome, ` + username + `!</h3><ul><li><a href="/page/about">About</a></li><li><a href="/page/demo">Demo</a></li><li><a href="/user/profile">Profile</a></li><li><a href="/user/logout">Logout</a></li>`
		
		if isAdmin || isEngineer {
			nav += ` | <a href="/metadata/tables">Database Tables</a>`
			sidebar += `<li><a href="/metadata/tables">Database Tables</a></li>`
		}
		
		sidebar += `</ul>`
		page.Navigation = nav
		page.Sidebar = sidebar
	} else {
		page.Navigation = `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a> | <a href="/config">Config</a>`
		page.Sidebar = `<h3>Quick Links</h3><ul><li><a href="/page/about">About</a></li><li><a href="/page/demo">Demo</a></li><li><a href="/user/login">Login</a></li><li><a href="/config">Config</a></li></ul>`
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

	// Get user ID for permission checking
	var userID int
	if h.sm.IsAuthenticated(r) {
		session, err := h.sm.GetSessionFromRequest(r)
		if err == nil {
			userID = session.UserID
		}
	}

	// Get page with permission check
	page, err := h.db.GetPageWithPermissionCheck(path, userID)
	if err != nil {
		database.LogSQLError(err)
		if err.Error() == "access denied" {
			RenderMessage(w, "403 Forbidden", "Access Denied", "error", "You do not have permission to access this page.", "/", "Go Home", http.StatusForbidden)
		} else {
			RenderMessage(w, "404 Not Found", "Page Not Found", "error", "The page you requested does not exist.", "/", "Go Home", http.StatusNotFound)
		}
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
			
			// Check if user is admin or engineer
			session, err := h.sm.GetSessionFromRequest(r)
			var isAdmin, isEngineer bool
			if err == nil {
				isAdmin, _ = h.db.IsUserInGroup(session.UserID, "admin")
				isEngineer, _ = h.db.IsUserInGroup(session.UserID, "engineer")
			}
			
			// Build navigation with admin/engineer-specific links
			nav := `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/profile">Profile</a> | <a href="/user/logout">Logout</a>`
			
			if isAdmin || isEngineer {
				nav += ` | <a href="/metadata/tables">Database Tables</a>`
			}
			
			page.Navigation = nav
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
			.container { max-width: 100%; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
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
			.container { max-width: 100%; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
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

// Handler to display and edit config settings and .env file
func (h *PageHandler) HandleConfigPage(w http.ResponseWriter, r *http.Request) {
	// Only allow admin or engineer
	if !h.sm.IsAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}
	session, err := h.sm.GetSessionFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}
	isAdmin, _ := h.db.IsUserInGroup(session.UserID, "admin")
	isEngineer, _ := h.db.IsUserInGroup(session.UserID, "engineer")
	if !isAdmin && !isEngineer {
		RenderMessage(w, "Access Denied", "Access Denied", "error", "You do not have permission to access this page.", "/", "Go Home", http.StatusForbidden)
		return
	}

	var message string
	var messageType string
	var envContent string
	// Load .env file content
	if data, err := ioutil.ReadFile(".env"); err == nil {
		envContent = string(data)
	}

	// Add a shared function to reload .env and update config pointer
	ReloadEnvConfig := func(cfg *config.Config) (success bool, errMsg string) {
		newCfg := config.LoadConfig()
		if newCfg == nil {
			return false, "Failed to reload .env file."
		}
		*cfg = *newCfg
		return true, ""
	}

	if r.Method == "POST" {
		r.ParseForm()
		newEnv := r.FormValue("env_content")
		if err := ioutil.WriteFile(".env", []byte(newEnv), 0644); err != nil {
			message = "Failed to save .env file: " + err.Error()
			messageType = "error"
		} else {
			success, errMsg := ReloadEnvConfig(h.cfg)
			if success {
				message = ".env file saved and reloaded successfully."
				messageType = "success"
				envContent = newEnv
			} else {
				message = errMsg
				messageType = "error"
			}
		}
	}

	cfg := h.cfg
	html := `<html><head><title>Config Settings</title>
	<style>body{font-family:sans-serif;background:#f5f5f5;padding:2rem;}table{background:white;border-radius:8px;box-shadow:0 2px 10px rgba(0,0,0,0.1);padding:2rem;}th,td{padding:0.5rem 1rem;text-align:left;}th{background:#f0f0f0;}tr:nth-child(even){background:#fafafa;}textarea{width:100%;height:300px;font-family:monospace;font-size:1rem;margin-bottom:1rem;}button{padding:0.5rem 1.5rem;font-size:1rem;border-radius:4px;border:none;background:#667eea;color:white;cursor:pointer;}button:hover{background:#556cd6;} .msg-success{color:green;} .msg-error{color:red;}</style>
	</head><body><h1>Config Settings</h1>`
	if message != "" {
		html += `<div class='msg-` + messageType + `'>` + message + `</div><br>`
	}
	html += `<form method='POST'><h2>Edit .env File</h2><textarea name='env_content'>` + template.HTMLEscapeString(envContent) + `</textarea><br><button type='submit'>Save & Reload</button></form><br>`
	html += `<h2>Current Config</h2><table><tr><th>Key</th><th>Value</th></tr>` +
		configRow("MySQLHost", cfg.MySQLHost) +
		configRow("MySQLPort", cfg.MySQLPort) +
		configRow("MySQLUser", cfg.MySQLUser) +
		configRow("MySQLPassword", mask(cfg.MySQLPassword)) +
		configRow("MySQLDatabase", cfg.MySQLDatabase) +
		configRow("DebuggingMode", fmt.Sprintf("%v", cfg.DebuggingMode)) +
		configRow("LoggingLevel", fmt.Sprintf("%d", cfg.LoggingLevel)) +
		configRow("TestAdminUsername", cfg.TestAdminUsername) +
		configRow("TestAdminPassword", mask(cfg.TestAdminPassword)) +
		configRow("TestCustomerUsername", cfg.TestCustomerUsername) +
		configRow("TestCustomerPassword", mask(cfg.TestCustomerPassword)) +
		configRow("TestWrongPassword", mask(cfg.TestWrongPassword)) +
		configRow("SMTPHost", cfg.SMTPHost) +
		configRow("SMTPPort", cfg.SMTPPort) +
		configRow("SMTPUsername", cfg.SMTPUsername) +
		configRow("SMTPPassword", mask(cfg.SMTPPassword)) +
		configRow("FromEmail", cfg.FromEmail) +
		configRow("FromName", cfg.FromName) +
		configRow("DKIMPrivateKeyFile", cfg.DKIMPrivateKeyFile) +
		configRow("DKIMSelector", cfg.DKIMSelector) +
		configRow("DKIMDomain", cfg.DKIMDomain) +
		configRow("ServerPort", cfg.ServerPort) +
		`</table><br><a href='/'>Back to Home</a></body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func configRow(key, value string) string {
	return "<tr><td>" + key + "</td><td>" + value + "</td></tr>"
}

func mask(s string) string {
	if s == "" { return "" }
	if len(s) <= 2 { return "*" }
	return s[:1] + strings.Repeat("*", len(s)-2) + s[len(s)-1:]
} 
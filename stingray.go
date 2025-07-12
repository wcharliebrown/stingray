package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Server struct {
	db     *Database
	server *http.Server
}

func NewServer(db *Database) *Server {
	mux := http.NewServeMux()
	server := &Server{db: db}

	// Page routes
	mux.HandleFunc("/", server.handleHome)
	mux.HandleFunc("/page/", server.handlePage)
	mux.HandleFunc("/pages", server.handlePages)
	mux.HandleFunc("/templates", server.handleTemplates)
	mux.HandleFunc("/template/", server.handleTemplate)
	mux.HandleFunc("/user/login", server.handleLogin)
	mux.HandleFunc("/user/login_post", server.handleLoginPost)
	mux.HandleFunc("/shutdown", server.handleShutdown)

	server.server = &http.Server{
		Addr:    ":6273",
		Handler: mux,
	}

	return server
}

func (s *Server) Start() error {
	log.Printf("Starting Sting Ray server on port 6273...")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server gracefully...")
	return s.server.Shutdown(ctx)
}

// Handler functions
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.renderMessage(w, "404 Not Found", "Page Not Found", "error", "The page you requested does not exist.", "/", "Go Home", http.StatusNotFound)
		return
	}

	page, err := s.db.GetPage("home")
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	html, err := renderPage(page)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/page/")
	if path == "" {
		s.renderMessage(w, "404 Not Found", "Page Not Found", "error", "The page you requested does not exist.", "/", "Go Home", http.StatusNotFound)
		return
	}

	page, err := s.db.GetPage(path)
	if err != nil {
		s.renderMessage(w, "404 Not Found", "Page Not Found", "error", "The page you requested does not exist.", "/", "Go Home", http.StatusNotFound)
		return
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
		return
	}

	html, err := renderPage(page)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handlePages(w http.ResponseWriter, r *http.Request) {
	pages, err := s.db.GetAllPages()
	if err != nil {
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

func (s *Server) handleTemplates(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleTemplate(w http.ResponseWriter, r *http.Request) {
	templateName := strings.TrimPrefix(r.URL.Path, "/template/")
	if templateName == "" {
		s.renderMessage(w, "404 Not Found", "Template Not Found", "error", "The requested template does not exist.", "/templates", "View Templates", http.StatusNotFound)
		return
	}

	templateContent, err := loadTemplate(templateName)
	if err != nil {
		s.renderMessage(w, "404 Not Found", "Template Not Found", "error", "The requested template does not exist.", "/templates", "View Templates", http.StatusNotFound)
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

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	page, err := s.db.GetPage("login")
	if err != nil {
		s.renderMessage(w, "Login Page Not Found", "Login Page Not Found", "error", "The login page could not be found.", "/", "Go Home", http.StatusNotFound)
		return
	}

	html, err := renderPage(page)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.renderMessage(w, "405 Method Not Allowed", "Method Not Allowed", "error", "Only POST is allowed for this endpoint.", "/user/login", "Back to Login", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var data struct {
		Title       string
		MetaDescription string
		Header      string
		HeaderClass string
		Message     string
		ButtonURL   string
		ButtonText  string
		Footer      string
	}
	data.Footer = "© 2024 Sting Ray CMS"

	if username == "admin" && password == "password" {
		data.Title = "Login Success - Sting Ray"
		data.MetaDescription = "Login successful"
		data.Header = "Login Successful!"
		data.HeaderClass = "success"
		data.Message = "Welcome, admin!"
		data.ButtonURL = "/"
		data.ButtonText = "Go Home"
	} else {
		data.Title = "Login Failed - Sting Ray"
		data.MetaDescription = "Login failed"
		data.Header = "Login Failed"
		data.HeaderClass = "error"
		data.Message = "Invalid username or password. Try admin/password."
		data.ButtonURL = "/user/login"
		data.ButtonText = "Try Again"
	}

	tmplContent, err := loadTemplate("message")
	if err != nil {
		http.Error(w, "Message template not found", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("message").Parse(tmplContent)
	if err != nil {
		http.Error(w, "Error parsing message template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Println("Shutdown requested via POST /shutdown")
		go func() {
			time.Sleep(1 * time.Second)
			os.Exit(0)
		}()
	}

	// Show shutdown page
	page, err := s.db.GetPage("shutdown")
	if err != nil {
		http.Error(w, "Shutdown page not found", http.StatusNotFound)
		return
	}

	html, err := renderPage(page)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) renderMessage(w http.ResponseWriter, title, header, headerClass, message, buttonURL, buttonText string, status int) {
	data := struct {
		Title       string
		MetaDescription string
		Header      string
		HeaderClass string
		Message     string
		ButtonURL   string
		ButtonText  string
		Footer      string
	}{
		Title: title,
		MetaDescription: header,
		Header: header,
		HeaderClass: headerClass,
		Message: message,
		ButtonURL: buttonURL,
		ButtonText: buttonText,
		Footer: "© 2024 Sting Ray CMS",
	}
	tmplContent, err := loadTemplate("message")
	if err != nil {
		http.Error(w, "Message template not found", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("message").Parse(tmplContent)
	if err != nil {
		http.Error(w, "Error parsing message template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	tmpl.Execute(w, data)
}

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize database
	db, err := NewDatabase(config.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create server
	server := NewServer(db)

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Println("Server started on http://localhost:6273")
	log.Println("Press Ctrl+C to stop the server")

	<-done
	log.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
} 
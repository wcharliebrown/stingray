package handlers

import (
	"html/template"
	"net/http"
	"stingray/database"
	"stingray/templates"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db *database.Database
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *database.Database) *AuthHandler {
	return &AuthHandler{db: db}
}

// HandleLogin handles the login page request
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	page, err := h.db.GetPage("login")
	if err != nil {
		RenderMessage(w, "Login Page Not Found", "Login Page Not Found", "error", "The login page could not be found.", "/", "Go Home", http.StatusNotFound)
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

// HandleLoginPost handles the login form submission
func (h *AuthHandler) HandleLoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		RenderMessage(w, "405 Method Not Allowed", "Method Not Allowed", "error", "Only POST is allowed for this endpoint.", "/user/login", "Back to Login", http.StatusMethodNotAllowed)
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

	tmplContent, err := templates.LoadTemplate("message")
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
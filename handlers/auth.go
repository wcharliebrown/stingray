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
	sm *SessionMiddleware
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *database.Database) *AuthHandler {
	return &AuthHandler{
		db: db,
		sm: NewSessionMiddleware(db),
	}
}

// HandleLogin handles the login page request
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to home
	if h.sm.IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

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
		// Create session
		session, err := h.db.CreateSession("admin", username, SessionDuration)
		if err != nil {
			data.Title = "Login Error - Sting Ray"
			data.MetaDescription = "Login error"
			data.Header = "Login Error"
			data.HeaderClass = "error"
			data.Message = "Failed to create session. Please try again."
			data.ButtonURL = "/user/login"
			data.ButtonText = "Try Again"
		} else {
			// Set session cookie
			h.sm.SetSessionCookie(w, session.SessionID)
			
			data.Title = "Login Success - Sting Ray"
			data.MetaDescription = "Login successful"
			data.Header = "Login Successful!"
			data.HeaderClass = "success"
			data.Message = "Welcome, admin! You are now logged in."
			data.ButtonURL = "/"
			data.ButtonText = "Go Home"
		}
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

// HandleLogout handles user logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session from request
	if session, err := h.sm.GetSessionFromRequest(r); err == nil {
		// Invalidate session in database
		h.db.InvalidateSession(session.SessionID)
	}

	// Clear session cookie
	h.sm.ClearSessionCookie(w)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleProfile shows user profile page (requires authentication)
func (h *AuthHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	session, err := h.sm.GetSessionFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	var data struct {
		Title       string
		MetaDescription string
		Header      string
		HeaderClass string
		Message     string
		ButtonURL   string
		ButtonText  string
		Footer      string
		Username    string
		UserID      string
		LoginTime   string
	}
	
	data.Title = "User Profile - Sting Ray"
	data.MetaDescription = "User profile page"
	data.Header = "User Profile"
	data.HeaderClass = "success"
	data.Message = "Welcome to your profile page!"
	data.ButtonURL = "/"
	data.ButtonText = "Go Home"
	data.Footer = "© 2024 Sting Ray CMS"
	data.Username = session.Username
	data.UserID = session.UserID
	data.LoginTime = session.CreatedAt.Format("2006-01-02 15:04:05")

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
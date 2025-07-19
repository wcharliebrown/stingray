package handlers

import (
	"html/template"
	"net/http"
	"stingray/database"
	"stingray/logging"
	"stingray/templates"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db     *database.Database
	sm     *SessionMiddleware
	logger *logging.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *database.Database, logger *logging.Logger) *AuthHandler {
	return &AuthHandler{
		db:     db,
		sm:     NewSessionMiddleware(db),
		logger: logger,
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
		database.LogSQLError(err)
		h.logger.LogError("Login page not found: %v", err)
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
	data.Footer = "© 2025 StingRay"

	// Get remote address for logging
	remoteAddr := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		remoteAddr = forwardedFor
	}

	// Authenticate user against database
	user, err := h.db.AuthenticateUser(username, password)
	if err != nil {
		database.LogSQLError(err)
		// Log failed login attempt
		if h.logger != nil {
			h.logger.LogLogin(username, remoteAddr, false)
		}
		data.Title = "Login Failed - Sting Ray"
		data.MetaDescription = "Login failed"
		data.Header = "Login Failed"
		data.HeaderClass = "error"
		data.Message = "Invalid username or password."
		data.ButtonURL = "/user/login"
		data.ButtonText = "Try Again"
	} else {
		// Create session
		session, err := h.db.CreateSession(user.ID, user.Username, SessionDuration)
		if err != nil {
			database.LogSQLError(err)
			// Log failed login attempt (authentication succeeded but session creation failed)
			if h.logger != nil {
				h.logger.LogLogin(username, remoteAddr, false)
			}
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
			
			// Log successful login
			if h.logger != nil {
				h.logger.LogLogin(username, remoteAddr, true)
			}
			
			data.Title = "Login Success - Sting Ray"
			data.MetaDescription = "Login successful"
			data.Header = "Login Successful!"
			data.HeaderClass = "success"
			data.Message = "Welcome, " + user.Username + "! You are now logged in."
			data.ButtonURL = "/"
			data.ButtonText = "Go Home"
		}
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
		UserID      int
		LoginTime   string
	}
	
	data.Title = "User Profile - Sting Ray"
	data.MetaDescription = "User profile page"
	data.Header = "User Profile"
	data.HeaderClass = "success"
	data.Message = "Welcome to your profile page!"
	data.ButtonURL = "/"
	data.ButtonText = "Go Home"
	data.Footer = "© 2025 StingRay"
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
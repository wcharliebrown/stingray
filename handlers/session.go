package handlers

import (
	"net/http"
	"stingray/database"
	"stingray/models"
	"time"
)

const (
	SessionCookieName = "stingray_session"
	SessionDuration   = 24 * time.Hour // 24 hours
)

// SessionMiddleware handles session management
type SessionMiddleware struct {
	db *database.Database
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(db *database.Database) *SessionMiddleware {
	return &SessionMiddleware{db: db}
}

// GetSessionFromRequest extracts and validates session from request
func (m *SessionMiddleware) GetSessionFromRequest(r *http.Request) (*models.Session, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, err
	}

	session, err := m.db.GetSession(cookie.Value)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// SetSessionCookie sets the session cookie in the response
func (m *SessionMiddleware) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(SessionDuration),
	})
}

// ClearSessionCookie removes the session cookie
func (m *SessionMiddleware) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-1 * time.Hour), // Expire immediately
		MaxAge:   -1,
	})
}

// IsAuthenticated checks if the request has a valid session
func (m *SessionMiddleware) IsAuthenticated(r *http.Request) bool {
	_, err := m.GetSessionFromRequest(r)
	return err == nil
}

// RequireAuth middleware that redirects to login if not authenticated
func (m *SessionMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !m.IsAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// OptionalAuth middleware that adds session info to request context if authenticated
func (m *SessionMiddleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if session, err := m.GetSessionFromRequest(r); err == nil {
			// Add session to request context or headers for template access
			r.Header.Set("X-User-ID", session.UserID)
			r.Header.Set("X-Username", session.Username)
		}
		next.ServeHTTP(w, r)
	}
} 
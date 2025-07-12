package handlers

import (
	"net/http"
	"stingray/database"
)

// RoleMiddleware handles role-based access control
type RoleMiddleware struct {
	db *database.Database
	sm *SessionMiddleware
}

// NewRoleMiddleware creates a new role middleware
func NewRoleMiddleware(db *database.Database) *RoleMiddleware {
	return &RoleMiddleware{
		db: db,
		sm: NewSessionMiddleware(db),
	}
}

// RequireAuth middleware ensures user is authenticated
func (rm *RoleMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rm.sm.IsAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// RequireGroup middleware ensures user is in a specific group
func (rm *RoleMiddleware) RequireGroup(groupName string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// First check if user is authenticated
			if !rm.sm.IsAuthenticated(r) {
				http.Redirect(w, r, "/user/login", http.StatusSeeOther)
				return
			}

			// Get session to get user ID
			session, err := rm.sm.GetSessionFromRequest(r)
			if err != nil {
				http.Redirect(w, r, "/user/login", http.StatusSeeOther)
				return
			}

			// Check if user is in required group
			isInGroup, err := rm.db.IsUserInGroup(session.UserID, groupName)
			if err != nil || !isInGroup {
				RenderMessage(w, "Access Denied", "Access Denied", "error", 
					"You do not have permission to access this page.", "/", "Go Home", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireAdmin middleware ensures user is in admin group
func (rm *RoleMiddleware) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return rm.RequireGroup("admin")(next)
}

// RequireCustomer middleware ensures user is in customers group
func (rm *RoleMiddleware) RequireCustomer(next http.HandlerFunc) http.HandlerFunc {
	return rm.RequireGroup("customers")(next)
} 
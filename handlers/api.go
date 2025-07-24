package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"stingray/database"
	"stingray/models"
	"stingray/config"
)

// APIHandler handles API requests
type APIHandler struct {
	db  *database.Database
	rm  *RoleMiddleware
	cfg *config.Config
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(db *database.Database, cfg *config.Config) *APIHandler {
	return &APIHandler{
		db:  db,
		rm:  NewRoleMiddleware(db),
		cfg: cfg,
	}
}

// API Response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HandleGetUsers returns all users (admin only)
func (h *APIHandler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check admin permissions
	h.rm.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
		users, err := h.db.GetAllUsers()
		if err != nil {
			database.LogSQLError(err)
			response := APIResponse{
				Success: false,
				Error:   "Failed to retrieve users: " + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Don't return passwords in API response
		type SafeUser struct {
			ID        int    `json:"id"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}

		safeUsers := make([]SafeUser, len(users))
		for i, user := range users {
			safeUsers[i] = SafeUser{
				ID:        user.ID,
				Username:  user.Username,
				Email:     user.Email,
				CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		response := APIResponse{
			Success: true,
			Data:    safeUsers,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})(w, r)
}

// HandleGetGroups returns all groups (admin only)
func (h *APIHandler) HandleGetGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check admin permissions
	h.rm.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
		groups, err := h.db.GetAllGroups()
		if err != nil {
			database.LogSQLError(err)
			response := APIResponse{
				Success: false,
				Error:   "Failed to retrieve groups: " + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := APIResponse{
			Success: true,
			Data:    groups,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})(w, r)
}

// HandleGetUserGroups returns groups for a specific user (admin only)
func (h *APIHandler) HandleGetUserGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check admin permissions
	h.rm.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			response := APIResponse{
				Success: false,
				Error:   "user_id parameter is required",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			response := APIResponse{
				Success: false,
				Error:   "Invalid user_id parameter",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		groups, err := h.db.GetUserGroups(userID)
		if err != nil {
			database.LogSQLError(err)
			response := APIResponse{
				Success: false,
				Error:   "Failed to retrieve user groups: " + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := APIResponse{
			Success: true,
			Data:    groups,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})(w, r)
}

// HandleGetCurrentUser returns current user info
func (h *APIHandler) HandleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	h.rm.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		session, err := h.rm.sm.GetSessionFromRequest(r)
		if err != nil {
			response := APIResponse{
				Success: false,
				Error:   "Failed to get session",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		user, err := h.db.GetUserByID(session.UserID)
		if err != nil {
			database.LogSQLError(err)
			response := APIResponse{
				Success: false,
				Error:   "Failed to get user: " + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		groups, err := h.db.GetUserGroups(user.ID)
		if err != nil {
			database.LogSQLError(err)
			response := APIResponse{
				Success: false,
				Error:   "Failed to get user groups: " + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		type UserInfo struct {
			ID        int           `json:"id"`
			Username  string        `json:"username"`
			Email     string        `json:"email"`
			CreatedAt string        `json:"created_at"`
			UpdatedAt string        `json:"updated_at"`
			Groups    []models.Group `json:"groups"`
		}

		userInfo := UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
			Groups:    groups,
		}

		response := APIResponse{
			Success: true,
			Data:    userInfo,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})(w, r)
}

// HandleReloadEnv reloads the .env file and updates the config
func (h *APIHandler) HandleReloadEnv(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.rm.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
		success, errMsg := ReloadEnvConfig(h.cfg)
		if !success {
			response := APIResponse{
				Success: false,
				Error:   errMsg,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
		response := APIResponse{
			Success: true,
			Message: ".env file reloaded successfully.",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})(w, r)
} 
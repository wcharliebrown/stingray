package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"stingray/config"
	"stingray/database"
	"stingray/email"
	"stingray/logging"
	"stingray/templates"
	"time"
)

// PasswordResetHandler handles password reset functionality
type PasswordResetHandler struct {
	db     *database.Database
	cfg    *config.Config
	email  *email.EmailService
	logger *logging.Logger
}

// NewPasswordResetHandler creates a new password reset handler
func NewPasswordResetHandler(db *database.Database, cfg *config.Config, logger *logging.Logger) *PasswordResetHandler {
	emailService, err := email.NewEmailService(
		cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword,
		cfg.FromEmail, cfg.FromName, cfg.DKIMPrivateKeyFile, cfg.DKIMSelector, cfg.DKIMDomain,
	)
	if err != nil {
		// Log error but continue without email service
		if logger != nil {
			logger.LogError("Failed to initialize email service: %v", err)
		} else {
			fmt.Printf("Warning: Failed to initialize email service: %v\n", err)
		}
		emailService = nil
	}

	return &PasswordResetHandler{
		db:     db,
		cfg:    cfg,
		email:  emailService,
		logger: logger,
	}
}

// HandlePasswordResetRequest handles the password reset request page
func (h *PasswordResetHandler) HandlePasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Show password reset request form
		page, err := h.db.GetPage("password-reset-request")
		if err != nil {
			database.LogSQLError(err)
			RenderMessage(w, "Password Reset", "Password Reset", "info", 
				"Enter your email address to receive a password reset link.", "/user/login", "Back to Login", http.StatusOK)
			return
		}

		html, err := templates.RenderPage(page)
		if err != nil {
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
		return
	}

	if r.Method != "POST" {
		RenderMessage(w, "405 Method Not Allowed", "Method Not Allowed", "error", 
			"Only POST is allowed for this endpoint.", "/user/password-reset-request", "Back to Reset Request", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		RenderMessage(w, "Password Reset Failed", "Password Reset Failed", "error", 
			"Email address is required.", "/user/password-reset-request", "Try Again", http.StatusBadRequest)
		return
	}

	// Check if user exists
	user, err := h.db.GetUserByEmail(email)
	if err != nil {
		// Don't reveal if user exists or not for security
		RenderMessage(w, "Password Reset Requested", "Password Reset Requested", "success", 
			"If an account with that email exists, a password reset link has been sent.", "/user/login", "Back to Login", http.StatusOK)
		return
	}

	// Generate reset token
	token, err := generateResetToken()
	if err != nil {
		database.LogSQLError(err)
		RenderMessage(w, "Password Reset Error", "Password Reset Error", "error", 
			"Failed to generate reset token. Please try again.", "/user/password-reset-request", "Try Again", http.StatusInternalServerError)
		return
	}

	// Set expiration (1 hour from now)
	expiresAt := time.Now().Add(1 * time.Hour)

	// Create reset token in database
	err = h.db.CreatePasswordResetToken(user.ID, email, token, expiresAt)
	if err != nil {
		database.LogSQLError(err)
		RenderMessage(w, "Password Reset Error", "Password Reset Error", "error", 
			"Failed to create reset token. Please try again.", "/user/password-reset-request", "Try Again", http.StatusInternalServerError)
		return
	}

	// Send password reset email
	if h.email != nil {
		resetURL := fmt.Sprintf("http://localhost:6273/user/password-reset-confirm?token=%s", token)
		err = h.email.SendPasswordResetEmail(email, resetURL)
		if err != nil {
			h.logger.LogError("Failed to send password reset email: %v", err)
			// Don't reveal the error to the user for security
			RenderMessage(w, "Password Reset Error", "Password Reset Error", "error", 
				"Failed to send reset email. Please try again.", "/user/password-reset-request", "Try Again", http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback for when email service is not available
		resetURL := fmt.Sprintf("/user/password-reset-confirm?token=%s", token)
		RenderMessage(w, "Password Reset Requested", "Password Reset Requested", "success", 
			fmt.Sprintf("If an account with that email exists, a password reset link has been sent. For testing, you can use this link: %s", resetURL), 
			"/user/login", "Back to Login", http.StatusOK)
		return
	}
	
	RenderMessage(w, "Password Reset Requested", "Password Reset Requested", "success", 
		"If an account with that email exists, a password reset link has been sent to your email address.", 
		"/user/login", "Back to Login", http.StatusOK)
}

// HandlePasswordResetConfirm handles the password reset confirmation page
func (h *PasswordResetHandler) HandlePasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		RenderMessage(w, "Invalid Reset Link", "Invalid Reset Link", "error", 
			"The password reset link is invalid or missing.", "/user/password-reset-request", "Request New Reset", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		// Verify token exists and is not expired
		resetToken, err := h.db.GetPasswordResetToken(token)
		if err != nil {
			RenderMessage(w, "Invalid Reset Link", "Invalid Reset Link", "error", 
				"The password reset link is invalid or has expired.", "/user/password-reset-request", "Request New Reset", http.StatusBadRequest)
			return
		}

		if resetToken.Used {
			RenderMessage(w, "Token Already Used", "Token Already Used", "error", 
				"This password reset link has already been used.", "/user/password-reset-request", "Request New Reset", http.StatusBadRequest)
			return
		}

		if time.Now().After(resetToken.ExpiresAt) {
			RenderMessage(w, "Token Expired", "Token Expired", "error", 
				"This password reset link has expired.", "/user/password-reset-request", "Request New Reset", http.StatusBadRequest)
			return
		}

		// Show password reset form
		page, err := h.db.GetPage("password-reset-confirm")
		if err != nil {
			database.LogSQLError(err)
			// Create a simple form if page doesn't exist
			html := fmt.Sprintf(`
				<!DOCTYPE html>
				<html>
				<head>
					<title>Reset Password - Sting Ray</title>
					<meta charset="utf-8">
					<style>
						body { font-family: Arial, sans-serif; margin: 40px; }
						.form-group { margin-bottom: 15px; }
						label { display: block; margin-bottom: 5px; }
						input[type="password"] { width: 300px; padding: 8px; }
						.btn { padding: 10px 20px; background: #007cba; color: white; border: none; cursor: pointer; }
						.btn:hover { background: #005a87; }
					</style>
				</head>
				<body>
					<h1>Reset Password</h1>
					<form method="post">
						<input type="hidden" name="token" value="%s">
						<div class="form-group">
							<label for="password">New Password:</label>
							<input type="password" id="password" name="password" required>
						</div>
						<div class="form-group">
							<label for="confirm_password">Confirm Password:</label>
							<input type="password" id="confirm_password" name="confirm_password" required>
						</div>
						<button type="submit" class="btn">Reset Password</button>
					</form>
					<p><a href="/user/login">Back to Login</a></p>
				</body>
				</html>`, token)
			
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(html))
			return
		}

		// Replace token placeholder in content
		page.MainContent = fmt.Sprintf(page.MainContent, token)

		html, err := templates.RenderPage(page)
		if err != nil {
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
		return
	}

	if r.Method != "POST" {
		RenderMessage(w, "405 Method Not Allowed", "Method Not Allowed", "error", 
			"Only POST is allowed for this endpoint.", "/user/password-reset-confirm?token="+token, "Back to Reset Form", http.StatusMethodNotAllowed)
		return
	}

	// Process password reset
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if password == "" {
		RenderMessage(w, "Password Reset Failed", "Password Reset Failed", "error", 
			"Password is required.", "/user/password-reset-confirm?token="+token, "Try Again", http.StatusBadRequest)
		return
	}

	if password != confirmPassword {
		RenderMessage(w, "Password Reset Failed", "Password Reset Failed", "error", 
			"Passwords do not match.", "/user/password-reset-confirm?token="+token, "Try Again", http.StatusBadRequest)
		return
	}

	// Verify token again
	resetToken, err := h.db.GetPasswordResetToken(token)
	if err != nil {
		RenderMessage(w, "Invalid Reset Link", "Invalid Reset Link", "error", 
			"The password reset link is invalid or has expired.", "/user/password-reset-request", "Request New Reset", http.StatusBadRequest)
		return
	}

	if resetToken.Used {
		RenderMessage(w, "Token Already Used", "Token Already Used", "error", 
			"This password reset link has already been used.", "/user/password-reset-request", "Request New Reset", http.StatusBadRequest)
		return
	}

	if time.Now().After(resetToken.ExpiresAt) {
		RenderMessage(w, "Token Expired", "Token Expired", "error", 
			"This password reset link has expired.", "/user/password-reset-request", "Request New Reset", http.StatusBadRequest)
		return
	}

	// Update user password
	err = h.db.UpdateUserPassword(resetToken.UserID, password)
	if err != nil {
		database.LogSQLError(err)
		RenderMessage(w, "Password Reset Error", "Password Reset Error", "error", 
			"Failed to update password. Please try again.", "/user/password-reset-confirm?token="+token, "Try Again", http.StatusInternalServerError)
		return
	}

	// Mark token as used
	err = h.db.MarkPasswordResetTokenUsed(token)
	if err != nil {
		database.LogSQLError(err)
		// Don't fail the reset if marking as used fails
	}

	RenderMessage(w, "Password Reset Successful", "Password Reset Successful", "success", 
		"Your password has been successfully reset. You can now login with your new password.", "/user/login", "Login", http.StatusOK)
}

// generateResetToken generates a secure random token for password reset
func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
} 
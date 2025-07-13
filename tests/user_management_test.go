package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"stingray/database"
	"stingray/handlers"
)

func setupTestDatabase(t *testing.T) *database.Database {
	// Use a test database
	dsn := "root:password@tcp(localhost:3306)/stingray_test?parseTime=true"
	db, err := database.NewDatabase(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func cleanupTestDatabase(t *testing.T, db *database.Database) {
	// Clean up test data
	db.GetDB().Exec("DELETE FROM user_groups")
	db.GetDB().Exec("DELETE FROM sessions")
	db.GetDB().Exec("DELETE FROM users")
	db.GetDB().Exec("DELETE FROM user_groups_table")
	db.GetDB().Exec("DELETE FROM pages")
}

func TestUserAuthentication(t *testing.T) {
	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)
	defer db.Close()

	// Test user authentication with environment passwords
	adminPassword := os.Getenv("TEST_ADMIN_PASSWORD")
	
	customerPassword := os.Getenv("TEST_CUSTOMER_PASSWORD")

	// Test admin authentication
	user, err := db.AuthenticateUser("admin", adminPassword)
	if err != nil {
		t.Fatalf("Failed to authenticate admin user: %v", err)
	}

	if user.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", user.Username)
	}

	if user.Email != "adminuser@servicecompany.net" {
		t.Errorf("Expected email 'adminuser@servicecompany.net', got '%s'", user.Email)
	}

	// Test invalid credentials
	_, err = db.AuthenticateUser("admin", "wrongpassword")
	if err == nil {
		t.Error("Expected authentication to fail with wrong password")
	}

	// Test customer authentication
	customer, err := db.AuthenticateUser("customer", customerPassword)
	if err != nil {
		t.Fatalf("Failed to authenticate customer user: %v", err)
	}

	if customer.Username != "customer" {
		t.Errorf("Expected username 'customer', got '%s'", customer.Username)
	}
}

func TestUserGroups(t *testing.T) {
	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)
	defer db.Close()

	// Get admin user
	admin, err := db.AuthenticateUser("admin", "admin123")
	if err != nil {
		t.Fatalf("Failed to authenticate admin user: %v", err)
	}

	// Get customer user
	customer, err := db.AuthenticateUser("customer", "customer123")
	if err != nil {
		t.Fatalf("Failed to authenticate customer user: %v", err)
	}

	// Test admin groups
	adminGroups, err := db.GetUserGroups(admin.ID)
	if err != nil {
		t.Fatalf("Failed to get admin groups: %v", err)
	}

	if len(adminGroups) != 1 {
		t.Errorf("Expected 1 group for admin, got %d", len(adminGroups))
	}

	if adminGroups[0].Name != "admin" {
		t.Errorf("Expected admin group name 'admin', got '%s'", adminGroups[0].Name)
	}

	// Test customer groups
	customerGroups, err := db.GetUserGroups(customer.ID)
	if err != nil {
		t.Fatalf("Failed to get customer groups: %v", err)
	}

	if len(customerGroups) != 1 {
		t.Errorf("Expected 1 group for customer, got %d", len(customerGroups))
	}

	if customerGroups[0].Name != "customers" {
		t.Errorf("Expected customer group name 'customers', got '%s'", customerGroups[0].Name)
	}

	// Test group membership checks
	isAdmin, err := db.IsUserInGroup(admin.ID, "admin")
	if err != nil {
		t.Fatalf("Failed to check admin group membership: %v", err)
	}
	if !isAdmin {
		t.Error("Admin should be in admin group")
	}

	isCustomer, err := db.IsUserInGroup(customer.ID, "customers")
	if err != nil {
		t.Fatalf("Failed to check customer group membership: %v", err)
	}
	if !isCustomer {
		t.Error("Customer should be in customers group")
	}

	// Test that admin is not in customers group
	isAdminInCustomers, err := db.IsUserInGroup(admin.ID, "customers")
	if err != nil {
		t.Fatalf("Failed to check admin in customers group: %v", err)
	}
	if isAdminInCustomers {
		t.Error("Admin should not be in customers group")
	}
}

func TestSessionManagement(t *testing.T) {
	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)
	defer db.Close()

	// Get admin user
	admin, err := db.AuthenticateUser("admin", "admin123")
	if err != nil {
		t.Fatalf("Failed to authenticate admin user: %v", err)
	}

	// Create session
	session, err := db.CreateSession(admin.ID, admin.Username, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.UserID != admin.ID {
		t.Errorf("Expected session user ID %d, got %d", admin.ID, session.UserID)
	}

	if session.Username != admin.Username {
		t.Errorf("Expected session username '%s', got '%s'", admin.Username, session.Username)
	}

	// Retrieve session
	retrievedSession, err := db.GetSession(session.SessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}

	if retrievedSession.SessionID != session.SessionID {
		t.Errorf("Expected session ID '%s', got '%s'", session.SessionID, retrievedSession.SessionID)
	}

	// Invalidate session
	err = db.InvalidateSession(session.SessionID)
	if err != nil {
		t.Fatalf("Failed to invalidate session: %v", err)
	}

	// Try to retrieve invalidated session
	_, err = db.GetSession(session.SessionID)
	if err == nil {
		t.Error("Expected error when retrieving invalidated session")
	}
}

func TestLoginHandler(t *testing.T) {
	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)
	defer db.Close()

	authHandler := handlers.NewAuthHandler(db)

	// Get admin password from environment
	adminPassword := os.Getenv("TEST_ADMIN_PASSWORD")

	// Test successful login
	formData := url.Values{}
	formData.Set("username", "admin")
	formData.Set("password", adminPassword)

	req := httptest.NewRequest("POST", "/user/login_post", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	authHandler.HandleLoginPost(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test failed login
	formData = url.Values{}
	formData.Set("username", "admin")
	formData.Set("password", "wrongpassword")

	req = httptest.NewRequest("POST", "/user/login_post", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	authHandler.HandleLoginPost(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that response contains error message
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "Invalid username or password") {
		t.Error("Expected error message in response")
	}
}

func TestRoleMiddleware(t *testing.T) {
	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)
	defer db.Close()

	roleMW := handlers.NewRoleMiddleware(db)

	// Test admin access to admin-only page
	admin, err := db.AuthenticateUser("admin", "admin123")
	if err != nil {
		t.Fatalf("Failed to authenticate admin user: %v", err)
	}

	session, err := db.CreateSession(admin.ID, admin.Username, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	req := httptest.NewRequest("GET", "/page/orders", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.SessionID,
	})
	w := httptest.NewRecorder()

	// Create a simple handler for testing
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Access granted"))
	}

	roleMW.RequireAdmin(testHandler)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test customer access to admin-only page (should be denied)
	customer, err := db.AuthenticateUser("customer", "customer123")
	if err != nil {
		t.Fatalf("Failed to authenticate customer user: %v", err)
	}

	customerSession, err := db.CreateSession(customer.ID, customer.Username, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create customer session: %v", err)
	}

	req = httptest.NewRequest("GET", "/page/orders", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: customerSession.SessionID,
	})
	w = httptest.NewRecorder()

	roleMW.RequireAdmin(testHandler)(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	// Test customer access to customer-only page
	req = httptest.NewRequest("GET", "/page/faq", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: customerSession.SessionID,
	})
	w = httptest.NewRecorder()

	roleMW.RequireCustomer(testHandler)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAPIEndpoints(t *testing.T) {
	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)
	defer db.Close()

	apiHandler := handlers.NewAPIHandler(db)

	// Test getting users (admin only)
	admin, err := db.AuthenticateUser("admin", "admin123")
	if err != nil {
		t.Fatalf("Failed to authenticate admin user: %v", err)
	}

	session, err := db.CreateSession(admin.ID, admin.Username, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/users", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.SessionID,
	})
	w := httptest.NewRecorder()

	apiHandler.HandleGetUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response handlers.APIResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}

	// Test getting current user
	req = httptest.NewRequest("GET", "/api/current-user", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.SessionID,
	})
	w = httptest.NewRecorder()

	apiHandler.HandleGetCurrentUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}

	// Test getting groups
	req = httptest.NewRequest("GET", "/api/groups", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.SessionID,
	})
	w = httptest.NewRecorder()

	apiHandler.HandleGetGroups(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}
}

func TestDatabaseOperations(t *testing.T) {
	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)
	defer db.Close()

	// Test getting all users
	users, err := db.GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	if len(users) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(users))
	}

	// Test getting all groups
	groups, err := db.GetAllGroups()
	if err != nil {
		t.Fatalf("Failed to get all groups: %v", err)
	}

	if len(groups) < 2 {
		t.Errorf("Expected at least 2 groups, got %d", len(groups))
	}

	// Test getting user by ID
	admin, err := db.AuthenticateUser("admin", "admin123")
	if err != nil {
		t.Fatalf("Failed to authenticate admin user: %v", err)
	}

	retrievedUser, err := db.GetUserByID(admin.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrievedUser.Username != admin.Username {
		t.Errorf("Expected username '%s', got '%s'", admin.Username, retrievedUser.Username)
	}
} 
package database

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"log"
	"path/filepath"
	"stingray/auth"
	"stingray/models"
	"strings"
	"time"
	"database/sql"
)

func (d *Database) initDatabase() error {
	// Create database if it doesn't exist
	createDBQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", "stingray")
	_, err := d.Exec(createDBQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create pages table
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS _page (
		id INT AUTO_INCREMENT PRIMARY KEY,
		slug VARCHAR(255) UNIQUE NOT NULL,
		title VARCHAR(255) NOT NULL,
		meta_description TEXT,
		header TEXT,
		navigation TEXT,
		main_content TEXT,
		sidebar TEXT,
		footer TEXT,
		css_class VARCHAR(255),
		scripts TEXT,
		template VARCHAR(100) DEFAULT 'default',
		read_groups TEXT,
		write_groups TEXT,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.Exec(createTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create groups table (no dependencies)
	createGroupsTableQuery := `
	CREATE TABLE IF NOT EXISTS _group (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		description TEXT,
		read_groups TEXT,
		write_groups TEXT,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_name (name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.Exec(createGroupsTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create users table (no dependencies)
	createUsersTableQuery := `
	CREATE TABLE IF NOT EXISTS _user (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		read_groups TEXT,
		write_groups TEXT,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_username (username),
		INDEX idx_email (email)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.Exec(createUsersTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create user_and_group table (depends on users and groups)
	createUserGroupsTableQuery := `
	CREATE TABLE IF NOT EXISTS _user_and_group (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		group_id INT NOT NULL,
		read_groups TEXT,
		write_groups TEXT,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_user_group (user_id, group_id),
		FOREIGN KEY (user_id) REFERENCES _user(id) ON DELETE CASCADE,
		FOREIGN KEY (group_id) REFERENCES _group(id) ON DELETE CASCADE,
		INDEX idx_user_id (user_id),
		INDEX idx_group_id (group_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.Exec(createUserGroupsTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create sessions table (depends on users)
	createSessionsTableQuery := `
	CREATE TABLE IF NOT EXISTS _session (
		id INT AUTO_INCREMENT PRIMARY KEY,
		session_id VARCHAR(255) UNIQUE NOT NULL,
		user_id INT NOT NULL,
		username VARCHAR(255) NOT NULL,
		read_groups TEXT,
		write_groups TEXT,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		INDEX idx_session_id (session_id),
		INDEX idx_expires_at (expires_at),
		INDEX idx_is_active (is_active),
		FOREIGN KEY (user_id) REFERENCES _user(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.Exec(createSessionsTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create _table_metadata table
	createTableMetadataQuery := `
	CREATE TABLE IF NOT EXISTS _table_metadata (
		id INT AUTO_INCREMENT PRIMARY KEY,
		table_name VARCHAR(255) UNIQUE NOT NULL,
		display_name VARCHAR(255) NOT NULL,
		description TEXT,
		read_groups TEXT,
		write_groups TEXT,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_table_name (table_name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.Exec(createTableMetadataQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create _field_metadata table
	createFieldMetadataQuery := `
	CREATE TABLE IF NOT EXISTS _field_metadata (
		id INT AUTO_INCREMENT PRIMARY KEY,
		table_name VARCHAR(255) NOT NULL,
		field_name VARCHAR(255) NOT NULL,
		display_name VARCHAR(255) NOT NULL,
		description TEXT,
		db_type VARCHAR(100) NOT NULL,
		html_input_type VARCHAR(100) NOT NULL,
		form_position INT DEFAULT 0,
		list_position INT DEFAULT 0,
		is_required BOOLEAN DEFAULT FALSE,
		is_read_only BOOLEAN DEFAULT FALSE,
		default_value TEXT,
		validation_rules TEXT,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_table_field (table_name, field_name),
		INDEX idx_table_name (table_name),
		INDEX idx_field_name (field_name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.Exec(createFieldMetadataQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create password reset tokens table
	createPasswordResetTokensQuery := `
	CREATE TABLE IF NOT EXISTS _password_reset_token (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		token VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		used BOOLEAN DEFAULT FALSE,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_token (token),
		INDEX idx_user_id (user_id),
		INDEX idx_expires_at (expires_at),
		INDEX idx_used (used),
		FOREIGN KEY (user_id) REFERENCES _user(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.Exec(createPasswordResetTokensQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize with default pages and users
	if err := d.initializePages(); err != nil {
		LogSQLError(err)
		return err
	}

	if err := d.initializeUsers(); err != nil {
		LogSQLError(err)
		return err
	}

	if err := d.initializeMetadata(); err != nil {
		LogSQLError(err)
		return err
	}

	// Update existing metadata to include engineer group access
	if err := d.updateExistingMetadataForEngineer(); err != nil {
		LogSQLError(err)
		return err
	}

	// Update existing metadata to include everyone group access
	if err := d.updateExistingMetadataForEveryone(); err != nil {
		LogSQLError(err)
		return err
	}

	// Run migration to add permission fields to existing tables
	if err := d.migrateAddPermissionFields(); err != nil {
		LogSQLError(err)
		return err
	}

	// Migrate existing db_type values to include proper length specifications
	if err := d.migrateUpdateDBTypes(); err != nil {
		LogSQLError(err)
		return err
	}

	return nil
}

func (d *Database) initializePages() error {
	pages := []models.Page{
		{
			Slug:           "home",
			Title:          "Welcome to Sting Ray",
			MetaDescription: "A modern content management system built with Go",
			Header:         "Welcome to Sting Ray",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Welcome to Sting Ray</h2><p>This is a modern content management system built with Go and MySQL. Features include dynamic page serving, template system, and RESTful API endpoints.</p><h3>New Features:</h3><ul><li><a href="/metadata/tables">Database Table Management</a> - View and edit database tables with metadata-driven forms</li><li>Role-based access control for table operations</li><li>Engineer mode for technical users</li></ul>`,
			Sidebar:        `<h3>Quick Links</h3><ul><li><a href="/page/about">About</a></li><li><a href="/page/demo">Demo</a></li><li><a href="/user/login">Login</a></li><li><a href="/metadata/tables">Database Tables</a></li></ul>`,
			Footer:         "© 2025 StingRay",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
			ReadGroups:     sql.NullString{String: "[\"everyone\"]", Valid: true},
			WriteGroups:    sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
		},
		{
			Slug:           "about",
			Title:          "About Sting Ray",
			MetaDescription: "Learn about the Sting Ray content management system",
			Header:         "About Sting Ray",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>About Sting Ray</h2><p>Sting Ray is a modern content management system built with Go and MySQL.</p><h3>Features:</h3><ul><li>Dynamic page serving</li><li>Template system with embedded templates</li><li>RESTful API endpoints</li><li>MySQL database backend</li><li>Responsive design</li></ul>`,
			Sidebar:        `<h3>Technology Stack</h3><ul><li>Go 1.24.4</li><li>MySQL 5.7+</li><li>HTML Templates</li><li>RESTful APIs</li></ul>`,
			Footer:         "© 2025 StingRay",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
			ReadGroups:     sql.NullString{String: "[\"everyone\"]", Valid: true},
			WriteGroups:    sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
		},
		{
			Slug:           "login",
			Title:          "Login - Sting Ray",
			MetaDescription: "User login page",
			Header:         "User Login",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Login</h2><p>Please enter your credentials to access the system.</p>{{template_login_form}}`,
			Sidebar:        `<h3>Need Help?</h3><p>Contact the administrator for login credentials.</p>`,
			Footer:         "© 2025 StingRay",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
			ReadGroups:     sql.NullString{String: "[\"everyone\"]", Valid: true},
			WriteGroups:    sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
		},
		{
			Slug:           "shutdown",
			Title:          "Server Shutdown - Sting Ray",
			MetaDescription: "Server shutdown confirmation",
			Header:         "Server Shutdown",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Server Shutdown</h2><p>The server is shutting down gracefully. Please wait...</p><div id="countdown">30</div><script>let count = 30; const timer = setInterval(() => { count--; document.getElementById('countdown').textContent = count; if (count <= 0) { clearInterval(timer); window.location.href = '/'; } }, 1000);</script>`,
			Sidebar:        `<h3>Shutdown Progress</h3><p>Server will be unavailable during shutdown.</p>`,
			Footer:         "© 2025 StingRay",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
			ReadGroups:     sql.NullString{String: "[\"everyone\"]", Valid: true},
			WriteGroups:    sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
		},
		{
			Slug:           "demo",
			Title:          "Template Demo - Sting Ray",
			MetaDescription: "Demonstration of embedded templates",
			Header:         "Template Demo",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Embedded Templates Demo</h2><p>This page demonstrates the embedded template system.</p><h3>Login Form Embedded:</h3>{{template_login_form}}<h3>Custom Content:</h3><p>This content is rendered with the modern template system.</p>`,
			Sidebar:        `<h3>Template Features</h3><ul><li>Embedded templates</li><li>Recursive processing</li><li>Multiple template support</li><li>Responsive design</li></ul>`,
			Footer:         "© 2025 StingRay",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
			ReadGroups:     sql.NullString{String: "[\"everyone\"]", Valid: true},
			WriteGroups:    sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
		},
		{
			Slug:           "orders",
			Title:          "Orders Management - Sting Ray",
			MetaDescription: "Admin orders management page",
			Header:         "Orders Management",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/page/orders">Orders</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Orders Management</h2><p>Welcome to the orders management system. This page is only accessible to administrators.</p><h3>Recent Orders:</h3><ul><li>Order #1001 - Customer A - $150.00</li><li>Order #1002 - Customer B - $75.50</li><li>Order #1003 - Customer C - $200.00</li></ul><h3>Quick Actions:</h3><ul><li><a href="#">View All Orders</a></li><li><a href="#">Create New Order</a></li><li><a href="#">Export Orders</a></li></ul>`,
			Sidebar:        `<h3>Order Statistics</h3><ul><li>Total Orders: 1,247</li><li>Pending Orders: 23</li><li>Completed Orders: 1,224</li><li>Total Revenue: $45,678.90</li></ul>`,
			Footer:         "© 2025 StingRay",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
			ReadGroups:     sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
			WriteGroups:    sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
		},
		{
			Slug:           "faq",
			Title:          "Frequently Asked Questions - Sting Ray",
			MetaDescription: "Customer FAQ page",
			Header:         "Frequently Asked Questions",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/page/faq">FAQ</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Frequently Asked Questions</h2><p>Welcome to our FAQ section. This page is only accessible to customers.</p><h3>General Questions:</h3><div class="faq-item"><h4>How do I place an order?</h4><p>You can place an order by contacting our sales team or using our online ordering system.</p></div><div class="faq-item"><h4>What are your payment terms?</h4><p>We accept payment upon delivery or within 30 days of invoice.</p></div><div class="faq-item"><h4>Do you offer shipping?</h4><p>Yes, we offer shipping to all locations within our service area.</p></div><h3>Technical Support:</h3><div class="faq-item"><h4>How do I get technical support?</h4><p>Contact our technical support team at support@company.com or call 1-800-SUPPORT.</p></div>`,
			Sidebar:        `<h3>Quick Contact</h3><ul><li>Sales: sales@company.com</li><li>Support: support@company.com</li><li>Phone: 1-800-COMPANY</li></ul><h3>Helpful Links</h3><ul><li><a href="#">Product Catalog</a></li><li><a href="#">Order Status</a></li><li><a href="#">Return Policy</a></li></ul>`,
			Footer:         "© 2025 StingRay",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
			ReadGroups:     sql.NullString{String: "[\"customers\", \"admin\", \"engineer\"]", Valid: true},
			WriteGroups:    sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
		},
		{
			Slug:           "password-reset-request",
			Title:          "Password Reset Request - Sting Ray",
			MetaDescription: "Request a password reset",
			Header:         "Password Reset Request",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Password Reset Request</h2><p>Enter your email address to receive a password reset link.</p><div class="card"><form action="/user/password-reset-request" method="post"><div class="form-group"><label for="email">Email Address:</label><input type="email" id="email" name="email" required></div><button type="submit" class="btn">Send Reset Link</button></form></div>`,
			Sidebar:        `<h3>Need Help?</h3><p>If you don't remember your email address, contact the administrator for assistance.</p>`,
			Footer:         "© 2025 StingRay",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
			ReadGroups:     sql.NullString{String: "[\"everyone\"]", Valid: true},
			WriteGroups:    sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
		},
		{
			Slug:           "password-reset-confirm",
			Title:          "Reset Password - Sting Ray",
			MetaDescription: "Reset your password",
			Header:         "Reset Password",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Reset Password</h2><p>Enter your new password below.</p><div class="card"><form method="post"><input type="hidden" name="token" value="%s"><div class="form-group"><label for="password">New Password:</label><input type="password" id="password" name="password" required></div><div class="form-group"><label for="confirm_password">Confirm Password:</label><input type="password" id="confirm_password" name="confirm_password" required></div><button type="submit" class="btn">Reset Password</button></form></div>`,
			Sidebar:        `<h3>Password Requirements</h3><ul><li>At least 8 characters long</li><li>Include uppercase and lowercase letters</li><li>Include numbers and special characters</li></ul>`,
			Footer:         "© 2025 StingRay",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
			ReadGroups:     sql.NullString{String: "[\"everyone\"]", Valid: true},
			WriteGroups:    sql.NullString{String: "[\"admin\", \"engineer\"]", Valid: true},
		},
	}

	for _, page := range pages {
		if err := d.createPageIfNotExists(page); err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
}

func (d *Database) createPageIfNotExists(page models.Page) error {
	// Check if page exists
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM _page WHERE slug = ?", page.Slug).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		_, err = d.Exec(`
			INSERT INTO _page (slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template, read_groups, write_groups)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			page.Slug, page.Title, page.MetaDescription, page.Header, page.Navigation,
			page.MainContent, page.Sidebar, page.Footer, page.CSSClass, page.Scripts, page.Template,
			page.ReadGroups, page.WriteGroups)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
}

func (d *Database) GetPage(slug string) (*models.Page, error) {
	var page models.Page
	var readGroups, writeGroups sql.NullString
	err := d.QueryRow(`
		SELECT id, slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template, read_groups, write_groups, created, modified
		FROM _page WHERE slug = ?`, slug).Scan(
		&page.ID, &page.Slug, &page.Title, &page.MetaDescription, &page.Header,
		&page.Navigation, &page.MainContent, &page.Sidebar, &page.Footer,
		&page.CSSClass, &page.Scripts, &page.Template, &readGroups, &writeGroups,
		&page.Created, &page.Modified)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	page.ReadGroups = readGroups
	page.WriteGroups = writeGroups
	return &page, nil
}

func (d *Database) GetAllPages() ([]models.Page, error) {
	rows, err := d.Query(`
		SELECT id, slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template, read_groups, write_groups, created, modified
		FROM _page ORDER BY slug`)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	defer rows.Close()

	var pages []models.Page
	for rows.Next() {
		var page models.Page
		var readGroups, writeGroups sql.NullString
		err := rows.Scan(
			&page.ID, &page.Slug, &page.Title, &page.MetaDescription, &page.Header,
			&page.Navigation, &page.MainContent, &page.Sidebar, &page.Footer,
			&page.CSSClass, &page.Scripts, &page.Template, &readGroups, &writeGroups,
			&page.Created, &page.Modified)
		if err != nil {
			LogSQLError(err)
			return nil, err
		}
		page.ReadGroups = readGroups
		page.WriteGroups = writeGroups
		pages = append(pages, page)
	}
	return pages, nil
}

// Session operations
func (d *Database) CreateSession(userID int, username string, duration time.Duration) (*models.Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(duration)

	_, err = d.Exec(`
		INSERT INTO _session (session_id, user_id, username, expires_at, is_active)
		VALUES (?, ?, ?, ?, TRUE)`,
		sessionID, userID, username, expiresAt)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}

	session := &models.Session{
		SessionID: sessionID,
		UserID:    userID,
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	return session, nil
}

func (d *Database) GetSession(sessionID string) (*models.Session, error) {
	var session models.Session
	var readGroups, writeGroups sql.NullString
	err := d.QueryRow(`
		SELECT id, session_id, user_id, username, read_groups, write_groups, created, expires_at, is_active
		FROM _session WHERE session_id = ? AND is_active = TRUE AND expires_at > NOW()`,
		sessionID).Scan(
		&session.ID, &session.SessionID, &session.UserID, &session.Username,
		&readGroups, &writeGroups, &session.CreatedAt, &session.ExpiresAt, &session.IsActive)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	
	// Handle NULL values
	if readGroups.Valid {
		session.ReadGroups = readGroups.String
	} else {
		session.ReadGroups = ""
	}
	if writeGroups.Valid {
		session.WriteGroups = writeGroups.String
	} else {
		session.WriteGroups = ""
	}
	
	return &session, nil
}

func (d *Database) InvalidateSession(sessionID string) error {
	_, err := d.Exec(`
		UPDATE _session SET is_active = FALSE WHERE session_id = ?`,
		sessionID)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

func (d *Database) CleanupExpiredSessions() error {
	_, err := d.Exec(`
		UPDATE _session SET is_active = FALSE WHERE expires_at <= NOW()`)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Exported for testing
func GenerateSessionIDForTest() (string, error) {
	return generateSessionID()
}

// User and Group Management Functions

func (d *Database) initializeUsers() error {
	// Create default groups
	groups := []models.Group{
		{Name: "admin", Description: "Administrator group with full access"},
		{Name: "customers", Description: "Customer group with limited access"},
		{Name: "engineer", Description: "Engineer group with technical access"},
		{Name: "everyone", Description: "Special group that includes all users including unauthenticated users"},
	}

	for _, group := range groups {
		if err := d.createGroupIfNotExists(group); err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Create default users with hashed passwords
	users := []struct {
		user   models.User
		groups []string
	}{
		{
			user: models.User{
				Username: "admin",
				Email:    "adminuser@servicecompany.net",
				Password: "", // Will be set below
			},
			groups: []string{"admin"},
		},
		{
			user: models.User{
				Username: "customer",
				Email:    "customeruser@company.com",
				Password: "", // Will be set below
			},
			groups: []string{"customers"},
		},
		{
			user: models.User{
				Username: "engineer",
				Email:    "engineeruser@servicecompany.net",
				Password: "", // Will be set below
			},
			groups: []string{"engineer"},
		},
	}

	for _, userData := range users {
		// Hash the password from environment
		var plainPassword string
		if userData.user.Username == "admin" {
			plainPassword = os.Getenv("TEST_ADMIN_PASSWORD")
		} else if userData.user.Username == "customer" {
			plainPassword = os.Getenv("TEST_CUSTOMER_PASSWORD")
		} else if userData.user.Username == "engineer" {
			plainPassword = os.Getenv("TEST_ENGINEER_PASSWORD")
			if plainPassword == "" {
				plainPassword = "engineer" // Default password if not set in environment
			}
		}
		
		// Hash the password before storing
		hashedPassword, err := auth.HashPassword(plainPassword)
		if err != nil {
			return fmt.Errorf("failed to hash password for user %s: %w", userData.user.Username, err)
		}
		userData.user.Password = hashedPassword
		
		if err := d.createUserIfNotExists(userData.user, userData.groups); err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
}

func (d *Database) createGroupIfNotExists(group models.Group) error {
	// Check if group exists
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM _group WHERE name = ?", group.Name).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		_, err = d.Exec(`
			INSERT INTO _group (name, description)
			VALUES (?, ?)`,
			group.Name, group.Description)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
}

func (d *Database) createUserIfNotExists(user models.User, groupNames []string) error {
	// Check if user exists
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM _user WHERE username = ?", user.Username).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		// Create user
		result, err := d.Exec(`
			INSERT INTO _user (username, email, password)
			VALUES (?, ?, ?)`,
			user.Username, user.Email, user.Password)
		if err != nil {
			LogSQLError(err)
			return err
		}

		userID, err := result.LastInsertId()
		if err != nil {
			LogSQLError(err)
			return err
		}

		// Add user to groups
		for _, groupName := range groupNames {
			if err := d.addUserToGroup(int(userID), groupName); err != nil {
				LogSQLError(err)
				return err
			}
		}
	}

	return nil
}

func (d *Database) addUserToGroup(userID int, groupName string) error {
	// Get group ID
	var groupID int
	err := d.QueryRow("SELECT id FROM _group WHERE name = ?", groupName).Scan(&groupID)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Check if user is already in group
	var count int
	err = d.QueryRow("SELECT COUNT(*) FROM _user_and_group WHERE user_id = ? AND group_id = ?", userID, groupID).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		_, err = d.Exec(`
			INSERT INTO _user_and_group (user_id, group_id)
			VALUES (?, ?)`,
			userID, groupID)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
}

// CreateUser creates a new user with a hashed password
func (d *Database) CreateUser(username, email, password string) error {
	// Hash the password before storing
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	result, err := d.Exec(`
		INSERT INTO _user (username, email, password)
		VALUES (?, ?, ?)`,
		username, email, hashedPassword)
	if err != nil {
		LogSQLError(err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		LogSQLError(err)
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	// Add user to default group (customers)
	if err := d.addUserToGroup(int(userID), "customers"); err != nil {
		LogSQLError(err)
		return fmt.Errorf("failed to add user to default group: %w", err)
	}

	return nil
}

// UpdateUserPassword updates a user's password with a new hash
func (d *Database) UpdateUserPassword(userID int, newPassword string) error {
	// Hash the new password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the password in database
	_, err = d.Exec("UPDATE _user SET password = ? WHERE id = ?", hashedPassword, userID)
	if err != nil {
		LogSQLError(err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (d *Database) AuthenticateUser(username, password string) (*models.User, error) {
	var user models.User
	err := d.QueryRow(`
		SELECT id, username, email, password, read_groups, write_groups, created, modified
		FROM _user WHERE username = ?`,
		username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.ReadGroups, &user.WriteGroups, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}

	// Check if password is in plain text (for migration)
	if !auth.IsHashFormat(user.Password) {
		// Migrate plain text password to hash
		if user.Password == password {
			// Password matches plain text, migrate to hash
			hashedPassword, err := auth.HashPassword(password)
			if err != nil {
				return nil, fmt.Errorf("failed to hash password: %w", err)
			}
			
			// Update the password in database
			_, err = d.Exec("UPDATE _user SET password = ? WHERE id = ?", hashedPassword, user.ID)
			if err != nil {
				LogSQLError(err)
				return nil, fmt.Errorf("failed to update password hash: %w", err)
			}
			
			user.Password = hashedPassword
			return &user, nil
		}
		return nil, fmt.Errorf("invalid password")
	}

	// Verify password against hash
	valid, err := auth.CheckPassword(password, user.Password)
	if err != nil {
		LogSQLError(err)
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	
	if !valid {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

func (d *Database) GetUserByID(userID int) (*models.User, error) {
	var user models.User
	err := d.QueryRow(`
		SELECT id, username, email, password, read_groups, write_groups, created, modified
		FROM _user WHERE id = ?`,
		userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.ReadGroups, &user.WriteGroups, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	return &user, nil
}

func (d *Database) GetUserGroups(userID int) ([]models.Group, error) {
	rows, err := d.Query(`
		SELECT g.id, g.name, g.description, g.read_groups, g.write_groups, g.created
		FROM _group g
		JOIN _user_and_group ug ON g.id = ug.group_id
		WHERE ug.user_id = ?
		ORDER BY g.name`,
		userID)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		err := rows.Scan(
			&group.ID, &group.Name, &group.Description, &group.ReadGroups, &group.WriteGroups, &group.CreatedAt)
		if err != nil {
			LogSQLError(err)
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (d *Database) IsUserInGroup(userID int, groupName string) (bool, error) {
	var count int
	err := d.QueryRow(`
		SELECT COUNT(*)
		FROM _user_and_group ug
		JOIN _group g ON ug.group_id = g.id
		WHERE ug.user_id = ? AND g.name = ?`,
		userID, groupName).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return false, err
	}
	return count > 0, nil
}

func (d *Database) GetAllGroups() ([]models.Group, error) {
	rows, err := d.Query(`
		SELECT id, name, description, read_groups, write_groups, created
		FROM _group
		ORDER BY name`)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		err := rows.Scan(
			&group.ID, &group.Name, &group.Description, &group.ReadGroups, &group.WriteGroups, &group.CreatedAt)
		if err != nil {
			LogSQLError(err)
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (d *Database) GetAllUsers() ([]models.User, error) {
	rows, err := d.Query(`
		SELECT id, username, email, password, read_groups, write_groups, created, modified
		FROM _user
		ORDER BY username`)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password,
			&user.ReadGroups, &user.WriteGroups, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			LogSQLError(err)
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
} 

// Metadata operations

// GetTableMetadata retrieves metadata for a specific table
func (d *Database) GetTableMetadata(tableName string) (*models.TableMetadata, error) {
	var metadata models.TableMetadata
	err := d.QueryRow(`
		SELECT id, table_name, display_name, description, read_groups, write_groups, created, modified
		FROM _table_metadata WHERE table_name = ?`,
		tableName).Scan(
		&metadata.ID, &metadata.TableName, &metadata.DisplayName, &metadata.Description,
		&metadata.ReadGroups, &metadata.WriteGroups, &metadata.CreatedAt, &metadata.UpdatedAt)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	return &metadata, nil
}

// GetAllTableMetadata retrieves metadata for all tables
func (d *Database) GetAllTableMetadata() ([]models.TableMetadata, error) {
	rows, err := d.Query(`
		SELECT id, table_name, display_name, description, read_groups, write_groups, created, modified
		FROM _table_metadata ORDER BY table_name`)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	defer rows.Close()

	var metadata []models.TableMetadata
	for rows.Next() {
		var item models.TableMetadata
		err := rows.Scan(
			&item.ID, &item.TableName, &item.DisplayName, &item.Description,
			&item.ReadGroups, &item.WriteGroups, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			LogSQLError(err)
			return nil, err
		}
		metadata = append(metadata, item)
	}
	return metadata, nil
}

// CreateTableMetadata creates new table metadata
func (d *Database) CreateTableMetadata(metadata *models.TableMetadata) error {
	_, err := d.Exec(`
		INSERT INTO _table_metadata (table_name, display_name, description, read_groups, write_groups)
		VALUES (?, ?, ?, ?, ?)`,
		metadata.TableName, metadata.DisplayName, metadata.Description,
		metadata.ReadGroups, metadata.WriteGroups)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// UpdateTableMetadata updates existing table metadata
func (d *Database) UpdateTableMetadata(metadata *models.TableMetadata) error {
	_, err := d.Exec(`
		UPDATE _table_metadata 
		SET display_name = ?, description = ?, read_groups = ?, write_groups = ?
		WHERE table_name = ?`,
		metadata.DisplayName, metadata.Description, metadata.ReadGroups, metadata.WriteGroups, metadata.TableName)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// GetFieldMetadata retrieves metadata for fields of a specific table
func (d *Database) GetFieldMetadata(tableName string) ([]models.FieldMetadata, error) {
	rows, err := d.Query(`
		SELECT id, table_name, field_name, display_name, description, db_type, html_input_type,
		       form_position, list_position, is_required, is_read_only, default_value, validation_rules,
		       created, modified
		FROM _field_metadata WHERE table_name = ? ORDER BY form_position, field_name`,
		tableName)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	defer rows.Close()

	var metadata []models.FieldMetadata
	for rows.Next() {
		var item models.FieldMetadata
		err := rows.Scan(
			&item.ID, &item.TableName, &item.FieldName, &item.DisplayName, &item.Description,
			&item.DBType, &item.HTMLInputType, &item.FormPosition, &item.ListPosition,
			&item.IsRequired, &item.IsReadOnly, &item.DefaultValue, &item.ValidationRules,
			&item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			LogSQLError(err)
			return nil, err
		}
		metadata = append(metadata, item)
	}
	return metadata, nil
}

// GetFieldMetadataByField retrieves metadata for a specific field
func (d *Database) GetFieldMetadataByField(tableName, fieldName string) (*models.FieldMetadata, error) {
	var metadata models.FieldMetadata
	err := d.QueryRow(`
		SELECT id, table_name, field_name, display_name, description, db_type, html_input_type,
		       form_position, list_position, is_required, is_read_only, default_value, validation_rules,
		       created, modified
		FROM _field_metadata WHERE table_name = ? AND field_name = ?`,
		tableName, fieldName).Scan(
		&metadata.ID, &metadata.TableName, &metadata.FieldName, &metadata.DisplayName, &metadata.Description,
		&metadata.DBType, &metadata.HTMLInputType, &metadata.FormPosition, &metadata.ListPosition,
		&metadata.IsRequired, &metadata.IsReadOnly, &metadata.DefaultValue, &metadata.ValidationRules,
		&metadata.CreatedAt, &metadata.UpdatedAt)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	return &metadata, nil
}

// CreateFieldMetadata creates new field metadata
func (d *Database) CreateFieldMetadata(metadata *models.FieldMetadata) error {
	_, err := d.Exec(`
		INSERT INTO _field_metadata (table_name, field_name, display_name, description, db_type, html_input_type,
		                           form_position, list_position, is_required, is_read_only, default_value, validation_rules)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metadata.TableName, metadata.FieldName, metadata.DisplayName, metadata.Description,
		metadata.DBType, metadata.HTMLInputType, metadata.FormPosition, metadata.ListPosition,
		metadata.IsRequired, metadata.IsReadOnly, metadata.DefaultValue, metadata.ValidationRules)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// UpdateFieldMetadata updates existing field metadata
func (d *Database) UpdateFieldMetadata(metadata *models.FieldMetadata) error {
	_, err := d.Exec(`
		UPDATE _field_metadata 
		SET display_name = ?, description = ?, db_type = ?, html_input_type = ?,
		    form_position = ?, list_position = ?, is_required = ?, is_read_only = ?,
		    default_value = ?, validation_rules = ?
		WHERE table_name = ? AND field_name = ?`,
		metadata.DisplayName, metadata.Description, metadata.DBType, metadata.HTMLInputType,
		metadata.FormPosition, metadata.ListPosition, metadata.IsRequired, metadata.IsReadOnly,
		metadata.DefaultValue, metadata.ValidationRules, metadata.TableName, metadata.FieldName)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// GetTableRows retrieves rows from a specific table with pagination
func (d *Database) GetTableRows(tableName string, page, pageSize int) ([]models.TableRow, int, error) {
	// Get total count
	var total int
	err := d.QueryRow("SELECT COUNT(*) FROM `"+tableName+"`").Scan(&total)
	if err != nil {
		LogSQLError(err)
		return nil, 0, err
	}

	// Get rows with pagination
	offset := (page - 1) * pageSize
	rows, err := d.Query("SELECT * FROM `"+tableName+"` LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		LogSQLError(err)
		return nil, 0, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		LogSQLError(err)
		return nil, 0, err
	}

	var tableRows []models.TableRow
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		err := rows.Scan(valuePtrs...)
		if err != nil {
			LogSQLError(err)
			return nil, 0, err
		}

		// Convert to map
		rowData := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				// Convert []uint8 to string for display
				if b, ok := val.([]uint8); ok {
					rowData[col] = string(b)
				} else {
					rowData[col] = val
				}
			}
		}

		// Get ID if it exists
		id := 0
		if idVal, exists := rowData["id"]; exists {
			if idInt, ok := idVal.(int64); ok {
				id = int(idInt)
			} else if idInt, ok := idVal.(int); ok {
				id = idInt
			}
		}

		tableRows = append(tableRows, models.TableRow{
			ID:   id,
			Data: rowData,
		})
	}

	return tableRows, total, nil
}

// GetTableRow retrieves a specific row from a table
func (d *Database) GetTableRow(tableName string, id int) (*models.TableRow, error) {
	rows, err := d.Query("SELECT * FROM `"+tableName+"` WHERE id = ?", id)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("row not found")
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		LogSQLError(err)
		return nil, err
	}

	// Create a slice of interface{} to hold the values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the row
	err = rows.Scan(valuePtrs...)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}

	// Convert to map
	rowData := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if val != nil {
			// Convert []uint8 to string for display
			if b, ok := val.([]uint8); ok {
				rowData[col] = string(b)
			} else {
				rowData[col] = val
			}
		}
	}

	return &models.TableRow{
		ID:   id,
		Data: rowData,
	}, nil
}

// CreateTableRow creates a new row in a table
func (d *Database) CreateTableRow(tableName string, data map[string]interface{}) error {
	// Build dynamic INSERT query
	var columns []string
	var placeholders []string
	var values []interface{}

	for col, val := range data {
		columns = append(columns, "`"+col+"`")
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := "INSERT INTO `" + tableName + "` (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	_, err := d.Exec(query, values...)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// UpdateTableRow updates an existing row in a table
func (d *Database) UpdateTableRow(tableName string, id int, data map[string]interface{}) error {
	// Build dynamic UPDATE query
	var setClauses []string
	var values []interface{}

	for col, val := range data {
		setClauses = append(setClauses, "`"+col+"` = ?")
		values = append(values, val)
	}

	values = append(values, id)
	query := "UPDATE `" + tableName + "` SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	_, err := d.Exec(query, values...)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// DeleteTableRow deletes a row from a table
func (d *Database) DeleteTableRow(tableName string, id int) error {
	_, err := d.Exec("DELETE FROM `"+tableName+"` WHERE id = ?", id)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// GetTableSchema retrieves the schema information for a table
func (d *Database) GetTableSchema(tableName string) ([]string, error) {
	rows, err := d.Query("DESCRIBE `" + tableName + "`")
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var field, typ, null, key, defaultVal, extra string
		err := rows.Scan(&field, &typ, &null, &key, &defaultVal, &extra)
		if err != nil {
			LogSQLError(err)
			return nil, err
		}
		columns = append(columns, field)
	}

	return columns, nil
}

// initializeMetadata creates default metadata for existing tables
func (d *Database) initializeMetadata() error {
	// Initialize metadata for users table
	usersMetadata := &models.TableMetadata{
		TableName:   "_user",
		DisplayName: "Users",
		Description: "User accounts and authentication information",
		ReadGroups:  `["admin", "engineer"]`,
		WriteGroups: `["admin", "engineer"]`,
	}

	if err := d.createTableMetadataIfNotExists(usersMetadata); err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize field metadata for users table
	usersFields := []models.FieldMetadata{
		{
			TableName:     "_user",
			FieldName:     "id",
			DisplayName:   "ID",
			Description:   "Unique identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  0,
			ListPosition:  0,
			IsRequired:    true,
			IsReadOnly:    true,
		},
		{
			TableName:     "_user",
			FieldName:     "username",
			DisplayName:   "Username",
			Description:   "User login name",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  1,
			ListPosition:  1,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_user",
			FieldName:     "email",
			DisplayName:   "Email",
			Description:   "User email address",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "email",
			FormPosition:  2,
			ListPosition:  2,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_user",
			FieldName:     "password",
			DisplayName:   "Password",
			Description:   "Hashed password",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "password",
			FormPosition:  3,
			ListPosition:  -1, // Don't show in list
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_user",
			FieldName:     "created",
			DisplayName:   "Created",
			Description:   "Account creation date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  4,
			ListPosition:  3,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_user",
			FieldName:     "modified",
			DisplayName:   "Modified",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  5,
			ListPosition:  4,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_user",
			FieldName:     "read_groups",
			DisplayName:   "Read Groups",
			Description:   "JSON array of groups that can read this user",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  6,
			ListPosition:  5,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_user",
			FieldName:     "write_groups",
			DisplayName:   "Write Groups",
			Description:   "JSON array of groups that can write to this user",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  7,
			ListPosition:  6,
			IsRequired:    false,
			IsReadOnly:    false,
		},
	}

	for _, field := range usersFields {
		if err := d.createFieldMetadataIfNotExists(&field); err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Initialize metadata for pages table
	pagesMetadata := &models.TableMetadata{
		TableName:   "_page",
		DisplayName: "Pages",
		Description: "Content pages and templates",
		ReadGroups:  `["admin", "customers", "engineer", "everyone"]`,
		WriteGroups: `["admin", "engineer"]`,
	}

	if err := d.createTableMetadataIfNotExists(pagesMetadata); err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize field metadata for pages table
	pagesFields := []models.FieldMetadata{
		{
			TableName:     "_page",
			FieldName:     "id",
			DisplayName:   "ID",
			Description:   "Unique identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  0,
			ListPosition:  0,
			IsRequired:    true,
			IsReadOnly:    true,
		},
		{
			TableName:     "_page",
			FieldName:     "slug",
			DisplayName:   "Slug",
			Description:   "URL-friendly identifier",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  1,
			ListPosition:  1,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "title",
			DisplayName:   "Title",
			Description:   "Page title",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  2,
			ListPosition:  2,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "meta_description",
			DisplayName:   "Meta Description",
			Description:   "SEO meta description",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  3,
			ListPosition:  -1, // Don't show in list
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "header",
			DisplayName:   "Header",
			Description:   "Page header content",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  4,
			ListPosition:  -1,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "navigation",
			DisplayName:   "Navigation",
			Description:   "Navigation menu HTML",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  5,
			ListPosition:  -1,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "main_content",
			DisplayName:   "Main Content",
			Description:   "Main page content",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  6,
			ListPosition:  -1,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "sidebar",
			DisplayName:   "Sidebar",
			Description:   "Sidebar content",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  7,
			ListPosition:  -1,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "footer",
			DisplayName:   "Footer",
			Description:   "Footer content",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  8,
			ListPosition:  -1,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "css_class",
			DisplayName:   "CSS Class",
			Description:   "CSS class for styling",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  9,
			ListPosition:  3,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "scripts",
			DisplayName:   "Scripts",
			Description:   "JavaScript code",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  10,
			ListPosition:  -1,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "template",
			DisplayName:   "Template",
			Description:   "Template name",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  11,
			ListPosition:  4,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "created",
			DisplayName:   "Created",
			Description:   "Page creation date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  12,
			ListPosition:  5,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_page",
			FieldName:     "modified",
			DisplayName:   "Modified",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  13,
			ListPosition:  6,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_page",
			FieldName:     "read_groups",
			DisplayName:   "Read Groups",
			Description:   "JSON array of groups that can read this page",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  14,
			ListPosition:  7,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_page",
			FieldName:     "write_groups",
			DisplayName:   "Write Groups",
			Description:   "JSON array of groups that can write to this page",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  15,
			ListPosition:  8,
			IsRequired:    false,
			IsReadOnly:    false,
		},
	}

	for _, field := range pagesFields {
		if err := d.createFieldMetadataIfNotExists(&field); err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Initialize metadata for groups table
	groupsMetadata := &models.TableMetadata{
		TableName:   "_group",
		DisplayName: "Groups",
		Description: "User groups for role-based access control",
		ReadGroups:  `["admin", "engineer"]`,
		WriteGroups: `["admin", "engineer"]`,
	}

	if err := d.createTableMetadataIfNotExists(groupsMetadata); err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize field metadata for groups table
	groupsFields := []models.FieldMetadata{
		{
			TableName:     "_group",
			FieldName:     "id",
			DisplayName:   "ID",
			Description:   "Unique identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  0,
			ListPosition:  0,
			IsRequired:    true,
			IsReadOnly:    true,
		},
		{
			TableName:     "_group",
			FieldName:     "name",
			DisplayName:   "Name",
			Description:   "Group name",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  1,
			ListPosition:  1,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_group",
			FieldName:     "description",
			DisplayName:   "Description",
			Description:   "Group description",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  2,
			ListPosition:  2,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_group",
			FieldName:     "created",
			DisplayName:   "Created",
			Description:   "Group creation date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  3,
			ListPosition:  3,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_group",
			FieldName:     "modified",
			DisplayName:   "Modified",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  4,
			ListPosition:  4,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_group",
			FieldName:     "read_groups",
			DisplayName:   "Read Groups",
			Description:   "JSON array of groups that can read this group",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  5,
			ListPosition:  5,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_group",
			FieldName:     "write_groups",
			DisplayName:   "Write Groups",
			Description:   "JSON array of groups that can write to this group",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  6,
			ListPosition:  6,
			IsRequired:    false,
			IsReadOnly:    false,
		},
	}

	for _, field := range groupsFields {
		if err := d.createFieldMetadataIfNotExists(&field); err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Initialize metadata for user_and_group table
	userGroupsMetadata := &models.TableMetadata{
		TableName:   "_user_and_group",
		DisplayName: "User Groups",
		Description: "User-group relationships for role assignment",
		ReadGroups:  `["admin", "engineer"]`,
		WriteGroups: `["admin", "engineer"]`,
	}

	if err := d.createTableMetadataIfNotExists(userGroupsMetadata); err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize field metadata for user_and_group table
	userGroupsFields := []models.FieldMetadata{
		{
			TableName:     "_user_and_group",
			FieldName:     "id",
			DisplayName:   "ID",
			Description:   "Unique identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  0,
			ListPosition:  0,
			IsRequired:    true,
			IsReadOnly:    true,
		},
		{
			TableName:     "_user_and_group",
			FieldName:     "user_id",
			DisplayName:   "User ID",
			Description:   "User identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  1,
			ListPosition:  1,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_user_and_group",
			FieldName:     "group_id",
			DisplayName:   "Group ID",
			Description:   "Group identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  2,
			ListPosition:  2,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_user_and_group",
			FieldName:     "created",
			DisplayName:   "Created",
			Description:   "Relationship creation date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  3,
			ListPosition:  3,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_user_and_group",
			FieldName:     "modified",
			DisplayName:   "Modified",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  4,
			ListPosition:  4,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_user_and_group",
			FieldName:     "read_groups",
			DisplayName:   "Read Groups",
			Description:   "JSON array of groups that can read this relationship",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  5,
			ListPosition:  5,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_user_and_group",
			FieldName:     "write_groups",
			DisplayName:   "Write Groups",
			Description:   "JSON array of groups that can write to this relationship",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  6,
			ListPosition:  6,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_user_and_group",
			FieldName:     "modified",
			DisplayName:   "Modified",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  4,
			ListPosition:  4,
			IsRequired:    false,
			IsReadOnly:    true,
		},
	}

	for _, field := range userGroupsFields {
		if err := d.createFieldMetadataIfNotExists(&field); err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Initialize metadata for sessions table
	sessionsMetadata := &models.TableMetadata{
		TableName:   "_session",
		DisplayName: "Sessions",
		Description: "User session information",
		ReadGroups:  `["admin", "engineer"]`,
		WriteGroups: `["admin", "engineer"]`,
	}

	if err := d.createTableMetadataIfNotExists(sessionsMetadata); err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize field metadata for sessions table
	sessionsFields := []models.FieldMetadata{
		{
			TableName:     "_session",
			FieldName:     "id",
			DisplayName:   "ID",
			Description:   "Unique identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  0,
			ListPosition:  0,
			IsRequired:    true,
			IsReadOnly:    true,
		},
		{
			TableName:     "_session",
			FieldName:     "session_id",
			DisplayName:   "Session ID",
			Description:   "Session identifier",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  1,
			ListPosition:  1,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_session",
			FieldName:     "user_id",
			DisplayName:   "User ID",
			Description:   "User identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  2,
			ListPosition:  2,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_session",
			FieldName:     "username",
			DisplayName:   "Username",
			Description:   "Username",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  3,
			ListPosition:  3,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_session",
			FieldName:     "created",
			DisplayName:   "Created",
			Description:   "Session creation date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  4,
			ListPosition:  4,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_session",
			FieldName:     "modified",
			DisplayName:   "Modified",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  5,
			ListPosition:  5,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_session",
			FieldName:     "expires_at",
			DisplayName:   "Expires At",
			Description:   "Session expiration date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  6,
			ListPosition:  6,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_session",
			FieldName:     "is_active",
			DisplayName:   "Is Active",
			Description:   "Whether session is active",
			DBType:        "BOOLEAN",
			HTMLInputType: "checkbox",
			FormPosition:  7,
			ListPosition:  7,
			IsRequired:    false,
			IsReadOnly:    false,
		},
	}

	for _, field := range sessionsFields {
		if err := d.createFieldMetadataIfNotExists(&field); err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Initialize metadata for table_metadata table
	tableMetadataMetadata := &models.TableMetadata{
		TableName:   "_table_metadata",
		DisplayName: "Table Metadata",
		Description: "Metadata for database tables",
		ReadGroups:  `["admin", "engineer"]`,
		WriteGroups: `["admin", "engineer"]`,
	}

	if err := d.createTableMetadataIfNotExists(tableMetadataMetadata); err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize field metadata for table_metadata table
	tableMetadataFields := []models.FieldMetadata{
		{
			TableName:     "_table_metadata",
			FieldName:     "id",
			DisplayName:   "ID",
			Description:   "Unique identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  0,
			ListPosition:  0,
			IsRequired:    true,
			IsReadOnly:    true,
		},
		{
			TableName:     "_table_metadata",
			FieldName:     "table_name",
			DisplayName:   "Table Name",
			Description:   "Database table name",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  1,
			ListPosition:  1,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_table_metadata",
			FieldName:     "display_name",
			DisplayName:   "Display Name",
			Description:   "Human-readable table name",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  2,
			ListPosition:  2,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_table_metadata",
			FieldName:     "description",
			DisplayName:   "Description",
			Description:   "Table description",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  3,
			ListPosition:  3,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_table_metadata",
			FieldName:     "read_groups",
			DisplayName:   "Read Groups",
			Description:   "JSON array of groups that can read this table",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  4,
			ListPosition:  4,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_table_metadata",
			FieldName:     "write_groups",
			DisplayName:   "Write Groups",
			Description:   "JSON array of groups that can write to this table",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  5,
			ListPosition:  5,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_table_metadata",
			FieldName:     "created",
			DisplayName:   "Created",
			Description:   "Metadata creation date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  6,
			ListPosition:  6,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_table_metadata",
			FieldName:     "modified",
			DisplayName:   "Modified",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  7,
			ListPosition:  7,
			IsRequired:    false,
			IsReadOnly:    true,
		},
	}

	for _, field := range tableMetadataFields {
		if err := d.createFieldMetadataIfNotExists(&field); err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Initialize metadata for field_metadata table
	fieldMetadataMetadata := &models.TableMetadata{
		TableName:   "_field_metadata",
		DisplayName: "Field Metadata",
		Description: "Metadata for database table fields",
		ReadGroups:  `["admin", "engineer"]`,
		WriteGroups: `["admin", "engineer"]`,
	}

	if err := d.createTableMetadataIfNotExists(fieldMetadataMetadata); err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize field metadata for field_metadata table
	fieldMetadataFields := []models.FieldMetadata{
		{
			TableName:     "_field_metadata",
			FieldName:     "id",
			DisplayName:   "ID",
			Description:   "Unique identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  0,
			ListPosition:  0,
			IsRequired:    true,
			IsReadOnly:    true,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "table_name",
			DisplayName:   "Table Name",
			Description:   "Database table name",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  1,
			ListPosition:  1,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "field_name",
			DisplayName:   "Field Name",
			Description:   "Database field name",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  2,
			ListPosition:  2,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "display_name",
			DisplayName:   "Display Name",
			Description:   "Human-readable field name",
			DBType:        "VARCHAR(255)",
			HTMLInputType: "text",
			FormPosition:  3,
			ListPosition:  3,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "description",
			DisplayName:   "Description",
			Description:   "Field description",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  4,
			ListPosition:  4,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "db_type",
			DisplayName:   "DB Type",
			Description:   "Database field type",
			DBType:        "VARCHAR(100)",
			HTMLInputType: "text",
			FormPosition:  5,
			ListPosition:  5,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "html_input_type",
			DisplayName:   "HTML Input Type",
			Description:   "HTML input type for forms",
			DBType:        "VARCHAR(100)",
			HTMLInputType: "text",
			FormPosition:  6,
			ListPosition:  6,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "form_position",
			DisplayName:   "Form Position",
			Description:   "Position in edit form",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  7,
			ListPosition:  7,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "list_position",
			DisplayName:   "List Position",
			Description:   "Position in table listing",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  8,
			ListPosition:  8,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "is_required",
			DisplayName:   "Is Required",
			Description:   "Whether field is required",
			DBType:        "BOOLEAN",
			HTMLInputType: "checkbox",
			FormPosition:  9,
			ListPosition:  9,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "is_read_only",
			DisplayName:   "Is Read Only",
			Description:   "Whether field is read-only",
			DBType:        "BOOLEAN",
			HTMLInputType: "checkbox",
			FormPosition:  10,
			ListPosition:  10,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "default_value",
			DisplayName:   "Default Value",
			Description:   "Default value for the field",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  11,
			ListPosition:  11,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "validation_rules",
			DisplayName:   "Validation Rules",
			Description:   "JSON string with validation rules",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  12,
			ListPosition:  12,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "created",
			DisplayName:   "Created",
			Description:   "Metadata creation date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  13,
			ListPosition:  13,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "_field_metadata",
			FieldName:     "modified",
			DisplayName:   "Modified",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  14,
			ListPosition:  14,
			IsRequired:    false,
			IsReadOnly:    true,
		},
	}

	for _, field := range fieldMetadataFields {
		if err := d.createFieldMetadataIfNotExists(&field); err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
}

// updateExistingMetadataForEngineer updates existing metadata to include engineer group access
func (d *Database) updateExistingMetadataForEngineer() error {
	// Tables that should be accessible to engineers
	engineerTables := []string{"_user", "_group", "_user_and_group", "_session", "_table_metadata", "_field_metadata", "_page"}
	
	for _, tableName := range engineerTables {
		// Update read groups to include engineer
		if tableName == "_page" {
			// Special handling for pages table - add engineer and everyone to existing read groups
			_, err := d.Exec(`
				UPDATE _table_metadata 
				SET read_groups = '["admin", "customers", "engineer", "everyone"]'
				WHERE table_name = ? AND read_groups = '["admin", "customers"]'`,
				tableName)
			if err != nil {
				LogSQLError(err)
				return err
			}
		} else {
			// Standard handling for other tables
			_, err := d.Exec(`
				UPDATE _table_metadata 
				SET read_groups = '["admin", "engineer"]'
				WHERE table_name = ? AND read_groups = '["admin"]'`,
				tableName)
			if err != nil {
				LogSQLError(err)
				return err
			}
		}
		
		// Update write groups to include engineer
		_, err := d.Exec(`
			UPDATE _table_metadata 
			SET write_groups = '["admin", "engineer"]'
			WHERE table_name = ? AND write_groups = '["admin"]'`,
			tableName)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
}

// updateExistingMetadataForEveryone updates existing metadata to include everyone group access
func (d *Database) updateExistingMetadataForEveryone() error {
	// Get all table metadata
	tableMetadata, err := d.GetAllTableMetadata()
	if err != nil {
		LogSQLError(err)
		return err
	}

	for _, table := range tableMetadata {
		tableName := table.TableName
		
		// Update read groups to include everyone for pages table
		if tableName == "_page" {
			// Special handling for pages table - add everyone to existing read groups
			_, err := d.Exec(`
				UPDATE _table_metadata 
				SET read_groups = '["admin", "customers", "engineer", "everyone"]'
				WHERE table_name = ? AND read_groups NOT LIKE '%everyone%'`,
				tableName)
			if err != nil {
				LogSQLError(err)
				return err
			}
		}
	}

	return nil
}

// createTableMetadataIfNotExists creates table metadata if it doesn't exist
func (d *Database) createTableMetadataIfNotExists(metadata *models.TableMetadata) error {
	// Check if metadata exists
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM _table_metadata WHERE table_name = ?", metadata.TableName).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		if err := d.CreateTableMetadata(metadata); err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
}

// createFieldMetadataIfNotExists creates field metadata if it doesn't exist
func (d *Database) createFieldMetadataIfNotExists(metadata *models.FieldMetadata) error {
	// Check if metadata exists
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM _field_metadata WHERE table_name = ? AND field_name = ?", 
		metadata.TableName, metadata.FieldName).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		if err := d.CreateFieldMetadata(metadata); err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
} 

// LogSQLError logs SQL errors to logs/db_errors.log, creating the directory if needed
func LogSQLError(err error) {
	if err == nil {
		return
	}
	logDir := "logs"
	logFile := filepath.Join(logDir, "db_errors.log")
	if _, statErr := os.Stat(logDir); os.IsNotExist(statErr) {
		os.MkdirAll(logDir, 0755)
	}
	f, fileErr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		log.Printf("Failed to open log file: %v", fileErr)
		return
	}
	defer f.Close()
	logger := log.New(f, "SQL_ERROR: ", log.LstdFlags)
	logger.Println(err.Error())
} 

// Permission checking functions

// CheckUserReadPermission checks if a user has read permission for a specific row
func (d *Database) CheckUserReadPermission(userID int, readGroups sql.NullString) (bool, error) {
	groupsStr := ""
	if readGroups.Valid {
		groupsStr = readGroups.String
	}
	if groupsStr == "" {
		return true, nil // No restrictions
	}

	// Parse read groups
	var groups []string
	if err := json.Unmarshal([]byte(groupsStr), &groups); err != nil {
		return false, fmt.Errorf("failed to parse read groups: %w", err)
	}

	// Check if user is in any of the read groups
	for _, group := range groups {
		// Everyone is automatically in the 'everyone' group
		if group == "everyone" {
			return true, nil
		}
		// For authenticated users, check their groups
		if inGroup, _ := d.IsUserInGroup(userID, group); inGroup {
			return true, nil
		}
	}

	return false, nil
}

// CheckUserWritePermission checks if a user has write permission for a specific row
func (d *Database) CheckUserWritePermission(userID int, writeGroups sql.NullString) (bool, error) {
	groupsStr := ""
	if writeGroups.Valid {
		groupsStr = writeGroups.String
	}
	if groupsStr == "" {
		return true, nil // No restrictions
	}

	// Parse write groups
	var groups []string
	if err := json.Unmarshal([]byte(groupsStr), &groups); err != nil {
		return false, fmt.Errorf("failed to parse write groups: %w", err)
	}

	// Check if user is in any of the write groups
	for _, group := range groups {
		// Everyone is automatically in the 'everyone' group
		if group == "everyone" {
			return true, nil
		}
		// For authenticated users, check their groups
		if inGroup, _ := d.IsUserInGroup(userID, group); inGroup {
			return true, nil
		}
	}

	return false, nil
}

// GetPageWithPermissionCheck gets a page and checks if the user has read permission
func (d *Database) GetPageWithPermissionCheck(slug string, userID int) (*models.Page, error) {
	page, err := d.GetPage(slug)
	if err != nil {
		return nil, err
	}

	// Check read permission
	hasPermission, err := d.CheckUserReadPermission(userID, page.ReadGroups)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, fmt.Errorf("access denied")
	}

	return page, nil
}

// GetAllPagesWithPermissionCheck gets all pages that the user has read permission for
func (d *Database) GetAllPagesWithPermissionCheck(userID int) ([]models.Page, error) {
	allPages, err := d.GetAllPages()
	if err != nil {
		return nil, err
	}

	var accessiblePages []models.Page
	for _, page := range allPages {
		hasPermission, err := d.CheckUserReadPermission(userID, page.ReadGroups)
		if err != nil {
			continue // Skip pages with invalid permission data
		}
		if hasPermission {
			accessiblePages = append(accessiblePages, page)
		}
	}

	return accessiblePages, nil
}

// Migration function to add permission fields to existing tables
func (d *Database) migrateAddPermissionFields() error {
	// Helper function to check if column exists
	columnExists := func(tableName, columnName string) (bool, error) {
		var count int
		err := d.QueryRow(`
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = ? 
			AND COLUMN_NAME = ?`, tableName, columnName).Scan(&count)
		return count > 0, err
	}

	// Add permission fields to _page table if they don't exist
	exists, err := columnExists("_page", "read_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _page ADD COLUMN read_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	exists, err = columnExists("_page", "write_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _page ADD COLUMN write_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Add permission fields to _user table if they don't exist
	exists, err = columnExists("_user", "read_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _user ADD COLUMN read_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	exists, err = columnExists("_user", "write_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _user ADD COLUMN write_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Add permission fields to _group table if they don't exist
	exists, err = columnExists("_group", "read_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _group ADD COLUMN read_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	exists, err = columnExists("_group", "write_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _group ADD COLUMN write_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Add permission fields to _session table if they don't exist
	exists, err = columnExists("_session", "read_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _session ADD COLUMN read_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	exists, err = columnExists("_session", "write_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _session ADD COLUMN write_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Add permission fields to _user_and_group table if they don't exist
	exists, err = columnExists("_user_and_group", "read_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _user_and_group ADD COLUMN read_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	exists, err = columnExists("_user_and_group", "write_groups")
	if err != nil {
		LogSQLError(err)
		return err
	}
	if !exists {
		_, err = d.Exec(`ALTER TABLE _user_and_group ADD COLUMN write_groups TEXT`)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Update existing pages with default permissions
	_, err = d.Exec(`
		UPDATE _page SET 
		read_groups = '["everyone"]',
		write_groups = '["admin", "engineer"]'
		WHERE read_groups IS NULL OR read_groups = ''`)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Update existing users with default permissions
	_, err = d.Exec(`
		UPDATE _user SET 
		read_groups = '["admin", "engineer"]',
		write_groups = '["admin", "engineer"]'
		WHERE read_groups IS NULL OR read_groups = ''`)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Update existing groups with default permissions
	_, err = d.Exec(`
		UPDATE _group SET 
		read_groups = '["admin", "engineer"]',
		write_groups = '["admin", "engineer"]'
		WHERE read_groups IS NULL OR read_groups = ''`)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Update existing sessions with default permissions
	_, err = d.Exec(`
		UPDATE _session SET 
		read_groups = '["admin", "engineer"]',
		write_groups = '["admin", "engineer"]'
		WHERE read_groups IS NULL OR read_groups = ''`)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Update existing user_and_group records with default permissions
	_, err = d.Exec(`
		UPDATE _user_and_group SET 
		read_groups = '["admin", "engineer"]',
		write_groups = '["admin", "engineer"]'
		WHERE read_groups IS NULL OR read_groups = ''`)
	if err != nil {
		LogSQLError(err)
		return err
	}

	return nil
}

// DeleteTableMetadata deletes table metadata and all related field metadata
func (d *Database) DeleteTableMetadata(tableName string) error {
	// Start a transaction
	tx, err := d.Begin()
	if err != nil {
		LogSQLError(err)
		return err
	}
	defer tx.Rollback()

	// Delete all field metadata for this table
	_, err = tx.Exec("DELETE FROM _field_metadata WHERE table_name = ?", tableName)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Delete table metadata
	_, err = tx.Exec("DELETE FROM _table_metadata WHERE table_name = ?", tableName)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Drop the actual table
	_, err = tx.Exec("DROP TABLE IF EXISTS `" + tableName + "`")
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// DeleteFieldMetadata deletes metadata for a specific field
func (d *Database) DeleteFieldMetadata(tableName, fieldName string) error {
	_, err := d.Exec("DELETE FROM _field_metadata WHERE table_name = ? AND field_name = ?", tableName, fieldName)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// CreateTableWithMetadata creates a new table with metadata and field metadata
func (d *Database) CreateTableWithMetadata(tableName, displayName, description, readGroups, writeGroups string, fields []models.FieldMetadata) error {
	// Start a transaction
	tx, err := d.Begin()
	if err != nil {
		LogSQLError(err)
		return err
	}
	defer tx.Rollback()

	// Create the actual table
	createTableSQL := "CREATE TABLE `" + tableName + "` ("
	createTableSQL += "id INT AUTO_INCREMENT PRIMARY KEY, "
	
	// Add management fields
	createTableSQL += "created TIMESTAMP DEFAULT CURRENT_TIMESTAMP, "
	createTableSQL += "modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, "
	createTableSQL += "read_groups TEXT, "
	createTableSQL += "write_groups TEXT"
	
	// Add custom fields
	for _, field := range fields {
		if field.FieldName != "id" && field.FieldName != "created" && field.FieldName != "modified" && 
		   field.FieldName != "read_groups" && field.FieldName != "write_groups" {
			createTableSQL += ", `" + field.FieldName + "` " + field.DBType
			if field.IsRequired {
				createTableSQL += " NOT NULL"
			}
			if field.DefaultValue != "" {
				createTableSQL += " DEFAULT '" + field.DefaultValue + "'"
			}
		}
	}
	
	createTableSQL += ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"

	_, err = tx.Exec(createTableSQL)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create table metadata
	tableMetadata := &models.TableMetadata{
		TableName:   tableName,
		DisplayName: displayName,
		Description: description,
		ReadGroups:  readGroups,
		WriteGroups: writeGroups,
	}

	_, err = tx.Exec(`
		INSERT INTO _table_metadata (table_name, display_name, description, read_groups, write_groups)
		VALUES (?, ?, ?, ?, ?)`,
		tableMetadata.TableName, tableMetadata.DisplayName, tableMetadata.Description,
		tableMetadata.ReadGroups, tableMetadata.WriteGroups)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create management field metadata
	managementFields := []models.FieldMetadata{
		{
			TableName:     tableName,
			FieldName:     "id",
			DisplayName:   "ID",
			Description:   "Unique identifier",
			DBType:        "INT",
			HTMLInputType: "number",
			FormPosition:  0,
			ListPosition:  0,
			IsRequired:    true,
			IsReadOnly:    true,
		},
		{
			TableName:     tableName,
			FieldName:     "created",
			DisplayName:   "Created",
			Description:   "Record creation date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  -1, // Don't show in form
			ListPosition:  -1, // Don't show in list
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     tableName,
			FieldName:     "modified",
			DisplayName:   "Modified",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  -1, // Don't show in form
			ListPosition:  -1, // Don't show in list
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     tableName,
			FieldName:     "read_groups",
			DisplayName:   "Read Groups",
			Description:   "JSON array of groups that can read this record",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  -1, // Don't show in form
			ListPosition:  -1, // Don't show in list
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     tableName,
			FieldName:     "write_groups",
			DisplayName:   "Write Groups",
			Description:   "JSON array of groups that can write to this record",
			DBType:        "TEXT",
			HTMLInputType: "textarea",
			FormPosition:  -1, // Don't show in form
			ListPosition:  -1, // Don't show in list
			IsRequired:    false,
			IsReadOnly:    false,
		},
	}

	// Insert management field metadata
	for _, field := range managementFields {
		_, err = tx.Exec(`
			INSERT INTO _field_metadata (table_name, field_name, display_name, description, db_type, html_input_type,
			                           form_position, list_position, is_required, is_read_only, default_value, validation_rules)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			field.TableName, field.FieldName, field.DisplayName, field.Description,
			field.DBType, field.HTMLInputType, field.FormPosition, field.ListPosition,
			field.IsRequired, field.IsReadOnly, field.DefaultValue, field.ValidationRules)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	// Insert custom field metadata
	for _, field := range fields {
		if field.FieldName != "id" && field.FieldName != "created" && field.FieldName != "modified" && 
		   field.FieldName != "read_groups" && field.FieldName != "write_groups" {
			_, err = tx.Exec(`
				INSERT INTO _field_metadata (table_name, field_name, display_name, description, db_type, html_input_type,
				                           form_position, list_position, is_required, is_read_only, default_value, validation_rules)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				field.TableName, field.FieldName, field.DisplayName, field.Description,
				field.DBType, field.HTMLInputType, field.FormPosition, field.ListPosition,
				field.IsRequired, field.IsReadOnly, field.DefaultValue, field.ValidationRules)
			if err != nil {
				LogSQLError(err)
				return err
			}
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// Password Reset Functions

// CreatePasswordResetToken creates a new password reset token for a user
func (d *Database) CreatePasswordResetToken(userID int, email string, token string, expiresAt time.Time) error {
	_, err := d.Exec(`
		INSERT INTO _password_reset_token (user_id, token, email, expires_at)
		VALUES (?, ?, ?, ?)`,
		userID, token, email, expiresAt)
	if err != nil {
		LogSQLError(err)
		return fmt.Errorf("failed to create password reset token: %w", err)
	}
	return nil
}

// GetPasswordResetToken retrieves a password reset token by token string
func (d *Database) GetPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	err := d.QueryRow(`
		SELECT id, user_id, token, email, expires_at, used, created, modified
		FROM _password_reset_token WHERE token = ?`,
		token).Scan(
		&resetToken.ID, &resetToken.UserID, &resetToken.Token, &resetToken.Email,
		&resetToken.ExpiresAt, &resetToken.Used, &resetToken.CreatedAt, &resetToken.UpdatedAt)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	return &resetToken, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used
func (d *Database) MarkPasswordResetTokenUsed(token string) error {
	_, err := d.Exec("UPDATE _password_reset_token SET used = TRUE WHERE token = ?", token)
	if err != nil {
		LogSQLError(err)
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	return nil
}

// CleanupExpiredPasswordResetTokens removes expired password reset tokens
func (d *Database) CleanupExpiredPasswordResetTokens() error {
	_, err := d.Exec("DELETE FROM _password_reset_token WHERE expires_at < NOW()")
	if err != nil {
		LogSQLError(err)
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return nil
}

// GetUserByEmail retrieves a user by email address
func (d *Database) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := d.QueryRow(`
		SELECT id, username, email, password, read_groups, write_groups, created, modified
		FROM _user WHERE email = ?`,
		email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.ReadGroups, &user.WriteGroups, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	return &user, nil
}

// migrateUpdateDBTypes updates existing db_type values to include proper length specifications
func (d *Database) migrateUpdateDBTypes() error {
	// Update VARCHAR fields to include length specification
	_, err := d.Exec(`
		UPDATE _field_metadata 
		SET db_type = 'VARCHAR(255)' 
		WHERE db_type = 'VARCHAR'`)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Update INT fields to include length specification (optional, but consistent)
	_, err = d.Exec(`
		UPDATE _field_metadata 
		SET db_type = 'INT' 
		WHERE db_type = 'INT'`)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Update DECIMAL fields to include precision and scale
	_, err = d.Exec(`
		UPDATE _field_metadata 
		SET db_type = 'DECIMAL(10,2)' 
		WHERE db_type = 'DECIMAL'`)
	if err != nil {
		LogSQLError(err)
		return err
	}

	return nil
}
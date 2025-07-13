package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"log"
	"path/filepath"
	"stingray/auth"
	"stingray/models"
	"strings"
	"time"
)

func (d *Database) initDatabase() error {
	// Create database if it doesn't exist
	createDBQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", "stingray")
	_, err := d.db.Exec(createDBQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create pages table
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS pages (
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
		template VARCHAR(100) DEFAULT 'default'
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.db.Exec(createTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create groups table (no dependencies)
	createGroupsTableQuery := `
	CREATE TABLE IF NOT EXISTS user_groups_table (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_name (name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.db.Exec(createGroupsTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create users table (no dependencies)
	createUsersTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_username (username),
		INDEX idx_email (email)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.db.Exec(createUsersTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create user_groups table (depends on users and groups)
	createUserGroupsTableQuery := `
	CREATE TABLE IF NOT EXISTS user_groups (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		group_id INT NOT NULL,
		UNIQUE KEY unique_user_group (user_id, group_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (group_id) REFERENCES user_groups_table(id) ON DELETE CASCADE,
		INDEX idx_user_id (user_id),
		INDEX idx_group_id (group_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.db.Exec(createUserGroupsTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create sessions table (depends on users)
	createSessionsTableQuery := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INT AUTO_INCREMENT PRIMARY KEY,
		session_id VARCHAR(255) UNIQUE NOT NULL,
		user_id INT NOT NULL,
		username VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		INDEX idx_session_id (session_id),
		INDEX idx_expires_at (expires_at),
		INDEX idx_is_active (is_active),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.db.Exec(createSessionsTableQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create table_metadata table
	createTableMetadataQuery := `
	CREATE TABLE IF NOT EXISTS table_metadata (
		id INT AUTO_INCREMENT PRIMARY KEY,
		table_name VARCHAR(255) UNIQUE NOT NULL,
		display_name VARCHAR(255) NOT NULL,
		description TEXT,
		read_groups TEXT,
		write_groups TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_table_name (table_name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.db.Exec(createTableMetadataQuery)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Create field_metadata table
	createFieldMetadataQuery := `
	CREATE TABLE IF NOT EXISTS field_metadata (
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
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_table_field (table_name, field_name),
		INDEX idx_table_name (table_name),
		INDEX idx_field_name (field_name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.db.Exec(createFieldMetadataQuery)
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

	return d.initializeMetadata()
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
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "about",
			Title:          "About Sting Ray",
			MetaDescription: "Learn about the Sting Ray content management system",
			Header:         "About Sting Ray",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>About Sting Ray</h2><p>Sting Ray is a modern content management system built with Go and MySQL.</p><h3>Features:</h3><ul><li>Dynamic page serving</li><li>Template system with embedded templates</li><li>RESTful API endpoints</li><li>MySQL database backend</li><li>Responsive design</li></ul>`,
			Sidebar:        `<h3>Technology Stack</h3><ul><li>Go 1.24.4</li><li>MySQL 5.7+</li><li>HTML Templates</li><li>RESTful APIs</li></ul>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "login",
			Title:          "Login - Sting Ray",
			MetaDescription: "User login page",
			Header:         "User Login",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Login</h2><p>Please enter your credentials to access the system.</p>{{template_login_form}}`,
			Sidebar:        `<h3>Need Help?</h3><p>Contact the administrator for login credentials.</p>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "shutdown",
			Title:          "Server Shutdown - Sting Ray",
			MetaDescription: "Server shutdown confirmation",
			Header:         "Server Shutdown",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Server Shutdown</h2><p>The server is shutting down gracefully. Please wait...</p><div id="countdown">30</div><script>let count = 30; const timer = setInterval(() => { count--; document.getElementById('countdown').textContent = count; if (count <= 0) { clearInterval(timer); window.location.href = '/'; } }, 1000);</script>`,
			Sidebar:        `<h3>Shutdown Progress</h3><p>Server will be unavailable during shutdown.</p>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "demo",
			Title:          "Template Demo - Sting Ray",
			MetaDescription: "Demonstration of embedded templates",
			Header:         "Template Demo",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Embedded Templates Demo</h2><p>This page demonstrates the embedded template system.</p><h3>Login Form Embedded:</h3>{{template_login_form}}<h3>Custom Content:</h3><p>This content is rendered with the modern template system.</p>`,
			Sidebar:        `<h3>Template Features</h3><ul><li>Embedded templates</li><li>Recursive processing</li><li>Multiple template support</li><li>Responsive design</li></ul>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "orders",
			Title:          "Orders Management - Sting Ray",
			MetaDescription: "Admin orders management page",
			Header:         "Orders Management",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/page/orders">Orders</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Orders Management</h2><p>Welcome to the orders management system. This page is only accessible to administrators.</p><h3>Recent Orders:</h3><ul><li>Order #1001 - Customer A - $150.00</li><li>Order #1002 - Customer B - $75.50</li><li>Order #1003 - Customer C - $200.00</li></ul><h3>Quick Actions:</h3><ul><li><a href="#">View All Orders</a></li><li><a href="#">Create New Order</a></li><li><a href="#">Export Orders</a></li></ul>`,
			Sidebar:        `<h3>Order Statistics</h3><ul><li>Total Orders: 1,247</li><li>Pending Orders: 23</li><li>Completed Orders: 1,224</li><li>Total Revenue: $45,678.90</li></ul>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
		},
		{
			Slug:           "faq",
			Title:          "Frequently Asked Questions - Sting Ray",
			MetaDescription: "Customer FAQ page",
			Header:         "Frequently Asked Questions",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/page/faq">FAQ</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Frequently Asked Questions</h2><p>Welcome to our FAQ section. This page is only accessible to customers.</p><h3>General Questions:</h3><div class="faq-item"><h4>How do I place an order?</h4><p>You can place an order by contacting our sales team or using our online ordering system.</p></div><div class="faq-item"><h4>What are your payment terms?</h4><p>We accept payment upon delivery or within 30 days of invoice.</p></div><div class="faq-item"><h4>Do you offer shipping?</h4><p>Yes, we offer shipping to all locations within our service area.</p></div><h3>Technical Support:</h3><div class="faq-item"><h4>How do I get technical support?</h4><p>Contact our technical support team at support@company.com or call 1-800-SUPPORT.</p></div>`,
			Sidebar:        `<h3>Quick Contact</h3><ul><li>Sales: sales@company.com</li><li>Support: support@company.com</li><li>Phone: 1-800-COMPANY</li></ul><h3>Helpful Links</h3><ul><li><a href="#">Product Catalog</a></li><li><a href="#">Order Status</a></li><li><a href="#">Return Policy</a></li></ul>`,
			Footer:         "© 2024 Sting Ray CMS",
			CSSClass:       "modern",
			Scripts:        "",
			Template:       "modern",
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
	err := d.db.QueryRow("SELECT COUNT(*) FROM pages WHERE slug = ?", page.Slug).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		_, err = d.db.Exec(`
			INSERT INTO pages (slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			page.Slug, page.Title, page.MetaDescription, page.Header, page.Navigation,
			page.MainContent, page.Sidebar, page.Footer, page.CSSClass, page.Scripts, page.Template)
		if err != nil {
			LogSQLError(err)
			return err
		}
	}

	return nil
}

func (d *Database) GetPage(slug string) (*models.Page, error) {
	var page models.Page
	err := d.db.QueryRow(`
		SELECT id, slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template
		FROM pages WHERE slug = ?`, slug).Scan(
		&page.ID, &page.Slug, &page.Title, &page.MetaDescription, &page.Header,
		&page.Navigation, &page.MainContent, &page.Sidebar, &page.Footer,
		&page.CSSClass, &page.Scripts, &page.Template)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	return &page, nil
}

func (d *Database) GetAllPages() ([]models.Page, error) {
	rows, err := d.db.Query(`
		SELECT id, slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template
		FROM pages ORDER BY slug`)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	defer rows.Close()

	var pages []models.Page
	for rows.Next() {
		var page models.Page
		err := rows.Scan(
			&page.ID, &page.Slug, &page.Title, &page.MetaDescription, &page.Header,
			&page.Navigation, &page.MainContent, &page.Sidebar, &page.Footer,
			&page.CSSClass, &page.Scripts, &page.Template)
		if err != nil {
			LogSQLError(err)
			return nil, err
		}
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

	_, err = d.db.Exec(`
		INSERT INTO sessions (session_id, user_id, username, expires_at, is_active)
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
	err := d.db.QueryRow(`
		SELECT id, session_id, user_id, username, created_at, expires_at, is_active
		FROM sessions WHERE session_id = ? AND is_active = TRUE AND expires_at > NOW()`,
		sessionID).Scan(
		&session.ID, &session.SessionID, &session.UserID, &session.Username,
		&session.CreatedAt, &session.ExpiresAt, &session.IsActive)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	return &session, nil
}

func (d *Database) InvalidateSession(sessionID string) error {
	_, err := d.db.Exec(`
		UPDATE sessions SET is_active = FALSE WHERE session_id = ?`,
		sessionID)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

func (d *Database) CleanupExpiredSessions() error {
	_, err := d.db.Exec(`
		UPDATE sessions SET is_active = FALSE WHERE expires_at <= NOW()`)
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
	}

	for _, userData := range users {
		// Hash the password from environment
		var plainPassword string
		if userData.user.Username == "admin" {
			plainPassword = os.Getenv("TEST_ADMIN_PASSWORD")
		} else {
			plainPassword = os.Getenv("TEST_CUSTOMER_PASSWORD")
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
	err := d.db.QueryRow("SELECT COUNT(*) FROM user_groups_table WHERE name = ?", group.Name).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		_, err = d.db.Exec(`
			INSERT INTO user_groups_table (name, description)
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
	err := d.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", user.Username).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		// Create user
		result, err := d.db.Exec(`
			INSERT INTO users (username, email, password)
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
	err := d.db.QueryRow("SELECT id FROM user_groups_table WHERE name = ?", groupName).Scan(&groupID)
	if err != nil {
		LogSQLError(err)
		return err
	}

	// Check if user is already in group
	var count int
	err = d.db.QueryRow("SELECT COUNT(*) FROM user_groups WHERE user_id = ? AND group_id = ?", userID, groupID).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return err
	}

	if count == 0 {
		_, err = d.db.Exec(`
			INSERT INTO user_groups (user_id, group_id)
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
	result, err := d.db.Exec(`
		INSERT INTO users (username, email, password)
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
	_, err = d.db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, userID)
	if err != nil {
		LogSQLError(err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (d *Database) AuthenticateUser(username, password string) (*models.User, error) {
	var user models.User
	err := d.db.QueryRow(`
		SELECT id, username, email, password, created_at, updated_at
		FROM users WHERE username = ?`,
		username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)
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
			_, err = d.db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, user.ID)
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
	err := d.db.QueryRow(`
		SELECT id, username, email, password, created_at, updated_at
		FROM users WHERE id = ?`,
		userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		LogSQLError(err)
		return nil, err
	}
	return &user, nil
}

func (d *Database) GetUserGroups(userID int) ([]models.Group, error) {
	rows, err := d.db.Query(`
		SELECT g.id, g.name, g.description, g.created_at
		FROM user_groups_table g
		JOIN user_groups ug ON g.id = ug.group_id
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
			&group.ID, &group.Name, &group.Description, &group.CreatedAt)
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
	err := d.db.QueryRow(`
		SELECT COUNT(*)
		FROM user_groups ug
		JOIN user_groups_table g ON ug.group_id = g.id
		WHERE ug.user_id = ? AND g.name = ?`,
		userID, groupName).Scan(&count)
	if err != nil {
		LogSQLError(err)
		return false, err
	}
	return count > 0, nil
}

func (d *Database) GetAllGroups() ([]models.Group, error) {
	rows, err := d.db.Query(`
		SELECT id, name, description, created_at
		FROM user_groups_table
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
			&group.ID, &group.Name, &group.Description, &group.CreatedAt)
		if err != nil {
			LogSQLError(err)
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (d *Database) GetAllUsers() ([]models.User, error) {
	rows, err := d.db.Query(`
		SELECT id, username, email, password, created_at, updated_at
		FROM users
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
			&user.CreatedAt, &user.UpdatedAt)
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
	err := d.db.QueryRow(`
		SELECT id, table_name, display_name, description, read_groups, write_groups, created_at, updated_at
		FROM table_metadata WHERE table_name = ?`,
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
	rows, err := d.db.Query(`
		SELECT id, table_name, display_name, description, read_groups, write_groups, created_at, updated_at
		FROM table_metadata ORDER BY table_name`)
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
	_, err := d.db.Exec(`
		INSERT INTO table_metadata (table_name, display_name, description, read_groups, write_groups)
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
	_, err := d.db.Exec(`
		UPDATE table_metadata 
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
	rows, err := d.db.Query(`
		SELECT id, table_name, field_name, display_name, description, db_type, html_input_type,
		       form_position, list_position, is_required, is_read_only, default_value, validation_rules,
		       created_at, updated_at
		FROM field_metadata WHERE table_name = ? ORDER BY form_position, field_name`,
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
	err := d.db.QueryRow(`
		SELECT id, table_name, field_name, display_name, description, db_type, html_input_type,
		       form_position, list_position, is_required, is_read_only, default_value, validation_rules,
		       created_at, updated_at
		FROM field_metadata WHERE table_name = ? AND field_name = ?`,
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
	_, err := d.db.Exec(`
		INSERT INTO field_metadata (table_name, field_name, display_name, description, db_type, html_input_type,
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
	_, err := d.db.Exec(`
		UPDATE field_metadata 
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
	err := d.db.QueryRow("SELECT COUNT(*) FROM `"+tableName+"`").Scan(&total)
	if err != nil {
		LogSQLError(err)
		return nil, 0, err
	}

	// Get rows with pagination
	offset := (page - 1) * pageSize
	rows, err := d.db.Query("SELECT * FROM `"+tableName+"` LIMIT ? OFFSET ?", pageSize, offset)
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
	rows, err := d.db.Query("SELECT * FROM `"+tableName+"` WHERE id = ?", id)
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
	_, err := d.db.Exec(query, values...)
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
	_, err := d.db.Exec(query, values...)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// DeleteTableRow deletes a row from a table
func (d *Database) DeleteTableRow(tableName string, id int) error {
	_, err := d.db.Exec("DELETE FROM `"+tableName+"` WHERE id = ?", id)
	if err != nil {
		LogSQLError(err)
		return err
	}
	return nil
}

// GetTableSchema retrieves the schema information for a table
func (d *Database) GetTableSchema(tableName string) ([]string, error) {
	rows, err := d.db.Query("DESCRIBE `" + tableName + "`")
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
		TableName:   "users",
		DisplayName: "Users",
		Description: "User accounts and authentication information",
		ReadGroups:  `["admin"]`,
		WriteGroups: `["admin"]`,
	}

	if err := d.createTableMetadataIfNotExists(usersMetadata); err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize field metadata for users table
	usersFields := []models.FieldMetadata{
		{
			TableName:     "users",
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
			TableName:     "users",
			FieldName:     "username",
			DisplayName:   "Username",
			Description:   "User login name",
			DBType:        "VARCHAR",
			HTMLInputType: "text",
			FormPosition:  1,
			ListPosition:  1,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "users",
			FieldName:     "email",
			DisplayName:   "Email",
			Description:   "User email address",
			DBType:        "VARCHAR",
			HTMLInputType: "email",
			FormPosition:  2,
			ListPosition:  2,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "users",
			FieldName:     "password",
			DisplayName:   "Password",
			Description:   "Hashed password",
			DBType:        "VARCHAR",
			HTMLInputType: "password",
			FormPosition:  3,
			ListPosition:  -1, // Don't show in list
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "users",
			FieldName:     "created_at",
			DisplayName:   "Created At",
			Description:   "Account creation date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  4,
			ListPosition:  3,
			IsRequired:    false,
			IsReadOnly:    true,
		},
		{
			TableName:     "users",
			FieldName:     "updated_at",
			DisplayName:   "Updated At",
			Description:   "Last update date",
			DBType:        "TIMESTAMP",
			HTMLInputType: "datetime-local",
			FormPosition:  5,
			ListPosition:  4,
			IsRequired:    false,
			IsReadOnly:    true,
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
		TableName:   "pages",
		DisplayName: "Pages",
		Description: "Content pages and templates",
		ReadGroups:  `["admin", "customers"]`,
		WriteGroups: `["admin"]`,
	}

	if err := d.createTableMetadataIfNotExists(pagesMetadata); err != nil {
		LogSQLError(err)
		return err
	}

	// Initialize field metadata for pages table
	pagesFields := []models.FieldMetadata{
		{
			TableName:     "pages",
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
			TableName:     "pages",
			FieldName:     "slug",
			DisplayName:   "Slug",
			Description:   "URL-friendly identifier",
			DBType:        "VARCHAR",
			HTMLInputType: "text",
			FormPosition:  1,
			ListPosition:  1,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "pages",
			FieldName:     "title",
			DisplayName:   "Title",
			Description:   "Page title",
			DBType:        "VARCHAR",
			HTMLInputType: "text",
			FormPosition:  2,
			ListPosition:  2,
			IsRequired:    true,
			IsReadOnly:    false,
		},
		{
			TableName:     "pages",
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
			TableName:     "pages",
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
			TableName:     "pages",
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
			TableName:     "pages",
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
			TableName:     "pages",
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
			TableName:     "pages",
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
			TableName:     "pages",
			FieldName:     "css_class",
			DisplayName:   "CSS Class",
			Description:   "CSS class for styling",
			DBType:        "VARCHAR",
			HTMLInputType: "text",
			FormPosition:  9,
			ListPosition:  3,
			IsRequired:    false,
			IsReadOnly:    false,
		},
		{
			TableName:     "pages",
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
			TableName:     "pages",
			FieldName:     "template",
			DisplayName:   "Template",
			Description:   "Template name",
			DBType:        "VARCHAR",
			HTMLInputType: "text",
			FormPosition:  11,
			ListPosition:  4,
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

	return nil
}

// createTableMetadataIfNotExists creates table metadata if it doesn't exist
func (d *Database) createTableMetadataIfNotExists(metadata *models.TableMetadata) error {
	// Check if metadata exists
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM table_metadata WHERE table_name = ?", metadata.TableName).Scan(&count)
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
	err := d.db.QueryRow("SELECT COUNT(*) FROM field_metadata WHERE table_name = ? AND field_name = ?", 
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
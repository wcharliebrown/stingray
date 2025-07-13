package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"stingray/models"
	"time"
)

func (d *Database) initDatabase() error {
	// Create database if it doesn't exist
	createDBQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", "stingray")
	_, err := d.db.Exec(createDBQuery)
	if err != nil {
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
		return err
	}

	// Create sessions table
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
		return err
	}

	// Create groups table
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
		return err
	}

	// Create users table
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
		return err
	}

	// Create user_groups table (many-to-many relationship)
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
		return err
	}

	// Initialize with default pages and users
	if err := d.initializePages(); err != nil {
		return err
	}

	return d.initializeUsers()
}

func (d *Database) initializePages() error {
	pages := []models.Page{
		{
			Slug:           "home",
			Title:          "Welcome to Sting Ray",
			MetaDescription: "A modern content management system built with Go",
			Header:         "Welcome to Sting Ray",
			Navigation:     `<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/user/login">Login</a>`,
			MainContent:    `<h2>Welcome to Sting Ray</h2><p>This is a modern content management system built with Go and MySQL. Features include dynamic page serving, template system, and RESTful API endpoints.</p>`,
			Sidebar:        `<h3>Quick Links</h3><ul><li><a href="/page/about">About</a></li><li><a href="/page/demo">Demo</a></li><li><a href="/user/login">Login</a></li></ul>`,
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
		return err
	}

	if count == 0 {
		_, err = d.db.Exec(`
			INSERT INTO pages (slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			page.Slug, page.Title, page.MetaDescription, page.Header, page.Navigation,
			page.MainContent, page.Sidebar, page.Footer, page.CSSClass, page.Scripts, page.Template)
		return err
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
		return nil, err
	}
	return &page, nil
}

func (d *Database) GetAllPages() ([]models.Page, error) {
	rows, err := d.db.Query(`
		SELECT id, slug, title, meta_description, header, navigation, main_content, sidebar, footer, css_class, scripts, template
		FROM pages ORDER BY slug`)
	if err != nil {
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
		return nil, err
	}
	return &session, nil
}

func (d *Database) InvalidateSession(sessionID string) error {
	_, err := d.db.Exec(`
		UPDATE sessions SET is_active = FALSE WHERE session_id = ?`,
		sessionID)
	return err
}

func (d *Database) CleanupExpiredSessions() error {
	_, err := d.db.Exec(`
		UPDATE sessions SET is_active = FALSE WHERE expires_at <= NOW()`)
	return err
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
			return err
		}
	}

	// Create default users
	users := []struct {
		user   models.User
		groups []string
	}{
		{
			user: models.User{
				Username: "admin",
				Email:    "adminuser@servicecompany.net",
				Password: os.Getenv("ADMIN_PASSWORD"), // In production, this should be hashed
			},
			groups: []string{"admin"},
		},
		{
			user: models.User{
				Username: "customer",
				Email:    "customeruser@company.com",
				Password: os.Getenv("CUSTOMER_PASSWORD"), // In production, this should be hashed
			},
			groups: []string{"customers"},
		},
	}

	for _, userData := range users {
		if err := d.createUserIfNotExists(userData.user, userData.groups); err != nil {
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
		return err
	}

	if count == 0 {
		_, err = d.db.Exec(`
			INSERT INTO user_groups_table (name, description)
			VALUES (?, ?)`,
			group.Name, group.Description)
		return err
	}

	return nil
}

func (d *Database) createUserIfNotExists(user models.User, groupNames []string) error {
	// Check if user exists
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", user.Username).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Create user
		result, err := d.db.Exec(`
			INSERT INTO users (username, email, password)
			VALUES (?, ?, ?)`,
			user.Username, user.Email, user.Password)
		if err != nil {
			return err
		}

		userID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		// Add user to groups
		for _, groupName := range groupNames {
			if err := d.addUserToGroup(int(userID), groupName); err != nil {
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
		return err
	}

	// Check if user is already in group
	var count int
	err = d.db.QueryRow("SELECT COUNT(*) FROM user_groups WHERE user_id = ? AND group_id = ?", userID, groupID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = d.db.Exec(`
			INSERT INTO user_groups (user_id, group_id)
			VALUES (?, ?)`,
			userID, groupID)
		return err
	}

	return nil
}

func (d *Database) AuthenticateUser(username, password string) (*models.User, error) {
	var user models.User
	err := d.db.QueryRow(`
		SELECT id, username, email, password, created_at, updated_at
		FROM users WHERE username = ? AND password = ?`,
		username, password).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		err := rows.Scan(
			&group.ID, &group.Name, &group.Description, &group.CreatedAt)
		if err != nil {
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
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		err := rows.Scan(
			&group.ID, &group.Name, &group.Description, &group.CreatedAt)
		if err != nil {
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
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
} 
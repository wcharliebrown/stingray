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
		user_id VARCHAR(255) NOT NULL,
		username VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		INDEX idx_session_id (session_id),
		INDEX idx_expires_at (expires_at),
		INDEX idx_is_active (is_active)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err = d.db.Exec(createSessionsTableQuery)
	if err != nil {
		return err
	}

	// Initialize with default pages
	return d.initializePages()
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
func (d *Database) CreateSession(userID, username string, duration time.Duration) (*models.Session, error) {
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
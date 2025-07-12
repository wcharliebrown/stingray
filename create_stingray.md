# Comprehensive Prompt: Recreate Sting Ray Web Application

## Project Overview
Create a modern web application called "Sting Ray" built with Go that serves as a content management system with a MySQL database backend. The application should be a simple, dependency-free web platform with RESTful API endpoints and template-based HTML rendering.

## Core Architecture

### Technology Stack
- **Backend**: Go 1.24.4
- **Database**: MySQL 5.7+
- **Templates**: HTML templates with embedded template support
- **Server**: HTTP server running on port 6273 (0x6273 in hex = 'bs')
- **Dependencies**: Only `github.com/go-sql-driver/mysql v1.9.3`

### Project Structure
```
stingray/
├── go.mod
├── go.sum
├── stingray.go          # Main application entry point
├── config.go            # Database configuration
├── database.go          # Database operations and page management
├── templates/           # HTML template files
│   ├── default          # Default template (324 lines)
│   ├── simple           # Simple template (128 lines)
│   ├── modern           # Modern template (198 lines)
│   ├── modern_header    # Header component (84 lines)
│   ├── modern_footer    # Footer component (98 lines)
│   └── login_form       # Login form component (13 lines)
├── env.example          # Environment variables example
├── .gitignore           # Git ignore file
└── README.md            # Project documentation
```

## Database Design

### MySQL Database Configuration
- Database name: `stingray`
- Character set: `utf8mb4`
- Collation: `utf8mb4_unicode_ci`

### Pages Table Schema
```sql
CREATE TABLE pages (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### Initial Data
The application should automatically populate the database with these pages:
- **home**: Welcome page with modern template
- **about**: About page with feature list
- **login**: User login page with embedded form
- **shutdown**: Server shutdown confirmation page
- **demo**: Embedded templates demonstration page

## Core Features

### 1. HTTP Server with Graceful Shutdown
- Server runs on port 6273
- Graceful shutdown with 30-second timeout
- Signal handling for SIGINT and SIGTERM
- POST `/shutdown` endpoint to trigger shutdown

### 2. Page Management System
- Dynamic page serving from database
- Support for multiple HTML templates
- Embedded template system (e.g., `{{template_login_form}}`)
- JSON and HTML response formats

### 3. API Endpoints

#### Page Endpoints
- `GET /` - Home page
- `GET /page/{slug}` - Dynamic page serving
- `GET /pages?response_format=json` - List all pages
- `GET /pages` - HTML page listing

#### Template Endpoints
- `GET /templates?response_format=json` - List available templates
- `GET /template/{name}?response_format=json` - Get specific template

#### User Endpoints
- `GET /user/login` - Login page
- `POST /user/login_post` - Login form submission

### 4. Template System
The application supports multiple HTML templates with embedded template functionality:

#### Template Files Required
1. **default** (324 lines) - Full-featured template with header, navigation, main content, sidebar, footer
2. **simple** (128 lines) - Minimal template for simple pages
3. **modern** (198 lines) - Modern design with embedded header/footer
4. **modern_header** (84 lines) - Header component
5. **modern_footer** (98 lines) - Footer component
6. **login_form** (13 lines) - Login form component

#### Embedded Template System
- Support for `{{template_name}}` syntax in content
- Recursive template processing
- Graceful handling of missing templates

### 5. Configuration Management
Environment variables with defaults:
- `MYSQL_HOST` (default: localhost)
- `MYSQL_PORT` (default: 3306)
- `MYSQL_USER` (default: root)
- `MYSQL_PASSWORD` (default: password)
- `MYSQL_DATABASE` (default: stingray)

## Implementation Details

### Database Operations (database.go)
- Automatic database and table creation
- Page CRUD operations
- Template loading and processing
- Embedded template resolution
- Response format handling (HTML/JSON)

### Configuration (config.go)
- Environment variable parsing
- MySQL DSN generation
- Default value handling

### Main Application (stingray.go)
- HTTP route definitions
- Graceful shutdown implementation
- Server lifecycle management

### Template System
- File-based template loading
- HTML template parsing
- Embedded template processing
- CSS-in-HTML styling
- Responsive design support

## Styling and Design

### CSS Framework
- Custom CSS grid system (12-column)
- Responsive breakpoints (mobile, tablet, desktop)
- Modern gradient backgrounds
- Card-based layout
- Hover effects and transitions

### Design Elements
- Gradient headers (#667eea to #764ba2)
- White content cards with shadows
- Modern typography (system fonts)
- Consistent spacing and padding
- Mobile-first responsive design

## Development Requirements

### Prerequisites
- Go 1.24.4 or later
- MySQL 5.7 or later
- Git for version control

### Build and Run
```bash
go mod tidy
go run .
```

### Database Setup
- MySQL server running
- Environment variables configured (optional, defaults provided)
- Automatic database creation on first run

## Key Implementation Notes

1. **Error Handling**: Graceful error handling with appropriate HTTP status codes
2. **Security**: Input validation and SQL injection prevention
3. **Performance**: Efficient database queries and template caching
4. **Maintainability**: Clean code structure with separation of concerns
5. **Extensibility**: Easy to add new pages, templates, and endpoints

## Testing and Validation

The application should support:
- JSON API testing with `curl`
- HTML page rendering in browsers
- Template system validation
- Database connectivity testing
- Graceful shutdown testing

This comprehensive prompt provides all the necessary details to recreate the Sting Ray web application with its full functionality, including the database schema, template system, API endpoints, and modern web design. 
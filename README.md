# Stingray API

![Banana Seat](Banana_Seat.png)

Stingray is a simple, fun web API written in Go. It is designed to be open source, dependency-free, and to have zero supply-chain attack surface. Perfect for learning, hacking, or deploying as a minimal, trustworthy service.

<table>
    <tr>
        <td><img src="Gauge_in_green.png" alt="No dependencies" width="200"/></td>
        <td><img src="Gauge_in_green.png" alt="Lines of Code" width="200"/></td>
        <td><img src="Gauge_in_red.png" alt="Pages per second" width="200"/></td>
    </tr>
    <tr>
    <td>Only 1 Dependency</td>
    <td>4550 Lines of Code</td>
    <td>5000 Pages/sec</td>
    </tr>
</table>

## Features

- 🚀 **Simple**: Minimalist codebase, easy to read and extend.
- 🦀 **Go-based**: Built with Go and MySQL for reliable data storage.
- 🔒 **Secure**: Configurable database connections with environment variables.
- 🌐 **Web API**: Exposes a RESTful API for easy integration.
- 🗄️ **Database**: MySQL backend with automatic schema creation.
- 👐 **Open Source**: MIT licensed, contributions welcome!
- 🔐 **Authentication**: Session-based user authentication system.
- 👥 **Role-Based Access Control**: User groups and permission management with public access via 'everyone' group.
- 📄 **Dynamic Content**: Template-driven page rendering with embedded templates.
- 🎨 **Multiple Templates**: Support for various HTML templates (default, simple, modern).
- 🔄 **Session Management**: Automatic session cleanup and expiration handling.
- 📊 **RESTful APIs**: JSON endpoints for user and group management.
- 🛡️ **Middleware**: Authentication and authorization middleware.
- ⚙️ **Configuration**: Environment-based configuration management.
- 🧪 **Testing**: Comprehensive test suite included.
- 🔐 **Secure Password Hashing**: Argon2id password hashing with automatic migration.
- 📝 **Configurable Forms**: Dynamic form generation with field metadata and engineer mode.

## Current Status

- Stability 10/10
- Usefulness 4/10

## Getting Started

### Prerequisites
- [Go](https://golang.org/dl/) (any recent version)
- [MySQL](https://dev.mysql.com/downloads/) (5.7 or later)

### Build & Run

```bash
git clone https://github.com/yourusername/stingray.git
cd stingray
go mod tidy
go run .
```

The API will start on `http://localhost:6273` by default.

### Database Setup

The application uses MySQL for data storage. See [MYSQL_SETUP.md](MYSQL_SETUP.md) for detailed setup instructions.

**Quick Start**:
1. Install and start MySQL server
2. Set environment variables (optional, defaults provided)
3. Run the application - it will automatically create the database and tables

### Environment Variables

Configure the application using environment variables. You can either set them directly or use a `.env` file:

#### Option 1: Direct Environment Variables
```bash
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=don't forget to set this :-)
export MYSQL_DATABASE=stingray
```

#### Option 2: Using .env File (Recommended)
1. Copy the example configuration:
   ```bash
   ./setup_env.sh
   ```
2. Edit the `.env` file to set your database credentials and test passwords
3. The application will automatically load the `.env` file

#### Test Credentials
For testing purposes, you can configure test user credentials in your `.env` file:
```bash
TEST_ADMIN_USERNAME=admin
TEST_ADMIN_PASSWORD=see .env file
TEST_CUSTOMER_USERNAME=customer
TEST_CUSTOMER_PASSWORD=see .env file
TEST_WRONG_PASSWORD=see .env file
```

**Security Note**: Change the default test passwords in production environments!

## Usage

### Web Interface

Visit `http://localhost:6273` to access the web interface. The system includes:

- **Home Page**: Welcome page with dynamic navigation
- **About Page**: Information about the system
- **Login System**: User authentication with session management
- **User Profile**: Personal profile page for authenticated users
- **Role-Based Pages**: Different content for admin and customer users

### API Endpoints

#### Authentication
- `GET /user/login` - Login page
- `POST /user/login_post` - Process login
- `GET /user/logout` - Logout user
- `GET /user/profile` - User profile (requires auth)

#### Content Management
- `GET /` - Home page
- `GET /page/{slug}` - Dynamic page content
- `GET /pages` - List all pages
- `GET /templates` - List available templates
- `GET /template/{name}` - Get template content

#### RESTful APIs
- `GET /api/users` - Get all users (admin only)
- `GET /api/groups` - Get all groups (admin only)
- `GET /api/user-groups?user_id={id}` - Get user groups (admin only)
- `GET /api/current-user` - Get current user info (requires auth)

#### Engineer-mode Database Management
- `GET /metadata/tables` - List all database tables (requires auth)
- `GET /metadata/tables?engineer=true` - Engineer view showing all tables (engineer group only)
- `GET /metadata/table/{table}` - View table data with pagination (requires auth)
- `GET /metadata/edit/{table}/{id}` - Edit table row (requires auth)
- `GET /metadata/edit/{table}/new` - Create new table row (requires auth)
- `POST /metadata/edit/{table}/{id}` - Update table row (requires auth)
- `GET /metadata/delete/{table}/{id}` - Delete table row (requires auth)

#### Role-Based Access
- `GET /page/orders` - Orders management (admin only)
- `GET /page/faq` - FAQ page (customer only)

### Template System

The system supports multiple HTML templates:

- **default**: Standard template with navigation and sidebar
- **simple**: Minimal template for basic content
- **modern**: Modern design with enhanced styling
- **modern_header**: Header component template
- **modern_footer**: Footer component template
- **login_form**: Login form component
- **message**: Message display template

Templates support embedded template references using `{{template_name}}` syntax.

### Public Access Control

The system uses the special 'everyone' group to mark assets as publicly accessible:

- **Public Pages**: Pages with 'everyone' in read_groups are accessible to all users, including unauthenticated visitors
- **Public Tables**: Database tables with 'everyone' in read_groups can be viewed by all users
- **Public Forms**: Forms and data with 'everyone' permissions are accessible without authentication
- **Examples**: About page, login page, and other public content use the 'everyone' group

### User Management

The system includes a complete user management system:

- **User Authentication**: Secure login with password verification
- **Session Management**: Automatic session creation and cleanup
- **User Groups**: Role-based access control with groups
- **Default Users**: Pre-configured admin and customer accounts

### Configurable Forms System

The system provides a powerful metadata-driven form generation system:

- **Field Metadata**: Each database field has configurable metadata including:
  - Display name and description
  - HTML input type (text, email, password, textarea, select)
  - Form position and list position
  - Required/read-only flags
  - Default values and validation rules
- **Dynamic Form Generation**: Forms are automatically generated based on field metadata
- **Engineer Mode**: Toggle between user-friendly labels and technical field names
- **Admin/Engineer Toggle**: Engineers can switch between normal view and engineer view showing all database tables
- **Role-Based Permissions**: Granular access control for table operations
- **CRUD Operations**: Create, read, update, and delete operations for any table
- **Pagination**: Built-in pagination for large datasets
- **JSON API**: All form operations available via JSON API endpoints

### Database Features

- **Automatic Schema Creation**: Tables created on first run
- **Session Cleanup**: Automatic cleanup of expired sessions
- **User Management**: Complete user and group management
- **Page Storage**: Dynamic page content storage and retrieval
- **Metadata-Driven Forms**: Configurable form generation based on field metadata
- **Role-Based Access Control**: Granular permissions for table read/write operations
- **Engineer Mode**: Technical view with raw field names and database types
- **Engineer Toggle**: Engineers can view all database tables regardless of permissions

## Development

### Testing

Run the test suite:

```bash
go test ./...
```

Or use the Makefile:

```bash
make test
```

#### User System Testing
Test the user management system with secure credentials:

```bash
./test_user_system.sh
```

This script will:
- Load test credentials from environment variables or `.env` file
- Test login functionality for admin and customer users
- Verify role-based access control
- Test API endpoints with proper authentication
- Clean up after testing

**Security**: The test script now uses environment variables instead of hardcoded passwords!

### Building

Build the application:

```bash
make build
```

### Running

Run the application:

```bash
make run
```

## Why Stingray?
- **Educational**: Great for learning Go and web APIs.
- **Trustworthy**: No hidden dependencies, no risk of supply-chain attacks.
- **Fun**: Tinker, extend, and make it your own!

## Contributing

Contributions are welcome! Please open issues or pull requests.

## License

MIT License. See [LICENSE](LICENSE) for details.

## TODOs

- [x] Add database for saving user and route data (MySQL)
- [x] Allow templates to contain other templates
- [X] Add more API endpoints
- [X] Write unit tests
- [x] Add authentication/authorization
- [x] Implement session management
- [x] Add role-based access control
- [x] Create RESTful API endpoints
- [x] Add template system with embedded templates
- [X] Improve documentation
- [ ] Add an admin template and switch administration pages
- [ ] Add settings table so features can be dynamically set without restart
- [ ] add password reset
- [ ] add 2FA with 6-digit code sent to email
- [ ] make public nav menu dynamic
- [ ] add left nav for admin pages
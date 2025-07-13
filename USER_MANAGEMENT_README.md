# User and Group Management System

This document describes the comprehensive user and group management system implemented in Sting Ray CMS.

## Overview

The system provides:
- User authentication with email/password
- Role-based access control through groups
- Session management
- Protected pages for different user roles
- RESTful API endpoints for user management

## Database Schema

### Tables

#### `_user`
- `id` (INT, PRIMARY KEY, AUTO_INCREMENT)
- `username` (VARCHAR(255), UNIQUE, NOT NULL)
- `email` (VARCHAR(255), UNIQUE, NOT NULL)
- `password` (VARCHAR(255), NOT NULL)
- `created_at` (TIMESTAMP, DEFAULT CURRENT_TIMESTAMP)
- `updated_at` (TIMESTAMP, DEFAULT CURRENT_TIMESTAMP ON UPDATE)

#### `_group`
- `id` (INT, PRIMARY KEY, AUTO_INCREMENT)
- `name` (VARCHAR(255), UNIQUE, NOT NULL)
- `description` (TEXT)
- `created_at` (TIMESTAMP, DEFAULT CURRENT_TIMESTAMP)

#### `_user_and_group` (Many-to-Many Relationship)
- `id` (INT, PRIMARY KEY, AUTO_INCREMENT)
- `user_id` (INT, FOREIGN KEY REFERENCES _user(id))
- `group_id` (INT, FOREIGN KEY REFERENCES _group(id))
- UNIQUE KEY on (user_id, group_id)

#### `_session`
- `id` (INT, PRIMARY KEY, AUTO_INCREMENT)
- `session_id` (VARCHAR(255), UNIQUE, NOT NULL)
- `user_id` (INT, NOT NULL, FOREIGN KEY REFERENCES _user(id))
- `username` (VARCHAR(255), NOT NULL)
- `created_at` (TIMESTAMP, DEFAULT CURRENT_TIMESTAMP)
- `expires_at` (TIMESTAMP, NOT NULL)
- `is_active` (BOOLEAN, DEFAULT TRUE)

## Default Users

On startup, the system creates two default users:

### Admin User
- **Username**: `admin`
- **Email**: `adminuser@servicecompany.net`
- **Password**: `see .env file`
- **Groups**: `admin`

### Customer User
- **Username**: `customer`
- **Email**: `customeruser@company.com`
- **Password**: `see .env file`
- **Groups**: `customers`

### Engineer User
- **Username**: `engineer`
- **Email**: `engineeruser@servicecompany.net`
- **Password**: `see .env file`
- **Groups**: `engineer`

## Default Groups

### Admin Group
- **Name**: `admin`
- **Description**: "Administrator group with full access"
- **Permissions**: Access to all pages and API endpoints

### Customers Group
- **Name**: `customers`
- **Description**: "Customer group with limited access"
- **Permissions**: Access to FAQ page and basic functionality

### Engineer Group
- **Name**: `engineer`
- **Description**: "Engineer group with technical access"
- **Permissions**: Access to database management, metadata tables, and engineer toggle functionality

### Everyone Group (Special)
- **Name**: `everyone`
- **Description**: "Special group that includes all users including unauthenticated users"
- **Purpose**: Used to mark assets as publicly accessible
- **Usage**: 
  - Public pages (about, login, home) have 'everyone' in read_groups
  - Public tables can be viewed by all users when 'everyone' is in read_groups
  - Allows unauthenticated access to specific content
  - Automatically includes all users (authenticated and unauthenticated)

## Protected Pages

### Orders Page (`/page/orders`)
- **Access**: Admin group only
- **Content**: Orders management interface with statistics and actions
- **Features**: View orders, create new orders, export data

### FAQ Page (`/page/faq`)
- **Access**: Customers group only
- **Content**: Frequently asked questions and support information
- **Features**: General questions, technical support, contact information

### Database Tables Page (`/metadata/tables`)
- **Access**: Admin and Engineer groups only
- **Content**: Database table management interface
- **Features**: 
  - View all accessible database tables
  - Engineer toggle for viewing all tables (engineer group only)
  - Edit and delete table data
  - Create new table records

### Public Pages
- **Access**: Everyone (including unauthenticated users)
- **Content**: Public pages like about, login, home
- **Features**:
  - Accessible without authentication
  - Marked with 'everyone' group in read_groups
  - Examples: About page, login page, home page

## API Endpoints

### Authentication Required
All API endpoints require authentication via session cookie.

### `/api/users` (GET)
- **Access**: Admin group only
- **Response**: List of all users (passwords excluded)
- **Example Response**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "username": "admin",
      "email": "adminuser@servicecompany.net",
      "created_at": "2024-01-01 12:00:00",
      "updated_at": "2024-01-01 12:00:00"
    }
  ]
}
```

### `/api/groups` (GET)
- **Access**: Admin group only
- **Response**: List of all groups
- **Example Response**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "admin",
      "description": "Administrator group with full access",
      "created_at": "2024-01-01 12:00:00"
    }
  ]
}
```

### `/api/user-groups` (GET)
- **Access**: Admin group only
- **Parameters**: `user_id` (required)
- **Response**: Groups for a specific user
- **Example Response**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "admin",
      "description": "Administrator group with full access",
      "created_at": "2024-01-01 12:00:00"
    }
  ]
}
```

### `/api/current-user` (GET)
- **Access**: Any authenticated user
- **Response**: Current user information with groups
- **Example Response**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "email": "adminuser@servicecompany.net",
    "created_at": "2024-01-01 12:00:00",
    "updated_at": "2024-01-01 12:00:00",
    "groups": [
      {
        "id": 1,
        "name": "admin",
        "description": "Administrator group with full access",
        "created_at": "2024-01-01 12:00:00"
      }
    ]
  }
}
```

## Middleware

### RoleMiddleware
Provides role-based access control:

- `RequireAuth()` - Ensures user is authenticated
- `RequireGroup(groupName)` - Ensures user is in specific group
- `RequireAdmin()` - Ensures user is in admin group
- `RequireCustomer()` - Ensures user is in customers group
- `RequireEngineer()` - Ensures user is in engineer group

### SessionMiddleware
Handles session management:

- `IsAuthenticated()` - Checks if user is logged in
- `GetSessionFromRequest()` - Retrieves session from request
- `SetSessionCookie()` - Sets session cookie
- `ClearSessionCookie()` - Removes session cookie

## Database Operations

### User Management
- `AuthenticateUser(username, password)` - Authenticates user credentials
- `GetUserByID(userID)` - Retrieves user by ID
- `GetAllUsers()` - Retrieves all users
- `createUserIfNotExists(user, groupNames)` - Creates user with groups

### Group Management
- `GetUserGroups(userID)` - Gets groups for a user
- `IsUserInGroup(userID, groupName)` - Checks group membership
- `GetAllGroups()` - Retrieves all groups
- `createGroupIfNotExists(group)` - Creates group if not exists

### Session Management
- `CreateSession(userID, username, duration)` - Creates new session
- `GetSession(sessionID)` - Retrieves session
- `InvalidateSession(sessionID)` - Invalidates session
- `CleanupExpiredSessions()` - Cleans up expired sessions

## Testing

### Running Tests
```bash
# Run all tests
go test ./tests/...

# Run specific test file
go test ./tests/user_management_test.go

# Run with verbose output
go test -v ./tests/...
```

### Test Script
```bash
# Run the automated test script
./test_user_system.sh
```

The test script verifies:
- User authentication
- Role-based access control
- API endpoints
- Database initialization
- Session management

## Security Considerations

### Password Storage
- **Current**: Plain text passwords (for development)
- **Production**: Should use bcrypt or similar hashing

### Session Security
- Session IDs are cryptographically random
- Sessions expire automatically
- Sessions can be invalidated

### Access Control
- Role-based access control on all protected endpoints
- Group membership verified on each request
- Proper error handling for unauthorized access

## Usage Examples

### Login
```bash
curl -X POST http://localhost:6273/user/login_post \
  -d "username=admin&password=see .env file" \
  -H "Content-Type: application/x-www-form-urlencoded"
```

### Access Protected Page
```bash
# Get session cookie first
curl -c cookies.txt http://localhost:6273/user/login

# Access admin-only page
curl -b cookies.txt http://localhost:6273/page/orders

# Access customer-only page
curl -b cookies.txt http://localhost:6273/page/faq
```

### API Usage
```bash
# Get all users (admin only)
curl -H "Cookie: session_id=YOUR_SESSION_ID" \
  http://localhost:6273/api/users

# Get current user info
curl -H "Cookie: session_id=YOUR_SESSION_ID" \
  http://localhost:6273/api/current-user
```

## Future Enhancements

1. **Password Hashing**: Implement bcrypt for secure password storage
2. **Password Reset**: Add password reset functionality
3. **User Registration**: Allow self-registration with approval
4. **Audit Logging**: Track user actions and access
5. **Advanced Permissions**: Fine-grained permission system
6. **LDAP Integration**: Support for LDAP authentication
7. **OAuth**: Support for OAuth providers
8. **Rate Limiting**: Prevent brute force attacks
9. **Two-Factor Authentication**: Add 2FA support
10. **Session Management**: Admin interface for session management

## Troubleshooting

### Common Issues

1. **Database Connection**: Ensure MySQL is running and accessible
2. **Table Creation**: Check that tables are created on startup
3. **Session Issues**: Clear browser cookies if sessions aren't working
4. **Permission Denied**: Verify user is in correct group
5. **API Errors**: Check authentication and group membership

### Debug Mode
Enable debug logging by setting environment variables:
```bash
export DEBUG=true
go run main.go
```

## Dependencies

- Go 1.24.4+
- MySQL 5.7+
- `github.com/go-sql-driver/mysql`

## License

This user management system is part of the Sting Ray CMS project. 
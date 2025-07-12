# Session Management in Sting Ray

This document describes the session management functionality implemented in Sting Ray.

## Overview

Sting Ray now includes comprehensive session management using database-backed sessions and secure cookies. Sessions are stored in a `sessions` table and managed through HTTP cookies.

## Features

### Session Management
- **Database-backed sessions**: Sessions are stored in MySQL with automatic expiration
- **Secure cookies**: HttpOnly cookies with SameSite protection
- **Session expiration**: Configurable session duration (default: 24 hours)
- **Automatic cleanup**: Expired sessions are automatically cleaned up

### Authentication Flow
1. **Login**: Users authenticate with username/password
2. **Session creation**: A new session is created in the database
3. **Cookie setting**: A secure session cookie is set in the browser
4. **Session validation**: All subsequent requests validate the session cookie
5. **Logout**: Sessions are invalidated and cookies are cleared

## Database Schema

### Sessions Table
```sql
CREATE TABLE sessions (
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
);
```

## API Endpoints

### Authentication Endpoints
- `GET /user/login` - Login page
- `POST /user/login_post` - Process login form
- `GET /user/logout` - Logout user
- `GET /user/profile` - User profile (requires authentication)

### Session Middleware
- `RequireAuth` - Redirects to login if not authenticated
- `OptionalAuth` - Adds session info to request if authenticated

## Configuration

### Session Duration
Default session duration is 24 hours. This can be modified in `handlers/session.go`:

```go
const SessionDuration = 24 * time.Hour
```

### Cookie Settings
Session cookies are configured with:
- **HttpOnly**: `true` (prevents XSS attacks)
- **Secure**: `false` (set to `true` in production with HTTPS)
- **SameSite**: `StrictMode` (CSRF protection)
- **Path**: `/` (available across the site)

## Usage Examples

### Login
```bash
curl -X POST http://localhost:6273/user/login_post \
  -d "username=admin&password=password" \
  -H "Content-Type: application/x-www-form-urlencoded"
```

### Access Protected Page
```bash
curl -b cookies.txt http://localhost:6273/user/profile
```

### Logout
```bash
curl -b cookies.txt http://localhost:6273/user/logout
```

## Security Features

### Session Security
- **Cryptographically secure session IDs**: 32-byte random values
- **Database validation**: Sessions are validated against the database
- **Automatic expiration**: Sessions expire after configured duration
- **Secure cookies**: HttpOnly and SameSite protection

### Protection Against Common Attacks
- **Session hijacking**: Secure cookie settings
- **CSRF attacks**: SameSite cookie attribute
- **XSS attacks**: HttpOnly cookies
- **Session fixation**: New session ID on login

## Testing

### Manual Testing
1. Start the server: `go run main.go`
2. Visit `http://localhost:6273/user/login`
3. Login with `admin`/`password`
4. Verify session cookie is set
5. Access protected pages
6. Test logout functionality

### Automated Testing
Run the test script:
```bash
./test_session.sh
```

### Unit Testing
Run session tests:
```bash
go test ./tests -v
```

## Maintenance

### Session Cleanup
Expired sessions are automatically cleaned up every hour. The cleanup process:
1. Marks expired sessions as inactive
2. Logs cleanup completion
3. Runs in a background goroutine

### Manual Cleanup
To manually clean up expired sessions:
```sql
UPDATE sessions SET is_active = FALSE WHERE expires_at <= NOW();
```

## Troubleshooting

### Common Issues

1. **Session not persisting**
   - Check cookie settings in browser
   - Verify database connection
   - Check session table exists

2. **Login redirects immediately**
   - Session may already exist
   - Check session validation logic

3. **Profile page not accessible**
   - Verify session cookie is set
   - Check session is active in database

### Debug Information
Session information is logged during:
- Session creation
- Session validation
- Session cleanup
- Login/logout operations

## Future Enhancements

### Planned Features
- **Session refresh**: Extend session on activity
- **Multiple sessions**: Allow multiple sessions per user
- **Session analytics**: Track session usage
- **Remember me**: Long-term sessions option
- **Session sharing**: Cross-device session management

### Security Improvements
- **HTTPS enforcement**: Require HTTPS in production
- **Session rotation**: Rotate session IDs periodically
- **Rate limiting**: Prevent session brute force
- **Audit logging**: Log all session activities 
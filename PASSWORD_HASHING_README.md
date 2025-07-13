# Argon2 Password Hashing Implementation

This document describes the Argon2 password hashing implementation in the Sting Ray CMS.

## Overview

The application now uses **Argon2id** for password hashing, which is the winner of the Password Hashing Competition (PHC) and considered the most secure password hashing algorithm available.

## Security Features

### Argon2id Algorithm
- **Memory-hard**: Resistant to GPU/ASIC attacks
- **Configurable parameters**: Memory, iterations, parallelism
- **Built-in salt generation**: Each password gets a unique salt
- **Constant-time comparison**: Prevents timing attacks

### Default Parameters
```go
Memory:      64 MB      // Memory usage
Iterations:  3          // Time cost
Parallelism: 2          // Parallel threads
SaltLength:  16 bytes   // Salt size
KeyLength:   32 bytes   // Hash output size
```

### Hash Format
Argon2 hashes are stored in the following format:
```
$argon2id$v=19$m=65536,t=3,p=2$salt$hash
```

Where:
- `argon2id`: Algorithm variant (recommended)
- `v=19`: Version
- `m=65536`: Memory in KB (64MB)
- `t=3`: Iterations
- `p=2`: Parallelism
- `salt`: Base64-encoded salt
- `hash`: Base64-encoded hash

## Implementation Details

### Files Created/Modified

#### New Files:
- `auth/password.go` - Argon2 password hashing utilities
- `tests/password_test.go` - Comprehensive password hashing tests
- `test_password_hashing.sh` - Integration test script
- `PASSWORD_HASHING_README.md` - This documentation

#### Modified Files:
- `database/operations.go` - Updated authentication and user creation
- `tests/user_management_test.go` - Updated tests for hashed passwords
- `go.mod` - Added golang.org/x/crypto dependency

### Key Functions

#### Password Hashing
```go
// Hash a password
hash, err := auth.HashPassword("mypassword")

// Hash with custom parameters
params := &auth.Argon2Params{
    Memory:      32 * 1024, // 32 MB
    Iterations:  2,
    Parallelism: 1,
}
hash, err := auth.HashPasswordWithParams("mypassword", params)
```

#### Password Verification
```go
// Verify a password against its hash
valid, err := auth.CheckPassword("mypassword", hash)
if valid {
    // Password is correct
}
```

#### Migration Support
```go
// Check if password is in hash format
if !auth.IsHashFormat(user.Password) {
    // Migrate plain text password
    hash, err := auth.MigratePlainTextPassword(plainPassword)
}
```

## Migration Strategy

### Automatic Migration
The system automatically migrates plain text passwords to hashed format:

1. **First Login**: When a user logs in with a plain text password
2. **Hash Generation**: The password is hashed using Argon2
3. **Database Update**: The hash replaces the plain text password
4. **Future Logins**: All subsequent logins use hash verification

### Migration Process
```go
func (d *Database) AuthenticateUser(username, password string) (*models.User, error) {
    // Get user from database
    user, err := d.db.GetUser(username)
    
    // Check if password is plain text
    if !auth.IsHashFormat(user.Password) {
        if user.Password == password {
            // Migrate to hash
            hash, err := auth.HashPassword(password)
            d.db.UpdatePassword(user.ID, hash)
            return user, nil
        }
        return nil, fmt.Errorf("invalid password")
    }
    
    // Verify against hash
    valid, err := auth.CheckPassword(password, user.Password)
    if !valid {
        return nil, fmt.Errorf("invalid password")
    }
    
    return user, nil
}
```

## Database Changes

### User Creation
New users are created with hashed passwords:

```go
func (d *Database) CreateUser(username, email, password string) error {
    // Hash password before storing
    hashedPassword, err := auth.HashPassword(password)
    
    // Store hashed password
    _, err = d.db.Exec(`
        INSERT INTO users (username, email, password)
        VALUES (?, ?, ?)`,
        username, email, hashedPassword)
}
```

### Password Updates
```go
func (d *Database) UpdateUserPassword(userID int, newPassword string) error {
    // Hash the new password
    hashedPassword, err := auth.HashPassword(newPassword)
    
    // Update in database
    _, err = d.db.Exec("UPDATE users SET password = ? WHERE id = ?", 
                       hashedPassword, userID)
}
```

## Testing

### Unit Tests
Run the password hashing tests:
```bash
go test ./tests/... -run "TestHash|TestCheck|TestIsHash|TestMigrate" -v
```

### Integration Tests
Run the integration test script:
```bash
./test_password_hashing.sh
```

### Test Coverage
The tests cover:
- ✅ Basic password hashing
- ✅ Password verification
- ✅ Hash uniqueness (salt generation)
- ✅ Hash format validation
- ✅ Custom parameters
- ✅ Plain text migration
- ✅ Various password types (Unicode, special chars, etc.)

## Security Best Practices

### 1. Strong Passwords
- Enforce minimum length (8+ characters)
- Require complexity (uppercase, lowercase, numbers, symbols)
- Check against common password lists

### 2. Rate Limiting
- Limit login attempts per IP
- Implement exponential backoff
- Log failed attempts

### 3. HTTPS
- Always use HTTPS in production
- Set Secure flag on cookies
- Use HSTS headers

### 4. Session Security
- Use HttpOnly cookies
- Set appropriate SameSite policy
- Implement session timeout

### 5. Password Policies
```go
func ValidatePassword(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("password too short")
    }
    
    // Add more validation rules
    return nil
}
```

## Configuration

### Environment Variables
Set strong passwords in your `.env` file:
```bash
TEST_ADMIN_PASSWORD=your_strong_admin_password_here
TEST_CUSTOMER_PASSWORD=your_strong_customer_password_here
```

### Custom Parameters
You can adjust Argon2 parameters based on your hardware:

```go
// For high-security applications
params := &auth.Argon2Params{
    Memory:      128 * 1024, // 128 MB
    Iterations:  4,
    Parallelism: 4,
}

// For resource-constrained environments
params := &auth.Argon2Params{
    Memory:      32 * 1024,  // 32 MB
    Iterations:  2,
    Parallelism: 1,
}
```

## Performance Considerations

### Hash Generation Time
- Default parameters: ~100-200ms
- Adjustable based on security requirements
- Memory usage: 64MB per hash operation

### Verification Time
- Same as generation time
- Constant-time comparison prevents timing attacks

### Scaling
- Argon2 is memory-hard, making it resistant to parallel attacks
- Parameters can be tuned for your hardware
- Consider using different parameters for different user tiers

## Troubleshooting

### Common Issues

#### 1. Hash Verification Fails
- Check if password is being hashed correctly
- Verify hash format in database
- Ensure salt is being generated properly

#### 2. Migration Not Working
- Check if `auth.IsHashFormat()` is working correctly
- Verify database update queries
- Check for transaction rollbacks

#### 3. Performance Issues
- Reduce memory/iteration parameters
- Consider using bcrypt for lower-end systems
- Monitor hash generation times

### Debug Commands
```bash
# Test password hashing directly
go run -c 'package main; import "stingray/auth"; func main() { hash, _ := auth.HashPassword("test"); println(hash) }'

# Check hash format
go run -c 'package main; import "stingray/auth"; func main() { println(auth.IsHashFormat("$argon2id$...")) }'
```

## Future Enhancements

### 1. Password Policies
- Minimum length requirements
- Complexity requirements
- Password history
- Expiration policies

### 2. Advanced Features
- Password reset functionality
- Two-factor authentication
- Account lockout policies
- Audit logging

### 3. Performance Optimizations
- Caching for frequently accessed users
- Batch password updates
- Async hash generation

## References

- [Argon2 Specification](https://password-hashing.net/submissions/specs/Argon-v3.pdf)
- [Password Hashing Competition](https://password-hashing.net/)
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [Go crypto/argon2 Documentation](https://pkg.go.dev/golang.org/x/crypto/argon2)

## Conclusion

The Argon2 password hashing implementation provides:

✅ **State-of-the-art security** with Argon2id  
✅ **Automatic migration** from plain text passwords  
✅ **Configurable parameters** for different security needs  
✅ **Comprehensive testing** with full coverage  
✅ **Production-ready** implementation  
✅ **Future-proof** design with upgrade paths  

This implementation follows security best practices and provides a solid foundation for user authentication in the Sting Ray CMS. 
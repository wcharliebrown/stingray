# Public Access Control - 'everyone' Group

## Overview

The 'everyone' group is a special system group that allows marking assets as publicly accessible. This enables unauthenticated users to access specific content while maintaining security for protected resources.

## Purpose

The 'everyone' group serves as a mechanism to:
- Mark pages as publicly accessible
- Allow unauthenticated access to specific database tables
- Provide granular control over what content is public vs. private
- Maintain security while enabling public content

## How It Works

### Automatic Inclusion
- **All Users**: Every user (authenticated and unauthenticated) is automatically considered to be in the 'everyone' group
- **No Assignment**: Users don't need to be explicitly added to the 'everyone' group
- **System Group**: The 'everyone' group is a special system group, not a regular user group

### Permission Checking
When checking access permissions, the system:
1. Checks if the user is authenticated
2. If authenticated, checks the user's actual groups
3. Always checks if 'everyone' is in the read_groups or write_groups
4. Grants access if the user is in any of the required groups OR if 'everyone' is specified

## Usage Examples

### Public Pages
```sql
-- Example: About page accessible to everyone
INSERT INTO _page (slug, title, read_groups, write_groups) 
VALUES ('about', 'About Us', '["everyone"]', '["admin", "engineer"]');
```

### Public Tables
```sql
-- Example: Public table accessible to everyone
INSERT INTO _table_metadata (table_name, display_name, read_groups, write_groups)
VALUES ('public_announcements', 'Public Announcements', '["everyone"]', '["admin"]');
```

### Mixed Access
```sql
-- Example: Page accessible to both everyone and specific groups
INSERT INTO _page (slug, title, read_groups, write_groups)
VALUES ('news', 'Company News', '["everyone", "customers", "admin"]', '["admin"]');
```

## Database Schema

### Page Permissions
```sql
CREATE TABLE _page (
    -- ... other fields ...
    read_groups TEXT,   -- JSON array: ["everyone", "admin", "customers"]
    write_groups TEXT,  -- JSON array: ["admin", "engineer"]
    -- ... other fields ...
);
```

### Table Permissions
```sql
CREATE TABLE _table_metadata (
    -- ... other fields ...
    read_groups TEXT,   -- JSON array: ["everyone", "admin", "engineer"]
    write_groups TEXT,  -- JSON array: ["admin", "engineer"]
    -- ... other fields ...
);
```

## Default Public Content

### Public Pages
The system creates several public pages by default:
- **Home Page** (`/`): Welcome page accessible to everyone
- **About Page** (`/page/about`): Information about the system
- **Login Page** (`/user/login`): Login form for authentication

### Public Tables
Some tables may be marked as public for demonstration purposes:
- **Pages Table**: Often marked as public to allow viewing page content
- **Public Announcements**: Example table that could be public

## Implementation Details

### Permission Checking Logic
```go
// Pseudo-code for permission checking
func hasAccess(userGroups []string, requiredGroups []string) bool {
    for _, group := range requiredGroups {
        if group == "everyone" {
            return true  // Everyone has access
        }
        if contains(userGroups, group) {
            return true  // User is in required group
        }
    }
    return false
}
```

### Database Operations
- **Read Groups**: Stored as JSON arrays in TEXT columns
- **Write Groups**: Stored as JSON arrays in TEXT columns
- **NULL Handling**: Proper handling of NULL values in permission fields
- **JSON Parsing**: Safe parsing of JSON permission arrays

## Security Considerations

### Public vs. Private
- **Public Content**: Only mark content as public if it's truly meant for everyone
- **Sensitive Data**: Never mark sensitive tables or pages with 'everyone' access
- **Write Permissions**: Be especially careful with write_groups - rarely should 'everyone' have write access

### Best Practices
- **Principle of Least Privilege**: Only grant the minimum necessary access
- **Regular Audits**: Periodically review what content is marked as public
- **Clear Documentation**: Document why specific content is public
- **Testing**: Test both authenticated and unauthenticated access

## API Endpoints

### Public Page Access
```bash
# Access public page (no authentication required)
curl http://localhost:6273/page/about

# Access public page with authentication (still works)
curl -H "Cookie: session_id=YOUR_SESSION" http://localhost:6273/page/about
```

### Public Table Access
```bash
# Access public table data (if table has 'everyone' in read_groups)
curl http://localhost:6273/metadata/table/public_announcements

# Access with authentication (still works)
curl -H "Cookie: session_id=YOUR_SESSION" http://localhost:6273/metadata/table/public_announcements
```

## Configuration

### Setting Public Permissions
```sql
-- Make a page public
UPDATE _page SET read_groups = '["everyone"]' WHERE slug = 'about';

-- Make a table public
UPDATE _table_metadata SET read_groups = '["everyone"]' WHERE table_name = 'public_announcements';

-- Add 'everyone' to existing permissions
UPDATE _page SET read_groups = '["everyone", "admin", "customers"]' WHERE slug = 'news';
```

### Environment Variables
No specific environment variables are needed for the 'everyone' group as it's a system group.

## Testing

### Test Public Access
```bash
# Test unauthenticated access to public page
curl http://localhost:6273/page/about

# Test authenticated access to public page
curl -H "Cookie: session_id=YOUR_SESSION" http://localhost:6273/page/about

# Test access to protected page (should fail)
curl http://localhost:6273/page/orders
```

### Test Table Access
```bash
# Test unauthenticated access to public table
curl http://localhost:6273/metadata/table/public_announcements

# Test authenticated access to protected table
curl -H "Cookie: session_id=YOUR_SESSION" http://localhost:6273/metadata/table/_user
```

## Troubleshooting

### Common Issues

#### Public Content Not Accessible
- **Issue**: Public page/table not accessible to unauthenticated users
- **Check**: Verify 'everyone' is in the read_groups field
- **Solution**: Update the read_groups to include 'everyone'

#### JSON Format Issues
- **Issue**: Permission checking fails due to malformed JSON
- **Check**: Verify read_groups/write_groups are valid JSON arrays
- **Solution**: Fix JSON format: `'["everyone", "admin"]'`

#### NULL Value Errors
- **Issue**: Database errors related to NULL permission fields
- **Check**: Look for NULL values in read_groups/write_groups columns
- **Solution**: Update NULL values to empty arrays: `'[]'`

### Debug Commands
```bash
# Check page permissions
curl http://localhost:6273/page/about?response_format=json | jq '.read_groups'

# Check table permissions
curl -H "Cookie: session_id=YOUR_SESSION" "http://localhost:6273/metadata/tables?response_format=json" | jq '.tables[] | {name: .TableName, read_groups: .ReadGroups}'
```

## Migration and Updates

### Adding Public Access
```sql
-- Add 'everyone' to existing page permissions
UPDATE _page 
SET read_groups = JSON_ARRAY_APPEND(read_groups, '$', 'everyone')
WHERE slug IN ('about', 'home', 'login');
```

### Removing Public Access
```sql
-- Remove 'everyone' from page permissions
UPDATE _page 
SET read_groups = JSON_REMOVE(read_groups, '$[0]')
WHERE JSON_CONTAINS(read_groups, '"everyone"');
```

## Future Enhancements

### Potential Improvements
- **Audit Logging**: Track when public content is accessed
- **Rate Limiting**: Limit access to public content to prevent abuse
- **Caching**: Cache public content for better performance
- **Analytics**: Track public content usage patterns

### Integration Ideas
- **CDN Integration**: Serve public content through CDN
- **Caching Headers**: Add appropriate cache headers for public content
- **Public API**: Create dedicated public API endpoints
- **Content Management**: Admin interface for managing public content 
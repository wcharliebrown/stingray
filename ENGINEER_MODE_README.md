# Engineer Toggle Functionality

## Overview

The Engineer Toggle is a specialized feature that allows users in the Engineer group to view all database tables regardless of their individual permissions. This provides engineers with full visibility into the database structure while maintaining security for other users.

## Features

### Toggle Interface
- **For Engineers**: Enabled toggle with "Admin View" and "Engineer View" options
- **For Non-Engineers**: Disabled toggle buttons with explanatory text
- **Visual Design**: Modern toggle buttons with proper styling and hover effects

### View Modes
- **Admin View (Normal Mode)**: Shows tables based on user's read permissions (including 'everyone' group access)
- **Engineer View**: Shows all tables in the database, bypassing permission filtering

### Navigation Integration
- **Engineers & Admins**: Can see "Database Tables" link in navigation and sidebar
- **Customers**: Cannot see the metadata tables link
- **Consistent**: Works across all pages (home, about, etc.)

## User Experience

### Engineer User Experience
1. **Login**: Engineer logs in with engineer credentials
2. **Navigation**: Sees "Database Tables" link in navigation
3. **Access**: Can access `/metadata/tables` page
4. **Toggle**: Sees enabled toggle with "Admin View" and "Engineer View" options
5. **Engineer Mode**: Can switch to engineer view to see all tables
6. **Visual Feedback**: Sees "Engineer Mode" notice when in engineer view

### Non-Engineer User Experience
1. **Login**: User logs in with non-engineer credentials
2. **Navigation**: Does not see "Database Tables" link (unless admin)
3. **Access**: Can access `/metadata/tables` page (if admin)
4. **Toggle**: Sees disabled toggle buttons with explanation
5. **Limited Access**: Only sees tables they have permission to access

## Technical Implementation

### Database Changes
- **Session Fix**: Fixed NULL value handling in session retrieval for read_groups/write_groups columns
- **User Groups**: Engineer group with appropriate permissions
- **Table Metadata**: All tables accessible to engineer group

### Code Changes

#### handlers/metadata.go
- Added engineer group membership checking
- Implemented toggle functionality with `engineer=true` parameter
- Enhanced HTML template with toggle interface
- Added engineer mode notice

#### handlers/pages.go
- Updated navigation logic to include engineer group
- Added engineer group membership checking for navigation links

#### database/operations.go
- Fixed GetSession function to handle NULL values properly
- Added proper NULL value handling for read_groups/write_groups

### API Endpoints

#### GET /metadata/tables
- **Normal Mode**: Returns tables filtered by user permissions
- **Engineer Mode**: Returns all tables in database
- **Response Format**: JSON with tables, is_engineer, and engineer_mode flags

#### GET /metadata/tables?engineer=true
- **Access**: Engineer group only
- **Response**: All database tables regardless of permissions
- **Visual**: Shows "Engineer Mode" notice in HTML

## Security Considerations

### Access Control
- **Engineer Group Only**: Toggle functionality restricted to engineer group
- **Permission Bypass**: Engineer mode shows all tables but doesn't grant write access
- **Public Access**: Normal mode respects 'everyone' group permissions for public tables
- **Session Validation**: Proper session handling with NULL value fixes

### User Interface
- **Clear Messaging**: Non-engineers see explanation of why toggle is disabled
- **Visual Indicators**: Engineer mode clearly indicated with notice
- **Consistent Design**: Toggle matches existing design system

## Testing

### Test Script
```bash
./test_engineer_toggle.sh
```

### Test Coverage
- Engineer login and authentication
- Toggle functionality (enabled/disabled states)
- Table access in both modes
- Navigation link visibility
- HTML interface functionality

### Expected Results
- **Engineer User**: 
  - `is_engineer: true`
  - Can access engineer mode
  - Sees all tables in engineer mode
  - Navigation shows "Database Tables" link
- **Customer User**:
  - `is_engineer: false`
  - Cannot access engineer mode
  - Sees limited tables
  - Navigation does not show "Database Tables" link
- **Admin User**:
  - `is_engineer: false`
  - Cannot access engineer mode
  - Sees all tables (due to admin permissions)
  - Navigation shows "Database Tables" link

## Usage Examples

### Engineer Accessing Tables
```bash
# Login as engineer
curl -X POST http://localhost:6273/user/login_post \
  -d "username=engineer&password=engineer" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -c cookies.txt

# Access normal view
curl -b cookies.txt "http://localhost:6273/metadata/tables?response_format=json"

# Access engineer view
curl -b cookies.txt "http://localhost:6273/metadata/tables?engineer=true&response_format=json"
```

### Customer Access (Limited)
```bash
# Login as customer
curl -X POST http://localhost:6273/user/login_post \
  -d "username=customer&password=customer" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -c cookies.txt

# Access tables (limited view)
curl -b cookies.txt "http://localhost:6273/metadata/tables?response_format=json"
```

## Configuration

### Default Users
The system creates an engineer user by default:
- **Username**: `engineer`
- **Email**: `engineeruser@servicecompany.net`
- **Password**: `engineer` (or set via `TEST_ENGINEER_PASSWORD` environment variable)
- **Group**: `engineer`

### Environment Variables
```bash
# Optional: Set engineer password
export TEST_ENGINEER_PASSWORD=your_secure_password
```

## Troubleshooting

### Common Issues

#### Session Not Recognized
- **Issue**: User shows as not authenticated
- **Solution**: Check database session table for NULL values in read_groups/write_groups
- **Fix**: Restart server to apply session NULL handling fix

#### Toggle Not Working
- **Issue**: Engineer toggle not appearing or not functional
- **Solution**: Verify user is in engineer group
- **Check**: Use API to verify group membership

#### Navigation Links Missing
- **Issue**: "Database Tables" link not appearing
- **Solution**: Check if user is in admin or engineer group
- **Verify**: Test with known engineer user

### Debug Commands
```bash
# Check if user is in engineer group
curl -b cookies.txt "http://localhost:6273/api/current-user" | jq '.data.groups[].name'

# Test engineer toggle functionality
curl -b cookies.txt "http://localhost:6273/metadata/tables?response_format=json" | jq '.is_engineer'
```

## Future Enhancements

### Potential Improvements
- **Audit Logging**: Track when engineer mode is used
- **Time Limits**: Temporary engineer mode access
- **Granular Permissions**: More specific engineer permissions
- **UI Enhancements**: Better visual indicators for engineer mode
- **Export Features**: Allow engineers to export table schemas

### Integration Ideas
- **Database Schema Viewer**: Show table relationships
- **Query Builder**: Visual query interface for engineers
- **Data Export**: Export table data for analysis
- **Schema Documentation**: Auto-generate database documentation 
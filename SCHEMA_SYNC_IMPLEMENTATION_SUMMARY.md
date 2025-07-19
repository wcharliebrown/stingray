# Field Metadata Schema Synchronization Implementation Summary

## Overview

This implementation adds automatic synchronization between field metadata and the actual database schema. When field metadata is created, updated, or deleted, the corresponding database schema is automatically updated to maintain consistency.

## What Was Implemented

### 1. Database Helper Functions

Added to `database/operations.go`:

- `getCurrentFieldType(tableName, fieldName string) (string, error)` - Gets the current data type of a field from the database
- `fieldExists(tableName, fieldName string) (bool, error)` - Checks if a field exists in a table
- `alterTableField(tableName, fieldName, newDBType string, isRequired bool, defaultValue string) error` - Modifies a field in the database
- `addTableField(tableName, fieldName, dbType string, isRequired bool, defaultValue string) error` - Adds a new field to a table
- `dropTableField(tableName, fieldName string) error` - Removes a field from a table

### 2. Enhanced Field Metadata Functions

Modified existing functions in `database/operations.go`:

#### `CreateFieldMetadata(metadata *models.FieldMetadata) error`
- **Before**: Only inserted metadata into `_field_metadata` table
- **After**: Also adds the field to the actual database table using `ALTER TABLE ADD COLUMN`
- **Protection**: Skips management fields (id, created, modified, read_groups, write_groups)
- **Transaction**: Wraps both operations in a database transaction

#### `UpdateFieldMetadata(metadata *models.FieldMetadata) error`
- **Before**: Only updated metadata in `_field_metadata` table
- **After**: Also updates the database schema when DBType, IsRequired, or DefaultValue changes
- **Comparison**: Compares current metadata with new metadata to detect changes
- **Protection**: Skips management fields
- **Transaction**: Wraps both operations in a database transaction

#### `DeleteFieldMetadata(tableName, fieldName string) error`
- **Before**: Only deleted metadata from `_field_metadata` table
- **After**: Also removes the field from the actual database table using `ALTER TABLE DROP COLUMN`
- **Protection**: Skips management fields
- **Transaction**: Wraps both operations in a database transaction

### 3. API Endpoints

Added to `handlers/metadata.go`:

#### `HandleFieldMetadata(w http.ResponseWriter, r *http.Request)`
- **POST /api/metadata/field** - Create new field metadata
- **PUT /api/metadata/field/{table}/{field}** - Update existing field metadata
- **DELETE /api/metadata/field/{table}/{field}** - Delete field metadata
- **Security**: Requires authentication and admin/engineer permissions
- **JSON**: Accepts and returns JSON data

### 4. Server Routes

Added to `server.go`:

```go
// Field metadata API routes
mux.HandleFunc("/api/metadata/field", loggingMW.Wrap(sessionMW.RequireAuth(server.metadataHandler.HandleFieldMetadata)))
mux.HandleFunc("/api/metadata/field/", loggingMW.Wrap(sessionMW.RequireAuth(server.metadataHandler.HandleFieldMetadata)))
```

## Key Features

### 1. Automatic Schema Updates
- When you change a field's `DBType` from `INT` to `TIMESTAMP`, the database schema is automatically updated
- When you add a new field metadata, the field is automatically added to the database table
- When you delete field metadata, the field is automatically removed from the database table

### 2. Transaction Safety
- All operations use database transactions
- If any part fails, all changes are rolled back
- Ensures data consistency between metadata and schema

### 3. Security
- Authentication required for all operations
- Admin or Engineer permissions required
- Management fields are protected from modification
- SQL injection protection through parameterized queries

### 4. Error Handling
- Comprehensive error logging
- Graceful failure handling
- Field existence validation

## Example Usage

### Changing Field Type
```bash
curl -X PUT "http://localhost:8080/api/metadata/field/_page/footer" \
  -H "Content-Type: application/json" \
  -d '{
    "table_name": "_page",
    "field_name": "footer",
    "display_name": "Footer",
    "description": "Page footer content",
    "db_type": "TIMESTAMP",
    "html_input_type": "datetime-local",
    "form_position": 5,
    "list_position": 5,
    "is_required": false,
    "is_read_only": false,
    "default_value": "",
    "validation_rules": ""
  }'
```

This automatically executes:
```sql
ALTER TABLE _page MODIFY COLUMN footer TIMESTAMP NULL
```

### Adding New Field
```bash
curl -X POST "http://localhost:8080/api/metadata/field" \
  -H "Content-Type: application/json" \
  -d '{
    "table_name": "_page",
    "field_name": "new_field",
    "display_name": "New Field",
    "description": "A new field",
    "db_type": "VARCHAR(255)",
    "html_input_type": "text",
    "form_position": 10,
    "list_position": 10,
    "is_required": false,
    "is_read_only": false,
    "default_value": "",
    "validation_rules": ""
  }'
```

This automatically executes:
```sql
ALTER TABLE _page ADD COLUMN new_field VARCHAR(255) NULL
```

### Deleting Field
```bash
curl -X DELETE "http://localhost:8080/api/metadata/field/_page/new_field"
```

This automatically executes:
```sql
ALTER TABLE _page DROP COLUMN new_field
```

## Protected Fields

The following management fields are protected from schema modifications:
- `id` - Primary key field
- `created` - Creation timestamp
- `modified` - Modification timestamp
- `read_groups` - Read permissions
- `write_groups` - Write permissions

## Database Compatibility

This implementation is designed for MySQL/MariaDB databases and uses:
- `INFORMATION_SCHEMA.COLUMNS` for field information
- `ALTER TABLE` statements for schema modifications
- Transaction support for data consistency

## Testing

Two test scripts were created:
- `test_field_metadata_schema_sync.sh` - Full API testing (requires authentication)
- `test_schema_sync_simple.sh` - Simple verification script

## Files Modified

1. `database/operations.go` - Added helper functions and enhanced existing functions
2. `handlers/metadata.go` - Added API handler for field metadata operations
3. `server.go` - Added API routes
4. `FIELD_METADATA_SCHEMA_SYNC_README.md` - Comprehensive documentation
5. `test_field_metadata_schema_sync.sh` - API test script
6. `test_schema_sync_simple.sh` - Simple verification script

## Benefits

1. **Consistency**: Metadata and database schema are always in sync
2. **Automation**: No manual ALTER TABLE commands needed
3. **Safety**: Transaction-based operations with rollback capability
4. **Security**: Protected management fields and proper authentication
5. **API-First**: RESTful API for programmatic access
6. **Error Handling**: Comprehensive error logging and validation

The implementation is now ready for use and will automatically keep field metadata and database schema synchronized. 
# Field Metadata Schema Synchronization

This document describes the automatic synchronization between field metadata and the actual database schema in the Stingray application.

## Overview

When field metadata is created, updated, or deleted, the system now automatically updates the corresponding database schema to maintain consistency between the metadata and the actual database structure.

## Features

### 1. Create Field Metadata
When creating new field metadata:
- The field is automatically added to the actual database table
- Management fields (`id`, `created`, `modified`, `read_groups`, `write_groups`) are excluded from schema changes
- The operation is wrapped in a transaction for data consistency

### 2. Update Field Metadata
When updating field metadata:
- If the `DBType`, `IsRequired`, or `DefaultValue` changes, the database schema is automatically updated
- The system compares the current metadata with the new metadata to detect changes
- Management fields are protected from schema modifications
- The operation is wrapped in a transaction for data consistency

### 3. Delete Field Metadata
When deleting field metadata:
- The corresponding field is automatically removed from the database table
- Management fields are protected from deletion
- The operation is wrapped in a transaction for data consistency

## Implementation Details

### Helper Functions

The following helper functions were added to support schema synchronization:

- `getCurrentFieldType(tableName, fieldName string) (string, error)` - Gets the current data type of a field
- `fieldExists(tableName, fieldName string) (bool, error)` - Checks if a field exists in a table
- `alterTableField(tableName, fieldName, newDBType string, isRequired bool, defaultValue string) error` - Modifies a field in the database
- `addTableField(tableName, fieldName, dbType string, isRequired bool, defaultValue string) error` - Adds a new field to a table
- `dropTableField(tableName, fieldName string) error` - Removes a field from a table

### Protected Fields

The following management fields are protected from schema modifications:
- `id` - Primary key field
- `created` - Creation timestamp
- `modified` - Modification timestamp
- `read_groups` - Read permissions
- `write_groups` - Write permissions

### Transaction Safety

All schema modification operations are wrapped in database transactions to ensure:
- Data consistency between metadata and schema
- Rollback capability if any operation fails
- Atomic operations (all changes succeed or all fail)

## Example Usage

### Creating a New Field

```go
metadata := &models.FieldMetadata{
    TableName:     "_page",
    FieldName:     "footer",
    DisplayName:   "Footer",
    Description:   "Page footer content",
    DBType:        "TEXT",
    HTMLInputType: "textarea",
    FormPosition:  5,
    ListPosition:  5,
    IsRequired:    false,
    IsReadOnly:    false,
    DefaultValue:  "",
    ValidationRules: "",
}

err := db.CreateFieldMetadata(metadata)
// This will:
// 1. Add the 'footer' column to the '_page' table
// 2. Insert the field metadata record
```

### Updating Field Type

```go
metadata := &models.FieldMetadata{
    TableName:     "_page",
    FieldName:     "footer",
    DisplayName:   "Footer",
    Description:   "Page footer content",
    DBType:        "TIMESTAMP", // Changed from TEXT to TIMESTAMP
    HTMLInputType: "datetime-local",
    FormPosition:  5,
    ListPosition:  5,
    IsRequired:    false,
    IsReadOnly:    false,
    DefaultValue:  "",
    ValidationRules: "",
}

err := db.UpdateFieldMetadata(metadata)
// This will:
// 1. Execute: ALTER TABLE _page MODIFY COLUMN footer TIMESTAMP NULL
// 2. Update the field metadata record
```

### Deleting a Field

```go
err := db.DeleteFieldMetadata("_page", "footer")
// This will:
// 1. Execute: ALTER TABLE _page DROP COLUMN footer
// 2. Delete the field metadata record
```

## Testing

Use the provided test script to verify the functionality:

```bash
./test_field_metadata_schema_sync.sh
```

This script tests:
1. Creating new field metadata and verifying schema update
2. Updating field metadata (changing DB type) and verifying schema update
3. Updating field metadata (changing to TIMESTAMP) and verifying schema update
4. Deleting field metadata and verifying schema update

## Error Handling

The system includes comprehensive error handling:
- SQL errors are logged using `LogSQLError`
- Transactions automatically rollback on errors
- Field existence is checked before operations
- Management fields are protected from modifications

## Security Considerations

- All SQL operations use parameterized queries to prevent SQL injection
- Management fields are protected from accidental modification
- Transactions ensure data consistency
- Error logging helps with debugging and monitoring

## Database Compatibility

This implementation is designed for MySQL/MariaDB databases and uses:
- `INFORMATION_SCHEMA.COLUMNS` for field information
- `ALTER TABLE` statements for schema modifications
- Transaction support for data consistency

## Future Enhancements

Potential improvements could include:
- Support for additional database types (PostgreSQL, SQLite)
- Validation of data type compatibility
- Backup creation before schema changes
- Audit logging of schema modifications
- Support for more complex field modifications (renaming, constraints) 
#!/bin/bash

# Test script to verify field metadata schema synchronization
# This script tests creating, updating, and deleting field metadata
# and ensures the database schema is updated accordingly

echo "Testing Field Metadata Schema Synchronization"
echo "=============================================="

# Test 1: Create a new field metadata and verify it's added to the database
echo ""
echo "Test 1: Creating new field metadata"
echo "-----------------------------------"

# Create a test field metadata
curl -X POST "http://localhost:8080/api/metadata/field" \
  -H "Content-Type: application/json" \
  -d '{
    "table_name": "_page",
    "field_name": "test_field",
    "display_name": "Test Field",
    "description": "A test field for schema sync testing",
    "db_type": "VARCHAR(255)",
    "html_input_type": "text",
    "form_position": 10,
    "list_position": 10,
    "is_required": false,
    "is_read_only": false,
    "default_value": "",
    "validation_rules": ""
  }'

echo ""
echo "Verifying field was added to database schema..."
# Check if the field exists in the database
mysql -u root -p -e "DESCRIBE _page;" | grep test_field

# Test 2: Update field metadata (change DB type) and verify schema is updated
echo ""
echo "Test 2: Updating field metadata (changing DB type)"
echo "--------------------------------------------------"

# Update the field to change its type from VARCHAR to INT
curl -X PUT "http://localhost:8080/api/metadata/field/_page/test_field" \
  -H "Content-Type: application/json" \
  -d '{
    "table_name": "_page",
    "field_name": "test_field",
    "display_name": "Test Field Updated",
    "description": "A test field with updated type",
    "db_type": "INT",
    "html_input_type": "number",
    "form_position": 10,
    "list_position": 10,
    "is_required": true,
    "is_read_only": false,
    "default_value": "0",
    "validation_rules": ""
  }'

echo ""
echo "Verifying field type was updated in database schema..."
# Check if the field type was updated
mysql -u root -p -e "DESCRIBE _page;" | grep test_field

# Test 3: Update field metadata (change to TIMESTAMP) and verify schema is updated
echo ""
echo "Test 3: Updating field metadata (changing to TIMESTAMP)"
echo "-------------------------------------------------------"

# Update the field to change its type to TIMESTAMP
curl -X PUT "http://localhost:8080/api/metadata/field/_page/test_field" \
  -H "Content-Type: application/json" \
  -d '{
    "table_name": "_page",
    "field_name": "test_field",
    "display_name": "Test Field Timestamp",
    "description": "A test field with timestamp type",
    "db_type": "TIMESTAMP",
    "html_input_type": "datetime-local",
    "form_position": 10,
    "list_position": 10,
    "is_required": false,
    "is_read_only": false,
    "default_value": "",
    "validation_rules": ""
  }'

echo ""
echo "Verifying field type was updated to TIMESTAMP in database schema..."
# Check if the field type was updated
mysql -u root -p -e "DESCRIBE _page;" | grep test_field

# Test 4: Delete field metadata and verify it's removed from the database
echo ""
echo "Test 4: Deleting field metadata"
echo "-------------------------------"

# Delete the field metadata
curl -X DELETE "http://localhost:8080/api/metadata/field/_page/test_field"

echo ""
echo "Verifying field was removed from database schema..."
# Check if the field was removed
mysql -u root -p -e "DESCRIBE _page;" | grep test_field || echo "Field successfully removed"

echo ""
echo "Test completed!" 
#!/bin/bash

# Test script for Engineer Toggle functionality
echo "Testing Engineer Toggle Functionality"
echo "===================================="

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
fi

# Set default test credentials if not provided in environment
export TEST_ADMIN_USERNAME=${TEST_ADMIN_USERNAME:-admin}
export TEST_ADMIN_PASSWORD=${TEST_ADMIN_PASSWORD:-admin}
export TEST_CUSTOMER_USERNAME=${TEST_CUSTOMER_USERNAME:-customer}
export TEST_CUSTOMER_PASSWORD=${TEST_CUSTOMER_PASSWORD:-customer}
export TEST_ENGINEER_USERNAME=${TEST_ENGINEER_USERNAME:-engineer}
export TEST_ENGINEER_PASSWORD=${TEST_ENGINEER_PASSWORD:-engineer}

echo "Using test credentials:"
echo "  Admin: $TEST_ADMIN_USERNAME/$TEST_ADMIN_PASSWORD"
echo "  Customer: $TEST_CUSTOMER_USERNAME/$TEST_CUSTOMER_PASSWORD"
echo "  Engineer: $TEST_ENGINEER_USERNAME/$TEST_ENGINEER_PASSWORD"
echo ""

# Start the server in background
echo "Starting server..."
go run . &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo ""
echo "Testing Engineer Toggle Functionality"
echo "-----------------------------------"

# Test engineer login and metadata access
echo "Testing engineer login..."
curl -s -X POST http://localhost:6273/user/login_post \
  -d "username=$TEST_ENGINEER_USERNAME&password=$TEST_ENGINEER_PASSWORD" \
  -H "Content-Type: application/x-www-form-urlencoded" -c engineer_cookies.txt > /dev/null

echo "Testing engineer metadata access (normal mode)..."
ENGINEER_RESPONSE=$(curl -s -b engineer_cookies.txt "http://localhost:6273/metadata/tables?response_format=json")
echo "Engineer normal mode - is_engineer: $(echo $ENGINEER_RESPONSE | jq -r '.is_engineer')"
echo "Engineer normal mode - engineer_mode: $(echo $ENGINEER_RESPONSE | jq -r '.engineer_mode')"
echo "Engineer normal mode - table count: $(echo $ENGINEER_RESPONSE | jq '.tables | length')"

echo "Testing engineer metadata access (engineer mode)..."
ENGINEER_ENGINEER_MODE_RESPONSE=$(curl -s -b engineer_cookies.txt "http://localhost:6273/metadata/tables?engineer=true&response_format=json")
echo "Engineer engineer mode - is_engineer: $(echo $ENGINEER_ENGINEER_MODE_RESPONSE | jq -r '.is_engineer')"
echo "Engineer engineer mode - engineer_mode: $(echo $ENGINEER_ENGINEER_MODE_RESPONSE | jq -r '.engineer_mode')"
echo "Engineer engineer mode - table count: $(echo $ENGINEER_ENGINEER_MODE_RESPONSE | jq '.tables | length')"

# Test customer login and metadata access
echo ""
echo "Testing customer login..."
curl -s -X POST http://localhost:6273/user/login_post \
  -d "username=$TEST_CUSTOMER_USERNAME&password=$TEST_CUSTOMER_PASSWORD" \
  -H "Content-Type: application/x-www-form-urlencoded" -c customer_cookies.txt > /dev/null

echo "Testing customer metadata access..."
CUSTOMER_RESPONSE=$(curl -s -b customer_cookies.txt "http://localhost:6273/metadata/tables?response_format=json")
echo "Customer - is_engineer: $(echo $CUSTOMER_RESPONSE | jq -r '.is_engineer')"
echo "Customer - engineer_mode: $(echo $CUSTOMER_RESPONSE | jq -r '.engineer_mode')"
echo "Customer - table count: $(echo $CUSTOMER_RESPONSE | jq '.tables | length')"

# Test admin login and metadata access
echo ""
echo "Testing admin login..."
curl -s -X POST http://localhost:6273/user/login_post \
  -d "username=$TEST_ADMIN_USERNAME&password=$TEST_ADMIN_PASSWORD" \
  -H "Content-Type: application/x-www-form-urlencoded" -c admin_cookies.txt > /dev/null

echo "Testing admin metadata access..."
ADMIN_RESPONSE=$(curl -s -b admin_cookies.txt "http://localhost:6273/metadata/tables?response_format=json")
echo "Admin - is_engineer: $(echo $ADMIN_RESPONSE | jq -r '.is_engineer')"
echo "Admin - engineer_mode: $(echo $ADMIN_RESPONSE | jq -r '.engineer_mode')"
echo "Admin - table count: $(echo $ADMIN_RESPONSE | jq '.tables | length')"

# Test navigation links
echo ""
echo "Testing navigation links..."
echo "Engineer navigation contains 'Database Tables': $(curl -s -b engineer_cookies.txt "http://localhost:6273/" | grep -q "Database Tables" && echo "YES" || echo "NO")"
echo "Customer navigation contains 'Database Tables': $(curl -s -b customer_cookies.txt "http://localhost:6273/" | grep -q "Database Tables" && echo "YES" || echo "NO")"
echo "Admin navigation contains 'Database Tables': $(curl -s -b admin_cookies.txt "http://localhost:6273/" | grep -q "Database Tables" && echo "YES" || echo "NO")"

# Test HTML toggle functionality
echo ""
echo "Testing HTML toggle functionality..."
echo "Engineer toggle is enabled: $(curl -s -b engineer_cookies.txt "http://localhost:6273/metadata/tables" | grep -q 'href="?engineer=true"' && echo "YES" || echo "NO")"
echo "Customer toggle is disabled: $(curl -s -b customer_cookies.txt "http://localhost:6273/metadata/tables" | grep -q 'disabled' && echo "YES" || echo "NO")"

echo ""
echo "Test Summary"
echo "============"
echo "Engineer toggle functionality test completed."
echo ""
echo "Expected behavior:"
echo "- Engineers should see toggle and have access to all tables"
echo "- Customers should see disabled toggle and limited table access"
echo "- Admins should see toggle and have access to all tables"
echo "- Navigation should show 'Database Tables' link for engineers and admins"

# Stop the server
echo ""
echo "Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "Test completed!" 
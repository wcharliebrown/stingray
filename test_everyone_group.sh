#!/bin/bash

# Test script for 'everyone' group functionality
echo "Testing 'everyone' group functionality"
echo "====================================="

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
fi

# Start the server in background
echo "Starting server..."
go run . &
SERVER_PID=$!

# Wait for server to start
sleep 5

echo ""
echo "Testing Public Access to Pages"
echo "-----------------------------"

# Test public access to about page
echo "Testing public access to /page/about..."
curl -s http://localhost:6273/page/about | grep -q "About Sting Ray" && echo "✓ Public access to about page works" || echo "✗ Public access to about page failed"

# Test public access to login page
echo "Testing public access to /user/login..."
curl -s http://localhost:6273/user/login | grep -q "Login" && echo "✓ Public access to login page works" || echo "✗ Public access to login page failed"

echo ""
echo "Testing Metadata Tables Access"
echo "-----------------------------"

# Test unauthenticated access to metadata tables
echo "Testing unauthenticated access to metadata tables..."
METADATA_RESPONSE=$(curl -s "http://localhost:6273/metadata/tables?response_format=json")
echo "Metadata response: $METADATA_RESPONSE"

# Test unauthenticated access to pages table data
echo "Testing unauthenticated access to pages table..."
PAGES_RESPONSE=$(curl -s "http://localhost:6273/metadata/table/_page?response_format=json")
echo "Pages table response: $PAGES_RESPONSE"

echo ""
echo "Testing Authenticated Access"
echo "---------------------------"

# Get session cookie for admin
ADMIN_COOKIE=$(curl -s -c - http://localhost:6273/user/login | grep session_id | awk '{print $7}')

# Test authenticated access to metadata tables
echo "Testing authenticated access to metadata tables..."
AUTH_METADATA_RESPONSE=$(curl -s -H "Cookie: session_id=$ADMIN_COOKIE" "http://localhost:6273/metadata/tables?response_format=json")
echo "Authenticated metadata response: $AUTH_METADATA_RESPONSE"

echo ""
echo "Testing Groups API"
echo "-----------------"

# Test groups API with admin session
echo "Testing groups API with admin session..."
GROUPS_RESPONSE=$(curl -s -H "Cookie: session_id=$ADMIN_COOKIE" "http://localhost:6273/api/groups")
echo "Groups response: $GROUPS_RESPONSE"

echo ""
echo "Test Summary"
echo "============"
echo "All tests completed. Check the output above for any failures."

# Stop the server
echo ""
echo "Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "Test completed!" 
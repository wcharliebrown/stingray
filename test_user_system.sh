#!/bin/bash

# Test script for user and group management system
echo "Testing Sting Ray User Management System"
echo "========================================"

# Start the server in background
echo "Starting server..."
go run main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo ""
echo "Testing Login System"
echo "-------------------"

# Test admin login
echo "Testing admin login..."
curl -s -X POST http://localhost:6273/user/login_post \
  -d "username=admin&password=admin123" \
  -H "Content-Type: application/x-www-form-urlencoded" | grep -q "Login Successful" && echo "✓ Admin login successful" || echo "✗ Admin login failed"

# Test customer login
echo "Testing customer login..."
curl -s -X POST http://localhost:6273/user/login_post \
  -d "username=customer&password=customer123" \
  -H "Content-Type: application/x-www-form-urlencoded" | grep -q "Login Successful" && echo "✓ Customer login successful" || echo "✗ Customer login failed"

# Test failed login
echo "Testing failed login..."
curl -s -X POST http://localhost:6273/user/login_post \
  -d "username=admin&password=wrongpassword" \
  -H "Content-Type: application/x-www-form-urlencoded" | grep -q "Invalid username or password" && echo "✓ Failed login handled correctly" || echo "✗ Failed login not handled correctly"

echo ""
echo "Testing API Endpoints"
echo "-------------------"

# Get session cookie for admin
ADMIN_COOKIE=$(curl -s -c - http://localhost:6273/user/login | grep session_id | awk '{print $7}')

# Test API endpoints with admin session
echo "Testing API endpoints with admin session..."

# Test get users API
echo "Testing /api/users..."
curl -s -H "Cookie: session_id=$ADMIN_COOKIE" http://localhost:6273/api/users | grep -q "success" && echo "✓ Users API working" || echo "✗ Users API failed"

# Test get groups API
echo "Testing /api/groups..."
curl -s -H "Cookie: session_id=$ADMIN_COOKIE" http://localhost:6273/api/groups | grep -q "success" && echo "✓ Groups API working" || echo "✗ Groups API failed"

# Test get current user API
echo "Testing /api/current-user..."
curl -s -H "Cookie: session_id=$ADMIN_COOKIE" http://localhost:6273/api/current-user | grep -q "success" && echo "✓ Current user API working" || echo "✗ Current user API failed"

echo ""
echo "Testing Role-Based Access Control"
echo "-------------------------------"

# Test admin access to orders page
echo "Testing admin access to orders page..."
curl -s -H "Cookie: session_id=$ADMIN_COOKIE" http://localhost:6273/page/orders | grep -q "Orders Management" && echo "✓ Admin can access orders page" || echo "✗ Admin cannot access orders page"

# Test admin access to FAQ page
echo "Testing admin access to FAQ page..."
curl -s -H "Cookie: session_id=$ADMIN_COOKIE" http://localhost:6273/page/faq | grep -q "Access Denied" && echo "✓ Admin correctly denied access to FAQ page" || echo "✗ Admin incorrectly allowed access to FAQ page"

# Get session cookie for customer
CUSTOMER_COOKIE=$(curl -s -c - http://localhost:6273/user/login | grep session_id | awk '{print $7}')

# Test customer access to FAQ page
echo "Testing customer access to FAQ page..."
curl -s -H "Cookie: session_id=$CUSTOMER_COOKIE" http://localhost:6273/page/faq | grep -q "Frequently Asked Questions" && echo "✓ Customer can access FAQ page" || echo "✗ Customer cannot access FAQ page"

# Test customer access to orders page
echo "Testing customer access to orders page..."
curl -s -H "Cookie: session_id=$CUSTOMER_COOKIE" http://localhost:6273/page/orders | grep -q "Access Denied" && echo "✓ Customer correctly denied access to orders page" || echo "✗ Customer incorrectly allowed access to orders page"

echo ""
echo "Testing Database Initialization"
echo "-----------------------------"

# Test that default users were created
echo "Testing default user creation..."
curl -s -X POST http://localhost:6273/user/login_post \
  -d "username=admin&password=admin123" \
  -H "Content-Type: application/x-www-form-urlencoded" | grep -q "adminuser@servicecompany.net" && echo "✓ Admin user created with correct email" || echo "✗ Admin user not created correctly"

curl -s -X POST http://localhost:6273/user/login_post \
  -d "username=customer&password=customer123" \
  -H "Content-Type: application/x-www-form-urlencoded" | grep -q "customeruser@company.com" && echo "✓ Customer user created with correct email" || echo "✗ Customer user not created correctly"

echo ""
echo "Test Summary"
echo "============"
echo "All tests completed. Check the output above for any failures."
echo ""
echo "Default users created:"
echo "- admin/admin123 (adminuser@servicecompany.net) - admin group"
echo "- customer/customer123 (customeruser@company.com) - customers group"
echo ""
echo "Protected pages:"
echo "- /page/orders - Admin only"
echo "- /page/faq - Customer only"
echo ""
echo "API endpoints:"
echo "- /api/users - Admin only"
echo "- /api/groups - Admin only"
echo "- /api/user-groups - Admin only"
echo "- /api/current-user - Authenticated users"

# Stop the server
echo ""
echo "Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "Test completed!" 
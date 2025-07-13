#!/bin/bash

echo "Testing Sting Ray Session Functionality"
echo "========================================"

# Test 1: Check if server is running
echo "1. Checking if server is running..."
if curl -s http://localhost:6273 > /dev/null; then
    echo "   ✓ Server is running"
else
    echo "   ✗ Server is not running. Please start the server first."
    exit 1
fi

# Test 2: Test login page
echo "2. Testing login page..."
if curl -s http://localhost:6273/user/login | grep -q "Login"; then
    echo "   ✓ Login page is accessible"
else
    echo "   ✗ Login page is not accessible"
fi

# Test 3: Test login with correct credentials
echo "3. Testing login with correct credentials..."
LOGIN_RESPONSE=$(curl -s -c cookies.txt -X POST http://localhost:6273/user/login_post \
    -d "username=${TEST_ADMIN_USERNAME:-admin}&password=${TEST_ADMIN_PASSWORD:-admin}" \
    -H "Content-Type: application/x-www-form-urlencoded")

if echo "$LOGIN_RESPONSE" | grep -q "Login Successful"; then
    echo "   ✓ Login successful"
else
    echo "   ✗ Login failed"
fi

# Test 4: Test session cookie
echo "4. Testing session cookie..."
if grep -q "stingray_session" cookies.txt; then
    echo "   ✓ Session cookie was set"
else
    echo "   ✗ Session cookie was not set"
fi

# Test 5: Test profile page (requires authentication)
echo "5. Testing profile page..."
PROFILE_RESPONSE=$(curl -s -b cookies.txt http://localhost:6273/user/profile)

if echo "$PROFILE_RESPONSE" | grep -q "User Profile"; then
    echo "   ✓ Profile page accessible with session"
else
    echo "   ✗ Profile page not accessible with session"
fi

# Test 6: Test logout
echo "6. Testing logout..."
LOGOUT_RESPONSE=$(curl -s -b cookies.txt http://localhost:6273/user/logout)

if echo "$LOGOUT_RESPONSE" | grep -q "Welcome to Sting Ray"; then
    echo "   ✓ Logout successful (redirected to home)"
else
    echo "   ✗ Logout may have failed"
fi

# Test 7: Test profile page after logout (should redirect to login)
echo "7. Testing profile page after logout..."
PROFILE_AFTER_LOGOUT=$(curl -s -b cookies.txt http://localhost:6273/user/profile)

if echo "$PROFILE_AFTER_LOGOUT" | grep -q "Login"; then
    echo "   ✓ Profile page correctly redirects to login after logout"
else
    echo "   ✗ Profile page does not redirect after logout"
fi

echo ""
echo "Session functionality test completed!"
echo "Check the output above for any failures."

# Clean up
rm -f cookies.txt 
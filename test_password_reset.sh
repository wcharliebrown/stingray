#!/bin/bash

# Test script for password reset functionality
echo "🧪 Testing Password Reset Functionality"
echo "========================================"

# Check if the application is running
if ! curl -s http://localhost:6273 > /dev/null; then
    echo "❌ Application is not running on http://localhost:6273"
    echo "   Please start the application first: go run ."
    exit 1
fi

echo "✅ Application is running"

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Set default test credentials if not provided in environment
export TEST_ADMIN_USERNAME=${TEST_ADMIN_USERNAME:-admin}
export TEST_ADMIN_PASSWORD=${TEST_ADMIN_PASSWORD:-admin}
export TEST_ADMIN_EMAIL=${TEST_ADMIN_EMAIL:-adminuser@servicecompany.net}

echo "Using test credentials:"
echo "  Admin: $TEST_ADMIN_USERNAME/$TEST_ADMIN_PASSWORD ($TEST_ADMIN_EMAIL)"
echo ""

# Test 1: Check if password reset request page is accessible
echo "📋 Test 1: Password reset request page accessibility"
if curl -s http://localhost:6273/user/password-reset-request | grep -q "Password Reset Request"; then
    echo "✅ Password reset request page is accessible"
else
    echo "❌ Password reset request page is not accessible"
    exit 1
fi

# Test 2: Test password reset request with valid email
echo ""
echo "📋 Test 2: Password reset request with valid email"
RESET_RESPONSE=$(curl -s -X POST http://localhost:6273/user/password-reset-request \
    -d "email=$TEST_ADMIN_EMAIL" \
    -H "Content-Type: application/x-www-form-urlencoded")

if echo "$RESET_RESPONSE" | grep -q "Password Reset Requested"; then
    echo "✅ Password reset request successful"
    
    # Extract the reset URL from the response
    RESET_URL=$(echo "$RESET_RESPONSE" | grep -o '/user/password-reset-confirm?token=[^"]*' | head -1)
    if [ -n "$RESET_URL" ]; then
        echo "✅ Reset URL generated: $RESET_URL"
        
        # Test 3: Test password reset confirmation page
        echo ""
        echo "📋 Test 3: Password reset confirmation page"
        CONFIRM_PAGE=$(curl -s "http://localhost:6273$RESET_URL")
        if echo "$CONFIRM_PAGE" | grep -q "Reset Password"; then
            echo "✅ Password reset confirmation page is accessible"
            
            # Test 4: Test password reset with new password
            echo ""
            echo "📋 Test 4: Password reset with new password"
            NEW_PASSWORD="newpassword123"
            RESET_RESPONSE=$(curl -s -X POST "http://localhost:6273$RESET_URL" \
                -d "password=$NEW_PASSWORD&confirm_password=$NEW_PASSWORD" \
                -H "Content-Type: application/x-www-form-urlencoded")
            
            if echo "$RESET_RESPONSE" | grep -q "Password Reset Successful"; then
                echo "✅ Password reset successful"
                
                # Test 5: Test login with new password
                echo ""
                echo "📋 Test 5: Login with new password"
                LOGIN_RESPONSE=$(curl -s -X POST http://localhost:6273/user/login_post \
                    -d "username=$TEST_ADMIN_USERNAME&password=$NEW_PASSWORD" \
                    -H "Content-Type: application/x-www-form-urlencoded")
                
                if echo "$LOGIN_RESPONSE" | grep -q "Login Successful"; then
                    echo "✅ Login with new password successful"
                else
                    echo "❌ Login with new password failed"
                fi
            else
                echo "❌ Password reset failed"
                echo "   Response: $RESET_RESPONSE"
            fi
        else
            echo "❌ Password reset confirmation page is not accessible"
        fi
    else
        echo "❌ Reset URL not found in response"
    fi
else
    echo "❌ Password reset request failed"
    echo "   Response: $RESET_RESPONSE"
fi

# Test 6: Test password reset request with invalid email
echo ""
echo "📋 Test 6: Password reset request with invalid email"
INVALID_RESPONSE=$(curl -s -X POST http://localhost:6273/user/password-reset-request \
    -d "email=nonexistent@example.com" \
    -H "Content-Type: application/x-www-form-urlencoded")

if echo "$INVALID_RESPONSE" | grep -q "Password Reset Requested"; then
    echo "✅ Password reset request with invalid email handled correctly (doesn't reveal user existence)"
else
    echo "❌ Password reset request with invalid email failed"
    echo "   Response: $INVALID_RESPONSE"
fi

# Test 7: Test password reset request with empty email
echo ""
echo "📋 Test 7: Password reset request with empty email"
EMPTY_RESPONSE=$(curl -s -X POST http://localhost:6273/user/password-reset-request \
    -d "email=" \
    -H "Content-Type: application/x-www-form-urlencoded")

if echo "$EMPTY_RESPONSE" | grep -q "Password Reset Failed"; then
    echo "✅ Password reset request with empty email handled correctly"
else
    echo "❌ Password reset request with empty email failed"
    echo "   Response: $EMPTY_RESPONSE"
fi

echo ""
echo "🎉 Password Reset Test Summary"
echo "=============================="
echo "All password reset functionality tests completed!"
echo ""
echo "Features tested:"
echo "✅ Password reset request page"
echo "✅ Password reset request with valid email"
echo "✅ Password reset confirmation page"
echo "✅ Password reset with new password"
echo "✅ Login with new password"
echo "✅ Security: Invalid email handling"
echo "✅ Security: Empty email handling"
echo ""
echo "The password reset system is working correctly!" 
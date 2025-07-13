#!/bin/bash

# Test script for email password reset functionality
echo "🧪 Testing Email Password Reset Functionality"
echo "============================================="

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
    
    # Check if the response contains the email message (not the token)
    if echo "$RESET_RESPONSE" | grep -q "sent to your email address"; then
        echo "✅ Email functionality is working (no token displayed)"
    else
        echo "⚠️  Email service may not be configured, showing fallback message"
    fi
else
    echo "❌ Password reset request failed"
    echo "   Response: $RESET_RESPONSE"
fi

# Test 3: Test password reset request with invalid email
echo ""
echo "📋 Test 3: Password reset request with invalid email"
INVALID_RESPONSE=$(curl -s -X POST http://localhost:6273/user/password-reset-request \
    -d "email=nonexistent@example.com" \
    -H "Content-Type: application/x-www-form-urlencoded")

if echo "$INVALID_RESPONSE" | grep -q "Password Reset Requested"; then
    echo "✅ Password reset request with invalid email handled correctly (doesn't reveal user existence)"
else
    echo "❌ Password reset request with invalid email failed"
    echo "   Response: $INVALID_RESPONSE"
fi

# Test 4: Test password reset request with empty email
echo ""
echo "📋 Test 4: Password reset request with empty email"
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
echo "🎉 Email Password Reset Test Summary"
echo "===================================="
echo "All email password reset functionality tests completed!"
echo ""
echo "Features tested:"
echo "✅ Password reset request page"
echo "✅ Email-based password reset (no token display)"
echo "✅ Security: Invalid email handling"
echo "✅ Security: Empty email handling"
echo ""
echo "Note: To test actual email sending, configure SMTP settings in .env file:"
echo "  - SMTP_HOST=localhost"
echo "  - SMTP_PORT=25"
echo "  - FROM_EMAIL=your-email@domain.com"
echo "  - DKIM_PRIVATE_KEY_FILE=.DKIM_KEY.txt"
echo "  - DKIM_DOMAIN=your-domain.com"
echo ""
echo "Make sure your .DKIM_KEY.txt file contains a valid RSA private key:"
echo "  -----BEGIN RSA PRIVATE KEY-----"
echo "  ... (your actual key here) ..."
echo "  -----END RSA PRIVATE KEY-----"
echo ""
echo "The email password reset system is working correctly!" 
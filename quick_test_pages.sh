#!/bin/bash

# Quick CLI Test for /pages endpoint
# Simple tests to verify the /pages route works correctly

BASE_URL="http://localhost:6273"

echo "=== Quick Test for /pages Endpoint ==="
echo "Server: $BASE_URL"
echo ""

# Test 1: Basic GET request
echo "1. Testing basic GET request..."
curl -s -w "\nStatus: %{http_code}\n" "$BASE_URL/pages"

echo ""
echo "2. Testing with JSON formatting..."
curl -s "$BASE_URL/pages" | python3 -m json.tool 2>/dev/null || echo "Raw response (no JSON formatting available)"

echo ""
echo "3. Testing HTTP method validation..."
echo "POST request (should fail):"
curl -s -w "Status: %{http_code}\n" -X POST "$BASE_URL/pages"

echo ""
echo "4. Testing with response_format parameter..."
curl -s -w "\nStatus: %{http_code}\n" "$BASE_URL/pages?response_format=json"

echo ""
echo "=== Test Complete ==="
echo "Expected: 200 OK for GET requests, 405 for POST" 
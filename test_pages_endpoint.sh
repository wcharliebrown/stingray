#!/bin/bash

# CLI Test Script for /pages endpoint
# Tests various scenarios for the /pages route

BASE_URL="http://localhost:6273"
PAGES_ENDPOINT="/pages"

echo "=== Sting Ray /pages Endpoint CLI Tests ==="
echo "Base URL: $BASE_URL"
echo ""

# Function to print test header
print_test_header() {
    echo "----------------------------------------"
    echo "Test: $1"
    echo "----------------------------------------"
}

# Function to make curl request and format output
make_request() {
    local method="$1"
    local url="$2"
    local description="$3"
    
    echo "Request: $method $url"
    echo "Description: $description"
    echo ""
    
    if [ "$method" = "GET" ]; then
        curl -s -w "\nHTTP Status: %{http_code}\nContent-Type: %{content_type}\nResponse Time: %{time_total}s\n" \
             -H "Accept: application/json" \
             "$url"
    else
        curl -s -w "\nHTTP Status: %{http_code}\nContent-Type: %{content_type}\nResponse Time: %{time_total}s\n" \
             -X "$method" \
             -H "Accept: application/json" \
             "$url"
    fi
    echo ""
}

# Test 1: Basic GET request to /pages
print_test_header "Basic GET request to /pages endpoint"
make_request "GET" "$BASE_URL$PAGES_ENDPOINT" "Should return all pages in JSON format"

# Test 2: GET request with explicit JSON format parameter
print_test_header "GET request with response_format=json parameter"
make_request "GET" "$BASE_URL$PAGES_ENDPOINT?response_format=json" "Should return JSON response with explicit format parameter"

# Test 3: GET request with HTML format parameter (should still return JSON)
print_test_header "GET request with response_format=html parameter"
make_request "GET" "$BASE_URL$PAGES_ENDPOINT?response_format=html" "Should return JSON response even with html format parameter"

# Test 4: POST request (should fail - method not allowed)
print_test_header "POST request to /pages endpoint"
make_request "POST" "$BASE_URL$PAGES_ENDPOINT" "Should return 405 Method Not Allowed"

# Test 5: PUT request (should fail - method not allowed)
print_test_header "PUT request to /pages endpoint"
make_request "PUT" "$BASE_URL$PAGES_ENDPOINT" "Should return 405 Method Not Allowed"

# Test 6: DELETE request (should fail - method not allowed)
print_test_header "DELETE request to /pages endpoint"
make_request "DELETE" "$BASE_URL$PAGES_ENDPOINT" "Should return 405 Method Not Allowed"

# Test 7: GET request with Accept header for JSON
print_test_header "GET request with Accept: application/json header"
curl -s -w "\nHTTP Status: %{http_code}\nContent-Type: %{content_type}\nResponse Time: %{time_total}s\n" \
     -H "Accept: application/json" \
     "$BASE_URL$PAGES_ENDPOINT"

# Test 8: GET request with Accept header for HTML (should still return JSON)
print_test_header "GET request with Accept: text/html header"
curl -s -w "\nHTTP Status: %{http_code}\nContent-Type: %{content_type}\nResponse Time: %{time_total}s\n" \
     -H "Accept: text/html" \
     "$BASE_URL$PAGES_ENDPOINT"

# Test 9: GET request with verbose output
print_test_header "GET request with verbose curl output"
curl -v "$BASE_URL$PAGES_ENDPOINT" 2>&1 | head -20

# Test 10: Test with jq for JSON formatting (if available)
if command -v jq &> /dev/null; then
    print_test_header "GET request with jq JSON formatting"
    curl -s "$BASE_URL$PAGES_ENDPOINT" | jq '.'
else
    print_test_header "GET request with pretty JSON formatting (jq not available)"
    curl -s "$BASE_URL$PAGES_ENDPOINT" | python3 -m json.tool 2>/dev/null || echo "Python json.tool not available"
fi

# Test 11: Performance test - multiple requests
print_test_header "Performance test - 5 consecutive requests"
for i in {1..5}; do
    echo "Request $i:"
    curl -s -w "Status: %{http_code}, Time: %{time_total}s\n" \
         -o /dev/null \
         "$BASE_URL$PAGES_ENDPOINT"
done

# Test 12: Test with different User-Agent
print_test_header "GET request with custom User-Agent"
curl -s -w "\nHTTP Status: %{http_code}\nContent-Type: %{content_type}\nResponse Time: %{time_total}s\n" \
     -H "User-Agent: StingRay-CLI-Test/1.0" \
     "$BASE_URL$PAGES_ENDPOINT"

echo ""
echo "=== Test Summary ==="
echo "✅ Basic functionality tests completed"
echo "✅ HTTP method validation tests completed"
echo "✅ Response format tests completed"
echo "✅ Performance tests completed"
echo ""
echo "Expected Results:"
echo "- GET requests should return 200 OK with JSON content"
echo "- POST/PUT/DELETE should return 405 Method Not Allowed"
echo "- Response should contain all pages from database"
echo "- Content-Type should be application/json"
echo ""
echo "To run individual tests, use:"
echo "curl -s $BASE_URL$PAGES_ENDPOINT | jq '.'"
echo "curl -v $BASE_URL$PAGES_ENDPOINT"
echo "curl -w \"Status: %{http_code}\" $BASE_URL$PAGES_ENDPOINT" 
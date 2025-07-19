#!/bin/bash

echo "Testing modern template with header and footer..."

# Start the server in the background
./stingray &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Test a page that uses the modern template
curl -s "http://localhost:6273/page/about" | grep -q "modern_header\|modern_footer"

if [ $? -eq 0 ]; then
    echo "❌ Template references still present - embedding not working"
    curl -s "http://localhost:6273/page/about" | grep -E "(template_modern_header|template_modern_footer)"
else
    echo "✅ Template embedding working - no template references found"
fi

# Check if header and footer content is actually rendered
curl -s "http://localhost:6273/page/about" | grep -q "<header class=\"header\">"
if [ $? -eq 0 ]; then
    echo "✅ Header content is being rendered"
else
    echo "❌ Header content not found"
fi

curl -s "http://localhost:6273/page/about" | grep -q "<footer class=\"footer\">"
if [ $? -eq 0 ]; then
    echo "✅ Footer content is being rendered"
else
    echo "❌ Footer content not found"
fi

# Kill the server
kill $SERVER_PID 2>/dev/null

echo "Test completed." 
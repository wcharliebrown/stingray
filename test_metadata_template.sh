#!/bin/bash

echo "Testing metadata template with header and footer..."

# Test the metadata tables page
echo "Testing /metadata/tables:"
curl -s "http://localhost:6273/metadata/tables" | grep -q "{{template_"
if [ $? -eq 0 ]; then
    echo "❌ Template references still present"
else
    echo "✅ Template references processed"
fi

# Check if header and footer content is actually rendered
curl -s "http://localhost:6273/metadata/tables" | grep -q "<header class=\"header\">"
if [ $? -eq 0 ]; then
    echo "✅ Header content is being rendered"
else
    echo "❌ Header content not found"
fi

curl -s "http://localhost:6273/metadata/tables" | grep -q "<footer class=\"footer\">"
if [ $? -eq 0 ]; then
    echo "✅ Footer content is being rendered"
else
    echo "❌ Footer content not found"
fi

# Check if navigation links are working
curl -s "http://localhost:6273/metadata/tables" | grep -q "<a href=\"/\">Home</a>"
if [ $? -eq 0 ]; then
    echo "✅ Navigation links are working"
else
    echo "❌ Navigation links not found"
fi

echo "Test completed." 
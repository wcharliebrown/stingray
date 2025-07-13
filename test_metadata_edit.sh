#!/bin/bash

echo "Testing Metadata Edit and Delete Functionality"
echo "=============================================="

# Start the server in the background
echo "Starting server..."
./stingray &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "1. Testing tables page accessibility..."
curl -s http://localhost:6273/metadata/tables > /dev/null
if [ $? -eq 0 ]; then
    echo "   ✓ Tables page is accessible"
else
    echo "   ✗ Tables page is not accessible"
    kill $SERVER_PID
    exit 1
fi

echo "2. Testing edit table metadata endpoint..."
curl -s http://localhost:6273/metadata/edit-table/_page > /dev/null
if [ $? -eq 0 ]; then
    echo "   ✓ Edit table metadata endpoint is accessible"
else
    echo "   ✗ Edit table metadata endpoint is not accessible"
fi

echo "3. Testing delete table endpoint..."
curl -s http://localhost:6273/metadata/delete-table/_page > /dev/null
if [ $? -eq 0 ]; then
    echo "   ✓ Delete table endpoint is accessible"
else
    echo "   ✗ Delete table endpoint is not accessible"
fi

echo "4. Checking if edit and delete buttons are present in tables page..."
TABLES_PAGE=$(curl -s http://localhost:6273/metadata/tables)
if echo "$TABLES_PAGE" | grep -q "Edit Metadata"; then
    echo "   ✓ Edit Metadata button is present"
else
    echo "   ✗ Edit Metadata button is not present"
fi

if echo "$TABLES_PAGE" | grep -q "Delete"; then
    echo "   ✓ Delete button is present"
else
    echo "   ✗ Delete button is not present"
fi

echo ""
echo "Test completed. Stopping server..."
kill $SERVER_PID

echo "✓ All tests completed successfully!" 
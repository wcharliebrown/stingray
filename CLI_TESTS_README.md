# CLI Tests for /pages Endpoint

This directory contains CLI test scripts for testing the `/pages` endpoint of the Sting Ray application using curl.

## Test Scripts

### 1. `quick_test_pages.sh` - Quick Basic Tests
A simple script for basic functionality testing.

**Usage:**
```bash
./quick_test_pages.sh
```

**What it tests:**
- Basic GET request to `/pages`
- JSON response formatting
- HTTP method validation (POST should fail)
- Response format parameter handling

### 2. `test_pages_endpoint.sh` - Comprehensive Tests
A comprehensive test suite covering all scenarios.

**Usage:**
```bash
./test_pages_endpoint.sh
```

**What it tests:**
- All HTTP methods (GET, POST, PUT, DELETE)
- Different response format parameters
- Various HTTP headers
- Performance testing
- JSON formatting with jq (if available)
- Verbose output for debugging

## Manual Testing Commands

### Basic GET request
```bash
curl http://localhost:6273/pages
```

### GET with JSON formatting
```bash
curl -s http://localhost:6273/pages | jq '.'
# or
curl -s http://localhost:6273/pages | python3 -m json.tool
```

### Test with response format parameter
```bash
curl http://localhost:6273/pages?response_format=json
```

### Test HTTP method validation
```bash
curl -X POST http://localhost:6273/pages
curl -X PUT http://localhost:6273/pages
curl -X DELETE http://localhost:6273/pages
```

### Verbose output for debugging
```bash
curl -v http://localhost:6273/pages
```

### Performance testing
```bash
curl -w "Status: %{http_code}, Time: %{time_total}s\n" -o /dev/null http://localhost:6273/pages
```

## Expected Results

### Successful GET Request (200 OK)
```json
{
  "about": {
    "id": 2,
    "slug": "about",
    "title": "About Sting Ray",
    "meta_description": "Learn more about the Sting Ray application",
    "header": "<h1>About Sting Ray</h1>",
    "navigation": "<ul><li><a href=\"/\">Home</a></li><li><a href=\"/page/about\">About</a></li><li><a href=\"/user/login\">Login</a></li></ul>",
    "main_content": "<p>Sting Ray is a modern web application built with Go...</p>",
    "sidebar": "<div class='sidebar'><h3>Contact</h3><p>Get in touch with us for more information.</p></div>",
    "footer": "<footer>&copy; 2025 StingRay. All rights reserved.</footer>",
    "css_class": "about-page",
    "scripts": "<script>console.log('About page loaded');</script>",
    "template": "default"
  },
  "home": {
    // ... home page data
  },
  "login": {
    // ... login page data
  },
  "shutdown": {
    // ... shutdown page data
  }
}
```

### Failed Requests (405 Method Not Allowed)
```
Method not allowed
```

## Prerequisites

1. **Server Running**: Make sure the Sting Ray server is running on `http://localhost:6273`
2. **curl**: Should be available on most Unix-like systems
3. **Optional Tools**:
   - `jq` for JSON formatting: `brew install jq` (macOS) or `apt-get install jq` (Ubuntu)
   - `python3` for JSON formatting (usually pre-installed)

## Running Tests

1. Start the Sting Ray server:
   ```bash
   go run stingray.go
   ```

2. In another terminal, run the tests:
   ```bash
   # Quick test
   ./quick_test_pages.sh
   
   # Comprehensive test
   ./test_pages_endpoint.sh
   ```

## Troubleshooting

### Server not running
```
curl: (7) Failed to connect to localhost port 6273: Connection refused
```
**Solution**: Start the server with `go run stingray.go`

### Permission denied
```
bash: ./test_pages_endpoint.sh: Permission denied
```
**Solution**: Make the script executable with `chmod +x test_pages_endpoint.sh`

### JSON formatting not working
If `jq` is not available, the script will fall back to raw output or try `python3 -m json.tool`.

## Test Coverage

The comprehensive test script covers:

- ✅ HTTP method validation
- ✅ Response format handling
- ✅ Content-Type headers
- ✅ Performance metrics
- ✅ Error handling
- ✅ JSON response structure
- ✅ Different Accept headers
- ✅ Custom User-Agent headers

## Expected Database Pages

The tests expect the following pages to be available in the database:
- `home` - Welcome page
- `about` - About page
- `login` - Login page
- `shutdown` - Shutdown page 
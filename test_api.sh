#!/bin/bash

# Link Shortener API Test Script
# Tests all endpoints with various authentication scenarios

[ -f .env.local ] && set -a && . .env.local && set +a
set -e

BASE_URL="http://localhost:8080"
# Note: Update this token to match your .env.local AUTH_TOKEN
VALID_TOKEN="${AUTH_TOKEN}"
INVALID_TOKEN="invalid-token"

echo "🧪 Link Shortener API Test Suite"
echo "=================================="
echo

# Test 1: Health check (public)
echo "1. Testing health endpoint (public)..."
response=$(curl -s "$BASE_URL/health")
if [[ "$response" == *"healthy"* ]]; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed: $response"
    exit 1
fi
echo

# Test 2: Dashboard without auth (should work - now public)
echo "2. Testing dashboard without auth (should work)..."
response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/")
if [[ "$response" == "200" ]]; then
    echo "✅ Dashboard accessible without auth"
else
    echo "❌ Dashboard failed without auth: HTTP $response"
    exit 1
fi
echo

# Test 3: Analytics without auth (should work - now public)
echo "3. Testing analytics without auth (should work)..."
response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/analytics")
if [[ "$response" == "200" ]]; then
    echo "✅ Analytics accessible without auth"
else
    echo "❌ Analytics failed without auth: HTTP $response"
    exit 1
fi
echo

# Test 4: Create short URL without auth (should fail)
echo "4. Testing create short URL without auth (should fail)..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "original_url=https://example.com" \
    "$BASE_URL/shorten")
if [[ "$response" == "401" ]]; then
    echo "✅ Create short URL properly requires auth"
else
    echo "❌ Create short URL should require auth: HTTP $response"
    exit 1
fi
echo

# Test 5: Create short URL with invalid token (should fail)
echo "5. Testing create short URL with invalid token (should fail)..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Authorization: Bearer $INVALID_TOKEN" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "original_url=https://example.com" \
    "$BASE_URL/shorten")
if [[ "$response" == "401" ]]; then
    echo "✅ Invalid token properly rejected"
else
    echo "❌ Invalid token should be rejected: HTTP $response"
    exit 1
fi
echo

# Test 6: Create short URL with valid token (should work)
echo "6. Testing create short URL with valid token (should work)..."
response=$(curl -s -X POST \
    -H "Authorization: Bearer $VALID_TOKEN" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "original_url=https://test-example.com&created_by=test-script" \
    "$BASE_URL/shorten")
if [[ "$response" == *"Short link created successfully"* ]]; then
    echo "✅ Create short URL with valid token works"
    # Extract the short code from the response
    short_code=$(echo "$response" | grep -o 'value="[^"]*' | grep -o '[^"]*$' | grep -o '[^/]*$')
    echo "   Created short code: $short_code"
else
    echo "❌ Create short URL with valid token failed: $response"
    exit 1
fi
echo

# Test 7: Test redirect (public - should work)
if [[ -n "$short_code" ]]; then
    echo "7. Testing redirect for created link (should return 302)..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/$short_code")
    if [[ "$response" == "302" ]]; then
        echo "✅ Redirect returns proper 302 status"
    else
        echo "❌ Redirect should return 302: HTTP $response"
        exit 1
    fi
else
    echo "7. Skipping redirect test (no short code available)"
fi
echo

# Test 8: Test redirect for non-existent code (should 404)
echo "8. Testing redirect for non-existent code (should 404)..."
response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/nonexistent")
if [[ "$response" == "404" ]]; then
    echo "✅ Non-existent code properly returns 404"
else
    echo "❌ Non-existent code should return 404: HTTP $response"
    exit 1
fi
echo

# Test 9: Test malformed authorization header
echo "9. Testing malformed authorization header (should fail)..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Authorization: NotBearer $VALID_TOKEN" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "original_url=https://example.com" \
    "$BASE_URL/shorten")
if [[ "$response" == "401" ]]; then
    echo "✅ Malformed auth header properly rejected"
else
    echo "❌ Malformed auth header should be rejected: HTTP $response"
    exit 1
fi
echo

# Test 10: Test API endpoint with JSON accept header
echo "10. Testing analytics with JSON accept header..."
response=$(curl -s -H "Accept: application/json" "$BASE_URL/analytics")
if [[ "$response" == *"\"total_links\""* ]]; then
    echo "✅ JSON API response works"
else
    echo "❌ JSON API response failed: $response"
    exit 1
fi
echo

echo "🎉 All tests passed!"
echo "Summary:"
echo "- ✅ Public endpoints accessible without auth"
echo "- ✅ Protected endpoints require valid bearer token"
echo "- ✅ Invalid tokens properly rejected"
echo "- ✅ Link creation and redirect functionality works"
echo "- ✅ JSON API responses work"
echo
echo "Authentication setup:"
echo "- Dashboard: PUBLIC (no auth required)"
echo "- Analytics: PUBLIC (no auth required)" 
echo "- Link creation: PROTECTED (bearer token required)"
echo "- Redirects: PUBLIC (no auth required)"
echo "- Health check: PUBLIC (no auth required)"
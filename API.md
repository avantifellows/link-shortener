# Link Shortener API Documentation

A RESTful API for creating and managing short links with authentication and analytics.

## Base URL

```
Production: https://lnk.avantifellows.org
Testing:    https://temp-lnk.avantifellows.org
Local:      http://localhost:8080
```

## Authentication

Protected endpoints require a Bearer token in the Authorization header:

```bash
Authorization: Bearer YOUR_AUTH_TOKEN
```

Get your auth token from your administrator or environment configuration.

## API Endpoints

### 🔒 Create Short URL (Protected)

**POST** `/shorten`

Create a new short URL. Requires authentication.

#### Request Headers
```
Authorization: Bearer YOUR_AUTH_TOKEN
Content-Type: application/x-www-form-urlencoded
Accept: application/json  # For JSON response
```

#### Request Body (Form Data)
```
original_url=https://example.com/very/long/url
custom_code=my-custom-code    # Optional: 3-20 chars, alphanumeric + hyphens/underscores
created_by=username           # Optional: identifier for creator
```

#### Request Body (JSON)
```json
{
  "original_url": "https://example.com/very/long/url",
  "custom_code": "my-custom-code",
  "created_by": "username"
}
```

#### Response (Success - 200)
```json
{
  "short_code": "abc123",
  "short_url": "https://lnk.avantifellows.org/abc123",
  "original_url": "https://example.com/very/long/url"
}
```

#### Response (Error - 400/401)
```json
{
  "error": "Custom code already exists"
}
```

#### curl Examples
```bash
# Form submission
curl -X POST https://lnk.avantifellows.org/shorten \
  -H "Authorization: Bearer YOUR_AUTH_TOKEN" \
  -d "original_url=https://example.com&custom_code=test&created_by=api-user"

# JSON submission  
curl -X POST https://lnk.avantifellows.org/shorten \
  -H "Authorization: Bearer YOUR_AUTH_TOKEN" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com","custom_code":"test","created_by":"api-user"}'
```

---

### 🌐 Get Analytics (Public)

**GET** `/analytics`

Retrieve analytics data for all links. No authentication required.

#### Request Headers
```
Accept: application/json
```

#### Response (200)
```json
{
  "links": [
    {
      "short_code": "abc123",
      "original_url": "https://example.com",
      "created_at": "2025-08-20T10:30:00Z",
      "created_by": "username",
      "click_count": 42,
      "last_accessed": "2025-08-20T15:45:00Z"
    }
  ],
  "total_links": 1,
  "total_clicks": 42,
  "recent_clicks": [
    {
      "id": 1,
      "short_code": "abc123", 
      "timestamp": "2025-08-20T15:45:00Z",
      "user_agent": "Mozilla/5.0...",
      "ip_address": "192.168.1.1",
      "referrer": "https://google.com"
    }
  ]
}
```

#### curl Example
```bash
curl -H "Accept: application/json" https://lnk.avantifellows.org/analytics
```

---

### 🌐 Health Check (Public)

**GET** `/health`

Check if the service is running. No authentication required.

#### Response (200)
```json
{
  "status": "healthy"
}
```

#### curl Example
```bash
curl https://lnk.avantifellows.org/health
```

---

### 🌐 URL Redirect (Public)

**GET** `/{short_code}`

Redirect to the original URL. Automatically tracks click analytics. No authentication required.

#### Parameters
- `short_code` - The short code to redirect (e.g., "abc123")

#### Response
- **302 Found** - Redirects to original URL
- **404 Not Found** - Short code doesn't exist

#### Example
```bash
# Browser redirect
https://lnk.avantifellows.org/abc123
# → Redirects to original URL

# Follow redirect with curl
curl -L https://lnk.avantifellows.org/abc123
```

---

### 🌐 Dashboard (Public)

**GET** `/`

Web interface for viewing analytics and creating links. No authentication required for viewing.

#### Example
```
https://lnk.avantifellows.org/
```

## Error Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 302 | Redirect (for short URLs) |
| 400 | Bad Request (invalid URL, custom code exists, etc.) |
| 401 | Unauthorized (missing/invalid token) |
| 404 | Not Found (invalid short code) |
| 405 | Method Not Allowed |
| 500 | Internal Server Error |

## Rate Limiting

Currently no rate limiting is implemented. For production use, consider:
- IP-based rate limiting
- Token-based quotas
- Nginx rate limiting

## Security Notes

1. **Bearer Tokens**: Store securely, never commit to version control
2. **HTTPS Only**: Always use HTTPS in production
3. **Input Validation**: All URLs are validated before storage
4. **Click Tracking**: IP addresses are logged for analytics

## SDK Examples

### JavaScript/Node.js
```javascript
const authToken = process.env.AUTH_TOKEN;
const baseUrl = 'https://lnk.avantifellows.org';

async function createShortUrl(originalUrl, customCode = '', createdBy = '') {
  const response = await fetch(`${baseUrl}/shorten`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${authToken}`,
      'Accept': 'application/json',
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      original_url: originalUrl,
      custom_code: customCode,
      created_by: createdBy
    })
  });
  
  return response.json();
}

// Usage
createShortUrl('https://example.com', 'my-link', 'api-user')
  .then(result => console.log('Short URL:', result.short_url));
```

### Python
```python
import requests
import os

auth_token = os.getenv('AUTH_TOKEN')
base_url = 'https://lnk.avantifellows.org'

def create_short_url(original_url, custom_code='', created_by=''):
    response = requests.post(
        f'{base_url}/shorten',
        headers={
            'Authorization': f'Bearer {auth_token}',
            'Accept': 'application/json'
        },
        data={
            'original_url': original_url,
            'custom_code': custom_code,
            'created_by': created_by
        }
    )
    return response.json()

# Usage
result = create_short_url('https://example.com', 'my-link', 'api-user')
print(f"Short URL: {result['short_url']}")
```

### Go
```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "strings"
)

type CreateShortURLResponse struct {
    ShortCode   string `json:"short_code"`
    ShortURL    string `json:"short_url"`
    OriginalURL string `json:"original_url"`
}

func createShortURL(originalURL, customCode, createdBy string) (*CreateShortURLResponse, error) {
    authToken := os.Getenv("AUTH_TOKEN")
    baseURL := "https://lnk.avantifellows.org"
    
    data := url.Values{}
    data.Set("original_url", originalURL)
    data.Set("custom_code", customCode)
    data.Set("created_by", createdBy)
    
    req, _ := http.NewRequest("POST", baseURL+"/shorten", strings.NewReader(data.Encode()))
    req.Header.Set("Authorization", "Bearer "+authToken)
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result CreateShortURLResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}

// Usage
func main() {
    result, _ := createShortURL("https://example.com", "my-link", "api-user")
    fmt.Printf("Short URL: %s\n", result.ShortURL)
}
```

## Support

For issues or questions about the API:
1. Check the [main README](README.md) for setup instructions
2. Review the [test suite](test_api.sh) for working examples
3. Contact your system administrator for authentication tokens
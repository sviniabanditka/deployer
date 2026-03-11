# API Reference — Authentication

Base URL: `https://api.deployer.dev/api/v1`

All API requests (except auth and webhooks) require a Bearer token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

## Register

Create a new account.

```
POST /auth/register
```

```bash
curl -X POST https://api.deployer.dev/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "securepassword"
  }'
```

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Display name |
| `email` | string | Yes | Valid email address |
| `password` | string | Yes | Minimum 8 characters |

**Response** `201 Created`:

```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "jane@example.com",
    "name": "Jane Doe",
    "created_at": "2026-03-10T12:00:00Z"
  }
}
```

## Login

Authenticate and receive access and refresh tokens.

```
POST /auth/login
```

```bash
curl -X POST https://api.deployer.dev/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "password": "securepassword"
  }'
```

**Response** `200 OK`:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

If the user has 2FA enabled, the response is:

```json
{
  "requires_2fa": true,
  "temp_token": "temp-token-string"
}
```

## Refresh token

Exchange a refresh token for a new token pair.

```
POST /auth/refresh
```

```bash
curl -X POST https://api.deployer.dev/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{ "refresh_token": "eyJhbGciOiJIUzI1NiIs..." }'
```

**Response** `200 OK`:

```json
{
  "access_token": "new-access-token",
  "refresh_token": "new-refresh-token"
}
```

## Two-Factor Authentication (2FA)

### Enable 2FA

Generates a TOTP secret and QR code URL. Requires verification before activation.

```
POST /auth/2fa/enable
```

```bash
curl -X POST https://api.deployer.dev/api/v1/auth/2fa/enable \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code_url": "otpauth://totp/Deployer:jane@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Deployer"
}
```

### Verify 2FA setup

Confirm the TOTP code to activate 2FA.

```
POST /auth/2fa/verify
```

```bash
curl -X POST https://api.deployer.dev/api/v1/auth/2fa/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "code": "123456" }'
```

### Validate 2FA during login

Complete login when 2FA is required.

```
POST /auth/2fa/validate
```

```bash
curl -X POST https://api.deployer.dev/api/v1/auth/2fa/validate \
  -H "Content-Type: application/json" \
  -d '{
    "temp_token": "temp-token-from-login",
    "code": "123456"
  }'
```

**Response** `200 OK`:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Disable 2FA

```
POST /auth/2fa/disable
```

```bash
curl -X POST https://api.deployer.dev/api/v1/auth/2fa/disable \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "code": "123456" }'
```

## GDPR

### Export user data

Download all data associated with your account.

```
GET /auth/export-data
```

```bash
curl https://api.deployer.dev/api/v1/auth/export-data \
  -H "Authorization: Bearer $TOKEN" \
  -o user-data-export.json
```

### Delete account

Permanently delete your account and all associated data.

```
DELETE /auth/account
```

```bash
curl -X DELETE https://api.deployer.dev/api/v1/auth/account \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "password": "your-password" }'
```

## Rate limiting

Public auth endpoints are rate-limited. Authenticated endpoints have higher limits based on your plan. When rate-limited, the API returns `429 Too Many Requests`.

## Error format

All errors follow this format:

```json
{
  "error": "description of the error"
}
```

Common HTTP status codes:

| Code | Meaning |
|------|---------|
| `400` | Bad request (validation error) |
| `401` | Unauthorized (missing or invalid token) |
| `403` | Forbidden (quota exceeded or insufficient permissions) |
| `404` | Resource not found |
| `409` | Conflict (e.g., duplicate email) |
| `429` | Rate limit exceeded |
| `500` | Internal server error |

# API Reference — Apps

All app endpoints require authentication.

## Create app

```
POST /apps
```

```bash
curl -X POST https://api.deployer.dev/api/v1/apps \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "name": "my-app" }'
```

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | App name (1-100 characters) |

**Response** `201 Created`:

```json
{
  "app": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "...",
    "name": "my-app",
    "slug": "my-app",
    "status": "created",
    "created_at": "2026-03-10T12:00:00Z"
  }
}
```

Returns `403` if the app quota for your plan is exceeded.

## List apps

```
GET /apps
```

```bash
curl https://api.deployer.dev/api/v1/apps \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "apps": [
    {
      "id": "...",
      "name": "my-app",
      "slug": "my-app",
      "status": "running",
      "created_at": "2026-03-10T12:00:00Z"
    }
  ]
}
```

## Get app

```
GET /apps/:id
```

```bash
curl https://api.deployer.dev/api/v1/apps/{app_id} \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "app": {
    "id": "...",
    "name": "my-app",
    "slug": "my-app",
    "status": "running",
    "created_at": "2026-03-10T12:00:00Z"
  }
}
```

## Delete app

```
DELETE /apps/:id
```

```bash
curl -X DELETE https://api.deployer.dev/api/v1/apps/{app_id} \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{ "message": "app deleted" }
```

## Deploy

Upload a ZIP archive to deploy.

```
POST /apps/:id/deploy
```

```bash
curl -X POST https://api.deployer.dev/api/v1/apps/{app_id}/deploy \
  -H "Authorization: Bearer $TOKEN" \
  -F "archive=@app.zip"
```

**Response** `202 Accepted`:

```json
{
  "message": "deployment queued",
  "deployment": {
    "id": "d1234567-...",
    "app_id": "...",
    "version": 3,
    "status": "pending",
    "image_tag": "registry.deployer.dev/my-app:3"
  }
}
```

## Stop app

```
POST /apps/:id/stop
```

```bash
curl -X POST https://api.deployer.dev/api/v1/apps/{app_id}/stop \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "message": "app stopped",
  "app": { "id": "...", "status": "stopped" }
}
```

## Start app

Restart a stopped app using its latest deployment.

```
POST /apps/:id/start
```

```bash
curl -X POST https://api.deployer.dev/api/v1/apps/{app_id}/start \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "message": "app started",
  "app": { "id": "...", "status": "running" }
}
```

## Update environment variables

```
PUT /apps/:id/env
```

```bash
curl -X PUT https://api.deployer.dev/api/v1/apps/{app_id}/env \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vars": {
      "NODE_ENV": "production",
      "SECRET": "abc123"
    }
  }'
```

::: warning
This replaces all environment variables. Include existing variables you want to keep.
:::

**Response** `200 OK`:

```json
{ "message": "env vars updated" }
```

## Get app logs

Fetch recent logs (HTTP) or stream live logs (WebSocket).

```
GET /apps/:id/logs
```

```bash
curl https://api.deployer.dev/api/v1/apps/{app_id}/logs \
  -H "Authorization: Bearer $TOKEN"
```

For real-time streaming, connect via WebSocket:

```
ws://api.deployer.dev/api/v1/apps/{app_id}/logs
```

## Get app stats

Get CPU and memory usage for the running container.

```
GET /apps/:id/stats
```

```bash
curl https://api.deployer.dev/api/v1/apps/{app_id}/stats \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "stats": {
    "cpu_percent": 2.5,
    "memory_usage": 67108864,
    "memory_limit": 536870912,
    "network_rx": 1024000,
    "network_tx": 512000
  }
}
```

## Git integration

### Connect repository

```
POST /apps/:id/git/connect
```

```bash
curl -X POST https://api.deployer.dev/api/v1/apps/{app_id}/git/connect \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "github",
    "repo_url": "https://github.com/user/repo",
    "branch": "main",
    "access_token": "ghp_xxxx"
  }'
```

### Get connection

```
GET /apps/:id/git
```

### Disconnect repository

```
DELETE /apps/:id/git/disconnect
```

# API Reference — Databases

All database endpoints require authentication.

## Create database

```
POST /databases
```

```bash
curl -X POST https://api.deployer.dev/api/v1/databases \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-postgres",
    "engine": "postgres",
    "version": "16",
    "app_id": "550e8400-..."
  }'
```

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Database name (1-100 characters) |
| `engine` | string | Yes | `postgres`, `mysql`, `mongodb`, or `redis` |
| `version` | string | No | Engine version (uses latest if omitted) |
| `app_id` | string | No | Link to an app on creation |

**Response** `201 Created`:

```json
{
  "id": "d1234567-...",
  "name": "my-postgres",
  "engine": "postgres",
  "version": "16",
  "status": "running",
  "host": "db-abc123.internal",
  "port": 5432,
  "username": "deployer",
  "password": "auto-generated-password",
  "database_name": "mydb",
  "connection_url": "postgres://deployer:auto-generated-password@db-abc123.internal:5432/mydb"
}
```

Returns `403` if the database quota for your plan is exceeded.

## List databases

```
GET /databases
```

```bash
curl https://api.deployer.dev/api/v1/databases \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "databases": [
    {
      "id": "...",
      "name": "my-postgres",
      "engine": "postgres",
      "status": "running"
    }
  ]
}
```

## Get database

Returns full connection details.

```
GET /databases/:id
```

```bash
curl https://api.deployer.dev/api/v1/databases/{db_id} \
  -H "Authorization: Bearer $TOKEN"
```

## Delete database

```
DELETE /databases/:id
```

```bash
curl -X DELETE https://api.deployer.dev/api/v1/databases/{db_id} \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{ "message": "database deleted" }
```

## Stop database

```
POST /databases/:id/stop
```

```bash
curl -X POST https://api.deployer.dev/api/v1/databases/{db_id}/stop \
  -H "Authorization: Bearer $TOKEN"
```

## Start database

```
POST /databases/:id/start
```

```bash
curl -X POST https://api.deployer.dev/api/v1/databases/{db_id}/start \
  -H "Authorization: Bearer $TOKEN"
```

## Link to app

Sets `DATABASE_URL` (or `REDIS_URL`) on the linked app.

```
POST /databases/:id/link
```

```bash
curl -X POST https://api.deployer.dev/api/v1/databases/{db_id}/link \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "app_id": "550e8400-..." }'
```

**Response** `200 OK`:

```json
{ "message": "database linked to app" }
```

## Unlink from app

```
POST /databases/:id/unlink
```

```bash
curl -X POST https://api.deployer.dev/api/v1/databases/{db_id}/unlink \
  -H "Authorization: Bearer $TOKEN"
```

## Create backup

```
POST /databases/:id/backups
```

```bash
curl -X POST https://api.deployer.dev/api/v1/databases/{db_id}/backups \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `201 Created`:

```json
{
  "id": "b1234567-...",
  "database_id": "d1234567-...",
  "status": "completed",
  "size": "12.4 MB",
  "created_at": "2026-03-10T14:30:00Z"
}
```

## List backups

```
GET /databases/:id/backups
```

```bash
curl https://api.deployer.dev/api/v1/databases/{db_id}/backups \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "backups": [
    {
      "id": "b1234567-...",
      "status": "completed",
      "size": "12.4 MB",
      "created_at": "2026-03-10T14:30:00Z"
    }
  ]
}
```

## Restore backup

```
POST /databases/:id/backups/:backup_id/restore
```

```bash
curl -X POST https://api.deployer.dev/api/v1/databases/{db_id}/backups/{backup_id}/restore \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{ "message": "backup restored" }
```

# Database Backups

Create, list, and restore backups for any managed database.

## Create a backup

### CLI

```bash
deployer db backup <db-id>
```

### API

```bash
curl -X POST https://api.deployer.dev/api/v1/databases/{db_id}/backups \
  -H "Authorization: Bearer $TOKEN"
```

Response:

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

### CLI

```bash
deployer db backups <db-id>
```

```
ID                                    DATE                      SIZE      STATUS
--                                    ----                      ----      ------
b1234567-...                          2026-03-10T14:30:00Z      12.4 MB   completed
b2345678-...                          2026-03-09T14:30:00Z      12.1 MB   completed
```

### API

```bash
curl https://api.deployer.dev/api/v1/databases/{db_id}/backups \
  -H "Authorization: Bearer $TOKEN"
```

## Restore a backup

::: danger
Restoring a backup overwrites all current data in the database.
:::

### CLI

```bash
deployer db restore <db-id> <backup-id>
```

You will be asked to confirm before the restore proceeds.

### API

```bash
curl -X POST https://api.deployer.dev/api/v1/databases/{db_id}/backups/{backup_id}/restore \
  -H "Authorization: Bearer $TOKEN"
```

## Backup methods by engine

| Engine | Method | Format |
|--------|--------|--------|
| PostgreSQL | `pg_dump` | Custom format |
| MySQL | `mysqldump` | SQL |
| MongoDB | `mongodump` | BSON |
| Redis | `BGSAVE` | RDB snapshot |

## Automatic backups

Automatic daily backups are available on the Pro plan. They are retained for 7 days (Pro) or 30 days (Business).

## Tips

- Create a backup before running migrations.
- Create a backup before restoring a different backup.
- Backup size counts toward your plan's storage limit.
- Backups are stored encrypted at rest.

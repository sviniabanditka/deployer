# Managed Databases

Deployer provides managed database instances that you can create, link to apps, and back up with a single command.

## Supported engines

| Engine | Versions | Default port |
|--------|----------|-------------|
| PostgreSQL | 14, 15, 16 | 5432 |
| MySQL | 8.0, 8.4 | 3306 |
| MongoDB | 6.0, 7.0 | 27017 |
| Redis | 7.0, 7.2 | 6379 |

## Create a database

### CLI

```bash
deployer db create postgres
deployer db create mysql --name my-mysql-db
deployer db create mongodb --app <app-id>
```

### API

```bash
curl -X POST https://api.deployer.dev/api/v1/databases \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-database",
    "engine": "postgres",
    "version": "16"
  }'
```

The response includes connection credentials:

```json
{
  "id": "...",
  "name": "my-database",
  "engine": "postgres",
  "status": "running",
  "host": "db-abc123.internal",
  "port": 5432,
  "username": "deployer",
  "password": "generated-password",
  "database_name": "mydb",
  "connection_url": "postgres://deployer:generated-password@db-abc123.internal:5432/mydb"
}
```

## Link a database to an app

Linking sets `DATABASE_URL` (or `REDIS_URL`) on the app automatically.

### CLI

```bash
deployer db link <db-id> <app-id>
```

### API

```bash
curl -X POST https://api.deployer.dev/api/v1/databases/{db_id}/link \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "app_id": "<app-id>" }'
```

To unlink:

```bash
deployer db unlink <db-id>
```

## List databases

```bash
deployer db list
```

```
ID                                    NAME          ENGINE    STATUS   CONNECTION URL
--                                    ----          ------    ------   --------------
550e8400-...                          my-postgres   postgres  running  postgres://deployer:****@db-abc.internal:5432/mydb
661f9500-...                          my-redis      redis     running  redis://:****@db-def.internal:6379
```

## Stop and start

```bash
deployer db stop <db-id>
deployer db start <db-id>
```

Stopped databases do not consume compute resources but retain their data.

## Delete a database

```bash
deployer db delete <db-id>
```

::: danger
Deleting a database destroys all data permanently. Create a backup first.
:::

## Next steps

- [PostgreSQL guide](/guide/db-postgres)
- [MySQL guide](/guide/db-mysql)
- [MongoDB guide](/guide/db-mongodb)
- [Redis guide](/guide/db-redis)
- [Backups](/guide/db-backups)

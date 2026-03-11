# PostgreSQL

## Create a PostgreSQL database

```bash
deployer db create postgres --name my-postgres
```

Specify a version:

```bash
curl -X POST https://api.deployer.dev/api/v1/databases \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-postgres",
    "engine": "postgres",
    "version": "16"
  }'
```

Available versions: `14`, `15`, `16` (default: latest).

## Connection details

After creation, the API returns:

```json
{
  "host": "db-abc123.internal",
  "port": 5432,
  "username": "deployer",
  "password": "auto-generated",
  "database_name": "mydb",
  "connection_url": "postgres://deployer:auto-generated@db-abc123.internal:5432/mydb"
}
```

## Connect from your app

Link the database to your app to inject `DATABASE_URL` automatically:

```bash
deployer db link <db-id> <app-id>
```

Then read it in your code:

```js
// Node.js
const { Pool } = require('pg')
const pool = new Pool({ connectionString: process.env.DATABASE_URL })
```

```python
# Python / Django
import dj_database_url
DATABASES = { 'default': dj_database_url.config(default=os.environ['DATABASE_URL']) }
```

```go
// Go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

## Connect with psql

If your database is exposed externally (self-hosted setup), connect directly:

```bash
psql "postgres://deployer:password@db-abc123.internal:5432/mydb"
```

## Performance tips

- Use connection pooling (e.g., PgBouncer or your framework's built-in pool).
- Keep connections under your plan's limit.
- Add indexes for frequently queried columns.
- Run `EXPLAIN ANALYZE` on slow queries.

## Backups

See [Backups](/guide/db-backups) for creating and restoring PostgreSQL backups.

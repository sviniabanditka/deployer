# MySQL

## Create a MySQL database

```bash
deployer db create mysql --name my-mysql
```

With a specific version:

```bash
curl -X POST https://api.deployer.dev/api/v1/databases \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-mysql",
    "engine": "mysql",
    "version": "8.4"
  }'
```

Available versions: `8.0`, `8.4` (default: latest).

## Connection details

```json
{
  "host": "db-def456.internal",
  "port": 3306,
  "username": "deployer",
  "password": "auto-generated",
  "database_name": "mydb",
  "connection_url": "mysql://deployer:auto-generated@db-def456.internal:3306/mydb"
}
```

## Connect from your app

Link the database:

```bash
deployer db link <db-id> <app-id>
```

Then use `DATABASE_URL` in your code:

```js
// Node.js (mysql2)
const mysql = require('mysql2/promise')
const connection = await mysql.createConnection(process.env.DATABASE_URL)
```

```python
# Python / Django
DATABASES = {
    'default': dj_database_url.config(default=os.environ['DATABASE_URL'])
}
```

```php
// Laravel (.env)
// DATABASE_URL is parsed automatically by Laravel's database config
```

## Connect with mysql client

```bash
mysql -h db-def456.internal -P 3306 -u deployer -p mydb
```

## Backups

See [Backups](/guide/db-backups) for creating and restoring MySQL backups.

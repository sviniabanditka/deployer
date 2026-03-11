# MongoDB

## Create a MongoDB database

```bash
deployer db create mongodb --name my-mongo
```

With a specific version:

```bash
curl -X POST https://api.deployer.dev/api/v1/databases \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-mongo",
    "engine": "mongodb",
    "version": "7.0"
  }'
```

Available versions: `6.0`, `7.0` (default: latest).

## Connection details

```json
{
  "host": "db-ghi789.internal",
  "port": 27017,
  "username": "deployer",
  "password": "auto-generated",
  "database_name": "mydb",
  "connection_url": "mongodb://deployer:auto-generated@db-ghi789.internal:27017/mydb"
}
```

## Connect from your app

Link the database:

```bash
deployer db link <db-id> <app-id>
```

```js
// Node.js (mongoose)
const mongoose = require('mongoose')
await mongoose.connect(process.env.DATABASE_URL)
```

```python
# Python (pymongo)
from pymongo import MongoClient
client = MongoClient(os.environ['DATABASE_URL'])
```

```go
// Go (mongo-driver)
client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("DATABASE_URL")))
```

## Connect with mongosh

```bash
mongosh "mongodb://deployer:password@db-ghi789.internal:27017/mydb"
```

## Backups

See [Backups](/guide/db-backups) for creating and restoring MongoDB backups.

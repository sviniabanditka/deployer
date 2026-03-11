# Redis

## Create a Redis instance

```bash
deployer db create redis --name my-redis
```

With a specific version:

```bash
curl -X POST https://api.deployer.dev/api/v1/databases \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-redis",
    "engine": "redis",
    "version": "7.2"
  }'
```

Available versions: `7.0`, `7.2` (default: latest).

## Connection details

```json
{
  "host": "db-jkl012.internal",
  "port": 6379,
  "password": "auto-generated",
  "connection_url": "redis://:auto-generated@db-jkl012.internal:6379"
}
```

## Connect from your app

Link the database:

```bash
deployer db link <db-id> <app-id>
```

When linked, the `REDIS_URL` environment variable is set on your app.

```js
// Node.js (ioredis)
const Redis = require('ioredis')
const redis = new Redis(process.env.REDIS_URL)
```

```python
# Python
import redis
r = redis.from_url(os.environ['REDIS_URL'])
```

```go
// Go (go-redis)
opt, _ := redis.ParseURL(os.Getenv("REDIS_URL"))
client := redis.NewClient(opt)
```

## Common use cases

- **Session storage** — store user sessions in Redis for fast access.
- **Caching** — cache database queries or API responses.
- **Job queues** — use with BullMQ, Celery, or Asynq.
- **Pub/Sub** — real-time messaging between services.

## Connect with redis-cli

```bash
redis-cli -h db-jkl012.internal -p 6379 -a 'password'
```

## Notes

- Redis data is stored in memory. Data may be lost on restart unless persistence is enabled.
- Deployer configures Redis with AOF persistence by default.
- Backups are supported. See [Backups](/guide/db-backups).

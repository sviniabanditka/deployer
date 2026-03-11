# Environment Variables

Environment variables let you configure your application without changing code. Use them for database URLs, API keys, feature flags, and other runtime configuration.

## Set variables with the CLI

```bash
# Set one or more variables
deployer env set DATABASE_URL=postgres://... SECRET_KEY=mysecret

# List all variables
deployer env list

# Remove a variable
deployer env unset SECRET_KEY
```

## Set variables with the API

```bash
curl -X PUT https://api.deployer.dev/api/v1/apps/{app_id}/env \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vars": {
      "DATABASE_URL": "postgres://user:pass@host:5432/db",
      "SECRET_KEY": "supersecret",
      "NODE_ENV": "production"
    }
  }'
```

::: warning
The `PUT` endpoint replaces all environment variables. Fetch current values first if you only want to add or update a subset.
:::

## Built-in variables

Deployer automatically injects these variables into every container:

| Variable | Description |
|----------|-------------|
| `PORT` | The port your app should listen on |

## Linked database variables

When you [link a database](/guide/databases) to your app, Deployer automatically sets the connection URL:

| Engine | Variable | Example |
|--------|----------|---------|
| PostgreSQL | `DATABASE_URL` | `postgres://user:pass@host:5432/dbname` |
| MySQL | `DATABASE_URL` | `mysql://user:pass@host:3306/dbname` |
| MongoDB | `DATABASE_URL` | `mongodb://user:pass@host:27017/dbname` |
| Redis | `REDIS_URL` | `redis://:pass@host:6379` |

## Best practices

- Never commit secrets to source control. Use environment variables instead.
- Use descriptive names: `STRIPE_SECRET_KEY` not `KEY1`.
- Keep values short and avoid multi-line content.
- Redeploy after changing variables — the new values take effect on the next container start.
- Use the web dashboard or CLI to manage variables; do not bake them into your Docker image.

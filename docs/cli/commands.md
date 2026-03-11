# CLI Command Reference

## Global flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `~/.deployer/config.json` |
| `--api-url` | API server URL | `http://localhost:3000/api/v1` |
| `-h, --help` | Show help | |

## Authentication

### `deployer register`

Create a new account interactively. Prompts for name, email, and password. Automatically logs you in after registration.

```bash
deployer register
```

### `deployer login`

Log in to an existing account. Prompts for email and password.

```bash
deployer login
```

## App management

### `deployer init`

Initialize a new app in the current directory. Creates the app on the server and writes a `.deployer.json` file locally.

```bash
deployer init
```

Prompts for app name (defaults to directory name). If the directory is already linked to an app, asks for confirmation before creating a new one.

### `deployer deploy`

Deploy the current directory. Archives the project, uploads it, and starts a build.

```bash
deployer deploy
```

The CLI polls the deployment status and displays progress.

### `deployer apps`

Manage applications. Defaults to `deployer apps list`.

#### `deployer apps list`

List all apps in your account.

```bash
deployer apps list
```

```
ID                                    NAME          SLUG          STATUS
--                                    ----          ----          ------
550e8400-...                          my-app        my-app        running
661f9500-...                          api-service   api-service   stopped
```

#### `deployer apps info [app-id]`

Show details for an app. Uses the current directory's app if no ID is given.

```bash
deployer apps info
deployer apps info 550e8400-...
```

#### `deployer apps delete [app-id]`

Delete an app. Prompts for confirmation.

```bash
deployer apps delete 550e8400-...
```

### `deployer start`

Start the app linked to the current directory (uses the latest deployment).

```bash
deployer start
```

### `deployer stop`

Stop the app linked to the current directory.

```bash
deployer stop
```

### `deployer logs`

View application logs.

```bash
deployer logs
deployer logs -f    # stream logs in real time
```

| Flag | Description |
|------|-------------|
| `-f, --follow` | Stream logs via WebSocket |

### `deployer scale`

Scale application processes (not yet available).

```bash
deployer scale web=3
```

## Environment variables

### `deployer env set`

Set one or more environment variables.

```bash
deployer env set KEY=VALUE
deployer env set DB_HOST=localhost DB_PORT=5432
```

### `deployer env list`

List all environment variables for the current app.

```bash
deployer env list
```

### `deployer env unset`

Remove one or more environment variables.

```bash
deployer env unset KEY1 KEY2
```

## Databases

### `deployer db create <engine>`

Create a managed database. Supported engines: `postgres`, `mysql`, `mongodb`, `redis`.

```bash
deployer db create postgres
deployer db create mysql --name my-mysql --app <app-id>
```

| Flag | Description |
|------|-------------|
| `--name` | Database name (auto-generated if omitted) |
| `--app` | Link to an app on creation |

### `deployer db list`

List all databases.

```bash
deployer db list
```

### `deployer db info <id>`

Show database details and connection information.

```bash
deployer db info <db-id>
```

### `deployer db delete <id>`

Delete a database. Prompts for confirmation.

```bash
deployer db delete <db-id>
```

### `deployer db stop <id>`

Stop a database container (data is preserved).

```bash
deployer db stop <db-id>
```

### `deployer db start <id>`

Start a stopped database.

```bash
deployer db start <db-id>
```

### `deployer db link <db-id> <app-id>`

Link a database to an app. Sets `DATABASE_URL` or `REDIS_URL` on the app.

```bash
deployer db link <db-id> <app-id>
```

### `deployer db unlink <db-id>`

Unlink a database from its app.

```bash
deployer db unlink <db-id>
```

### `deployer db backup <id>`

Create a backup of a database.

```bash
deployer db backup <db-id>
```

### `deployer db backups <id>`

List all backups for a database.

```bash
deployer db backups <db-id>
```

### `deployer db restore <db-id> <backup-id>`

Restore a database from a backup. Prompts for confirmation.

```bash
deployer db restore <db-id> <backup-id>
```

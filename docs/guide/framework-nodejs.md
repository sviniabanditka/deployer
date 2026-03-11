# Node.js / Express

## Project setup

Deployer detects Node.js projects by the presence of `package.json`. Ensure you have a `start` script:

```json
{
  "name": "my-app",
  "scripts": {
    "start": "node index.js"
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}
```

## Port binding

Your app must listen on the port provided by the `PORT` environment variable:

```js
const express = require('express')
const app = express()
const port = process.env.PORT || 3000

app.get('/', (req, res) => {
  res.json({ status: 'ok' })
})

app.listen(port, '0.0.0.0', () => {
  console.log(`Listening on port ${port}`)
})
```

::: warning
Always bind to `0.0.0.0`, not `127.0.0.1` or `localhost`. Container networking requires the app to accept connections on all interfaces.
:::

## Deploy

```bash
deployer init
deployer deploy
```

Deployer runs `npm ci --production` during the build and uses `npm start` to launch your app.

## Using TypeScript

Compile TypeScript before deploying, or add a build step:

```json
{
  "scripts": {
    "build": "tsc",
    "start": "node dist/index.js"
  }
}
```

Deployer runs `npm run build` automatically if a `build` script is present.

## Using a custom Dockerfile

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --production
COPY . .
EXPOSE 3000
CMD ["node", "index.js"]
```

## Environment variables

```bash
deployer env set NODE_ENV=production SESSION_SECRET=abc123
```

## Adding a database

```bash
deployer db create postgres --name my-db
deployer db link <db-id> <app-id>
```

This sets `DATABASE_URL` automatically. Use it with your ORM:

```js
// Sequelize
const sequelize = new Sequelize(process.env.DATABASE_URL)

// Knex
const knex = require('knex')({
  client: 'pg',
  connection: process.env.DATABASE_URL
})
```

## Common issues

| Problem | Solution |
|---------|----------|
| `ERR_MODULE_NOT_FOUND` | Ensure all dependencies are in `dependencies`, not only `devDependencies` |
| App crashes on start | Check `deployer logs` for the error message |
| Port timeout | Make sure your app reads `process.env.PORT` |

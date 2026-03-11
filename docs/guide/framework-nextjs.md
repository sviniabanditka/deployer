# Next.js

## Project setup

Deployer supports Next.js with standalone output mode for optimal container size.

### next.config.js

```js
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
}

module.exports = nextConfig
```

### package.json

```json
{
  "scripts": {
    "build": "next build",
    "start": "node .next/standalone/server.js"
  }
}
```

## Deploy

```bash
deployer init
deployer deploy
```

Deployer runs `npm ci`, then `npm run build`, and starts the app with `npm start`.

## Using a custom Dockerfile

For production deployments, a multi-stage Dockerfile is recommended:

```dockerfile
FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

ENV PORT=3000
EXPOSE 3000
CMD ["node", "server.js"]
```

## Environment variables

Next.js distinguishes between build-time and runtime environment variables:

- **Build-time** (`NEXT_PUBLIC_*`): Baked into the client bundle during `next build`.
- **Runtime** (`process.env.*`): Available in API routes and server components.

Set build-time variables before deploying:

```bash
deployer env set NEXT_PUBLIC_API_URL=https://api.example.com
deployer deploy
```

## Adding a database

```bash
deployer db create postgres
deployer db link <db-id> <app-id>
```

Use `DATABASE_URL` in API routes or server actions:

```js
// app/api/users/route.js
import { Pool } from 'pg'

const pool = new Pool({ connectionString: process.env.DATABASE_URL })

export async function GET() {
  const { rows } = await pool.query('SELECT * FROM users')
  return Response.json(rows)
}
```

## Common issues

| Problem | Solution |
|---------|----------|
| Missing static files | Copy `.next/static` and `public` in your Dockerfile |
| `NEXT_PUBLIC_*` not working | These are embedded at build time; set them before `deployer deploy` |
| Port mismatch | Ensure standalone server reads `process.env.PORT` |

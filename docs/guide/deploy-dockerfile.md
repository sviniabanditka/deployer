# Deploy with Dockerfile

For full control over your build, include a `Dockerfile` in your project root. Deployer will use it instead of auto-detecting the project type.

## Basic example

```dockerfile
FROM node:20-alpine

WORKDIR /app
COPY package*.json ./
RUN npm ci --production
COPY . .

EXPOSE 3000
CMD ["node", "index.js"]
```

Deploy as usual:

```bash
deployer deploy
```

## Requirements

Your Dockerfile must:

1. **Expose a port** — use the `EXPOSE` instruction or bind to `$PORT`.
2. **Listen on `0.0.0.0`** — do not bind to `127.0.0.1` or `localhost`.
3. **Use `CMD` or `ENTRYPOINT`** — Deployer needs to know how to start your app.

Deployer injects the `PORT` environment variable at runtime. Your app should read it:

```js
const port = process.env.PORT || 3000
```

## Multi-stage builds

Use multi-stage builds to keep images small:

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server .

# Run stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

## Build arguments

Deployer does not currently support custom `--build-arg` values. Use environment variables at runtime instead.

## .dockerignore

Include a `.dockerignore` file to keep your build context small:

```
.git
node_modules
*.md
.env
.deployer.json
```

## Tips

- Use `alpine`-based images to reduce build time and image size.
- Pin exact versions (`node:20.11-alpine` not `node:latest`).
- Place rarely-changing layers (like dependency install) before frequently-changing ones (like `COPY . .`).
- The build runs with a 10-minute timeout. Optimize slow builds by caching dependencies.

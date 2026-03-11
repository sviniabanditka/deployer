# What is Deployer?

Deployer is an open-source deployment platform that makes it simple to deploy and manage web applications. Upload a ZIP file, connect a Git repository, or provide a Dockerfile — Deployer builds, deploys, and runs your application with automatic SSL, subdomain routing, and monitoring out of the box.

## Who is it for?

- **Solo developers** who want to ship fast without managing servers.
- **Small teams** that need a self-hosted alternative to Heroku, Railway, or Render.
- **Agencies** deploying multiple client projects on shared infrastructure.
- **Learners** exploring deployment, containers, and infrastructure.

## Key Features

| Feature | Description |
|---------|-------------|
| **ZIP / Git / Dockerfile deploy** | Multiple deployment methods to suit your workflow |
| **Managed databases** | PostgreSQL, MySQL, MongoDB, Redis with one-click creation |
| **Auto-SSL** | Free TLS certificates via Traefik for every app |
| **Subdomain routing** | Each app gets `<slug>.deployer.dev` automatically |
| **Environment variables** | Securely manage config via UI or CLI |
| **Real-time logs** | Stream container logs over WebSocket |
| **CPU/RAM monitoring** | Per-container resource metrics |
| **CLI** | Full-featured command-line tool for all operations |
| **REST API** | Programmatic access to every feature |
| **2FA & GDPR** | Two-factor authentication, data export, account deletion |
| **Billing & plans** | Stripe-powered subscriptions with usage-based quotas |

## How It Compares

| | Deployer | Heroku | Railway | Render |
|--|---------|--------|---------|--------|
| Self-hosted | Yes | No | No | No |
| Open source | Yes | No | No | No |
| ZIP deploy | Yes | No | No | No |
| Git deploy | Yes | Yes | Yes | Yes |
| Managed DBs | Yes | Yes | Yes | Yes |
| Free tier | Yes | Limited | Limited | Yes |
| CLI | Yes | Yes | Yes | No |
| EU hosting | Yes | Add-on | No | EU region |

## Supported Languages & Frameworks

Deployer auto-detects your project type and builds it in a Docker container. Supported out of the box:

- **Node.js** — Express, Fastify, Koa, NestJS
- **Python** — Django, Flask, FastAPI
- **Go** — Standard library, Gin, Echo, Fiber
- **Next.js** — App Router & Pages Router with standalone output
- **PHP / Laravel** — Composer-based projects
- **Static sites** — HTML, CSS, JS served via nginx
- **Any Dockerfile** — Full control when you need it

## Architecture

Deployer consists of three components:

1. **API server** — Go (Fiber), handles REST endpoints and WebSocket log streaming
2. **Web dashboard** — Vue.js SPA for managing apps, databases, and billing
3. **CLI** — Go (Cobra), for terminal-based workflows

Infrastructure runs on Docker with Traefik as the reverse proxy, PostgreSQL for data storage, and Redis for task queues and rate limiting.

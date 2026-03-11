# Getting Started

This guide walks you through setting up Deployer and deploying your first application.

## Prerequisites

- A Deployer account (cloud or self-hosted instance)
- [Node.js](https://nodejs.org/) 18+ (for web projects) or your language runtime
- A terminal with `curl` available

## Step 1: Install the CLI

```bash
curl -fsSL https://get.deployer.dev | sh
```

This installs the `deployer` binary to `/usr/local/bin`. Verify the installation:

```bash
deployer --help
```

### Manual installation

Download the binary for your platform from the [GitHub releases](https://github.com/deployer/cli/releases) page and place it in your `PATH`.

## Step 2: Create an account

You can register via the web dashboard or the CLI:

```bash
deployer register
```

You will be prompted for your name, email, and password. After registration you are automatically logged in.

If you already have an account:

```bash
deployer login
```

## Step 3: Create an app

Navigate to your project directory and initialize a new app:

```bash
cd my-project
deployer init
```

This creates a `.deployer.json` file in your project root containing the app ID. The app name defaults to the directory name, and a unique subdomain is assigned automatically (e.g., `my-project.deployer.dev`).

## Step 4: Deploy

```bash
deployer deploy
```

The CLI archives your project, uploads it, and triggers a build. Deployer auto-detects your project type:

- **Node.js** — looks for `package.json`
- **Python** — looks for `requirements.txt` or `Pipfile`
- **Go** — looks for `go.mod`
- **Dockerfile** — uses your `Dockerfile` directly

The build output streams in your terminal. Once complete, your app is live at the subdomain shown in the output.

## Step 5: View your app

```bash
deployer apps info
```

Open the URL in your browser. Your app is running with automatic SSL.

## Next Steps

- [Quick Start](/guide/quick-start) — deploy a hello world app in 5 minutes
- [Environment Variables](/guide/env-vars) — configure your app
- [Managed Databases](/guide/databases) — add PostgreSQL, MySQL, MongoDB, or Redis
- [Git Deployments](/guide/deploy-git) — auto-deploy on push

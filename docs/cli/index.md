# CLI Installation

The Deployer CLI lets you deploy, manage, and monitor applications from your terminal.

## Quick install

```bash
curl -fsSL https://get.deployer.dev | sh
```

This detects your OS and architecture, downloads the correct binary, and installs it to `/usr/local/bin/deployer`.

## macOS

### Homebrew (coming soon)

```bash
brew install deployer/tap/deployer
```

### Manual

```bash
curl -L https://github.com/deployer/cli/releases/latest/download/deployer-darwin-arm64 -o deployer
chmod +x deployer
sudo mv deployer /usr/local/bin/
```

For Intel Macs, use `deployer-darwin-amd64`.

## Linux

### Script

```bash
curl -fsSL https://get.deployer.dev | sh
```

### Manual

```bash
curl -L https://github.com/deployer/cli/releases/latest/download/deployer-linux-amd64 -o deployer
chmod +x deployer
sudo mv deployer /usr/local/bin/
```

For ARM64 (e.g., Raspberry Pi 4), use `deployer-linux-arm64`.

## Windows

Download `deployer-windows-amd64.exe` from the [releases page](https://github.com/deployer/cli/releases) and add it to your PATH.

Or use PowerShell:

```powershell
Invoke-WebRequest -Uri "https://github.com/deployer/cli/releases/latest/download/deployer-windows-amd64.exe" -OutFile "deployer.exe"
Move-Item deployer.exe "$env:USERPROFILE\bin\deployer.exe"
```

## Verify installation

```bash
deployer --help
```

Expected output:

```
Deployer CLI - Build, deploy, and manage your applications.

Usage:
  deployer [command]

Available Commands:
  register    Create a new Deployer account
  login       Log in to your Deployer account
  init        Initialize a new app in the current directory
  deploy      Deploy the current directory to your app
  apps        Manage your applications
  logs        View application logs
  env         Manage environment variables
  db          Manage databases
  start       Start the application
  stop        Stop the application
  scale       Scale application processes

Flags:
      --api-url string   API server URL (default "http://localhost:3000/api/v1")
      --config string    config file (default "~/.deployer/config.json")
  -h, --help             help for deployer
```

## Configuration

The CLI stores its configuration at `~/.deployer/config.json`. This file contains:

- API URL
- Access token
- Refresh token
- Email

You can override the API URL per command:

```bash
deployer --api-url https://api.deployer.dev/api/v1 apps list
```

## Self-hosted API

If you run a self-hosted Deployer instance, point the CLI to your API:

```bash
deployer --api-url https://your-instance.com/api/v1 login
```

After login, the URL is saved in the config file and used for all subsequent commands.

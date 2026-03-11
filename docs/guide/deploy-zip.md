# Deploy with ZIP

The simplest way to deploy: upload a ZIP archive containing your application code.

## Using the CLI

The CLI handles archiving automatically. From your project directory:

```bash
deployer init    # first time only
deployer deploy
```

The CLI creates a ZIP of the current directory (excluding common patterns like `node_modules`, `.git`, `__pycache__`), uploads it, and starts the build.

### Excluded files

The following patterns are excluded by default:

- `.git/`
- `node_modules/`
- `__pycache__/`
- `.env`
- `.deployer.json`
- `*.log`

## Using the API

You can also deploy by sending a ZIP file directly to the API:

```bash
# Create a ZIP of your project
zip -r app.zip . -x "node_modules/*" ".git/*"

# Deploy
curl -X POST https://api.deployer.dev/api/v1/apps/{app_id}/deploy \
  -H "Authorization: Bearer $TOKEN" \
  -F "archive=@app.zip"
```

Response:

```json
{
  "message": "deployment queued",
  "deployment": {
    "id": "d1234567-...",
    "app_id": "a1234567-...",
    "version": 3,
    "status": "pending",
    "image_tag": "registry.deployer.dev/my-app:3"
  }
}
```

## Using the Web UI

1. Open your app in the dashboard.
2. Click **Deploy**.
3. Drag and drop a ZIP file or click to browse.
4. The build log streams in real time.

## Build process

After upload, Deployer:

1. Extracts the ZIP archive.
2. Detects the project type (Node.js, Python, Go, Dockerfile, static).
3. Generates a Dockerfile if one is not present.
4. Builds a container image.
5. Pushes the image to the internal registry.
6. Stops the old container (if any) and starts the new one.
7. Configures Traefik routing and SSL.

## Tips

- Keep your ZIP small. Exclude build artifacts and dependencies.
- Include a `Dockerfile` for full control over the build.
- Use `.deployerignore` (same syntax as `.gitignore`) to control what gets uploaded.
- Maximum upload size depends on your plan (100 MB on Free, 500 MB on Pro).

# API Reference — Deployments

Deployments are created when you upload a ZIP archive or push to a connected Git repository.

## Create a deployment

Deployments are created via the app deploy endpoint:

```
POST /apps/:id/deploy
```

```bash
curl -X POST https://api.deployer.dev/api/v1/apps/{app_id}/deploy \
  -H "Authorization: Bearer $TOKEN" \
  -F "archive=@app.zip"
```

**Response** `202 Accepted`:

```json
{
  "message": "deployment queued",
  "deployment": {
    "id": "d1234567-e89b-12d3-a456-426614174000",
    "app_id": "550e8400-e29b-41d4-a716-446655440000",
    "version": 5,
    "status": "pending",
    "image_tag": "registry.deployer.dev/my-app:5"
  }
}
```

## Deployment lifecycle

A deployment moves through these statuses:

| Status | Description |
|--------|-------------|
| `pending` | Deployment created, waiting for build |
| `building` | Container image is being built |
| `pushing` | Image pushed to internal registry |
| `deploying` | Old container stopped, new container starting |
| `running` | Deployment is live |
| `failed` | Build or deploy failed |

## How it works

1. You upload a ZIP archive to `POST /apps/:id/deploy`.
2. Deployer saves the archive and creates a deployment record with `pending` status.
3. A build task is enqueued to the async worker (powered by Asynq/Redis).
4. The worker extracts the archive, detects the project type, and builds a Docker image.
5. The image is pushed to the internal container registry.
6. The old container (if any) is stopped, and the new container is started with Traefik labels for routing.
7. Status updates to `running` on success or `failed` on error.

## Git-triggered deployments

When a connected repository receives a push to the configured branch:

1. GitHub/GitLab sends a webhook to Deployer.
2. Deployer verifies the webhook signature.
3. The repository is cloned and built.
4. The deployment follows the same lifecycle as ZIP deployments.

Webhook endpoints:

| Provider | URL |
|----------|-----|
| GitHub | `POST /webhooks/github` |
| GitLab | `POST /webhooks/gitlab` |

## Versioning

Each deployment increments the version number for the app. The image tag follows the pattern:

```
<registry_url>/<app_slug>:<version>
```

Example: `registry.deployer.dev/my-app:5`

## Build logs

Build logs are available through the app logs endpoint:

```bash
curl https://api.deployer.dev/api/v1/apps/{app_id}/logs \
  -H "Authorization: Bearer $TOKEN"
```

Or stream in real time via WebSocket:

```
ws://api.deployer.dev/api/v1/apps/{app_id}/logs
```

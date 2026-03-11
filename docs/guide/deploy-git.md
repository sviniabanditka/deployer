# Deploy with Git

Connect a GitHub or GitLab repository to auto-deploy on every push.

## Connect a repository

### Using the API

```bash
curl -X POST https://api.deployer.dev/api/v1/apps/{app_id}/git/connect \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "github",
    "repo_url": "https://github.com/yourname/yourrepo",
    "branch": "main",
    "access_token": "ghp_xxxxxxxxxxxx"
  }'
```

Supported providers: `github`, `gitlab`.

### Parameters

| Field | Required | Description |
|-------|----------|-------------|
| `provider` | Yes | `github` or `gitlab` |
| `repo_url` | Yes | Full HTTPS URL of the repository |
| `branch` | No | Branch to deploy (defaults to `main`) |
| `access_token` | Yes | Personal access token with repo read access |

## How auto-deploy works

1. When you connect a repo, Deployer registers a webhook on your repository.
2. On every push to the configured branch, the webhook fires.
3. Deployer clones the repo, builds the container image, and deploys it.

The webhook URL is:
- GitHub: `https://api.deployer.dev/api/v1/webhooks/github`
- GitLab: `https://api.deployer.dev/api/v1/webhooks/gitlab`

## Check connection status

```bash
curl https://api.deployer.dev/api/v1/apps/{app_id}/git \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "git_connection": {
    "id": "...",
    "app_id": "...",
    "provider": "github",
    "repo_url": "https://github.com/yourname/yourrepo",
    "branch": "main",
    "webhook_active": true,
    "last_deploy_at": "2026-03-10T14:30:00Z"
  }
}
```

## Disconnect a repository

```bash
curl -X DELETE https://api.deployer.dev/api/v1/apps/{app_id}/git/disconnect \
  -H "Authorization: Bearer $TOKEN"
```

This removes the webhook and stops auto-deploys. Existing deployments are not affected.

## Choosing a branch

You can only track one branch per app. To deploy multiple branches, create separate apps:

```bash
# Production
deployer init --name my-app-prod
# Connect to "main" branch

# Staging
deployer init --name my-app-staging
# Connect to "develop" branch
```

## Access tokens

### GitHub

Create a [fine-grained personal access token](https://github.com/settings/tokens?type=beta) with:
- **Repository access**: Only select repositories
- **Permissions**: Contents (read-only)

### GitLab

Create a [project access token](https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html) with:
- **Role**: Reporter
- **Scopes**: `read_repository`

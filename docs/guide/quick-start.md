# Quick Start

Deploy a Node.js application in under 5 minutes.

## 1. Install the CLI and log in

```bash
curl -fsSL https://get.deployer.dev | sh
deployer register   # or: deployer login
```

## 2. Create a project

```bash
mkdir hello-deployer && cd hello-deployer
npm init -y
npm install express
```

Create `index.js`:

```js
const express = require('express')
const app = express()
const port = process.env.PORT || 3000

app.get('/', (req, res) => {
  res.json({ message: 'Hello from Deployer!' })
})

app.listen(port, () => {
  console.log(`Server running on port ${port}`)
})
```

Add a start script to `package.json`:

```json
{
  "scripts": {
    "start": "node index.js"
  }
}
```

## 3. Initialize and deploy

```bash
deployer init
deployer deploy
```

Output:

```
Uploading... done
Building... done
Deployed successfully! Deployment ID: abc123

App 'hello-deployer' is live at https://hello-deployer.deployer.dev
```

## 4. Check status

```bash
deployer apps info
```

```
ID:        550e8400-e29b-41d4-a716-446655440000
Name:      hello-deployer
Slug:      hello-deployer
Status:    running
URL:       https://hello-deployer.deployer.dev
```

## 5. View logs

```bash
deployer logs
# or stream in real time:
deployer logs -f
```

## What just happened?

1. `deployer init` created the app on the server and saved the app ID locally.
2. `deployer deploy` zipped your project, uploaded it, and triggered a build.
3. Deployer detected a Node.js project, installed dependencies, built a container image, and started it.
4. Traefik automatically configured a subdomain and SSL certificate.

## Next steps

- Add a [database](/guide/databases)
- Set [environment variables](/guide/env-vars)
- Connect a [Git repository](/guide/deploy-git) for auto-deploy

# Static Sites

Deploy static HTML, CSS, and JavaScript sites. Deployer serves them with nginx.

## Project structure

```
my-site/
  index.html
  style.css
  script.js
  images/
    logo.png
```

No `package.json`, `go.mod`, or `requirements.txt` needed. Deployer auto-detects static projects when it finds an `index.html` in the root.

## Deploy

```bash
deployer init
deployer deploy
```

Deployer creates a lightweight nginx container to serve your files.

## Build tools (Vite, Hugo, Jekyll)

If your site requires a build step, include a `Dockerfile`:

### Vite

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Hugo

```dockerfile
FROM klakegg/hugo:0.121.0-ext-alpine AS builder
WORKDIR /src
COPY . .
RUN hugo --minify

FROM nginx:alpine
COPY --from=builder /src/public /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Single-page applications (SPA)

For SPAs with client-side routing, add an nginx config that redirects all requests to `index.html`:

Create `nginx.conf`:

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

Reference it in your Dockerfile:

```dockerfile
FROM nginx:alpine
COPY dist/ /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Custom 404 page

Place a `404.html` file in your project root. Deployer's default nginx config will serve it for missing pages.

## Tips

- Use a CDN in front of Deployer for global caching.
- Minify assets before deploying to reduce transfer size.
- Set long cache headers for hashed filenames (e.g., `app.a1b2c3.js`).

# Custom Domains

::: tip Coming Soon
Custom domain support is currently under development. This page describes the planned functionality.
:::

Every app deployed on Deployer gets a free subdomain at `<slug>.deployer.dev` with automatic SSL. Custom domain support will allow you to use your own domain name (e.g., `app.example.com`).

## Planned workflow

1. Add your domain in the dashboard or via the API.
2. Create a CNAME record pointing to your app's subdomain.
3. Deployer verifies DNS and provisions a TLS certificate.
4. Traffic to your custom domain is routed to your app.

## DNS configuration (planned)

```
Type:  CNAME
Name:  app
Value: my-app.deployer.dev
```

For root domains (apex), an `A` record pointing to Deployer's load balancer IP will be required.

## Timeline

Custom domains are on the roadmap for Q3 2026. Follow the [GitHub repository](https://github.com/deployer) for updates.

## Current workaround

If you need a custom domain now, set up a reverse proxy (e.g., Cloudflare, nginx) that forwards traffic to your `<slug>.deployer.dev` subdomain.

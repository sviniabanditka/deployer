# Pricing

Start free, upgrade when you need more.

## Plans

| | Free | Pro | Business |
|--|------|-----|----------|
| **Price** | $0/month | $15/month | $50/month |
| **Apps** | 2 | 10 | Unlimited |
| **Databases** | 1 | 5 | Unlimited |
| **Memory per app** | 256 MB | 1 GB | 4 GB |
| **CPU per app** | 0.5 vCPU | 2 vCPU | 4 vCPU |
| **Storage** | 1 GB | 10 GB | 100 GB |
| **Custom domains** | No | Yes | Yes |
| **Auto backups** | No | Daily (7-day retention) | Daily (30-day retention) |
| **Support** | Community | Email | Priority |
| **SSL** | Yes | Yes | Yes |
| **Monitoring** | Yes | Yes | Yes |

## Free tier

The free plan is perfect for side projects, demos, and learning. It includes:

- 2 apps
- 1 database
- 256 MB RAM per app
- 1 GB total storage
- Automatic SSL and subdomain

No credit card required.

## Upgrading

Upgrade from the dashboard or via the API:

```bash
curl -X POST https://api.deployer.dev/api/v1/billing/subscribe \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "plan": "pro" }'
```

Payment is handled by Stripe. You can manage your payment methods and invoices through the billing portal:

```bash
curl https://api.deployer.dev/api/v1/billing/portal \
  -H "Authorization: Bearer $TOKEN"
```

## Changing plans

Upgrade or downgrade at any time. Changes take effect immediately:

- **Upgrade**: New limits apply immediately. You are charged a prorated amount.
- **Downgrade**: New limits apply at the start of the next billing period.

## Cancellation

Cancel your subscription at any time. Your plan remains active until the end of the current billing period:

```bash
curl -X POST https://api.deployer.dev/api/v1/billing/cancel \
  -H "Authorization: Bearer $TOKEN"
```

After the period ends, your account reverts to the Free plan. Apps and databases that exceed Free tier limits are stopped (not deleted).

## Usage tracking

Check your current usage:

```bash
curl https://api.deployer.dev/api/v1/billing/usage \
  -H "Authorization: Bearer $TOKEN"
```

```json
{
  "usage": {
    "app_count": 3,
    "app_limit": 10,
    "db_count": 2,
    "db_limit": 5,
    "storage_used": 524288000,
    "storage_max": 10737418240
  }
}
```

## FAQ

**Can I try Pro features before paying?**
Not currently. The Free tier is generous enough to evaluate all core features.

**What happens if I exceed my limits?**
You will not be able to create new apps or databases until you upgrade or delete existing resources.

**Is there a yearly discount?**
Not yet. Contact us if you are interested in annual billing.

**Do you offer discounts for open-source projects?**
Yes. Contact support with a link to your repository.

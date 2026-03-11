# API Reference — Billing

All billing endpoints require authentication. Billing is powered by Stripe.

## List plans

```
GET /billing/plans
```

```bash
curl https://api.deployer.dev/api/v1/billing/plans \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "plans": [
    {
      "id": "...",
      "name": "free",
      "display_name": "Free",
      "price_cents": 0,
      "app_limit": 2,
      "db_limit": 1,
      "memory_limit": 268435456,
      "cpu_limit": 0.5,
      "storage_limit": 1073741824,
      "custom_domains": false,
      "priority_support": false
    },
    {
      "id": "...",
      "name": "pro",
      "display_name": "Pro",
      "price_cents": 1500,
      "app_limit": 10,
      "db_limit": 5,
      "memory_limit": 1073741824,
      "cpu_limit": 2.0,
      "storage_limit": 10737418240,
      "custom_domains": true,
      "priority_support": false
    }
  ]
}
```

## Get current subscription

```
GET /billing/subscription
```

```bash
curl https://api.deployer.dev/api/v1/billing/subscription \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "plan": {
    "name": "pro",
    "display_name": "Pro",
    "price_cents": 1500
  },
  "subscription": {
    "id": "...",
    "status": "active",
    "current_period_start": "2026-03-01T00:00:00Z",
    "current_period_end": "2026-04-01T00:00:00Z",
    "cancel_at_period_end": false
  }
}
```

## Subscribe to a plan

```
POST /billing/subscribe
```

```bash
curl -X POST https://api.deployer.dev/api/v1/billing/subscribe \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "plan": "pro" }'
```

**Response** `201 Created`:

```json
{
  "subscription": { "id": "...", "status": "active" },
  "client_secret": "pi_xxx_secret_xxx"
}
```

The `client_secret` is used with Stripe.js on the frontend to complete payment.

## Change plan

```
POST /billing/change-plan
```

```bash
curl -X POST https://api.deployer.dev/api/v1/billing/change-plan \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "plan": "business" }'
```

**Response** `200 OK`:

```json
{ "message": "plan changed" }
```

## Cancel subscription

Cancels at the end of the current billing period.

```
POST /billing/cancel
```

```bash
curl -X POST https://api.deployer.dev/api/v1/billing/cancel \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{ "message": "subscription will be canceled at end of billing period" }
```

## Resume subscription

Undo a pending cancellation.

```
POST /billing/resume
```

```bash
curl -X POST https://api.deployer.dev/api/v1/billing/resume \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{ "message": "subscription resumed" }
```

## Billing portal

Get a URL to the Stripe Customer Portal for managing payment methods and invoices.

```
GET /billing/portal
```

```bash
curl https://api.deployer.dev/api/v1/billing/portal \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{ "url": "https://billing.stripe.com/p/session/..." }
```

## List invoices

```
GET /billing/invoices
```

```bash
curl https://api.deployer.dev/api/v1/billing/invoices \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

```json
{
  "invoices": [
    {
      "id": "...",
      "amount_cents": 1500,
      "currency": "usd",
      "status": "paid",
      "invoice_url": "https://pay.stripe.com/invoice/...",
      "created_at": "2026-03-01T00:00:00Z"
    }
  ]
}
```

## Get usage

```
GET /billing/usage
```

```bash
curl https://api.deployer.dev/api/v1/billing/usage \
  -H "Authorization: Bearer $TOKEN"
```

**Response** `200 OK`:

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

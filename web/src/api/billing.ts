import client from './client'

export interface Plan {
  name: string
  displayName: string
  priceEur: number
  appLimit: number
  dbLimit: number
  memoryMb: number
  cpuCores: number
  storageGb: number
  customDomains: boolean
  prioritySupport: boolean
  features: string[]
}

export interface Subscription {
  id: string
  planName: string
  status: string
  currentPeriodStart: string
  currentPeriodEnd: string
  cancelAtPeriodEnd: boolean
}

export interface Invoice {
  id: string
  date: string
  description: string
  amountEur: number
  status: string
  stripeInvoiceUrl?: string
}

export interface UsageSummary {
  planName: string
  periodStart: string
  periodEnd: string
  apps: { used: number; limit: number }
  databases: { used: number; limit: number }
  memoryMb: { used: number; limit: number }
  storageGb: { used: number; limit: number }
}

export function fetchPlans() {
  return client.get<Plan[]>('/billing/plans')
}

export function fetchSubscription() {
  return client.get<{ plan: Plan; subscription: Subscription }>('/billing/subscription')
}

export function subscribe(planName: string) {
  return client.post<{ clientSecret: string; subscriptionId: string }>('/billing/subscribe', { planName })
}

export function changePlan(planName: string) {
  return client.post<void>('/billing/change-plan', { planName })
}

export function cancelSubscription() {
  return client.post<void>('/billing/cancel')
}

export function resumeSubscription() {
  return client.post<void>('/billing/resume')
}

export function getBillingPortalUrl() {
  return client.post<{ url: string }>('/billing/portal')
}

export function fetchInvoices() {
  return client.get<Invoice[]>('/billing/invoices')
}

export function fetchUsage() {
  return client.get<UsageSummary>('/billing/usage')
}

export interface BillingAddress {
  country: string
  postalCode: string
  vatNumber?: string
}

export function updateBillingAddress(data: BillingAddress) {
  return client.put<void>('/billing/address', data)
}

export function fetchBillingAddress() {
  return client.get<BillingAddress>('/billing/address')
}

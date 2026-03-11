import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as billingApi from '../api/billing'
import type { Plan, Subscription, Invoice, UsageSummary } from '../api/billing'

export const useBillingStore = defineStore('billing', () => {
  const plans = ref<Plan[]>([])
  const currentPlan = ref<Plan | null>(null)
  const subscription = ref<Subscription | null>(null)
  const invoices = ref<Invoice[]>([])
  const usage = ref<UsageSummary | null>(null)
  const loading = ref(false)

  async function fetchPlans() {
    loading.value = true
    try {
      const { data } = await billingApi.fetchPlans()
      plans.value = data
    } finally {
      loading.value = false
    }
  }

  async function fetchSubscription() {
    loading.value = true
    try {
      const { data } = await billingApi.fetchSubscription()
      currentPlan.value = data.plan
      subscription.value = data.subscription
    } finally {
      loading.value = false
    }
  }

  async function subscribe(planName: string) {
    const { data } = await billingApi.subscribe(planName)
    return data
  }

  async function changePlan(planName: string) {
    await billingApi.changePlan(planName)
    await fetchSubscription()
  }

  async function cancelSubscription() {
    await billingApi.cancelSubscription()
    await fetchSubscription()
  }

  async function resumeSubscription() {
    await billingApi.resumeSubscription()
    await fetchSubscription()
  }

  async function openBillingPortal() {
    const { data } = await billingApi.getBillingPortalUrl()
    window.open(data.url, '_blank')
  }

  async function fetchInvoices() {
    loading.value = true
    try {
      const { data } = await billingApi.fetchInvoices()
      invoices.value = data
    } finally {
      loading.value = false
    }
  }

  async function fetchUsage() {
    loading.value = true
    try {
      const { data } = await billingApi.fetchUsage()
      usage.value = data
    } finally {
      loading.value = false
    }
  }

  return {
    plans,
    currentPlan,
    subscription,
    invoices,
    usage,
    loading,
    fetchPlans,
    fetchSubscription,
    subscribe,
    changePlan,
    cancelSubscription,
    resumeSubscription,
    openBillingPortal,
    fetchInvoices,
    fetchUsage,
  }
})

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { loadStripe } from '@stripe/stripe-js'
import { useBillingStore } from '../stores/billing'
import PlanCard from '../components/PlanCard.vue'
import UsageBar from '../components/UsageBar.vue'
import type { Plan, BillingAddress } from '../api/billing'
import { updateBillingAddress, fetchBillingAddress } from '../api/billing'

const billing = useBillingStore()
const activeTab = ref<'plans' | 'usage' | 'invoices'>('plans')
const showConfirmModal = ref(false)
const pendingPlan = ref<Plan | null>(null)
const processing = ref(false)

// Billing address state
const billingAddress = ref<BillingAddress>({ country: '', postalCode: '', vatNumber: '' })
const addressSaving = ref(false)
const addressSaved = ref(false)
const addressError = ref('')

const euCountries = [
  { code: 'AT', name: 'Austria' },
  { code: 'BE', name: 'Belgium' },
  { code: 'BG', name: 'Bulgaria' },
  { code: 'HR', name: 'Croatia' },
  { code: 'CY', name: 'Cyprus' },
  { code: 'CZ', name: 'Czech Republic' },
  { code: 'DK', name: 'Denmark' },
  { code: 'EE', name: 'Estonia' },
  { code: 'FI', name: 'Finland' },
  { code: 'FR', name: 'France' },
  { code: 'DE', name: 'Germany' },
  { code: 'GR', name: 'Greece' },
  { code: 'HU', name: 'Hungary' },
  { code: 'IE', name: 'Ireland' },
  { code: 'IT', name: 'Italy' },
  { code: 'LV', name: 'Latvia' },
  { code: 'LT', name: 'Lithuania' },
  { code: 'LU', name: 'Luxembourg' },
  { code: 'MT', name: 'Malta' },
  { code: 'NL', name: 'Netherlands' },
  { code: 'PL', name: 'Poland' },
  { code: 'PT', name: 'Portugal' },
  { code: 'RO', name: 'Romania' },
  { code: 'SK', name: 'Slovakia' },
  { code: 'SI', name: 'Slovenia' },
  { code: 'ES', name: 'Spain' },
  { code: 'SE', name: 'Sweden' },
]

async function handleSaveAddress() {
  addressError.value = ''
  addressSaved.value = false
  addressSaving.value = true
  try {
    await updateBillingAddress(billingAddress.value)
    addressSaved.value = true
    setTimeout(() => { addressSaved.value = false }, 3000)
  } catch (e: any) {
    addressError.value = e.response?.data?.message || 'Failed to save billing address'
  } finally {
    addressSaving.value = false
  }
}

const tabs = [
  { key: 'plans' as const, label: 'Plans' },
  { key: 'usage' as const, label: 'Usage' },
  { key: 'invoices' as const, label: 'Invoices' },
]

const planOrder = ['free', 'starter', 'pro', 'business']

const currentPlanIndex = computed(() => {
  if (!billing.currentPlan) return 0
  return planOrder.indexOf(billing.currentPlan.name)
})

function isCurrentPlan(plan: Plan) {
  return billing.currentPlan?.name === plan.name
}

function isPopularPlan(plan: Plan) {
  return plan.name === 'pro'
}

async function handleSelectPlan(plan: Plan) {
  if (isCurrentPlan(plan)) return

  const targetIndex = planOrder.indexOf(plan.name)
  const isUpgrade = targetIndex > currentPlanIndex.value

  if (isUpgrade && currentPlanIndex.value === 0) {
    // Free -> paid: initiate Stripe checkout
    processing.value = true
    try {
      const { clientSecret } = await billing.subscribe(plan.name)
      const stripe = await loadStripe(import.meta.env.VITE_STRIPE_PUBLIC_KEY || '')
      if (stripe && clientSecret) {
        await stripe.confirmCardPayment(clientSecret)
        await billing.fetchSubscription()
      }
    } finally {
      processing.value = false
    }
  } else if (!isUpgrade) {
    // Downgrade: confirm first
    pendingPlan.value = plan
    showConfirmModal.value = true
  } else {
    // Upgrade between paid plans
    processing.value = true
    try {
      await billing.changePlan(plan.name)
    } finally {
      processing.value = false
    }
  }
}

async function confirmDowngrade() {
  if (!pendingPlan.value) return
  processing.value = true
  showConfirmModal.value = false
  try {
    await billing.changePlan(pendingPlan.value.name)
  } finally {
    processing.value = false
    pendingPlan.value = null
  }
}

function cancelDowngrade() {
  showConfirmModal.value = false
  pendingPlan.value = null
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function statusBadgeClass(status: string) {
  switch (status) {
    case 'paid':
      return 'bg-green-100 text-green-700'
    case 'open':
      return 'bg-yellow-100 text-yellow-700'
    case 'void':
    case 'uncollectible':
      return 'bg-red-100 text-red-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

onMounted(async () => {
  await Promise.all([
    billing.fetchPlans(),
    billing.fetchSubscription(),
    billing.fetchUsage(),
    billing.fetchInvoices(),
    fetchBillingAddress().then(({ data }) => {
      billingAddress.value = { country: data.country || '', postalCode: data.postalCode || '', vatNumber: data.vatNumber || '' }
    }).catch(() => {}),
  ])
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Billing</h1>
      <button
        @click="billing.openBillingPortal()"
        class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-indigo-600 bg-indigo-50 rounded-lg hover:bg-indigo-100 transition-colors"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
        </svg>
        Manage Billing
      </button>
    </div>

    <!-- Tabs -->
    <div class="border-b border-gray-200 mb-8">
      <nav class="flex gap-6">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          class="pb-3 text-sm font-medium transition-colors border-b-2"
          :class="
            activeTab === tab.key
              ? 'border-indigo-500 text-indigo-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
          "
        >
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <!-- Loading -->
    <div v-if="billing.loading" class="text-gray-500 py-12 text-center">Loading...</div>

    <!-- Plans tab -->
    <div v-else-if="activeTab === 'plans'">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        <PlanCard
          v-for="plan in billing.plans"
          :key="plan.name"
          :plan="plan"
          :is-current="isCurrentPlan(plan)"
          :is-popular="isPopularPlan(plan)"
          @select="handleSelectPlan"
        />
      </div>

      <!-- Subscription status -->
      <div v-if="billing.subscription" class="mt-8 rounded-xl bg-white shadow p-6">
        <h3 class="text-sm font-semibold text-gray-900 mb-3">Subscription Details</h3>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm">
          <div>
            <span class="text-gray-500">Status</span>
            <p class="mt-0.5 font-medium capitalize">{{ billing.subscription.status }}</p>
          </div>
          <div>
            <span class="text-gray-500">Current Period</span>
            <p class="mt-0.5 font-medium">
              {{ formatDate(billing.subscription.currentPeriodStart) }} &ndash;
              {{ formatDate(billing.subscription.currentPeriodEnd) }}
            </p>
          </div>
          <div>
            <span class="text-gray-500">Auto-renew</span>
            <p class="mt-0.5 font-medium">
              {{ billing.subscription.cancelAtPeriodEnd ? 'Cancels at period end' : 'Active' }}
            </p>
          </div>
        </div>
        <div class="mt-4 flex gap-3">
          <button
            v-if="!billing.subscription.cancelAtPeriodEnd && billing.currentPlan?.name !== 'free'"
            @click="billing.cancelSubscription()"
            class="text-sm text-red-600 hover:text-red-700 font-medium transition-colors"
          >
            Cancel Subscription
          </button>
          <button
            v-if="billing.subscription.cancelAtPeriodEnd"
            @click="billing.resumeSubscription()"
            class="text-sm text-indigo-600 hover:text-indigo-700 font-medium transition-colors"
          >
            Resume Subscription
          </button>
        </div>
      </div>

      <!-- Billing Address section -->
      <div class="mt-8 bg-white rounded-xl shadow p-6">
        <h3 class="text-sm font-semibold text-gray-900 mb-4">Billing Address</h3>
        <div class="space-y-4 max-w-md">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Country</label>
            <select
              v-model="billingAddress.country"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            >
              <option value="">Select country...</option>
              <option v-for="c in euCountries" :key="c.code" :value="c.code">{{ c.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Postal Code</label>
            <input
              v-model="billingAddress.postalCode"
              type="text"
              placeholder="12345"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">VAT Number <span class="text-gray-400 font-normal">(optional)</span></label>
            <input
              v-model="billingAddress.vatNumber"
              type="text"
              placeholder="e.g. DE123456789"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <div class="flex items-center gap-3">
            <button
              @click="handleSaveAddress"
              :disabled="addressSaving"
              class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
            >
              {{ addressSaving ? 'Saving...' : 'Save Address' }}
            </button>
            <span v-if="addressSaved" class="text-sm text-green-600 font-medium">Saved!</span>
            <span v-if="addressError" class="text-sm text-red-600">{{ addressError }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Usage tab -->
    <div v-else-if="activeTab === 'usage'">
      <div class="bg-white rounded-xl shadow p-6">
        <div v-if="billing.usage" class="space-y-6">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900">
                {{ billing.currentPlan?.displayName || 'Free' }} Plan
              </h3>
              <p class="text-sm text-gray-500 mt-0.5">
                {{ formatDate(billing.usage.periodStart) }} &ndash;
                {{ formatDate(billing.usage.periodEnd) }}
              </p>
            </div>
          </div>

          <div class="space-y-5">
            <UsageBar
              label="Apps"
              :current="billing.usage.apps.used"
              :limit="billing.usage.apps.limit"
              unit=""
              :unlimited="billing.usage.apps.limit === -1"
            />
            <UsageBar
              label="Databases"
              :current="billing.usage.databases.used"
              :limit="billing.usage.databases.limit"
              unit=""
              :unlimited="billing.usage.databases.limit === -1"
            />
            <UsageBar
              label="Memory"
              :current="billing.usage.memoryMb.used"
              :limit="billing.usage.memoryMb.limit"
              unit="MB"
              :unlimited="billing.usage.memoryMb.limit === -1"
            />
            <UsageBar
              label="Storage"
              :current="billing.usage.storageGb.used"
              :limit="billing.usage.storageGb.limit"
              unit="GB"
              :unlimited="billing.usage.storageGb.limit === -1"
            />
          </div>
        </div>
        <div v-else class="text-gray-500 py-8 text-center">
          No usage data available.
        </div>
      </div>
    </div>

    <!-- Invoices tab -->
    <div v-else-if="activeTab === 'invoices'">
      <div class="bg-white rounded-xl shadow overflow-hidden">
        <div v-if="billing.invoices.length === 0" class="text-gray-500 py-12 text-center">
          <svg class="mx-auto h-12 w-12 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p class="mt-3 text-sm font-medium">No invoices yet</p>
          <p class="mt-1 text-xs text-gray-400">Invoices will appear here once you subscribe to a paid plan.</p>
        </div>
        <table v-else class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Amount</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"></th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="invoice in billing.invoices" :key="invoice.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ formatDate(invoice.date) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                {{ invoice.description }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                &euro;{{ invoice.amountEur.toFixed(2) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium capitalize"
                  :class="statusBadgeClass(invoice.status)"
                >
                  {{ invoice.status }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                <a
                  v-if="invoice.stripeInvoiceUrl"
                  :href="invoice.stripeInvoiceUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-indigo-600 hover:text-indigo-500 font-medium"
                >
                  View
                </a>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Confirm downgrade modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-200"
        leave-active-class="transition-opacity duration-150"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
      >
        <div
          v-if="showConfirmModal"
          class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
          @click.self="cancelDowngrade"
        >
          <div class="bg-white rounded-xl shadow-xl max-w-md w-full mx-4 p-6">
            <h3 class="text-lg font-semibold text-gray-900">Confirm Plan Change</h3>
            <p class="mt-2 text-sm text-gray-600">
              Are you sure you want to switch to the
              <strong>{{ pendingPlan?.displayName }}</strong> plan? You may lose access to
              features available on your current plan.
            </p>
            <div class="mt-6 flex justify-end gap-3">
              <button
                @click="cancelDowngrade"
                class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
              >
                Cancel
              </button>
              <button
                @click="confirmDowngrade"
                :disabled="processing"
                class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
              >
                {{ processing ? 'Processing...' : 'Confirm Change' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Processing overlay -->
    <div
      v-if="processing"
      class="fixed inset-0 z-40 flex items-center justify-center bg-white/60"
    >
      <div class="text-center">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <p class="mt-3 text-sm font-medium text-gray-600">Processing...</p>
      </div>
    </div>
  </div>
</template>

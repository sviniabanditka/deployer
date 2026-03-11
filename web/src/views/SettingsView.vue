<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { changePassword, disable2FA, exportData, deleteAccount } from '../api/auth'
import TwoFactorSetup from '../components/TwoFactorSetup.vue'

const router = useRouter()
const auth = useAuthStore()

const activeTab = ref<'profile' | 'security' | 'data'>('profile')

// Profile / Password
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
})
const passwordError = ref('')
const passwordSuccess = ref('')
const passwordLoading = ref(false)

async function handleChangePassword() {
  passwordError.value = ''
  passwordSuccess.value = ''

  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    passwordError.value = 'New passwords do not match'
    return
  }
  if (passwordForm.newPassword.length < 8) {
    passwordError.value = 'Password must be at least 8 characters'
    return
  }

  passwordLoading.value = true
  try {
    await changePassword({
      currentPassword: passwordForm.currentPassword,
      newPassword: passwordForm.newPassword,
    })
    passwordSuccess.value = 'Password changed successfully'
    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
  } catch (e: any) {
    passwordError.value = e.response?.data?.message || 'Failed to change password'
  } finally {
    passwordLoading.value = false
  }
}

// Security / 2FA
const twoFAEnabled = ref(false)
const showSetup = ref(false)
const disableCode = ref('')
const disableError = ref('')
const disableLoading = ref(false)
const showDisableForm = ref(false)

function handleSetupComplete() {
  twoFAEnabled.value = true
  showSetup.value = false
}

async function handleDisable2FA() {
  disableError.value = ''
  if (!/^\d{6}$/.test(disableCode.value)) {
    disableError.value = 'Please enter a valid 6-digit code'
    return
  }
  disableLoading.value = true
  try {
    await disable2FA(disableCode.value)
    twoFAEnabled.value = false
    showDisableForm.value = false
    disableCode.value = ''
  } catch (e: any) {
    disableError.value = e.response?.data?.message || 'Failed to disable 2FA'
  } finally {
    disableLoading.value = false
  }
}

// Data & Privacy
const exportLoading = ref(false)
const deleteLoading = ref(false)
const showDeleteModal = ref(false)
const deleteConfirmText = ref('')
const deletePassword = ref('')
const deleteError = ref('')

async function handleExport() {
  exportLoading.value = true
  try {
    const blob = await exportData()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'my-data-export.json'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch {
    // silently fail
  } finally {
    exportLoading.value = false
  }
}

async function handleDeleteAccount() {
  deleteError.value = ''
  if (deleteConfirmText.value !== 'DELETE') {
    deleteError.value = 'Please type DELETE to confirm'
    return
  }
  if (!deletePassword.value) {
    deleteError.value = 'Please enter your password'
    return
  }
  deleteLoading.value = true
  try {
    await deleteAccount(deletePassword.value)
    auth.logout()
    router.push('/login')
  } catch (e: any) {
    deleteError.value = e.response?.data?.message || 'Failed to delete account'
  } finally {
    deleteLoading.value = false
  }
}

const tabs = [
  { key: 'profile', label: 'Profile' },
  { key: 'security', label: 'Security' },
  { key: 'data', label: 'Data & Privacy' },
] as const
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Account Settings</h1>

    <!-- Tabs -->
    <div class="border-b border-gray-200 mb-8">
      <nav class="flex gap-6">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          class="pb-3 text-sm font-medium border-b-2 transition-colors"
          :class="activeTab === tab.key
            ? 'border-indigo-600 text-indigo-600'
            : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
        >
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <!-- Profile Tab -->
    <div v-if="activeTab === 'profile'" class="space-y-8">
      <!-- User Info -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Profile Information</h2>
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              :value="auth.user?.name || ''"
              disabled
              class="w-full px-3 py-2 border border-gray-200 rounded-lg bg-gray-50 text-gray-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              :value="auth.user?.email || ''"
              disabled
              class="w-full px-3 py-2 border border-gray-200 rounded-lg bg-gray-50 text-gray-500"
            />
          </div>
        </div>
      </div>

      <!-- Change Password -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Change Password</h2>

        <div v-if="passwordError" class="mb-4 p-3 bg-red-50 text-red-700 rounded-lg text-sm">
          {{ passwordError }}
        </div>
        <div v-if="passwordSuccess" class="mb-4 p-3 bg-green-50 text-green-700 rounded-lg text-sm">
          {{ passwordSuccess }}
        </div>

        <form @submit.prevent="handleChangePassword" class="space-y-4 max-w-sm">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Current Password</label>
            <input
              v-model="passwordForm.currentPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">New Password</label>
            <input
              v-model="passwordForm.newPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Confirm New Password</label>
            <input
              v-model="passwordForm.confirmPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <button
            type="submit"
            :disabled="passwordLoading"
            class="px-5 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
          >
            {{ passwordLoading ? 'Updating...' : 'Update Password' }}
          </button>
        </form>
      </div>
    </div>

    <!-- Security Tab -->
    <div v-if="activeTab === 'security'" class="space-y-8">
      <!-- 2FA Section -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-1">Two-Factor Authentication</h2>
        <p class="text-sm text-gray-500 mb-5">Add an extra layer of security to your account.</p>

        <!-- 2FA Enabled State -->
        <div v-if="twoFAEnabled && !showSetup">
          <div class="flex items-center gap-3 mb-4">
            <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-green-100 text-green-700">
              <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
              Enabled
            </span>
            <span class="text-sm text-gray-600">Two-factor authentication is active on your account.</span>
          </div>

          <div v-if="!showDisableForm">
            <button
              @click="showDisableForm = true"
              class="px-4 py-2 text-sm font-medium text-red-700 bg-red-50 rounded-lg hover:bg-red-100 transition-colors"
            >
              Disable 2FA
            </button>
          </div>
          <div v-else class="space-y-3 max-w-sm">
            <div v-if="disableError" class="p-3 bg-red-50 text-red-700 rounded-lg text-sm">
              {{ disableError }}
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Enter 2FA code to disable</label>
              <input
                v-model="disableCode"
                type="text"
                inputmode="numeric"
                maxlength="6"
                placeholder="000000"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent font-mono tracking-widest"
              />
            </div>
            <div class="flex gap-2">
              <button
                @click="handleDisable2FA"
                :disabled="disableLoading"
                class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {{ disableLoading ? 'Disabling...' : 'Confirm Disable' }}
              </button>
              <button
                @click="showDisableForm = false; disableCode = ''; disableError = ''"
                class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>

        <!-- 2FA Setup Flow -->
        <div v-else-if="showSetup">
          <TwoFactorSetup
            @complete="handleSetupComplete"
            @cancel="showSetup = false"
          />
        </div>

        <!-- 2FA Not Enabled -->
        <div v-else>
          <button
            @click="showSetup = true"
            class="px-5 py-2.5 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors"
          >
            Enable Two-Factor Authentication
          </button>
        </div>
      </div>

      <!-- Active Sessions -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-1">Active Sessions</h2>
        <p class="text-sm text-gray-500 mb-4">Manage your active sessions across devices.</p>
        <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 text-sm text-gray-500 text-center">
          Session management coming soon.
        </div>
      </div>

      <!-- Login History -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-1">Login History</h2>
        <p class="text-sm text-gray-500 mb-4">Your recent account activity.</p>
        <div class="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <div class="flex items-center gap-3">
            <div class="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center">
              <svg class="w-4 h-4 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-800">Last login</p>
              <p class="text-xs text-gray-500">Current session</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Data & Privacy Tab -->
    <div v-if="activeTab === 'data'" class="space-y-8">
      <!-- Export Data -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-1">Export Your Data</h2>
        <p class="text-sm text-gray-500 mb-4">
          Download a copy of all your data in JSON format. This includes your profile, apps, databases, and billing history.
        </p>
        <button
          @click="handleExport"
          :disabled="exportLoading"
          class="px-5 py-2.5 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
        >
          {{ exportLoading ? 'Preparing export...' : 'Export My Data' }}
        </button>
      </div>

      <!-- GDPR Info -->
      <div class="bg-blue-50 border border-blue-200 rounded-xl p-6">
        <h3 class="text-sm font-semibold text-blue-900 mb-2">Your Privacy Rights</h3>
        <ul class="text-sm text-blue-800 space-y-1.5 list-disc list-inside">
          <li>You have the right to access, rectify, and delete your personal data.</li>
          <li>You can export your data at any time using the button above.</li>
          <li>You can delete your account permanently, which removes all associated data.</li>
          <li>For any privacy-related inquiries, contact us at privacy@deployer.dev.</li>
        </ul>
      </div>

      <!-- Danger Zone -->
      <div class="bg-white rounded-xl shadow-sm border-2 border-red-200 p-6">
        <h2 class="text-lg font-semibold text-red-700 mb-1">Danger Zone</h2>
        <p class="text-sm text-gray-500 mb-4">
          Permanently delete your account and all associated data. This action cannot be undone.
        </p>
        <button
          @click="showDeleteModal = true"
          class="px-5 py-2.5 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
        >
          Delete Account
        </button>
      </div>
    </div>

    <!-- Delete Account Modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-200"
        leave-active-class="transition-opacity duration-150"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
      >
        <div
          v-if="showDeleteModal"
          class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
          @click.self="showDeleteModal = false"
        >
          <div class="bg-white rounded-xl shadow-xl p-6 w-full max-w-md mx-4">
            <h3 class="text-lg font-semibold text-red-700">Delete Your Account</h3>
            <div class="mt-3 space-y-3">
              <div class="p-3 bg-red-50 rounded-lg">
                <p class="text-sm text-red-800 font-medium">This will permanently delete:</p>
                <ul class="mt-1.5 text-sm text-red-700 list-disc list-inside space-y-0.5">
                  <li>Your profile and personal information</li>
                  <li>All deployed applications</li>
                  <li>All databases and their data</li>
                  <li>Billing history and payment methods</li>
                </ul>
              </div>

              <div v-if="deleteError" class="p-3 bg-red-50 text-red-700 rounded-lg text-sm">
                {{ deleteError }}
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">
                  Type <span class="font-mono font-bold">DELETE</span> to confirm
                </label>
                <input
                  v-model="deleteConfirmText"
                  type="text"
                  placeholder="DELETE"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent font-mono"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Enter your password</label>
                <input
                  v-model="deletePassword"
                  type="password"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent"
                />
              </div>
            </div>

            <div class="mt-6 flex justify-end gap-3">
              <button
                @click="showDeleteModal = false; deleteConfirmText = ''; deletePassword = ''; deleteError = ''"
                class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
              >
                Cancel
              </button>
              <button
                @click="handleDeleteAccount"
                :disabled="deleteLoading || deleteConfirmText !== 'DELETE'"
                class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ deleteLoading ? 'Deleting...' : 'Delete Account' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

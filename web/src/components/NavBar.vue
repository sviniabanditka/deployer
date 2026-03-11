<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const mobileOpen = ref(false)
const userMenuOpen = ref(false)

async function handleLogout() {
  userMenuOpen.value = false
  await auth.logout()
  router.push('/login')
}

function closeUserMenu() {
  userMenuOpen.value = false
}
</script>

<template>
  <nav class="bg-indigo-700 shadow-lg">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-16">
        <div class="flex items-center gap-8">
          <router-link to="/" class="text-white text-lg font-bold tracking-tight">
            Deployer
          </router-link>
          <!-- Desktop nav -->
          <div class="hidden sm:flex gap-1">
            <router-link
              to="/"
              class="px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
              active-class="!bg-indigo-800 !text-white"
            >
              Dashboard
            </router-link>
            <router-link
              to="/apps"
              class="px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
              active-class="!bg-indigo-800 !text-white"
            >
              Apps
            </router-link>
            <router-link
              to="/databases"
              class="px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
              active-class="!bg-indigo-800 !text-white"
            >
              Databases
            </router-link>
            <router-link
              to="/billing"
              class="px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
              active-class="!bg-indigo-800 !text-white"
            >
              Billing
            </router-link>
            <router-link
              to="/status"
              class="px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
              active-class="!bg-indigo-800 !text-white"
            >
              Status
            </router-link>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <!-- User dropdown (desktop) -->
          <div v-if="auth.user" class="hidden sm:block relative">
            <button
              @click="userMenuOpen = !userMenuOpen"
              class="flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-indigo-600 transition-colors"
            >
              <div class="h-7 w-7 rounded-full bg-indigo-500 flex items-center justify-center text-xs font-medium text-white">
                {{ auth.user.name.charAt(0).toUpperCase() }}
              </div>
              <span class="text-sm text-indigo-100">{{ auth.user.name }}</span>
              <svg class="w-4 h-4 text-indigo-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>

            <!-- Dropdown overlay to close -->
            <div v-if="userMenuOpen" class="fixed inset-0 z-40" @click="closeUserMenu" />

            <!-- Dropdown menu -->
            <Transition
              enter-active-class="transition-all duration-150 ease-out"
              leave-active-class="transition-all duration-100 ease-in"
              enter-from-class="opacity-0 scale-95 -translate-y-1"
              leave-to-class="opacity-0 scale-95 -translate-y-1"
            >
              <div
                v-if="userMenuOpen"
                class="absolute right-0 mt-2 w-48 bg-white rounded-xl shadow-lg border border-gray-200 py-1 z-50"
              >
                <div class="px-4 py-2 border-b border-gray-100">
                  <p class="text-sm font-medium text-gray-900">{{ auth.user.name }}</p>
                  <p class="text-xs text-gray-500 truncate">{{ auth.user.email }}</p>
                </div>
                <router-link
                  to="/settings"
                  @click="closeUserMenu"
                  class="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                  Settings
                </router-link>
                <router-link
                  to="/billing"
                  @click="closeUserMenu"
                  class="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                  </svg>
                  Billing
                </router-link>
                <div class="border-t border-gray-100 my-1" />
                <button
                  @click="handleLogout"
                  class="flex items-center gap-2 w-full px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors"
                >
                  <svg class="w-4 h-4 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                  </svg>
                  Logout
                </button>
              </div>
            </Transition>
          </div>

          <!-- Mobile menu button -->
          <button
            @click="mobileOpen = !mobileOpen"
            class="sm:hidden p-2 text-indigo-200 hover:text-white transition-colors"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path v-if="!mobileOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Mobile menu -->
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      leave-active-class="transition-all duration-150 ease-in"
      enter-from-class="opacity-0 -translate-y-2"
      leave-to-class="opacity-0 -translate-y-2"
    >
      <div v-if="mobileOpen" class="sm:hidden border-t border-indigo-600">
        <div class="px-4 py-3 space-y-1">
          <router-link
            to="/"
            @click="mobileOpen = false"
            class="block px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
          >
            Dashboard
          </router-link>
          <router-link
            to="/apps"
            @click="mobileOpen = false"
            class="block px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
          >
            Apps
          </router-link>
          <router-link
            to="/databases"
            @click="mobileOpen = false"
            class="block px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
          >
            Databases
          </router-link>
          <router-link
            to="/billing"
            @click="mobileOpen = false"
            class="block px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
          >
            Billing
          </router-link>
          <router-link
            to="/settings"
            @click="mobileOpen = false"
            class="block px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
          >
            Settings
          </router-link>
          <router-link
            to="/status"
            @click="mobileOpen = false"
            class="block px-3 py-2 rounded-lg text-sm font-medium text-indigo-100 hover:bg-indigo-600 hover:text-white transition-colors"
          >
            Status
          </router-link>
        </div>
        <div v-if="auth.user" class="px-4 py-3 border-t border-indigo-600">
          <p class="text-sm text-indigo-200">{{ auth.user.name }}</p>
          <p v-if="auth.user.email" class="text-xs text-indigo-300">{{ auth.user.email }}</p>
          <button
            @click="handleLogout; mobileOpen = false"
            class="mt-2 w-full px-3 py-2 text-sm font-medium text-indigo-100 bg-indigo-800 rounded-lg hover:bg-indigo-900 transition-colors text-left"
          >
            Logout
          </button>
        </div>
      </div>
    </Transition>
  </nav>
</template>

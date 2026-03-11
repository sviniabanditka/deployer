<script setup lang="ts">
import NavBar from './components/NavBar.vue'
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <NavBar v-if="auth.isAuthenticated" />
    <main>
      <router-view />
    </main>
  </div>
</template>

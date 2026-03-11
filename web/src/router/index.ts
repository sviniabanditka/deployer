import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/LoginView.vue'),
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('../views/RegisterView.vue'),
    },
    {
      path: '/apps',
      name: 'AppList',
      component: () => import('../views/AppListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/apps/:id',
      name: 'AppDetail',
      component: () => import('../views/AppDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/databases',
      name: 'DatabaseList',
      component: () => import('../views/DatabaseListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/databases/:id',
      name: 'DatabaseDetail',
      component: () => import('../views/DatabaseDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/billing',
      name: 'Billing',
      component: () => import('../views/BillingView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/settings',
      name: 'Settings',
      component: () => import('../views/SettingsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/auth/callback',
      name: 'OAuthCallback',
      component: () => import('../views/OAuthCallbackView.vue'),
    },
    {
      path: '/status',
      name: 'Status',
      component: () => import('../views/StatusView.vue'),
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    next({ name: 'Login' })
  } else {
    next()
  }
})

export default router

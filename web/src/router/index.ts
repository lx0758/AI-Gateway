import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login/index.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/components/layout/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'Dashboard', component: () => import('@/views/Dashboard/index.vue') },
      { path: 'providers', name: 'Providers', component: () => import('@/views/Providers/index.vue') },
      { path: 'providers/:id', name: 'ProviderDetail', component: () => import('@/views/Providers/Detail.vue') },
      { path: 'models', name: 'Models', component: () => import('@/views/Models/index.vue') },
      { path: 'models/:id', name: 'ModelDetail', component: () => import('@/views/Models/Detail.vue') },
      { path: 'keys', name: 'Keys', component: () => import('@/views/Keys/index.vue') },
      { path: 'keys/:id', name: 'KeyDetail', component: () => import('@/views/Keys/Detail.vue') },
      { path: 'mcps', name: 'MCPs', component: () => import('@/views/MCPs/index.vue') },
      { path: 'mcps/:id', name: 'MCPDetail', component: () => import('@/views/MCPs/Detail.vue') },
      { path: 'model_usage', name: 'ModelUsage', component: () => import('@/views/ModelUsage/index.vue') },
      { path: 'mcp_usage', name: 'MCPUsage', component: () => import('@/views/MCPUsage/index.vue') },
      { path: 'settings', name: 'Settings', component: () => import('@/views/Settings/index.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  
  if (!userStore.user) {
    await userStore.fetchUser()
  }
  
  if (to.meta.requiresAuth !== false && !userStore.isLoggedIn) {
    next('/login')
  } else if (to.path === '/login' && userStore.isLoggedIn) {
    next('/')
  } else {
    next()
  }
})

export default router

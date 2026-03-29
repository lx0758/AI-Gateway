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
      { path: 'models', name: 'ModelMappings', component: () => import('@/views/Models/index.vue') },
      { path: 'api-keys', name: 'APIKeys', component: () => import('@/views/APIKeys/index.vue') },
      { path: 'usage', name: 'Usage', component: () => import('@/views/Usage/index.vue') },
      { path: 'settings', name: 'Settings', component: () => import('@/views/Settings/index.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth !== false && !userStore.isLoggedIn) {
    next('/login')
  } else if (to.path === '/login' && userStore.isLoggedIn) {
    next('/')
  } else {
    next()
  }
})

export default router

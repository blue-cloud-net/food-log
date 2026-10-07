import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { public: true } },
  { path: '/register', name: 'register', component: () => import('@/views/Register.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/components/layout/AppLayout.vue'),
    children: [
      { path: '', name: 'dashboard', component: () => import('@/views/Dashboard.vue') },
      { path: 'recipes', name: 'recipes', component: () => import('@/views/RecipeList.vue') },
      { path: 'recipes/new', name: 'recipe-new', component: () => import('@/views/RecipeEdit.vue') },
      { path: 'recipes/:id', name: 'recipe-detail', component: () => import('@/views/RecipeDetail.vue') },
      { path: 'recipes/:id/edit', name: 'recipe-edit', component: () => import('@/views/RecipeEdit.vue') },
      { path: 'restaurants', name: 'restaurants', component: () => import('@/views/RestaurantList.vue') },
      { path: 'restaurants/new', name: 'restaurant-new', component: () => import('@/views/RestaurantEdit.vue') },
      { path: 'restaurants/:id', name: 'restaurant-detail', component: () => import('@/views/RestaurantDetail.vue') },
      { path: 'restaurants/:id/edit', name: 'restaurant-edit', component: () => import('@/views/RestaurantEdit.vue') },
      { path: 'dishes', name: 'dishes', component: () => import('@/views/DishList.vue') },
      { path: 'search', name: 'search', component: () => import('@/views/SearchResults.vue') },
      { path: 'inventory', name: 'inventory', component: () => import('@/views/InventoryList.vue') },
      { path: 'inventory/new', name: 'inventory-new', component: () => import('@/views/InventoryEdit.vue') },
      { path: 'inventory/locations', name: 'inventory-locations', component: () => import('@/views/StorageLocationSettings.vue') },
      { path: 'inventory/:id/edit', name: 'inventory-edit', component: () => import('@/views/InventoryEdit.vue') },
      { path: 'tags', name: 'tags', component: () => import('@/views/TagSettings.vue') },
      { path: 'settings', name: 'settings', component: () => import('@/views/Settings.vue') },
      { path: 'settings/ai', name: 'settings-ai', component: () => import('@/views/SettingsAi.vue') },
      { path: 'settings/backup', name: 'settings-backup', component: () => import('@/views/SettingsBackup.vue') },
      { path: 'profile', name: 'profile', component: () => import('@/views/Profile.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.public && auth.isLoggedIn && (to.name === 'login' || to.name === 'register')) {
    return { name: 'dashboard' }
  }
})

export default router

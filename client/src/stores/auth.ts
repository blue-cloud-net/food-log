import { defineStore } from 'pinia'
import * as authApi from '@/api/auth'
import type { UserProfile } from '@/api/types'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: null as UserProfile | null
  }),
  getters: {
    isLoggedIn: (s) => !!s.token
  },
  actions: {
    async login(username: string, password: string) {
      const data = await authApi.login(username, password)
      this.token = data.token
      this.user = data.user
      localStorage.setItem('token', data.token)
    },
    async register(username: string, email: string, password: string) {
      const data = await authApi.register(username, email, password)
      this.token = data.token
      this.user = data.user
      localStorage.setItem('token', data.token)
    },
    async fetchMe() {
      if (!this.token) return
      try {
        this.user = await authApi.me()
      } catch {
        /* 401 时拦截器已处理跳转 */
      }
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
    }
  }
})

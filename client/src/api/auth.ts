import http from './http'
import type { UserProfile } from './types'

export interface AuthData {
  token: string
  user: UserProfile
}

export function login(username: string, password: string) {
  return http.post('/auth/login', { username, password }) as Promise<AuthData>
}

export function register(username: string, email: string, password: string) {
  return http.post('/auth/register', { username, email, password }) as Promise<AuthData>
}

export function me() {
  return http.get('/auth/me') as Promise<UserProfile>
}

export function updateMe(payload: {
  username?: string
  email?: string
  password?: string
  avatar_url?: string
}) {
  return http.put('/auth/me', payload) as Promise<UserProfile>
}

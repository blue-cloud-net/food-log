import http from './http'
import type { Paginated, Recipe, TagCategory } from './types'

export interface RecipeQuery {
  page?: number
  page_size?: number
  keyword?: string
  difficulty?: string
  tag?: string
  sort?: string
  favorite?: boolean
}

export function listRecipes(params: RecipeQuery = {}) {
  return http.get('/recipes', { params }) as Promise<Paginated<Recipe>>
}

export function getRecipe(id: string) {
  return http.get(`/recipes/${id}`) as Promise<Recipe>
}

export function createRecipe(payload: Partial<Recipe>) {
  return http.post('/recipes', payload) as Promise<Recipe>
}

export function updateRecipe(id: string, payload: Partial<Recipe>) {
  return http.put(`/recipes/${id}`, payload) as Promise<Recipe>
}

export function deleteRecipe(id: string) {
  return http.delete(`/recipes/${id}`) as Promise<{ deleted: boolean }>
}

export function getTagCategories() {
  return http.get('/recipes/tags') as Promise<TagCategory[]>
}

export function randomRecipe(params: { tag?: string; difficulty?: string } = {}) {
  return http.get('/recipes/random', { params }) as Promise<Recipe>
}

export function favoriteRecipe(id: string) {
  return http.post(`/recipes/${id}/favorite`) as Promise<{ favorited: boolean }>
}

export function unfavoriteRecipe(id: string) {
  return http.delete(`/recipes/${id}/favorite`) as Promise<{ favorited: boolean }>
}

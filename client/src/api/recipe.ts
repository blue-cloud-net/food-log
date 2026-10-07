import http from './http'
import type { Paginated, Recipe, TagCategory } from './types'

export interface RecipeQuery {
  page?: number
  page_size?: number
  keyword?: string
  difficulty?: string
  /** 菜谱级标签 id */
  tag?: string
  /** 食材级标签 id */
  ingredient_tag?: string
  sort?: string
  favorite?: boolean
  /** 已做/未做：'true' 只看已做，'false' 只看未做，不传为全部 */
  made?: 'true' | 'false'
  /** 只看喜欢（独立于收藏） */
  liked?: boolean
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

export function randomRecipe(params: { tag?: string; ingredient_tag?: string; difficulty?: string } = {}) {
  return http.get('/recipes/random', { params }) as Promise<Recipe>
}

export function favoriteRecipe(id: string) {
  return http.post(`/recipes/${id}/favorite`) as Promise<{ favorited: boolean }>
}

export function unfavoriteRecipe(id: string) {
  return http.delete(`/recipes/${id}/favorite`) as Promise<{ favorited: boolean }>
}

/** 标记「做过日期」；madeAt 为 null 表示取消已做 */
export function setRecipeMade(id: string, madeAt: string | null) {
  return http.put(`/recipes/${id}/made`, { made_at: madeAt }) as Promise<{ made_at: string | null }>
}

/** 标记「喜欢」（独立于收藏） */
export function setRecipeLiked(id: string, liked: boolean) {
  return http.put(`/recipes/${id}/like`, { liked }) as Promise<{ is_liked: boolean }>
}

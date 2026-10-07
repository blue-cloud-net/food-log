import http from './http'
import type { Dish, Paginated, Restaurant, RestaurantDetail, TagCategory } from './types'

export interface RestaurantQuery {
  page?: number
  page_size?: number
  keyword?: string
  /** 餐厅标签 id */
  tag?: string
  /** recommend | value | ambience | service */
  sort?: string
}

export function listRestaurants(params: RestaurantQuery = {}) {
  return http.get('/restaurants', { params }) as Promise<Paginated<Restaurant>>
}

/** dishTag 为菜品标签 id，用于在餐厅详情内过滤菜品列表 */
export function getRestaurant(id: string, dishTag?: string) {
  return http.get(`/restaurants/${id}`, {
    params: { dish_tag: dishTag || undefined }
  }) as Promise<RestaurantDetail>
}

export function createRestaurant(payload: Partial<Restaurant>) {
  return http.post('/restaurants', payload) as Promise<Restaurant>
}

export function updateRestaurant(id: string, payload: Partial<Restaurant>) {
  return http.put(`/restaurants/${id}`, payload) as Promise<Restaurant>
}

export function deleteRestaurant(id: string) {
  return http.delete(`/restaurants/${id}`) as Promise<{ deleted: boolean }>
}

export function addDish(restaurantId: string, payload: Partial<Dish>) {
  return http.post(`/restaurants/${restaurantId}/dishes`, payload) as Promise<Dish>
}

export function updateDish(id: string, payload: Partial<Dish>) {
  return http.put(`/dishes/${id}`, payload) as Promise<Dish>
}

export function deleteDish(id: string) {
  return http.delete(`/dishes/${id}`) as Promise<{ deleted: boolean }>
}

/** 餐厅标签词表（全局预设 + 本人自定义） */
export function getRestaurantTagCategories() {
  return http.get('/restaurants/tags') as Promise<TagCategory[]>
}

/** 菜品标签词表（全局预设 + 本人自定义） */
export function getDishTagCategories() {
  return http.get('/dishes/tags') as Promise<TagCategory[]>
}

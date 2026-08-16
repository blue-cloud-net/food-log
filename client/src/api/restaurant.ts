import http from './http'
import type { Dish, Paginated, Restaurant, RestaurantDetail } from './types'

export interface RestaurantQuery {
  page?: number
  page_size?: number
  keyword?: string
  cuisine_type?: string
  sort?: string
}

export function listRestaurants(params: RestaurantQuery = {}) {
  return http.get('/restaurants', { params }) as Promise<Paginated<Restaurant>>
}

export function getRestaurant(id: string) {
  return http.get(`/restaurants/${id}`) as Promise<RestaurantDetail>
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

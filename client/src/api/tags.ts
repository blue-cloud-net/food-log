import http from './http'
import type { Tag, TagCategory } from './types'

// 用户自定义标签维护（全局预设标签为只读，服务端会拒绝修改）

export function createTag(payload: { name: string; category_id?: string; mutex_group?: string }) {
  return http.post('/tags', payload) as Promise<Tag>
}

export function updateTag(id: string, payload: { name: string; mutex_group?: string; sort_order?: number }) {
  return http.put(`/tags/${id}`, payload) as Promise<Tag>
}

export function deleteTag(id: string) {
  return http.delete(`/tags/${id}`) as Promise<{ deleted: boolean }>
}

export function createTagCategory(payload: { name: string; color?: string }) {
  return http.post('/tag-categories', payload) as Promise<TagCategory>
}

export function updateTagCategory(id: string, payload: { name: string; color?: string; sort_order?: number }) {
  return http.put(`/tag-categories/${id}`, payload) as Promise<TagCategory>
}

export function deleteTagCategory(id: string) {
  return http.delete(`/tag-categories/${id}`) as Promise<{ deleted: boolean }>
}

// ===== 探店标签（餐厅 / 菜品各一套独立字典，与菜谱标签体系隔离）=====

export type ShopTagDomain = 'restaurant' | 'dish'

function shopTagBase(domain: ShopTagDomain) {
  return domain === 'restaurant' ? '/restaurant-tags' : '/dish-tags'
}

function shopCategoryBase(domain: ShopTagDomain) {
  return domain === 'restaurant' ? '/restaurant-tag-categories' : '/dish-tag-categories'
}

/** 新建自定义探店标签；category_id 为空时服务端自动归入「自定义」分类 */
export function createShopTag(
  domain: ShopTagDomain,
  payload: { name: string; category_id?: string; mutex_group?: string }
) {
  return http.post(shopTagBase(domain), payload) as Promise<Tag>
}

export function updateShopTag(
  domain: ShopTagDomain,
  id: string,
  payload: { name: string; mutex_group?: string; sort_order?: number }
) {
  return http.put(`${shopTagBase(domain)}/${id}`, payload) as Promise<Tag>
}

export function deleteShopTag(domain: ShopTagDomain, id: string) {
  return http.delete(`${shopTagBase(domain)}/${id}`) as Promise<{ deleted: boolean }>
}

export function createShopTagCategory(domain: ShopTagDomain, payload: { name: string; color?: string }) {
  return http.post(shopCategoryBase(domain), payload) as Promise<TagCategory>
}

export function updateShopTagCategory(
  domain: ShopTagDomain,
  id: string,
  payload: { name: string; color?: string; sort_order?: number }
) {
  return http.put(`${shopCategoryBase(domain)}/${id}`, payload) as Promise<TagCategory>
}

export function deleteShopTagCategory(domain: ShopTagDomain, id: string) {
  return http.delete(`${shopCategoryBase(domain)}/${id}`) as Promise<{ deleted: boolean }>
}

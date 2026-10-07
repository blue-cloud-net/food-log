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

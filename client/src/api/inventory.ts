import http from './http'
import type { InventoryItem, InventoryList } from './types'

export interface InventoryQuery {
  keyword?: string
  /** 存放位置 id */
  location_id?: string
  /** 存放位置大类：fridge / outside */
  area?: string
  /** 仅返回 N 天内到期（含已过期）的条目 */
  expiring_within_days?: number
  /** expire（默认，按过期日升序，未设置排最后）| created | name */
  sort?: string
}

export interface InventoryPayload {
  location_id: string
  name: string
  /** 数量（> 0） */
  quantity: number
  unit?: string
  category?: string
  /** YYYY-MM-DD，空串表示未设置 */
  expire_at?: string
  note?: string
  images?: string[]
}

export function listInventory(params: InventoryQuery = {}) {
  return http.get('/inventory', { params }) as Promise<InventoryList>
}

export function getInventoryItem(id: string) {
  return http.get(`/inventory/${id}`) as Promise<InventoryItem>
}

export function createInventoryItem(payload: InventoryPayload) {
  return http.post('/inventory', payload) as Promise<InventoryItem>
}

export function updateInventoryItem(id: string, payload: InventoryPayload) {
  return http.put(`/inventory/${id}`, payload) as Promise<InventoryItem>
}

export function deleteInventoryItem(id: string) {
  return http.delete(`/inventory/${id}`) as Promise<{ deleted: boolean }>
}

/** 批量删除，返回实际删除的条数（仅本人的条目会被删除） */
export function batchDeleteInventoryItems(ids: string[]) {
  return http.post('/inventory/batch-delete', { ids }) as Promise<{ deleted: number }>
}

/** 消耗一定数量；数量减到 0 时服务端会自动把该条移出库存 */
export function consumeInventoryItem(id: string, quantity: number) {
  return http.post(`/inventory/${id}/consume`, { quantity }) as Promise<{
    quantity: number
    removed: boolean
  }>
}

import http from './http'
import type { StorageLocation } from './types'

// 存放位置词表（全局预设 + 本人自定义）：预设位置只读，服务端会拒绝修改
export function listStorageLocations() {
  return http.get('/storage-locations') as Promise<StorageLocation[]>
}

export function createStorageLocation(payload: { area: string; name: string }) {
  return http.post('/storage-locations', payload) as Promise<StorageLocation>
}

export function updateStorageLocation(
  id: string,
  payload: { area: string; name: string; sort_order?: number }
) {
  return http.put(`/storage-locations/${id}`, payload) as Promise<StorageLocation>
}

export function deleteStorageLocation(id: string) {
  return http.delete(`/storage-locations/${id}`) as Promise<{ deleted: boolean }>
}

import http from './http'
import type { UploadFile } from './types'

/** 图片用途：决定服务端落地到 images/{type}/ 子目录 */
export type UploadKind = 'recipe' | 'restaurant' | 'inventory'

export function uploadImages(files: File[], type: UploadKind = 'recipe') {
  const form = new FormData()
  files.forEach((f) => form.append('images', f))
  form.append('type', type)
  return http.post('/upload', form) as Promise<{ files: UploadFile[] }>
}

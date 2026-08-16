import http from './http'
import type { UploadFile } from './types'

export function uploadImages(files: File[]) {
  const form = new FormData()
  files.forEach((f) => form.append('images', f))
  return http.post('/upload', form) as Promise<{ files: UploadFile[] }>
}

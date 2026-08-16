import http from './http'
import type { Recognition } from './types'

export function recognizeImage(imageUrl: string) {
  return http.post('/recipes/ai/recognize', { image_url: imageUrl }) as Promise<Recognition>
}

import http from './http'
import type { SearchResult } from './types'

export function globalSearch(keyword: string) {
  return http.get('/search', { params: { keyword } }) as Promise<SearchResult>
}

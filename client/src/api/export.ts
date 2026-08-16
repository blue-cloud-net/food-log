import http from './http'
import type { ExportData } from './types'

export function exportJson() {
  return http.get('/export', { params: { format: 'json' } }) as Promise<ExportData>
}

export function exportCsvBlob() {
  return http.get('/export', {
    params: { format: 'csv' },
    responseType: 'blob'
  }) as Promise<Blob>
}

export function importData(data: ExportData, mode: 'append' | 'overwrite' = 'append') {
  return http.post('/export/import', data, {
    params: { mode }
  }) as Promise<{ imported: boolean; mode: string }>
}

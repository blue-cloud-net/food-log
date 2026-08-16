import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 统一 axios 实例：请求自动带 JWT / X-AI-Key，响应自动解包 {code,message,data}
const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 20000
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  // 前端个人中心配置的 AI key 覆盖（可选）
  const aiKey = localStorage.getItem('ai_api_key')
  if (aiKey) {
    config.headers['X-AI-Key'] = aiKey
  }
  return config
})

http.interceptors.response.use(
  (res) => {
    const resp = res.data
    // 非标准包装（文件下载 Blob / CSV 文本 / 导出 JSON 对象）直接返回
    if (resp === null || typeof resp !== 'object' || !('code' in resp)) {
      return resp
    }
    if (resp.code === 0) {
      return resp.data
    }
    if (resp.code === 1002) {
      localStorage.removeItem('token')
      ElMessage.error('登录已过期，请重新登录')
      router.push('/login')
      return Promise.reject(new Error(resp.message))
    }
    ElMessage.error(resp.message || '请求失败')
    return Promise.reject(new Error(resp.message || '请求失败'))
  },
  (err) => {
    const status = err.response?.status
    if (status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
    }
    const msg = err.response?.data?.message || err.message || '网络错误'
    ElMessage.error(msg)
    return Promise.reject(err)
  }
)

export default http

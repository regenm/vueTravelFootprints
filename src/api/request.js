import axios from 'axios'
import { TOKEN_KEY, USER_KEY } from '@/utils/constants'
import { apiBaseURL } from '@/utils/env'

const request = axios.create({
  baseURL: apiBaseURL(),
  timeout: 30000
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  if (config.data instanceof FormData) {
    delete config.headers['Content-Type']
  } else if (!config.headers['Content-Type']) {
    config.headers['Content-Type'] = 'application/json'
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const data = response.data
    if (data && data.success === false) {
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return data
  },
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message || error.message || '网络错误'
    if (status === 401) {
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
      const path = window.location.pathname
      if (!path.startsWith('/login') && !path.startsWith('/s/')) {
        window.location.assign('/login')
      }
    }
    return Promise.reject(new Error(message))
  }
)

export default request

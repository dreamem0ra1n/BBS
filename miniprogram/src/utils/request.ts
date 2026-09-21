import { env } from '../config/env'
import { encodeForm } from './form'

export interface ApiResult<T> {
  success: boolean
  data: T
  message?: string
}

type RequestOptions = Omit<UniApp.RequestOptions, 'url'> & { url?: string }

export function request<T>(
  path: string,
  options: RequestOptions = {}
): Promise<T> {
  const method = (options.method || 'GET').toUpperCase() as UniApp.RequestOptions['method']
  const header: Record<string, string> = {
    ...(options.header as Record<string, string> | undefined),
  }
  const token = uni.getStorageSync('userToken')

  if (token) {
    header['X-User-Token'] = token
  }
  if (method !== 'GET' && method !== 'HEAD') {
    header['Content-Type'] = 'application/x-www-form-urlencoded'
  }

  let data = options.data || {}
  if (method !== 'GET' && method !== 'HEAD') {
    data = encodeForm(data as Record<string, unknown>)
  }

  return new Promise((resolve, reject) => {
    uni.request({
      ...options,
      url: `${env.apiBaseUrl}${path}`,
      method,
      data,
      header,
      timeout: options.timeout || 10000,
      success(response) {
        const body = (response.data || {}) as ApiResult<T>
        if (response.statusCode >= 200 && response.statusCode < 300 && body.success) {
          resolve(body.data)
          return
        }
        reject(new Error(body.message || `请求失败（${response.statusCode}）`))
      },
      fail: reject,
    })
  })
}

export function get<T>(path: string, data?: Record<string, unknown>): Promise<T> {
  return request<T>(path, { method: 'GET', data })
}

export function post<T>(path: string, data?: Record<string, unknown>): Promise<T> {
  return request<T>(path, { method: 'POST', data })
}

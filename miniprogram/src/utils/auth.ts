export interface LoginPayload {
  token: string
  user?: Record<string, unknown>
}

export function saveLogin(payload: LoginPayload): void {
  if (!payload?.token) {
    throw new Error('登录响应缺少 token')
  }
  uni.setStorageSync('userToken', payload.token)
  if (payload.user) {
    uni.setStorageSync('currentUser', payload.user)
  }
}

export function clearLogin(): void {
  uni.removeStorageSync('userToken')
  uni.removeStorageSync('currentUser')
}

export function getCurrentUser<T = Record<string, unknown>>(): T | null {
  return uni.getStorageSync('currentUser') || null
}

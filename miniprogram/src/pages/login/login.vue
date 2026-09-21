<template>
  <view class="page">
    <view class="card">
      <view class="title">登录</view>

      <view v-if="loading" class="muted">正在读取后端支持的登录方式…</view>

      <view v-else-if="loadError">
        <view class="error">{{ loadError }}</view>
        <button class="primary" size="mini" @click="loadLoginMethods">重试</button>
      </view>

      <template v-else>
        <view v-if="passwordEnabled">
          <view class="field">
            <text class="label">用户名</text>
            <input
              v-model="username"
              class="input"
              type="text"
              placeholder="本地调试账号：admin"
              :disabled="submitting"
            />
          </view>
          <view class="field">
            <text class="label">密码</text>
            <input
              v-model="password"
              class="input"
              password
              placeholder="本地调试密码：123456"
              :disabled="submitting"
            />
          </view>
          <button class="primary" :loading="submitting" @click="submitPassword">
            登录
          </button>
        </view>

        <button v-if="passportEnabled" class="link" @click="goPassport">
          使用 Passport 登录
        </button>

        <view v-if="!passwordEnabled && !passportEnabled" class="muted">
          后端没有启用任何登录方式，请检查 LoginMethods 配置。
        </view>
      </template>
    </view>

    <view class="muted hint">当前接口地址：{{ apiBaseUrl || '同源（H5 代理）' }}</view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { env } from '../../config/env'
import { saveLogin, type LoginPayload } from '../../utils/auth'
import { get, post } from '../../utils/request'

interface LoginMethods {
  passport?: boolean
  password?: boolean
}

interface PublicConfig {
  loginMethods?: LoginMethods
}

const apiBaseUrl = env.apiBaseUrl
const username = ref('')
const password = ref('')
const loading = ref(true)
const submitting = ref(false)
const loadError = ref('')
const passwordEnabled = ref(false)
const passportEnabled = ref(false)

async function loadLoginMethods(): Promise<void> {
  loading.value = true
  loadError.value = ''
  try {
    // 与 Site/Admin 一致：登录方式只以后端公开的 loginMethods 为准
    const config = await get<PublicConfig>('/api/config/configs')
    passwordEnabled.value = !!config.loginMethods?.password
    passportEnabled.value = !!config.loginMethods?.passport
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : '无法读取后端配置'
  } finally {
    loading.value = false
  }
}

async function submitPassword(): Promise<void> {
  if (submitting.value) return
  if (!username.value.trim()) {
    uni.showToast({ title: '请输入用户名', icon: 'none' })
    return
  }
  if (!password.value) {
    uni.showToast({ title: '请输入密码', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    const result = await post<LoginPayload>('/api/login/password', {
      username: username.value.trim(),
      password: password.value,
    })
    saveLogin(result)
    password.value = ''
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(backToPrevious, 600)
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '登录失败',
      icon: 'none',
    })
  } finally {
    submitting.value = false
  }
}

function backToPrevious(): void {
  if (getCurrentPages().length > 1) {
    uni.navigateBack()
    return
  }
  uni.reLaunch({ url: '/pages/index/index' })
}

function goPassport(): void {
  uni.navigateTo({ url: '/pages/passport/passport' })
}

onShow(() => loadLoginMethods())
</script>

<style lang="scss" scoped>
.title {
  margin-bottom: 24rpx;
  font-size: 40rpx;
  font-weight: 600;
}

.field {
  margin-bottom: 24rpx;
}

.label {
  display: block;
  margin-bottom: 12rpx;
  color: #6b7280;
  font-size: 24rpx;
}

.input {
  box-sizing: border-box;
  width: 100%;
  height: 80rpx;
  padding: 0 20rpx;
  background: #f6f7f9;
  border-radius: 8rpx;
  font-size: 28rpx;
}

.primary {
  margin-top: 8rpx;
  color: #ffffff;
  background: #2563eb;
}

.link {
  margin-top: 24rpx;
  color: #2563eb;
  background: transparent;
  font-size: 26rpx;
}

.error {
  margin-bottom: 16rpx;
  color: #e64c4c;
}

.hint {
  margin-top: 24rpx;
  text-align: center;
  word-break: break-all;
}
</style>

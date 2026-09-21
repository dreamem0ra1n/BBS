<template>
  <web-view :src="loginUrl" @message="onMessage" />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { env } from '../../config/env'
import { saveLogin, type LoginPayload } from '../../utils/auth'

const loginUrl = ref('')

onLoad(() => {
  loginUrl.value = `${env.passportUrl}?success=${encodeURIComponent(env.passportBridgeUrl)}`
})

/**
 * bridge 页面（site/pages/miniprogram/passport.vue）通过 wx.miniProgram.postMessage 回传凭证。
 * web-view 会把消息累积到 event.detail.data 数组里，取最后一条即可。
 */
function onMessage(event: any): void {
  const messages = event?.detail?.data
  const message = Array.isArray(messages) ? messages[messages.length - 1] : messages
  if (!message || message.type !== 'passport-login') return

  try {
    saveLogin(message as LoginPayload)
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 500)
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '登录失败',
      icon: 'none',
    })
  }
}
</script>

<template>
  <section class="bridge">
    <h1>{{ status }}</h1>
    <p v-if="error" class="error">{{ error }}</p>
  </section>
</template>

<script>
// 小程序 web-view 内嵌的 wx.miniProgram 由 JSSDK 提供
const WX_JSSDK_URL = 'https://res.wx.qq.com/open/js/jweixin-1.6.0.js'
// JSSDK 未就绪时的重试次数与间隔
const WX_WAIT_RETRIES = 10
const WX_WAIT_INTERVAL = 300

export default {
  layout: 'empty',
  data() {
    return {
      status: '正在完成 Passport 登录…',
      error: '',
    }
  },
  head() {
    return {
      title: 'Passport 登录',
      script: [
        {
          src: WX_JSSDK_URL,
        },
      ],
    }
  },
  mounted() {
    this.completeLogin()
  },
  methods: {
    async completeLogin() {
      const sessionToken = this.$route.query.SESSION_TOKEN
      if (!sessionToken) {
        this.fail('登录回调缺少 SESSION_TOKEN')
        return
      }

      // 只在下一次 signin 请求前有效：换到 userToken 后这个 cookie 就没用了
      this.$cookies.set('SESSION_TOKEN', sessionToken, {
        path: '/',
        maxAge: 300,
      })

      let loginResult
      try {
        // axios 的响应拦截器已经拆掉了 { success, data } 外层，这里直接拿到 data
        loginResult = await this.$axios.post('/api/login/signin')
      } catch (error) {
        this.fail(error.message || '无法换取 BBS 登录凭证')
        return
      }

      if (!loginResult || !loginResult.token) {
        this.fail('登录响应缺少 token')
        return
      }

      const message = {
        type: 'passport-login',
        token: loginResult.token,
        user: loginResult.user,
      }

      const wx = await this.waitForWx()
      if (!wx) {
        // 不在小程序 web-view 里（例如浏览器直接打开），登录已完成但没有回传通道
        this.status = '登录成功，请返回小程序'
        return
      }

      // postMessage 的消息会在 web-view 页面销毁时统一回传，所以紧接着返回小程序
      wx.miniProgram.postMessage({ data: message })
      this.status = '登录成功，正在返回小程序…'
      wx.miniProgram.navigateBack()
    },
    /**
     * JSSDK 可能比 mounted 稍晚就绪，轮询等待一小段时间
     * @returns {Promise<Object|null>} wx 对象，超时返回 null
     */
    waitForWx() {
      return new Promise((resolve) => {
        let retries = 0
        const check = () => {
          if (window.wx && window.wx.miniProgram) {
            resolve(window.wx)
            return
          }
          if (retries >= WX_WAIT_RETRIES) {
            resolve(null)
            return
          }
          retries += 1
          setTimeout(check, WX_WAIT_INTERVAL)
        }
        check()
      })
    },
    fail(message) {
      this.status = 'Passport 登录失败'
      this.error = message
    },
  },
}
</script>

<style lang="scss" scoped>
.bridge {
  padding: 120px 24px;
  text-align: center;

  h1 {
    font-size: 18px;
    line-height: 24px;
  }

  .error {
    margin-top: 16px;
    font-size: 14px;
    color: rgb(230, 76, 76);
  }
}
</style>

<style scoped>
.bridge {
  padding: 48px 24px;
  text-align: center;
}
</style>

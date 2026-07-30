import Vue from 'vue'

Vue.use({
  install(Vue, options) {
    Vue.prototype.$siteTitle = function (subTitle) {
      const siteTitle = this.$store.getters['config/siteTitle'] || ''
      if (subTitle) {
        return subTitle + (siteTitle ? ' | ' + siteTitle : '')
      }
      return siteTitle
    }

    Vue.prototype.$siteDescription = function () {
      return this.$store.getters['config/siteDescription']
    }

    Vue.prototype.$siteKeywords = function () {
      return this.$store.getters['config/siteKeywords']
    }

    Vue.prototype.$topicSiteTitle = function (topic) {
      if (topic.type === 0) {
        return this.$siteTitle(topic.title)
      } else {
        return this.$siteTitle(topic.content)
      }
    }

    /**
     * 链接跳转
     * @param path
     */
    Vue.prototype.$linkTo = function (path) {
      this.$router.push(path)
    }

    /**
     * 跳转到登录页
     * @param ref
     */
    Vue.prototype.$toSignin = function (ref) {
      const loginMethods = this.$store.state.config.config.loginMethods || {}

      // 如果启用了密码登录，则跳转到密码登录页面
      if (loginMethods.password) {
        const currentPath = this.$route && this.$route.fullPath
        const signinRoute = { path: '/user/signin' }
        const returnPath = ref || currentPath
        if (returnPath && !this.$isSigninUrl(returnPath)) {
          signinRoute.query = { ref: returnPath }
        }
        this.$router.push(signinRoute)
        return
      }

      // 如果启用了passport登录，则跳转到passport登录页面
      if (loginMethods.passport) {
        this.$toPassportSignin(ref)
        return
      }

      this.$msg({ type: 'error', message: '当前没有可用的登录方式' })
    }

    Vue.prototype.$toPassportSignin = function (ref) {
      if (!ref && process.client) {
        // 如果没配置refUrl，那么取当前地址
        ref = window.location.href
      }
      window.location.href =
        'https://www.qsc.zju.edu.cn/passport/v4/qsc/login?success=' + ref
    }

    /**
     * 是否是登陆页
     * @param ref
     * @returns {boolean}
     */
    Vue.prototype.$isSigninUrl = function (ref) {
      return ref === '/user/signin'
    }

    /**
     * 弹出错误消息，然后执行登录
     * @param message
     */
    Vue.prototype.$msgSignIn = function () {
      const that = this
      this.$msg({
        type: 'error',
        message: '请先登录',
        onClose() {
          that.$toSignin()
        },
      })
    }

    /**
     * 弹出消息然后执行函数
     * @param type 消息类型，success、error、info...
     * @param message 消息内容
     * @param then 要执行的函数
     */
    Vue.prototype.$msg = function ({
      type = 'success',
      message,
      duration = 800,
      onClose,
    }) {
      this.$message({
        duration: 800,
        type,
        message,
        onClose,
      })
    }
  },
})

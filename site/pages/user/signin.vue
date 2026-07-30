<template>
  <section class="main">
    <div class="container">
      <div class="main-body no-bg">
        <div class="widget signin">
          <div class="widget-header">登录</div>
          <div class="widget-content">
            <template v-if="loginMethods.password">
              <div class="field">
                <label class="label">用户名</label>
                <div class="control has-icons-left">
                  <input
                    v-model="username"
                    class="input is-success"
                    type="text"
                    placeholder="请输入用户名"
                    @keyup.enter="submitLogin"
                  />
                  <span class="icon is-small is-left"
                    ><i class="iconfont icon-username"
                  /></span>
                </div>
              </div>

              <div class="field">
                <label class="label">密码</label>
                <div class="control has-icons-left">
                  <input
                    v-model="password"
                    class="input"
                    type="password"
                    placeholder="请输入密码"
                    @keyup.enter="submitLogin"
                  />
                  <span class="icon is-small is-left"
                    ><i class="iconfont icon-password"
                  /></span>
                </div>
              </div>
              <div class="field">
                <button class="button is-success" @click="submitLogin">
                  登录
                </button>
                <nuxt-link class="button is-text" to="/user/signup">
                  没有账号？点击这里去注册&gt;&gt;
                </nuxt-link>
              </div>
            </template>

            <div v-if="loginMethods.passport" class="field">
              <button class="button is-link" @click="$toPassportSignin(ref)">
                通过求是潮 Passport 登录
              </button>
            </div>

            <div
              v-if="loginMethod.qq || loginMethod.github || loginMethod.osc"
              class="third-party-line"
            ></div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
export default {
  asyncData({ params, query }) {
    return {
      ref: query.ref,
    }
  },
  data() {
    return {
      username: '',
      password: '',
    }
  },
  head() {
    return {
      title: this.$siteTitle('登录'),
    }
  },
  computed: {
    currentUser() {
      return this.$store.state.user.current
    },
    isLogin() {
      return !!this.currentUser
    },
    loginMethods() {
      return this.$store.state.config.config.loginMethods || {}
    },
    loginMethod() {
      return this.$store.state.config.config.loginMethod || {}
    },
  },
  methods: {
    async submitLogin() {
      try {
        if (!this.username) {
          this.$message.error('请输入用户名')
          return
        }
        if (!this.password) {
          this.$message.error('请输入密码')
          return
        }
        const user = await this.$store.dispatch('user/signin', {
          username: this.username,
          password: this.password,
        })
        if (this.ref) {
          // 跳到登录前
          this.$linkTo(this.ref)
        } else {
          // 跳到个人主页
          this.$linkTo('/user/' + user.id)
        }
      } catch (e) {
        this.$message.error(e.message || e)
      }
    },
    /**
     * 如果已经登录了，那么直接跳转
     * @returns {boolean}
     */
    redirectIfLogined() {
      if (this.isLogin) {
        const me = this
        this.$msg({
          message: '登录成功',
          onClose() {
            if (me.ref && !me.$isSigninUrl(me.ref)) {
              me.$linkTo(me.ref)
            } else {
              me.$linkTo('/')
            }
          },
        })
        return true
      }
      return false
    },
  },
}
</script>

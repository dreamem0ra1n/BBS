export const state = () => ({
  current: null,
  userToken: null,
})

export const mutations = {
  setCurrent(state, user) {
    state.current = user
  },
  setUserToken(state, userToken) {
    state.userToken = userToken
  },
}

export const actions = {
  // 登录成功
  loginSuccess(context, { user }) {
    // const config = context.rootState.config.config
    // const cookieMaxAge = 86400 * config.tokenExpireDays
    // this.$cookies.set('userToken', token, { maxAge: cookieMaxAge, path: '/' })
    // context.commit('setUserToken', token)
    context.commit('setCurrent', user)
  },

  // 获取当前登录用户
  async getCurrentUser(context) {
    const user = await this.$axios.get('/api/user/current')
    context.commit('setCurrent', user)
    return user
  },

  // 登录

  signin(context) {
    const me = this
    this.$axios
      .post('/api/login/signin', { ref: null })
      .then(async (res) => {
        try {
          const info = await fetch(
            'https://www.qsc.zju.edu.cn/passport/v4/profile',
            {
              mode: 'cors',
              credentials: 'include',
            }
          )
          res.user.position =
            info.data.user.QscUser.department +
            '-' +
            info.data.user.QscUser.position
        } catch (e) {
          console.log(e)
        } finally {
          if (!res.user.position) {
            const pattern = res.user.roles[0].split('_')
            const role = pattern[0]
            res.user.position = res.user.department + '-' + role
          }
          context.commit('setCurrent', res.user)
          context.commit('setUserToken', res.token)
          const config = context.rootState.config.config
          this.$cookies.set('userToken', res.token, {
            maxAge: 86400 * config.tokenExpireDays,
            path: '/',
          })
        }
      })
      .catch((e) => {
        console.log(e)
        window.location.href =
          'https://www.qsc.zju.edu.cn/passport/v4/qsc/login?success=' +
          window.location.origin +
          window.location.pathname
        // me.state.env.currentURL
      })
  },

  // 退出登录
  async signout(context) {
    const userToken = this.$cookies.get('userToken')
    await this.$axios.get('/api/login/signout', {
      params: {
        userToken,
      },
    })
    context.commit('setUserToken', null)
    context.commit('setCurrent', null)
    this.$cookies.remove('userToken')
    this.$cookies.remove('SESSION_TOKEN')
    await fetch('https://www.qsc.zju.edu.cn/passport/v4/logout', {
      method: 'GET',
      credentials: 'include',
    })
    this.$forceUpdate()
  },
}

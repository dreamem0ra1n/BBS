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
    this.$axios
      .post('/api/login/signin', { ref: null })
      .then(async (res) => {
        const pattern = res.user.roles[0].split('_')
        const role = pattern[0]
        const node = await this.$axios.get(
          '/api/topic/node?nodeId=' + pattern[1]
        )
        if (node) {
          const session = node.name
          res.user.position = session + '-' + role
        } else res.user.position = res.user.roles[0]
        context.commit('setCurrent', res.user)
        context.commit('setUserToken', res.token)
        const config = this.$store.state.config.config
        this.$cookies.set('userToken', res.token, {
          maxAge: 86400 * config.tokenExpireDays,
          path: '/',
        })
      })
      .catch((e) => {
        console.log(e)
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
    await this.$axios.get('https://www.qsc.zju.edu.cn/passport/v4/logout')
    context.commit('setUserToken', null)
    context.commit('setCurrent', null)
    this.$cookies.remove('userToken')
    this.$cookies.remove('SESSION_TOKEN')
    this.$forceUpdate()
  },
}

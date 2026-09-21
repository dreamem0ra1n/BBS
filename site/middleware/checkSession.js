// 小程序 Passport 桥接页由 pages/miniprogram/passport.vue 自行消费 SESSION_TOKEN：
// 它需要从 URL 上读取凭证并调 /api/login/signin，换到 userToken 后回传给小程序。
// 如果这里先把 token 换掉并重定向到不带参数的地址，bridge 就拿不到凭证了。
const MINIPROGRAM_BRIDGE_PATH = '/miniprogram/passport'

export default async function (context) {
  if (context.route.path === MINIPROGRAM_BRIDGE_PATH) {
    return
  }
  // 若当前url中含有SESSION_TOKEN，说明从passport跳转回该网页，从 query 中提取 token 并存入cookie
  if (context.route.query.SESSION_TOKEN) {
    await context.app.$cookies.set(
      'SESSION_TOKEN',
      context.route.query.SESSION_TOKEN,
      {
        maxAge: 86400 * context.store.state.config.config.tokenExpireDays,
        path: '/',
      }
    )
    try {
      // 向 bbs 后端发送请求获取 userToken
      const response = await context.$axios.post('/api/login/signin')
      await context.app.$cookies.set('userToken', response.token, {
        maxAge: 86400 * context.store.state.config.config.tokenExpireDays,
        path: '/',
      })
      // 重定向至该页（无 SESSION_TOKEN 参数）
      context.redirect(context.route.path)
    } catch (e) {
      console.error(e)
      context.error({
        statusCode: 500,
        message: '500 Internal Error: ' + e.message,
      })
    }
  }
}

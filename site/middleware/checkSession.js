export default async function (context) {
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
      console.log('user token:', response.token)
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


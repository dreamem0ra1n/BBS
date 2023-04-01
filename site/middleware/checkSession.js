export default async function (context) {
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
      const response = await context.$axios.post('/api/login/signin')
      await context.app.$cookies.set('userToken', response.token, {
        maxAge: 86400 * context.store.state.config.config.tokenExpireDays,
        path: '/',
      })
    } catch (e) {
      console.log(e)
    } finally {
      context.redirect(context.route.path)
    }
  }
}

export default function (context) {
  if (context.route.query.SESSION_TOKEN) {
    context.$cookies.set('SESSION_TOKEN', context.route.query.SESSION_TOKEN, {
      maxAge: 86400 * context.store.state.config.config.tokenExpireDays,
      path: '/',
    })
    context.redirect(context.route.path)
  }
}

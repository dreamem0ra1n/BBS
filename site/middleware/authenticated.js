import UserHelper from '~/common/UserHelper'
export default function (context) {
  const user = context.store.state.user.current
  if (!user) {
    toSignIn(context)
    return
  }
  if (isAdminUrl(context)) {
    if (!UserHelper.isOwner(user) && !UserHelper.isAdmin(user)) {
      context.error({
        statusCode: 403,
        message: '403 forbidden',
      })
    }
  }
}
// 当前访问URL是否是管理后台
function isAdminUrl(context) {
  return context.route.path.indexOf('/admin') === 0
}
// 前往登录地址
function toSignIn(context) {
  const loginMethods = context.store.state.config.config.loginMethods || {}

  // 如果启用了密码登录，则跳转到密码登录页面
  if (loginMethods.password) {
    let ref = context.route.fullPath
    if (process.server && context.req) {
      ref = context.req.originalUrl.replace(/^\/bbs2/, '')
    }
    context.redirect('/user/signin?ref=' + encodeURIComponent(ref || '/'))
    return
  }

  // 如果启用了passport登录，则跳转到passport登录页面
  if (loginMethods.passport) {
    context.redirect(getSignInUrl(context))
    return
  }

  // 如果没有启用任何登录方式，则返回错误
  context.error({ statusCode: 503, message: 'No login method is enabled' })
}

// 获取登录跳转地址
function getSignInUrl(context) {
  let ref // 来源地址
  if (process.server) {
    // 服务端
    ref = context.req.originalUrl
  } else if (process.client) {
    // 客户端
    ref = context.route.path
  }
  let signinUrl =
    'https://www.qsc.zju.edu.cn/passport/v4/static/index.html#/login?success='
  if (ref) {
    signinUrl +=
      context.store.state.env.currentURL + '/bbs2' + encodeURIComponent(ref)
  }
  return signinUrl
}

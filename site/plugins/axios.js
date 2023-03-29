import qs from 'qs'
export default function ({ req, $axios, app, store }) {
  let prefix = ''
  try {
    if (window) prefix = '/bbs2'
  } catch (e) {
  } finally {
    // using the middleware in /middleware/login.js,the cookie is added on the request header
    // however, it seems that when the server transmits the request, the cookies are ignored
    // so I decided to pick userToken out manually here
    const ifToken = req?.rawHeaders.find((str) => str.includes('userToken'))
    let token
    if (ifToken) {
      token = ifToken.split('userToken=')[1].split(';')[0]
      app.$cookies.set('userToken', token)
      store.commit('user/setUserToken', token)
    }
    // the above part picks out userToken
    $axios.onRequest((config) => {
      config.url = prefix + config.url
      config.headers.common['X-Client'] = 'bbs-go-site'
      config.headers.post['Content-Type'] = 'application/x-www-form-urlencoded'
      // const userToken = app.$cookies.get('userToken')
      if (token) {
        config.headers.common['X-User-Token'] = token
      }
      config.transformRequest = [
        function (data) {
          if (process.client && data instanceof FormData) {
            // 如果是FormData就不转换
            return data
          }
          data = qs.stringify(data)
          return data
        },
      ]
    })
    $axios.onResponse((response) => {
      if (response.status !== 200) {
        return Promise.reject(response)
      }
      const jsonResult = response.data
      if (jsonResult.success) {
        return Promise.resolve(jsonResult.data)
      } else {
        return Promise.reject(jsonResult)
      }
    })
  }
}

import http from 'http'

export default function (req, res, next) {
  const cookies = req.rawHeaders.find(
    (str) => str.includes('SESSION_TOKEN=') && !str.includes('http')
  )
  const userToken = req.rawHeaders.find((str) => str.includes('userToken'))
  if (cookies && !userToken) {
    const options = {
      hostname: '127.0.0.1',
      path: '/api/login/signin',
      port: '8082',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': 0,
        Cookie: cookies,
      },
    }
    const request = http.request(options, (response) => {
      let rawData = ''
      response.on('data', (chunk) => {
        rawData += chunk
      })
      response.on('error', (err) => {
        console.log('middleware error')
        console.log(err)
        next()
      })
      response.on('end', () => {
        try {
          const parsedData = JSON.parse(rawData)
          const index = req.rawHeaders.indexOf(cookies)
          if (parsedData.data) {
            req.rawHeaders[index] =
              req.rawHeaders[index] + '; userToken=' + parsedData.data.token
          }
        } catch (e) {
          console.log(e)
        } finally {
          next()
        }
      })
    })
    // console.log(request)
    request.write('')
    request.end()
  } else next()
}

// 线上地址：发布包（yarn build:mp-weixin）使用
const PROD_API_BASE_URL = 'https://www.qsc.zju.edu.cn/bbs2'
const PROD_PASSPORT_URL =
  'https://www.qsc.zju.edu.cn/passport/v4/static/index.html#/login'
const PROD_PASSPORT_BRIDGE_URL =
  'https://www.qsc.zju.edu.cn/bbs2/miniprogram/passport'

// 本地调试地址：开发包（yarn dev:mp-weixin）直连本地后端。
// 开发者工具跑在 Windows 上，WSL2 会把 Windows 的 localhost 转发到 WSL 内监听的端口，
// 所以后端在 WSL 里跑时 127.0.0.1 就能通。真机预览要改成运行后端的电脑局域网 IP，
// 例如 http://192.168.1.10:8082，并让后端监听 0.0.0.0、放行 Windows 防火墙。
const DEV_API_BASE_URL = 'http://127.0.0.1:8082'

// H5 调试（yarn dev:h5）走 vite.config.ts 里的 server.proxy，用同源相对路径
const H5_DEV_API_BASE_URL = ''
// H5 调试 Passport 登录需要本地 site 在跑：cd site && yarn dev
const H5_DEV_PASSPORT_BRIDGE_URL =
  'http://localhost:3000/bbs2/miniprogram/passport'

const isDev = process.env.NODE_ENV === 'development'

let apiBaseUrl = isDev ? DEV_API_BASE_URL : PROD_API_BASE_URL
let passportBridgeUrl = PROD_PASSPORT_BRIDGE_URL

// 条件编译：仅 H5 平台保留这段，微信小程序产物里会被剥离
// #ifdef H5
if (isDev) {
  apiBaseUrl = H5_DEV_API_BASE_URL
  passportBridgeUrl = H5_DEV_PASSPORT_BRIDGE_URL
}
// #endif

export const env = {
  apiBaseUrl,
  passportUrl: PROD_PASSPORT_URL,
  passportBridgeUrl,
}

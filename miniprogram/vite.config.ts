import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

// H5 调试时后端地址，默认本地 Go 服务（server/bbs-go.yaml 里的 Port）
const apiTarget = process.env.BBS_API_PROXY || 'http://127.0.0.1:8082'

export default defineConfig({
  plugins: [uni()],
  // 只影响 yarn dev:h5 的开发服务器，构建产物里不包含代理配置
  server: {
    proxy: {
      // env.ts 在 H5 开发模式下把 apiBaseUrl 置空，请求落在同源 /api 上
      '/api': {
        target: apiTarget,
        changeOrigin: true,
      },
      // 兼容 apiBaseUrl 仍带 /bbs2 前缀的写法，映射关系与线上 Nginx 一致
      '/bbs2': {
        target: apiTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/bbs2/, ''),
      },
    },
  },
})

module.exports = {
  lintOnSave: true,
  publicPath: "/bbsadmin",
  css: {
    loaderOptions: {
      scss: {
        sassOptions: {
          quietDeps: true,
        },
      },
    },
  },
  configureWebpack: {
    performance: {
      maxAssetSize: 1200000,
      maxEntrypointSize: 1500000,
    },
    resolve: {
      fallback: {
        path: require.resolve("path-browserify"),
      },
    },
  },
  devServer: {
    port: 8080,
    allowedHosts: ["www.qsc.zju.edu.cn", "localhost"],
  },
};

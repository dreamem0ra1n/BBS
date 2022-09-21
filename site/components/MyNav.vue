<template>
  <div>
    <pc-nav class="pc-nav" />
    <mobile-nav class="mobile-nav" />
  </div>
</template>

<script>
export default {
  mounted() {
    this.$axios
      .post('/api/login/signin', { ref: window.location.href })
      .then((res) => {
        console.log(res)
        this.$store.commit('user/setCurrent', res.data)
        /* if (!res?.data?.success) {
          console.log(res)
          window.location =
            'https://www.qsc.zju.edu.cn/passport/v4/zju/login?success=' +
            window.location.href
        } */
      })
      .catch((e) => {
        console.log(e)
        if (e?.data?.success) {
          console.log(e)
          window.location =
            'https://www.qsc.zju.edu.cn/passport/v4/zju/login?success=' +
            window.location.href
        }
      })
  },
}
</script>

<style lang="scss">
// 当现实顶部导航栏的时候，顶部预留一段空白，防止重叠
body {
  @media screen and (max-width: 1024px) {
    padding-top: 46px;
  }
}
</style>

<style scoped>
@media screen and (max-width: 1024px) {
  .pc-nav {
    display: none !important;
  }
}

@media screen and (min-width: 1024px) {
  .mobile-nav {
    display: none !important;
  }
}
</style>

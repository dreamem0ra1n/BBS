<template>
  <div>
    <my-nav />

    <nuxt />

    <my-footer />
  </div>
</template>

<script>
export default {
  mounted() {
    // this.$store.commit('env/setDomain', window.location.origin)
    if (!this.$cookies.get('userToken'))
      this.$axios
        .post('/api/login/signin', { ref: window.location.href })
        .then((res) => {
          this.$store.commit('user/setCurrent', res.user)
          this.$store.commit('user/setUserToken', res.token)
          const config = this.$store.state.config.config
          this.$cookies.set('userToken', res.token, {
            maxAge: 86400 * config.tokenExpireDays,
            path: '/',
          })
          window.location = '/'
        })
        .catch((e) => {
          console.log(e)
        })
  },
}
</script>

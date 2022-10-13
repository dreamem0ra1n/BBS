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
    console.log(this.$route)
    if (!this.$cookies.get('userToken'))
      this.$axios
        .post('/api/login/signin', { ref: null })
        .then((res) => {
          this.$store.commit('user/setCurrent', res.user)
          this.$store.commit('user/setUserToken', res.token)
          const config = this.$store.state.config.config
          this.$cookies.set('userToken', res.token, {
            maxAge: 86400 * config.tokenExpireDays,
            path: '/',
          })
        })
        .catch((e) => {
          console.log(e)
        })
  },
}
</script>

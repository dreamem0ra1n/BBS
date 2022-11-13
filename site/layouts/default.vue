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
    if (this.$router?.currentRoute?.query?.SESSION_TOKEN) {
      this.$cookies.set(
        'SESSION_TOKEN',
        this.$router.currentRoute.query.SESSION_TOKEN,
        {
          maxAge: 86400 * this.$store.state.config.config.tokenExpireDays,
          path: '/',
        }
      )
      this.$router.push(this.$router.currentRoute.path)
    }
    // if (!this.$cookies.get('userToken'))
    this.$store.commit('env/setDomain', window.location.origin)
    this.$store.dispatch('user/signin')
  },
}
</script>

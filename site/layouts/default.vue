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
        .then(async (res) => {
          const pattern = res.user.roles[0].split('_')
          const role = pattern[0]
          const node = await this.$axios.get(
            '/api/topic/node?nodeId=' + pattern[1]
          )
          if (node) {
            const session = node.name
            res.user.position = session + '-' + role
          } else res.user.position = res.user.roles[0]
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

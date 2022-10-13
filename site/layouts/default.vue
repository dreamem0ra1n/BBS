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
    // if (!this.$cookies.get('userToken'))
    this.$store.dispatch('user/signin')
    console.log(this.$store.state.user)
  },
}
</script>

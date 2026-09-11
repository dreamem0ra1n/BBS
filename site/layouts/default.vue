<template>
  <div>
    <my-nav />

    <nuxt />

    <my-footer />

    <back-to-top />
  </div>
</template>

<script>
export default {
  mounted() {
    this.$store.commit('env/setDomain', window.location.origin)
  },
}
</script>

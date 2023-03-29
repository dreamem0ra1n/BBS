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
    this.$store.commit('env/setDomain', window.location.origin)
  },
}
</script>

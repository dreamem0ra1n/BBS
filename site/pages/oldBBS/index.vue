<template>
  <section class="main">
    <div class="container main-container left-main size-320">
      <div class="left-container">
        <div class="main-content no-padding no-bg topics-wrapper">
          <div class="topics-nav"><old-nav :data="data" /></div>
          <div class="topics-main"></div>
        </div>
      </div>
      <div class="right-container">
        <friend-links :links="links" />
      </div>
    </div>
  </section>
</template>
<script>
export default {
  async asyncData({ $axios, store }) {
    store.commit('env/setCurrentNodeId', -2) // 设置当前所在node
    const json = require('~/oldBBS.json')
    try {
      const [links] = await Promise.all([$axios.get('/api/link/toplinks')])
      return { links, data: json }
    } catch (e) {
      console.error(e)
    }
  },

  data() {},
  head() {
    return {
      title: this.$siteTitle(),
      meta: [
        {
          hid: 'description',
          name: 'description',
          content: this.$siteDescription(),
        },
        { hid: 'keywords', name: 'keywords', content: this.$siteKeywords() },
      ],
    }
  },
}
</script>

<style lang="scss" scoped></style>

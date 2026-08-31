<template>
  <div class="topics-main">
    <topic-sort :value="sort" />
    <load-more
      v-if="topicsPage"
      v-slot="{ results }"
      :init-data="topicsPage"
      :url="'/api/topic/topics?recommend=true&sort=' + sort"
    >
      <topic-list :topics="results" />
    </load-more>
  </div>
</template>

<script>
export default {
  async asyncData({ $axios, store, query }) {
    store.commit('env/setCurrentNodeId', -1) // 设置当前所在node
    try {
      const sort = query.sort === 'create' ? 'create' : 'comment'
      const [topicsPage] = await Promise.all([
        $axios.get('/api/topic/topics', { params: { recommend: true, sort } }),
      ])
      return { topicsPage, sort }
    } catch (e) {
      console.error(e)
    }
  },
  head() {
    return {
      title: this.$siteTitle('热门话题'),
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
  watchQuery: ['sort'],
}
</script>

<style lang="scss" scoped></style>

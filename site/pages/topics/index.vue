<template>
  <div class="topics-main">
    <tag-bar :node-id="0" />
    <topic-sort :value="sort" />
    <sticky-topics :node-id="0" />
    <load-more
      v-if="topicsPage"
      v-slot="{ results }"
      :init-data="topicsPage"
      :url="'/api/topic/topics?sort=' + sort"
    >
      <topic-list :topics="results" />
    </load-more>
  </div>
</template>

<script>
export default {
  async asyncData({ $axios, store, query }) {
    try {
      store.commit('env/setCurrentNodeId', 0) // 设置当前所在node
      const sort = query.sort === 'create' ? 'create' : 'comment'
      const [topicsPage] = await Promise.all([
        $axios.get('/api/topic/topics', { params: { sort } }),
      ])
      return { topicsPage, sort }
    } catch (e) {
      console.error(e)
    }
  },
  head() {
    return {
      title: this.$siteTitle('话题'),
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

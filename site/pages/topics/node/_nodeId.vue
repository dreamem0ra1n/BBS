<template>
  <div class="topics-main">
    <tag-bar :node-id="node.nodeId" />
    <topic-sort :value="sort" />
    <sticky-topics :node-id="node.nodeId" />
    <load-more
      v-if="topicsPage"
      v-slot="{ results }"
      :init-data="topicsPage"
      :url="'/api/topic/topics?nodeId=' + node.nodeId + '&sort=' + sort"
    >
      <topic-list :topics="results" />
    </load-more>
  </div>
</template>

<script>
export default {
  async asyncData({ $axios, params, store, query }) {
    const nodeId = parseInt(params.nodeId)
    const sort = query.sort === 'create' ? 'create' : 'comment'
    store.commit('env/setCurrentNodeId', nodeId) // 设置当前所在node
    const [node, topicsPage] = await Promise.all([
      $axios.get('/api/topic/node?nodeId=' + nodeId),
      $axios.get('/api/topic/topics', { params: { nodeId, sort } }),
    ])
    return {
      node,
      topicsPage,
      sort,
    }
  },
  head() {
    return {
      title: this.$siteTitle(this.node.name + ' - 话题'),
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
  mounted() {
    this.$store.commit('env/setCurrentNodeId', this.node.nodeId)
  },
}
</script>

<style lang="scss" scoped></style>

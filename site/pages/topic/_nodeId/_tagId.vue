<template>
  <div class="topics-main">
    <sticky-topics :node-id="node.nodeId" />
    <load-more
      v-if="topicsPage"
      v-slot="{ results }"
      :init-data="topicsPage"
      :url="'/api/topic/topics?nodeId=' + node.nodeId"
    >
      <topic-list :topics="results" />
    </load-more>
  </div>
</template>

<script>
export default {
  async asyncData({ $axios, params, store }) {
    const nodeId = parseInt(params.nodeId)
    const tagId = parseInt(params.tagId)
    store.commit('env/setCurrentNodeId', +nodeId) // 设置当前所在node
    store.commit('env/setCurrentTag', +tagId)
    console.log(nodeId)
    const [node, topicsPage, tag] = await Promise.all([
      $axios.get('/api/topic/node?nodeId=' + nodeId),
      $axios.post('/api/topic/topicsnt', {
        cursor: 0,
        nodeId,
        tagId,
      }),
      $axios.get('/api/tag/' + tagId),
    ])
    console.log(topicsPage)
    console.log(node)
    return {
      node,
      tag,
      topicsPage,
    }
  },
  head() {
    return {
      title: this.$siteTitle(this.node.name + ' - ' + this.tag.name),
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

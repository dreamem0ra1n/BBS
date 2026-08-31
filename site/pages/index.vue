<template>
  <section class="main">
    <div class="container main-container left-main size-320">
      <div class="left-container">
        <div class="main-content no-padding no-bg topics-wrapper">
          <div class="topics-nav"><topics-nav :nodes="nodes" /></div>
          <div class="topics-main">
            <topic-sort :value="sort" />
            <load-more
              v-if="topicsPage"
              v-slot="{ results }"
              :init-data="topicsPage"
              :url="'/api/topic/topics?sort=' + sort"
            >
              <topic-list :topics="results" />
            </load-more>
          </div>
        </div>
      </div>
      <div class="right-container">
        <site-notice />
        <check-in />
        <score-rank :score-rank="scoreRank" />
        <friend-links :links="links" />
        <site-stats :stats="stats" />
      </div>
    </div>
  </section>
</template>

<script>
export default {
  async asyncData({ $axios, store, query }) {
    store.commit('env/setCurrentNodeId', 0) // 设置当前所在node
    try {
      const sort = query.sort === 'create' ? 'create' : 'comment'
      const [nodes, topicsPage, scoreRank, links, stats] = await Promise.all([
        $axios.get('/api/topic/nodes'),
        $axios.get('/api/topic/topics', { params: { sort } }),
        $axios.get('/api/user/score/rank'),
        $axios.get('/api/link/toplinks'),
        $axios.get('/api/stats/site'),
      ])
      return { nodes, topicsPage, scoreRank, links, stats, sort }
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
  watchQuery: ['sort'],
}
</script>

<style lang="scss" scoped></style>

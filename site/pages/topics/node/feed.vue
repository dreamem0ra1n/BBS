<template>
  <div class="topics-main">
    <load-more
      v-if="user && topicsPage"
      v-slot="{ results }"
      :init-data="topicsPage"
      url="/api/feed/topics"
    >
      <topic-list v-if="results.length" :topics="results" />
      <div v-else class="feed-empty">暂无关注用户发布的帖子</div>
    </load-more>
    <div v-else-if="!user" class="feed-empty">
      登录后即可查看你关注的用户发布的帖子
      <button class="button is-primary is-small" @click="$toSignin()">
        去登录
      </button>
    </div>
  </div>
</template>

<script>
export default {
  async asyncData({ $axios, store }) {
    store.commit('env/setCurrentNodeId', -2) // 设置当前所在node
    let topicsPage
    try {
      if (store.state.user.current) {
        topicsPage = await $axios.get('/api/feed/topics')
      }
    } catch (e) {
      console.log(e.message || e)
    }
    return { topicsPage }
  },
  head() {
    return {
      title: this.$siteTitle('关注'),
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
  computed: {
    user() {
      return this.$store.state.user.current
    },
  },
}
</script>

<style lang="scss" scoped>
.feed-empty {
  padding: 32px 16px;
  border-radius: 3px;
  background: var(--bg-color);
  color: var(--text-color3);
  text-align: center;

  .button {
    margin-left: 10px;
  }
}
</style>

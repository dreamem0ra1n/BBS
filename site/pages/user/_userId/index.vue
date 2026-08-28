<template>
  <section class="main">
    <div class="container">
      <user-profile :user="user" />

      <div class="container main-container right-main size-320">
        <user-center-sidebar :user="user" />
        <div class="right-container">
          <div class="tabs-warp">
            <div class="tabs">
              <ul>
                <li :class="{ 'is-active': activeTab === 'topics' }">
                  <nuxt-link :to="tabLink('topics')">
                    <span class="icon is-small">
                      <i class="iconfont icon-topic" aria-hidden="true" />
                    </span>
                    <span>话题</span>
                  </nuxt-link>
                </li>
                <li :class="{ 'is-active': activeTab === 'favorites' }">
                  <nuxt-link :to="tabLink('favorites')">
                    <span class="icon is-small"
                      ><i class="iconfont icon-favorite" aria-hidden="true"
                    /></span>
                    <span>收藏</span>
                  </nuxt-link>
                </li>
                <li :class="{ 'is-active': activeTab === 'comments' }">
                  <nuxt-link :to="tabLink('comments')">
                    <span class="icon is-small"
                      ><i class="iconfont icon-comment" aria-hidden="true"
                    /></span>
                    <span>回复</span>
                  </nuxt-link>
                </li>
              </ul>
            </div>

            <div v-if="activeTab === 'topics'">
              <div
                v-if="
                  topicsPage && topicsPage.results && topicsPage.results.length
                "
              >
                <topic-list :topics="topicsPage.results" :show-avatar="false" />
                <pagination
                  :page="topicsPage.page"
                  :url-prefix="paginationUrlPrefix"
                />
              </div>
              <div v-else class="notification is-primary">暂无话题</div>
            </div>

            <div v-else-if="activeTab === 'favorites'">
              <ul
                v-if="favoritesPage.results && favoritesPage.results.length"
                class="favorite-list"
              >
                <li
                  v-for="favorite in favoritesPage.results"
                  :key="favorite.favoriteId"
                  class="favorite-item"
                >
                  <template v-if="favorite.deleted">
                    <div class="favorite-summary">收藏内容失效</div>
                  </template>
                  <template v-else>
                    <div class="favorite-title">
                      <a :href="favorite.url" target="_blank">{{
                        favorite.title
                      }}</a>
                    </div>
                    <div class="favorite-summary">{{ favorite.content }}</div>
                    <div class="favorite-meta">
                      <nuxt-link :to="'/user/' + favorite.user.id">{{
                        favorite.user.nickname
                      }}</nuxt-link>
                      <time>{{ favorite.createTime | prettyDate }}</time>
                    </div>
                  </template>
                </li>
              </ul>
              <div v-else class="notification is-primary">暂无收藏</div>
              <pagination
                v-if="favoritesPage.page"
                :page="favoritesPage.page"
                :url-prefix="paginationUrlPrefix"
              />
            </div>

            <div v-else-if="activeTab === 'comments'">
              <div class="comment-order">
                <span>排序：</span>
                <nuxt-link
                  :class="{ active: !ascOrder }"
                  :to="tabLink('comments', 0)"
                  >倒序</nuxt-link
                >
                <span> · </span>
                <nuxt-link
                  :class="{ active: ascOrder }"
                  :to="tabLink('comments', 1)"
                  >正序</nuxt-link
                >
              </div>
              <ul
                v-if="commentsPage.results && commentsPage.results.length"
                class="user-comments"
              >
                <li
                  v-for="comment in commentsPage.results"
                  :key="comment.commentId"
                  class="user-comment"
                >
                  <div class="comment-meta">
                    <time>{{ comment.createTime | prettyDate }}</time>
                    <a v-if="comment.entityUrl" :href="comment.entityUrl">{{
                      comment.entityTitle || '查看原文'
                    }}</a>
                  </div>
                  <div
                    class="comment-content"
                    v-html="commentDisplayContent(comment)"
                  />
                </li>
              </ul>
              <div v-else class="notification is-primary">暂无回复</div>
              <pagination
                v-if="commentsPage.page"
                :page="commentsPage.page"
                :url-prefix="paginationUrlPrefix"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
const defaultTab = 'topics'
const tabs = ['topics', 'favorites', 'comments']

export default {
  middleware: 'authenticated',
  async asyncData({ $axios, params, query, error }) {
    let user
    try {
      user = await $axios.get('/api/user/' + params.userId)
    } catch (err) {
      console.log(err)
      error({
        statusCode: 404,
        message: err.message || '系统错误',
      })
      return
    }

    const activeTab = tabs.includes(query.tab) ? query.tab : defaultTab
    const page = query.p || 1
    const ascOrder = query.asc_order === '1' ? 1 : 0
    let topicsPage = null
    let favoritesPage = null
    let commentsPage = null
    if (activeTab === 'topics') {
      topicsPage = await $axios.get('/api/topic/user/topics', {
        params: { userId: params.userId, page },
      })
    } else if (activeTab === 'favorites') {
      favoritesPage = await $axios.get('/api/user/favorites', {
        params: { userId: params.userId, page },
      })
    } else {
      commentsPage = await $axios.get('/api/comment/user/comments', {
        params: { userId: params.userId, page, asc_order: ascOrder },
      })
    }
    return {
      activeTab,
      ascOrder,
      user,
      topicsPage,
      favoritesPage,
      commentsPage,
    }
  },
  data() {
    return {}
  },
  head() {
    return {
      title: this.$siteTitle(this.user.nickname),
    }
  },
  computed: {
    currentUser() {
      return this.$store.state.user.current
    },
    isOwner() {
      const current = this.$store.state.user.current
      return this.user && current && this.user.id === current.id
    },
    paginationUrlPrefix() {
      const order =
        this.activeTab === 'comments' ? `&asc_order=${this.ascOrder}` : ''
      return `/user/${this.user.id}?tab=${this.activeTab}${order}&p=`
    },
  },
  watchQuery: ['tab', 'p', 'asc_order'],
  methods: {
    commentDisplayContent(comment) {
      let content = (comment.content || '').replace(/<img\b[^>]*>/gi, '[图片]')
      if (comment.imageList && comment.imageList.length) {
        const images = comment.imageList.map(() => '[图片]').join('\n')
        content = content ? `${content}\n${images}` : images
      }
      return content
    },
    tabLink(tab, order) {
      const query = { tab }
      if (tab === 'comments')
        query.asc_order = order == null ? this.ascOrder : order
      return { path: `/user/${this.user.id}`, query }
    },
  },
}
</script>

<style lang="scss" scoped>
.tabs-warp {
  background-color: var(--bg-color);
  padding: 0 10px 10px;

  .tabs {
    margin-bottom: 5px;
  }

  .favorite-list,
  .user-comments {
    margin: 0;
  }

  .favorite-item,
  .user-comment {
    padding: 8px 0;
    border-bottom: 1px solid var(--border-color);
  }

  .favorite-title a {
    color: var(--text-color3);
    font-size: 18px;
  }

  .favorite-summary {
    color: var(--text-color);
    font-size: 14px;
    padding-top: 6px;
  }

  .favorite-meta,
  .comment-meta {
    color: var(--text-color3);
    font-size: 13px;
    padding-top: 6px;

    a {
      color: var(--text-link-color);
      margin-right: 10px;
    }
  }

  .comment-order {
    text-align: right;
    padding: 5px 0;

    a {
      color: var(--text-link-color);
      &.active {
        font-weight: bold;
      }
    }
  }

  .comment-content {
    padding-top: 6px;
    word-break: break-word;
    white-space: pre-line;
  }
}
</style>

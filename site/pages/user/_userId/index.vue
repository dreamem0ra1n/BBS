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
                <li
                  v-if="isOwner"
                  :class="{ 'is-active': activeTab === 'comments' }"
                >
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
                <page-jump-pagination
                  :paging="topicsPage.page"
                  :page="topicsPage.page.page"
                  :url-prefix="paginationUrlPrefix"
                  placement="top"
                  standalone
                />
                <topic-list :topics="topicsPage.results" :show-avatar="false" />
                <page-jump-pagination
                  :paging="topicsPage.page"
                  :page="topicsPage.page.page"
                  :url-prefix="paginationUrlPrefix"
                  placement="bottom"
                  standalone
                />
              </div>
              <div v-else class="notification is-primary">暂无话题</div>
            </div>

            <div v-else-if="activeTab === 'favorites'">
              <div v-if="favoritesPrivate" class="notification is-primary">
                浪潮的秘密暂未公开:)
              </div>
              <page-jump-pagination
                v-if="!favoritesPrivate && favoritesPage.page"
                :paging="favoritesPage.page"
                :page="favoritesPage.page.page"
                :url-prefix="paginationUrlPrefix"
                placement="top"
                standalone
              />
              <ul
                v-if="
                  !favoritesPrivate &&
                  favoritesPage.results &&
                  favoritesPage.results.length
                "
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
              <div
                v-else-if="!favoritesPrivate"
                class="notification is-primary"
              >
                暂无收藏
              </div>
              <page-jump-pagination
                v-if="!favoritesPrivate && favoritesPage.page"
                :paging="favoritesPage.page"
                :page="favoritesPage.page.page"
                :url-prefix="paginationUrlPrefix"
                placement="bottom"
                standalone
              />
            </div>

            <div v-else-if="activeTab === 'comments'">
              <page-jump-pagination
                v-if="commentsPage.page"
                :paging="commentsPage.page"
                :page="commentsPage.page.page"
                :url-prefix="paginationUrlPrefix"
                placement="top"
                standalone
              />
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
                    class="comment-content content"
                    v-html="commentDisplayContent(comment)"
                  />
                </li>
              </ul>
              <div v-else class="notification is-primary">暂无回复</div>
              <page-jump-pagination
                v-if="commentsPage.page"
                :paging="commentsPage.page"
                :page="commentsPage.page.page"
                :url-prefix="paginationUrlPrefix"
                placement="bottom"
                standalone
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
  async asyncData({ $axios, params, query, error, store }) {
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

    const currentUser = store.state.user.current
    const isOwner = !!(currentUser && currentUser.id === user.id)
    const requestedTab = tabs.includes(query.tab) ? query.tab : defaultTab
    const activeTab =
      requestedTab === 'comments' && !isOwner ? defaultTab : requestedTab
    const page = query.p || 1
    let topicsPage = null
    let favoritesPage = null
    let commentsPage = null
    let favoritesPrivate = false
    if (activeTab === 'topics') {
      topicsPage = await $axios.get('/api/topic/user/topics', {
        params: { userId: params.userId, page },
      })
    } else if (activeTab === 'favorites') {
      favoritesPrivate = !isOwner && !user.publicFavorites
      if (favoritesPrivate) {
        favoritesPage = { results: [] }
      } else {
        favoritesPage = await $axios.get('/api/user/favorites', {
          params: { userId: params.userId, page },
        })
      }
    } else {
      commentsPage = await $axios.get('/api/comment/user/comments', {
        params: { userId: params.userId, page },
      })
    }
    return {
      activeTab,
      user,
      topicsPage,
      favoritesPage,
      commentsPage,
      favoritesPrivate,
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
      return `/user/${this.user.id}?tab=${this.activeTab}&p=`
    },
  },
  watchQuery: ['tab', 'p'],
  methods: {
    commentDisplayContent(comment) {
      let content = (comment.content || '').replace(/<img\b[^>]*>/gi, '[图片]')
      if (comment.imageList && comment.imageList.length) {
        const images = comment.imageList.map(() => '[图片]').join('\n')
        content = content ? `${content}\n${images}` : images
      }
      return content
    },
    tabLink(tab) {
      return { path: `/user/${this.user.id}`, query: { tab } }
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

  .comment-content {
    padding-top: 6px;
    word-break: break-word;
    white-space: pre-line;
  }
}
</style>

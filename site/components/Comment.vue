<template>
  <div class="comment-component">
    <div class="comment-header">
      <span v-if="commentCount > 0">{{ commentCount }}条评论</span>
      <span v-else>评论</span>
      <span class="comment-header-right">
        <span v-if="totalPages > 1" class="comment-pagination">
          <template v-for="(item, idx) in pageItems">
            <nuxt-link
              v-if="item.type === 'num' && item.n !== currentPage"
              :key="'n' + idx"
              class="comment-page-item"
              :to="pageUrl(item.n)"
              >{{ item.n }}</nuxt-link
            >
            <span
              v-else-if="item.type === 'num'"
              :key="'n' + idx"
              class="comment-page-item current"
              >{{ item.n }}</span
            >
            <nuxt-link
              v-else-if="item.type === 'prev' && item.n >= 1"
              :key="'p' + idx"
              class="comment-page-item"
              :to="pageUrl(item.n)"
              >&lt;</nuxt-link
            >
            <span
              v-else-if="item.type === 'prev'"
              :key="'p' + idx"
              class="comment-page-item disabled"
              >&lt;</span
            >
            <nuxt-link
              v-else-if="item.type === 'next' && item.n <= totalPages"
              :key="'x' + idx"
              class="comment-page-item"
              :to="pageUrl(item.n)"
              >&gt;</nuxt-link
            >
            <span v-else :key="'x' + idx" class="comment-page-item disabled"
              >&gt;</span
            >
          </template>
        </span>
        <span class="comment-order" @click="changeOrder">{{
          ascOrder === 1 ? '正序' : '倒序'
        }}</span>
      </span>
    </div>

    <template v-if="isLogin">
      <comment-input
        v-if="!noComment"
        ref="input"
        :mode="mode"
        :entity-id="entityId"
        :entity-type="entityType"
        @created="commentCreated"
      />
    </template>
    <div v-else class="comment-not-login">
      <div class="comment-login-div">
        请
        <a style="font-weight: 700" @click="toLogin">登录</a>后发表观点
      </div>
    </div>

    <comment-list
      ref="list"
      :entity-id="entityId"
      :entity-type="entityType"
      :comments-page="commentsPage"
      @reply="reply"
      @deleted="$emit('deleted')"
    />
  </div>
</template>

<script>
export default {
  props: {
    mode: {
      type: String,
      default: 'markdown',
    },
    entityType: {
      type: String,
      default: '',
      required: true,
    },
    entityId: {
      type: Number || String,
      default: 0,
      required: true,
    },
    commentsPage: {
      type: Object,
      default() {
        return {}
      },
    },
    page: {
      type: Number,
      default: 1,
    },
    commentCount: {
      type: Number,
      default: 0,
    },
    reGain: {
      type: Function,
      default: (order) => {},
    },
    noComment: {
      type: Boolean,
      default: false,
    },
  },
  computed: {
    isLogin() {
      return this.$store.state.user.current != null
    },
    user() {
      return this.$store.state.user.current
    },
    config() {
      return this.$store.state.config.config
    },
    ascOrder() {
      return this.$store.state.env.ascOrder
    },
    totalPages() {
      const p = this.commentsPage && this.commentsPage.page
      if (!p || !p.total) return 1
      const limit = p.limit || 10
      return Math.max(1, Math.ceil(p.total / limit))
    },
    currentPage() {
      return Math.min(Math.max(this.page || 1, 1), this.totalPages)
    },
    pageItems() {
      const total = this.totalPages
      const current = this.currentPage
      if (total <= 7) {
        const arr = []
        for (let i = 1; i <= total; i++) arr.push({ type: 'num', n: i })
        return arr
      }
      const items = [{ type: 'num', n: 1 }]
      items.push({ type: 'prev', n: current - 1 })
      let start = current - 2
      let end = current + 2
      if (start < 2) {
        end += 2 - start
        start = 2
      }
      if (end > total - 1) {
        start -= end - (total - 1)
        end = total - 1
      }
      if (start < 2) start = 2
      if (end > total - 1) end = total - 1
      for (let i = start; i <= end; i++) items.push({ type: 'num', n: i })
      items.push({ type: 'next', n: current + 1 })
      items.push({ type: 'num', n: total })
      return items
    },
  },
  methods: {
    pageUrl(page) {
      const base = '/topic/' + this.entityId
      return base + '/' + page
    },
    commentCreated() {
      this.$emit('reGain', this.ascOrder)
      this.$emit('created')
    },
    reply(quote) {
      this.$refs.input.reply(quote)
    },
    toLogin() {
      this.$toSignin()
    },
    changeOrder() {
      let order
      if (this.ascOrder === 0) order = 1
      else order = 0
      this.$emit('reGain', order)
    },
  },
}
</script>
<style lang="scss" scoped>
.comment-component {
  background-color: var(--bg-color);
  border-radius: 3px;
  .comment-header {
    padding-top: 20px;
    margin: 0 10px;
    color: var(--text-color);
    font-size: 16px;
    font-weight: 500;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .comment-header-right {
    display: flex;
    align-items: center;
    font-size: 14px;
    font-weight: 400;
  }

  .comment-pagination {
    display: flex;
    align-items: center;
    margin-right: 12px;
  }

  .comment-page-item {
    display: inline-block;
    min-width: 22px;
    height: 22px;
    line-height: 22px;
    text-align: center;
    margin: 0 2px;
    padding: 0 4px;
    color: var(--text-color3);
    cursor: pointer;
    border-radius: 3px;

    &:hover {
      color: var(--text-link-color);
    }

    &.current {
      color: #fff;
      background-color: var(--qsc-color);
    }

    &.disabled {
      color: var(--text-color4);
      cursor: not-allowed;
    }
  }

  .comment-not-login {
    margin: 10px;
    border: 1px solid var(--border-color);
    border-radius: 0;
    overflow: hidden;
    position: relative;
    padding: 10px;
    box-sizing: border-box;

    .comment-login-div {
      color: var(--text-color4);
      cursor: pointer;
      border-radius: 3px;
      padding: 0 10px;

      a {
        margin-left: 10px;
        margin-right: 10px;
      }
    }
  }
  .comment-order {
    cursor: pointer;
  }
}
</style>

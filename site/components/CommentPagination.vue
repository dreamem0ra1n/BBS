<template>
  <nav
    v-if="totalPages > 1"
    class="comment-pagination"
    :class="{
      'is-header': placement === 'top',
      'is-bottom': placement === 'bottom',
    }"
    role="navigation"
    aria-label="评论分页"
  >
    <div class="comment-page-links">
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
          aria-current="page"
          >{{ item.n }}</span
        >
        <nuxt-link
          v-else-if="item.type === 'prev' && item.n >= 1"
          :key="'p' + idx"
          class="comment-page-item"
          :to="pageUrl(item.n)"
          aria-label="上一页"
          >&lt;</nuxt-link
        >
        <span
          v-else-if="item.type === 'prev'"
          :key="'p' + idx"
          class="comment-page-item disabled"
          aria-disabled="true"
          >&lt;</span
        >
        <nuxt-link
          v-else-if="item.type === 'next' && item.n <= totalPages"
          :key="'x' + idx"
          class="comment-page-item"
          :to="pageUrl(item.n)"
          aria-label="下一页"
          >&gt;</nuxt-link
        >
        <span
          v-else
          :key="'x' + idx"
          class="comment-page-item disabled"
          aria-disabled="true"
          >&gt;</span
        >
      </template>
    </div>

    <form class="comment-page-jump" @submit.prevent="goToPage">
      <label :for="pageInputId">跳转至</label>
      <input
        :id="pageInputId"
        v-model="pageInput"
        class="comment-page-input"
        type="text"
        inputmode="numeric"
        pattern="[0-9]*"
        placeholder="页码"
        :maxlength="pageInputMaxLength"
        aria-label="页码"
        @input="onPageInput"
        @keydown.enter.prevent="goToPage"
      />
      <button type="submit" class="comment-page-submit">Go</button>
    </form>
  </nav>
</template>

<script>
export default {
  props: {
    commentsPage: {
      type: Object,
      default() {
        return {}
      },
      required: true,
    },
    entityId: {
      type: [Number, String],
      default: 0,
      required: true,
    },
    page: {
      type: Number,
      default: 1,
    },
    placement: {
      type: String,
      default: 'top',
    },
  },
  data() {
    return {
      pageInput: '',
    }
  },
  computed: {
    totalPages() {
      const paging = this.commentsPage && this.commentsPage.page
      if (!paging || !paging.total) {
        return 1
      }
      const limit = paging.limit || 10
      return Math.max(1, Math.ceil(paging.total / limit))
    },
    currentPage() {
      return Math.min(Math.max(this.page || 1, 1), this.totalPages)
    },
    pageInputMaxLength() {
      return String(this.totalPages).length
    },
    pageInputId() {
      return `comment-page-input-${this.entityId}-${this.placement}`
    },
    pageItems() {
      const total = this.totalPages
      const current = this.currentPage
      if (total <= 7) {
        const items = []
        for (let i = 1; i <= total; i++) {
          items.push({ type: 'num', n: i })
        }
        return items
      }

      const items = [
        { type: 'num', n: 1 },
        { type: 'prev', n: current - 1 },
      ]
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
      start = Math.max(start, 2)
      end = Math.min(end, total - 1)
      for (let i = start; i <= end; i++) {
        items.push({ type: 'num', n: i })
      }
      items.push({ type: 'next', n: current + 1 }, { type: 'num', n: total })
      return items
    },
  },
  watch: {
    currentPage: {
      immediate: true,
      handler(value) {
        this.pageInput = String(value)
      },
    },
  },
  methods: {
    pageUrl(page) {
      return `/topic/${this.entityId}/${page}`
    },
    onPageInput(event) {
      this.pageInput = event.target.value.slice(0, this.pageInputMaxLength)
    },
    goToPage() {
      const value = String(this.pageInput || '').trim()
      if (!value) {
        this.showPageError('请输入页码')
        return
      }
      if (!/^\d+$/.test(value)) {
        this.showPageError('页码只能输入数字')
        return
      }
      if (value.length > this.pageInputMaxLength) {
        this.showPageError(`页码不能超过${this.pageInputMaxLength}位`)
        return
      }

      const targetPage = Number(value)
      if (
        !Number.isInteger(targetPage) ||
        targetPage < 1 ||
        targetPage > this.totalPages
      ) {
        this.showPageError(`请输入1-${this.totalPages}之间的页码`)
        return
      }
      if (targetPage !== this.currentPage) {
        this.$router.push(this.pageUrl(targetPage))
      }
    },
    showPageError(message) {
      if (this.$message && this.$message.warning) {
        this.$message.warning(message)
      }
    },
  },
}
</script>

<style lang="scss" scoped>
.comment-pagination {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  color: var(--text-color3);
  font-size: 14px;

  &.is-header {
    margin-right: 12px;
  }

  &.is-bottom {
    justify-content: flex-end;
    margin-top: 8px;
    padding: 10px 10px 14px;
  }
}

.comment-page-links {
  display: flex;
  align-items: center;
}

.comment-page-item {
  display: inline-block;
  min-width: 22px;
  height: 22px;
  line-height: 22px;
  margin: 0 2px;
  padding: 0 4px;
  color: var(--text-color3);
  text-align: center;
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

.comment-page-jump {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}

.comment-page-input {
  display: inline-block;
  flex: 0 0 58px;
  width: 58px;
  min-width: 58px;
  height: 24px;
  padding: 0 4px;
  box-sizing: border-box;
  color: var(--text-color);
  text-align: center;
  background-color: var(--bg-color);
  border: 1px solid var(--text-color4);
  border-radius: 3px;

  &:focus {
    border-color: var(--text-link-color);
    outline: none;
  }
}

.comment-page-submit {
  height: 24px;
  padding: 0 7px;
  color: var(--text-color3);
  cursor: pointer;
  background-color: transparent;
  border: 1px solid var(--border-color);
  border-radius: 3px;

  &:hover {
    color: var(--text-link-color);
  }
}
</style>

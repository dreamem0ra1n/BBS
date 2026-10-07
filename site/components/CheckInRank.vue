<template>
  <div class="widget rank-card">
    <div class="widget-header">{{ title }}</div>
    <div class="widget-content rank-scroll" @scroll.passive="onScroll">
      <p v-if="loading && !results.length" class="rank-state">加载中…</p>
      <p v-else-if="error && !results.length" class="rank-state is-danger">
        {{ error }}
      </p>
      <p v-else-if="!results.length" class="rank-state">暂无数据</p>

      <div v-for="item in results" :key="item.rank" class="rank-row">
        <span class="rank-number">{{ item.rank }}</span>
        <avatar :user="item.user" :round="true" size="30" />
        <nuxt-link class="rank-name" :to="'/user/' + item.user.id">
          {{ item.user.nickname }}
        </nuxt-link>
        <span class="rank-value">
          {{
            today ? formatTime(item.checkInTime) : item.consecutiveDays + ' 天'
          }}
        </span>
      </div>

      <p v-if="error && results.length" class="rank-state is-danger">
        {{ error }}
      </p>
      <button
        v-if="!loading && (error || results.length < total)"
        class="button is-small is-fullwidth"
        @click="load"
      >
        {{ error ? '重试' : '加载更多' }}
      </button>
      <p v-else-if="loading" class="rank-state">加载中…</p>
      <p v-else-if="results.length" class="rank-state">没有更多了</p>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    title: { type: String, required: true },
    endpoint: { type: String, required: true },
    today: { type: Boolean, default: false },
  },
  data() {
    return { results: [], page: 1, total: 0, loading: false, error: '' }
  },
  mounted() {
    this.load()
  },
  methods: {
    refresh() {
      this.results = []
      this.page = 1
      this.total = 0
      this.error = ''
      this.load()
    },
    async load() {
      if (
        this.loading ||
        (!this.error && this.page > 1 && this.results.length >= this.total)
      )
        return
      this.loading = true
      this.error = ''
      try {
        const data = await this.$axios.get(this.endpoint, {
          params: { page: this.page, pageSize: 20 },
        })
        this.results = this.results.concat(data.results || [])
        this.total = data.page ? data.page.total : this.results.length
        this.page += 1
      } catch (error) {
        this.error = error.message || '排行榜加载失败'
      } finally {
        this.loading = false
      }
    },
    onScroll(event) {
      const element = event.target
      if (element.scrollTop + element.clientHeight >= element.scrollHeight - 24)
        this.load()
    },
    formatTime(timestamp) {
      return new Date(timestamp).toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
      })
    },
  },
}
</script>

<style lang="scss" scoped>
.rank-scroll {
  max-height: 360px;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}
.rank-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 0;
  border-bottom: 1px solid var(--border-color);
}
.rank-number {
  width: 22px;
  flex: none;
  color: var(--text-color3);
  text-align: center;
}
.rank-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rank-value {
  color: var(--text-color3);
  font-size: 0.8rem;
  white-space: nowrap;
}
.rank-state {
  color: var(--text-color3);
  text-align: center;
  padding: 12px 0;
  font-size: 0.85rem;
}
.rank-state.is-danger {
  color: #f14668;
}
</style>

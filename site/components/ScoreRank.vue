<template>
  <div v-if="scoreRank && scoreRank.length" class="widget">
    <div class="widget-header">
      <span class="widget-title">积分排行</span>
      <span class="score-rank-tabs">
        <button
          type="button"
          :class="{ active: activePeriod === 'all' }"
          :disabled="loading"
          @click="selectPeriod('all')"
        >
          历史总榜
        </button>
        <button
          type="button"
          :class="{ active: activePeriod === 'year' }"
          :disabled="loading"
          @click="selectPeriod('year')"
        >
          年度总榜
        </button>
      </span>
    </div>
    <div class="widget-content">
      <ul v-if="displayedScoreRank.length" class="score-rank">
        <li v-for="user in displayedScoreRank" :key="user.id">
          <avatar :user="user" size="35" round />
          <div class="score-user-info">
            <nuxt-link :to="'/user/' + user.id" class="score-nickname">{{
              user.nickname
            }}</nuxt-link>
            <p class="score-desc">
              {{ user.topicCount }} 帖子 • {{ user.commentCount }} 评论
            </p>
          </div>
          <div class="score-rank-info">
            <span class="score-user-score">
              <i class="iconfont icon-score" /><span>{{ user.score }}</span>
            </span>
          </div>
        </li>
      </ul>
      <div v-else class="score-rank-empty">近一年暂无注册用户</div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    scoreRank: {
      type: Array,
      default() {
        return null
      },
    },
  },
  data() {
    return {
      activePeriod: 'all',
      annualScoreRank: null,
      loading: false,
    }
  },
  computed: {
    displayedScoreRank() {
      return this.activePeriod === 'year'
        ? this.annualScoreRank || []
        : this.scoreRank
    },
  },
  methods: {
    async selectPeriod(period) {
      if (period === this.activePeriod || this.loading) {
        return
      }
      if (period === 'all') {
        this.activePeriod = period
        return
      }
      this.loading = true
      try {
        this.annualScoreRank = await this.$axios.get('/api/user/score/rank', {
          params: { period: 'year' },
        })
        this.activePeriod = period
      } catch (e) {
        this.$message.error(e.message || '加载年度排行失败')
      } finally {
        this.loading = false
      }
    },
  },
}
</script>

<style scoped lang="scss">
.score-rank-tabs {
  display: flex;
  gap: 8px;
  font-size: 12px;
  font-weight: 400;

  button {
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--text-color3);
    cursor: pointer;

    &.active {
      color: var(--text-link-color);
    }

    &:disabled {
      cursor: wait;
    }
  }
}

.score-rank-empty {
  padding: 20px 0;
  color: var(--text-color3);
  font-size: 12px;
  text-align: center;
}

.score-rank {
  li {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    list-style: none;
    font-size: 13px;
    position: relative;
    padding: 10px 0;

    &:not(:last-child) {
      border-bottom: 1px solid var(--border-color);
    }

    .score-user-info {
      width: 100%;
      margin-left: 9px;
      line-height: 1.4;
      font-size: 12px;
      .score-nickname {
        font-size: 14px;
        color: var(--text-color);
        line-height: 20px;

        &:hover {
          color: rgba(0, 166, 244, 0.8);
        }
      }
      .score-desc {
        font-size: 11px;
        color: var(--text-color3);
        line-height: 20px;
        display: block;
      }
    }

    .score-rank-info {
      width: 120px;
      .score-user-score {
        float: right;
        border-radius: 12px;
        color: var(--text-color3);
        height: 21px;
        line-height: 21px;
        padding: 0 6px;
        text-shadow: 0 0 1px #fff;
        background-color: var(--bg-color2);
        font-size: 0.75rem;
        align-items: center;
        i {
          margin-right: 3px;
        }
      }
    }
  }
}
</style>

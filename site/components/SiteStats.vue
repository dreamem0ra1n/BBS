<template>
  <div v-if="stats" class="widget">
    <div class="widget-header">统计信息</div>
    <div class="widget-content stats-grid">
      <div class="stats-item">
        <strong>{{ stats.totalTopics }}</strong>
        <span>历史总主题数</span>
      </div>
      <div class="stats-item">
        <strong>{{ stats.monthlyTopics }}</strong>
        <span>当月主题数</span>
      </div>
      <div class="stats-item">
        <strong>{{ stats.totalUsers }}</strong>
        <span>总用户数</span>
      </div>
      <div class="stats-item">
        <strong>{{ stats.monthlyActiveUsers }}</strong>
        <span>当月活跃用户数</span>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'SiteStats',
  props: {
    stats: {
      type: Object,
      default() {
        return null
      },
    },
  },
}
</script>

<style lang="scss" scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px 8px;
  padding: 14px 0;
}

.stats-item {
  display: flex;
  flex-direction: column;
  align-items: center;

  strong {
    color: var(--text-color3);
    font-size: 18px;
  }

  span {
    margin-top: 4px;
    color: var(--text-color4);
    font-size: 12px;
  }
}
</style>

<template>
  <div v-if="config.siteNotification" class="widget">
    <div class="widget-header">
      <span>公告</span>
      <button
        class="notice-toggle"
        type="button"
        @click="collapsed = !collapsed"
      >
        {{ collapsed ? '展开' : '收起' }}
      </button>
    </div>
    <div class="widget-content content notice-content" :class="{ collapsed }">
      <p v-html="config.siteNotification"></p>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      collapsed: true,
    }
  },
  computed: {
    config() {
      return this.$store.state.config.config
    },
  },
}
</script>

<style lang="scss" scoped>
.notice-content {
  font-size: 80%;
  overflow: hidden;

  &.collapsed {
    max-height: 120px;
  }
}

.notice-toggle {
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--text-color3);
  cursor: pointer;
  font-size: 12px;
}
</style>

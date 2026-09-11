<template>
  <div v-if="currentUser && enabled" class="widget online-users">
    <div class="widget-header">当前在线</div>
    <div v-if="users.length" class="widget-content">
      <div v-for="user in visibleUsers" :key="user.id" class="online-user">
        <avatar :user="user" :size="30" />
        <div class="online-user-info">
          <nuxt-link :to="'/user/' + user.id"
            ><strong>{{ user.nickname }}</strong></nuxt-link
          >
          <span>{{ user.greeting || '正在浏览 BBS' }}</span>
        </div>
      </div>
      <div
        v-if="expanded && (canCollapse || hiddenCount)"
        class="online-footer"
      >
        <span v-if="hiddenCount" class="online-note"
          >仅显示前 {{ users.length }} 位</span
        >
        <button
          v-if="canCollapse"
          type="button"
          class="online-toggle"
          @click="expanded = false"
        >
          收起
        </button>
      </div>
      <button
        v-else-if="hiddenCount"
        type="button"
        class="online-toggle"
        @click="expanded = true"
      >
        还有 {{ hiddenCount }} 人在线
      </button>
    </div>
    <div v-else class="widget-content empty">暂无在线用户</div>
  </div>
  <div v-else-if="!currentUser" class="widget online-users-login">请先登录</div>
</template>

<script>
const COLLAPSED_LIMIT = 10

export default {
  name: 'OnlineUsers',
  data() {
    return { expanded: false, collapsedLimit: COLLAPSED_LIMIT }
  },
  computed: {
    currentUser() {
      return this.$store.state.user.current
    },
    enabled() {
      return this.$store.state.presence.enabled
    },
    users() {
      return this.$store.state.presence.users || []
    },
    total() {
      return this.$store.state.presence.total || 0
    },
    totalCount() {
      return Math.max(this.total, this.users.length)
    },
    visibleUsers() {
      return this.expanded
        ? this.users
        : this.users.slice(0, this.collapsedLimit)
    },
    hiddenCount() {
      return Math.max(0, this.totalCount - this.visibleUsers.length)
    },
    canCollapse() {
      return this.users.length > this.collapsedLimit || this.hiddenCount > 0
    },
  },
  watch: {
    users(users) {
      if (users.length <= this.collapsedLimit) this.expanded = false
    },
  },
}
</script>

<style lang="scss" scoped>
.online-user {
  display: flex;
  align-items: center;
  padding: 7px 0;
  color: inherit;
}
.online-user-info {
  min-width: 0;
  margin-left: 9px;
  display: flex;
  flex-direction: column;
}
.online-user-info strong,
.online-user-info span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.online-user-info strong {
  font-size: 13px;
}
.online-user-info span {
  margin-top: 2px;
  color: var(--text-color3);
  font-size: 12px;
}
.online-toggle {
  display: block;
  margin-top: 6px;
  padding: 0;
  border: none;
  background: none;
  color: var(--text-color3);
  font-size: 12px;
  cursor: pointer;
}
.online-toggle:hover {
  color: var(--text-color2);
}
.online-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
}
.online-footer .online-toggle {
  margin-top: 0;
}
.online-note {
  color: var(--text-color4);
  font-size: 12px;
}
.empty,
.online-users-login {
  color: var(--text-color4);
  font-size: 13px;
}
</style>

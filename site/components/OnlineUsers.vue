<template>
  <div v-if="currentUser && enabled" class="widget online-users">
    <div class="widget-header">当前在线</div>
    <div v-if="users.length" class="widget-content">
      <div v-for="user in users" :key="user.id" class="online-user">
        <avatar :user="user" :size="30" />
        <div class="online-user-info">
          <nuxt-link :to="'/user/' + user.id"
            ><strong>{{ user.nickname }}</strong></nuxt-link
          >
          <span>{{ user.greeting || '正在浏览 BBS' }}</span>
        </div>
      </div>
    </div>
    <div v-else class="widget-content empty">暂无在线用户</div>
  </div>
  <div v-else-if="!currentUser" class="widget online-users-login">请先登录</div>
</template>

<script>
export default {
  name: 'OnlineUsers',
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
.empty,
.online-users-login {
  color: var(--text-color4);
  font-size: 13px;
}
</style>

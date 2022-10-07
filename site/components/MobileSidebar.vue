<template>
  <div class="mobile-sidebar">
    <transition name="fadeLeft">
      <div v-show="show" class="sidebar-container">
        <div v-if="siteNavs && siteNavs.length" class="sidebar-navs">
          <div
            v-for="(nav, index) in siteNavs"
            :key="index"
            class="sidebar-nav-item"
          >
            <i class="iconfont icon-nav" />
            <my-link :to="nav.url">{{ nav.title }}</my-link>
          </div>
        </div>
        <div class="sidebar-message">
          <i class="iconfont icon-message" />
          <my-link to="/user/messages">消息</my-link>
        </div>
        <template v-if="user">
          <div class="sidebar-userinfo">
            <i class="iconfont icon-username" />
            <span>{{ user.nickname }}</span>
          </div>
          <div class="sidebar-menus">
            <div class="sidebar-menu-item">
              <my-link :to="'/user/' + user.id">个人中心</my-link>
            </div>
            <div class="sidebar-menu-item">
              <my-link class="sidebar-menu-item" to="/user/favorites"
                >我的收藏</my-link
              >
            </div>
            <div class="sidebar-menu-item">
              <my-link class="sidebar-menu-item" to="/user/profile"
                >编辑资料</my-link
              >
            </div>
          </div>
        </template>
        <template v-else>
          <a class="sidebar-login-btn button is-primary" href="loginUrl"
            >登录
          </a>
        </template>
      </div>
    </transition>
  </div>
</template>

<script>
import UserHelper from '~/common/UserHelper'
export default {
  computed: {
    show() {
      return this.$store.state.env.showMobileSidebar
    },
    user() {
      return this.$store.state.user.current
    },
    isOwnerOrAdmin() {
      return UserHelper.isOwner(this.user) || UserHelper.isAdmin(this.user)
    },
    config() {
      return this.$store.state.config.config
    },
    siteNavs() {
      const config = this.$store.state.config.config
      return config.siteNavs || []
    },
    loginUrl() {
      return (
        'https://www.qsc.zju.edu.cn/passport/v4/?success=' +
        this.$store.state.env.currentDomain
      )
    },
  },
  methods: {
    async signout() {
      try {
        await this.$store.dispatch('user/signout')
        this.$linkTo('/')
      } catch (e) {
        console.error(e)
      }
    },
  },
}
</script>
<style lang="scss" scoped></style>

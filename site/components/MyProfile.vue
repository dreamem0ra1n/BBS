<template>
  <div class="widget">
    <div class="widget-header">
      <span>个人资料</span>
      <div class="slot">
        <nuxt-link v-if="isCurrentUser" to="/user/profile">
          编辑资料
        </nuxt-link>
      </div>
    </div>
    <div class="widget-content stable">
      <div
        v-for="(info, index) in infos"
        :key="index + info.attribute"
        class="str"
      >
        <div class="slabel">{{ info.title }}</div>
        <div class="svalue">
          {{ String(user[info.attribute]) }}
        </div>
      </div>

      <div v-if="user.homePage" class="str">
        <div class="slabel">主页</div>
        <div class="svalue">
          <a :href="user.homePage" target="_blank" rel="nofollow">{{
            user.homePage
          }}</a>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MyProfile',
  props: {
    user: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      infos: [
        { title: 'id', attribute: 'nickname' },
        { title: '职位', attribute: 'position' },
        { title: '签名', attribute: 'description' },
        { title: '专业', attribute: 'major' },
        { title: '生日', attribute: 'birthday' },
        { title: '手机号码', attribute: 'mobile' },
        { title: '微信', attribute: 'wechat' },
        { title: 'QQ', attribute: 'qq' },
      ],
    }
  },
  computed: {
    isCurrentUser() {
      const current = this.$store.state.user.current
      return current && this.user.id === current.id
    },
  },
}
</script>

<style scoped></style>

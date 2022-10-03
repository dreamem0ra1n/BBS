<template>
  <div class="widget">
    <div class="widget-header">
      <span>个人资料</span>
      <div class="slot">
        <nuxt-link to="/user/profile">编辑资料</nuxt-link>
      </div>
    </div>
    <div class="widget-content stable">
      <div
        class="str"
        v-for="(info, index) in infos"
        :key="index + info.attribute"
      >
        <div class="slabel">{{ info.title }}</div>
        <div class="svalue">{{ user[info.attribute] }}</div>
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
        { title: '昵称', attribute: 'nickname' },
        { title: '签名', attribute: 'description' },
        { title: '专业', attribute: 'major' },
        { title: '生日', attribute: 'birthday' },
        { title: '手机号码', attribute: 'mobile' },
        { title: '微信', attribute: 'wechat' },
        { title: 'QQ', attribute: 'qq' },
      ],
    }
  },
}
</script>

<style scoped></style>

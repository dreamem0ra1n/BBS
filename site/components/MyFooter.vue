<template>
  <footer class="footer">
    <div class="container content">
      <div>
        <nuxt-link to="/about">关于</nuxt-link>
        <nuxt-link to="/tags">标签</nuxt-link>
        <nuxt-link to="/links">友链</nuxt-link>
        <span @click="BG" class="go-bg">一键BG</span>
      </div>
      <div>
        © 2022 Powered by
        <a href="https://docs.bbs-go.com" target="_blank" class="light"
          >BBS-GO</a
        >
      </div>
    </div>
  </footer>
</template>

<script>
export default {
  methods: {
    async BG() {
      const me = this
      try {
        const topic = await this.$axios.post('/api/topic/create', {
          type: 0,
          nodeId: 11,
          title: '我要BG',
          content: '我要BG',
          access_lv: 1,
          imageList: '',
        })
        this.$msg({
          message: '提交成功',
          onClose() {
            me.$linkTo('/topic/' + topic.topicId)
          },
        })
      } catch (e) {
        this.$message.error(e.message || e)
      }
    },
  },
}
</script>

<style lang="scss" scoped>
.footer {
  font-size: 14px;
  color: var(--text-color3);
  background: none;
  text-align: left;
  margin: 0 10px;
  a {
    color: var(--text-color3);
    text-decoration: none;
  }

  .light {
    color: #eb5424; // TODO
    font-weight: bold;
  }
  .go-bg {
    transition: all 0.3s;
    cursor: pointer;
    opacity: 0%;
    color: #eb5424;
  }
  .go-bg:hover {
    opacity: 100%;
  }
}
</style>

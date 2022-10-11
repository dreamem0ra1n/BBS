<template>
  <footer class="footer">
    <div class="container content">
      <div>
        <my-link to="/about">关于</my-link>
        <my-link to="/tags">标签</my-link>
        <my-link to="/links">友链</my-link>
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
export default {}
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
}
</style>

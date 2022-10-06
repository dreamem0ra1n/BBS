<template>
  <div class="login-container">
    <el-form
      ref="loginForm"
      :model="loginForm"
      label-position="left"
      class="login-form"
      autocomplete="on"
    >
      <div class="title-container">
        <h3 class="title">登录</h3>
      </div>
      <el-button
        :loading="loading"
        type="primary"
        style="width: 100%; margin-bottom: 30px"
        @click.native.prevent="handleLogin"
      >
        Login
      </el-button>
    </el-form>
  </div>
</template>

<script>
export default {
  data() {
    return {
      capsTooltip: false,
      loading: false,
      redirect: undefined,
      otherQuery: {},
    };
  },
  watch: {
    $route: {
      handler(route) {
        const { query } = route;
        if (query) {
          this.redirect = query.redirect;
          this.otherQuery = this.getOtherQuery(query);
        }
      },
      immediate: true,
    },
  },
  mounted() {
    this.$store.dispatch("user/login");
  },
  methods: {
    handleLogin() {
      this.$refs.loginForm.validate((valid) => {
        if (valid) {
          this.loading = true;
          this.$store
            .dispatch("user/login")
            .then(() => {
              this.$router.push({ path: this.redirect || "/", query: this.otherQuery });
              this.loading = false;
            })
            .catch((e) => {
              this.loading = false;
            });
          return true;
        }
        return false;
      });
    },
    getOtherQuery(query) {
      return Object.keys(query).reduce((acc, cur) => {
        if (cur !== "redirect") {
          acc[cur] = query[cur];
        }
        return acc;
      }, {});
    },
  },
};
</script>

<style lang="scss">
.login-container {
  min-height: 100%;
  width: 100%;
  background-color: #fff;
  overflow: hidden;

  .login-form {
    position: relative;
    width: 520px;
    max-width: 100%;
    padding: 160px 35px 0;
    margin: 0 auto;
    overflow: hidden;

    .captcha-code {
      & > div {
        display: flex;
        .captcha-code-img {
          // margin-left: 10px;
          img {
            height: 36px;
          }
        }
      }
    }
  }

  .tips {
    font-size: 14px;
    color: #fff;
    margin-bottom: 10px;

    span {
      &:first-of-type {
        margin-right: 16px;
      }
    }
  }

  .title-container {
    position: relative;

    .title {
      font-size: 26px;
      color: #000;
      margin: 0 auto 40px auto;
      text-align: center;
      font-weight: bold;
    }
  }
}
</style>

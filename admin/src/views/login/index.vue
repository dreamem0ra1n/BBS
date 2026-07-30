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
        <h3 class="title">
          {{ loginMethods.password ? "账号密码登录" : "求是潮 Passport 登录" }}
        </h3>
      </div>
      <template v-if="loginMethods.password">
        <el-form-item prop="username">
          <el-input
            v-model.trim="loginForm.username"
            autocomplete="username"
            placeholder="用户名"
            @keyup.enter.native="handleLogin"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            autocomplete="current-password"
            placeholder="密码"
            show-password
            @keyup.enter.native="handleLogin"
          />
        </el-form-item>
      </template>
      <el-button
        v-if="configLoaded && loginMethods.password"
        :loading="loading"
        type="primary"
        style="width: 100%; margin-bottom: 30px"
        @click.native.prevent="handleLogin"
      >
        登录
      </el-button>
      <el-button
        v-if="configLoaded && loginMethods.passport"
        :loading="loading"
        type="primary"
        style="width: 100%; margin: 0 0 30px 0"
        @click.native.prevent="handlePassportLogin"
      >
        通过 Passport 登录
      </el-button>
      <p v-if="configLoaded && !hasLoginMethod" class="login-error">当前没有可用的登录方式</p>
    </el-form>
  </div>
</template>

<script>
export default {
  data() {
    return {
      loginForm: {
        username: "",
        password: "",
      },
      configLoaded: false,
      loginMethods: {},
      loading: false,
      redirect: undefined,
      otherQuery: {},
    };
  },
  computed: {
    hasLoginMethod() {
      return this.loginMethods.password || this.loginMethods.passport;
    },
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
  async mounted() {
    try {
      const config = await this.axios.get("/api/config/configs");
      this.loginMethods = config.loginMethods || {};
    } catch {
      this.loginMethods = {};
    } finally {
      this.configLoaded = true;
    }

    if (!this.loginMethods.password && this.loginMethods.passport) {
      this.handlePassportLogin();
    }
  },
  methods: {
    handleLogin() {
      if (!this.loginForm.username || !this.loginForm.password) {
        this.$message.error("请输入用户名和密码");
        return;
      }

      this.loading = true;
      this.$store
        .dispatch("user/login", this.loginForm)
        .then(() => {
          this.finishLogin();
          this.loading = false;
        })
        .catch(() => {
          this.loading = false;
        });
    },
    handlePassportLogin() {
      this.loading = true;
      this.$store
        .dispatch("user/passportLogin")
        .then(() => {
          this.finishLogin();
          this.loading = false;
        })
        .catch(() => {
          window.location =
            "https://www.qsc.zju.edu.cn/passport/v4/static/index.html#/login?success=" +
            "https://www.qsc.zju.edu.cn/bbsadmin";
          this.loading = false;
        });
    },
    finishLogin() {
      this.$router.push({ path: this.redirect || "/", query: this.otherQuery });
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

  .login-error {
    color: #f56c6c;
    text-align: center;
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

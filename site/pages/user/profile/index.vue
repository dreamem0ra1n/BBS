<template>
  <div class="widget no-margin">
    <div class="widget-header">
      <div>
        <i class="iconfont icon-setting" />
        <span>个人资料</span>
      </div>
      <nuxt-link :to="'/user/' + user.id" style="font-size: 13px">
        <i class="iconfont icon-return" />
        <span>返回个人主页</span>
      </nuxt-link>
    </div>
    <div class="widget-content">
      <!-- 头像 -->
      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">头像</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <avatar-edit
                v-model="user.avatar"
                @success="onAvatarUpdateSuccess"
                @error="onAvatarUpdateError"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 昵称 -->
      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">昵称</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <div>{{ form.nickname }}</div>
            </div>
          </div>
        </div>
      </div>
      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">职位</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <div>{{ form.position }}</div>
            </div>
          </div>
        </div>
      </div>
      <!-- 个性签名 -->
      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">个性签名</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <textarea
                v-model="form.description"
                class="textarea"
                rows="2"
                placeholder="一句话介绍你自己"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 个人主页 -->
      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">个人主页</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <input
                v-model="form.homePage"
                class="input"
                type="text"
                autocomplete="off"
                placeholder="请输入个人主页"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">专业</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <input
                v-model="form.major"
                class="input"
                type="text"
                autocomplete="off"
                placeholder="请输入专业"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">生日</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <input
                v-model="form.birthday"
                class="input"
                type="text"
                autocomplete="off"
                placeholder="请输入生日"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">手机号</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <input
                v-model="form.mobile"
                class="input"
                type="text"
                autocomplete="off"
                placeholder="请输入手机号"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">微信</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <input
                v-model="form.wechat"
                class="input"
                type="text"
                autocomplete="off"
                placeholder="请输入微信"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="field is-horizontal">
        <div class="field-label is-normal">
          <label class="label">QQ</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <input
                v-model="form.qq"
                class="input"
                type="text"
                autocomplete="off"
                placeholder="请输入QQ"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="field is-horizontal">
        <div class="field-label is-normal" />
        <div class="field-body">
          <div class="field">
            <div class="control">
              <a class="button is-success" @click="submitForm">保存</a>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  middleware: 'authenticated',
  /* async asyncData({ $axios }) {
    const user = await $axios.get('/api/user/current')
    const form = { ...user }
    return {
      user,
      form,
    } 
  }, */
  mounted() {
    this.reload()
    this.user = this.$store.state.user.current
  },
  data() {
    return {
      user: {},
      form: {
        nickname: '',
        homePage: '',
        description: '',
        major: '',
        birthday: '',
        mobile: '',
        wechat: '',
        qq: '',
        role: '',
      },
    }
  },
  head() {
    return {
      title: this.$siteTitle(this.user.nickname + ' - 个人资料'),
    }
  },
  methods: {
    async submitForm() {
      try {
        await this.$axios.post('/api/user/edit/' + this.user.id, this.form)
        await this.reload()
        this.$message.success('资料修改成功')
      } catch (e) {
        console.error(e)
        this.$message.error('资料修改失败：' + (e.message || e))
      }
    },
    onAvatarUpdateSuccess() {
      this.$message.success('头像更新成功')
    },
    onAvatarUpdateError(e) {
      this.$message.error('头像更新失败')
    },
    async reload() {
      const _user = await this.$axios.get('/api/user/current')
      const pattern = _user.roles[0].split('_')
      const role = pattern[0]
      const node = await this.$axios.get('/api/topic/node?nodeId=' + pattern[1])
      const session = node.name
      _user.position = session + '-' + role
      this.$store.commit('user/setCurrent', _user)

      this.form = { ..._user }
    },
  },
}
</script>

<style lang="scss" scoped></style>

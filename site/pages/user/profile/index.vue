<template>
  <div class="widget no-margin">
    <div class="widget-header">
      <div>
        <i class="iconfont icon-username" />
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
              <div>{{ position }}</div>
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
                v-model="birthdayInput"
                class="input"
                type="date"
                :max="today"
                autocomplete="off"
                @input="birthdayTouched = true"
              />
              <p v-if="hasInvalidBirthday" class="help is-warning">
                原已填写内容：{{
                  originalBirthday
                }}。格式不正确，生日通知无法发送；重新选择日期后将更新。
              </p>
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
              <button
                class="button is-success"
                :class="{ 'is-loading': saving }"
                :disabled="saving"
                @click="submitForm"
              >
                保存
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import UserHelper from '~/common/UserHelper'

export default {
  middleware: 'authenticated',
  data() {
    return {
      user: {},
      birthdayInput: '',
      birthdayTouched: false,
      originalBirthday: '',
      today: '',
      saving: false,
      form: {
        nickname: '',
        homePage: '',
        description: '',
        major: '',
        birthday: '',
        birthdayBlessingEnabled: false,
        birthdayBlessingNotifyEnabled: false,
        birthdayBlessingNotifyAvailable: false,
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
  computed: {
    position() {
      return UserHelper.getPosition(this.form)
    },
    hasInvalidBirthday() {
      return (
        !this.birthdayTouched &&
        !!this.originalBirthday &&
        !this.isValidBirthday(this.originalBirthday)
      )
    },
  },
  /* async asyncData({ $axios }) {
    const user = await $axios.get('/api/user/current')
    const form = { ...user }
    return {
      user,
      form,
    }
  }, */
  mounted() {
    this.today = this.formatDate(new Date())
    this.reload()
    this.user = this.$store.state.user.current
  },
  methods: {
    async submitForm() {
      this.saving = true
      try {
        const birthday = this.birthdayTouched
          ? this.birthdayInput
          : this.originalBirthday
        await this.$axios.post('/api/user/edit/' + this.user.id, {
          ...this.form,
          birthday,
        })
        await this.reload()
        this.$message.success('资料修改成功')
      } catch (e) {
        console.error(e)
        this.$message.error('资料修改失败：' + (e.message || e))
      } finally {
        this.saving = false
      }
    },
    onAvatarUpdateSuccess() {
      this.$message.success('头像更新成功')
    },
    onAvatarUpdateError(e) {
      this.$message.error('头像更新失败')
    },
    formatDate(date) {
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      return `${year}-${month}-${day}`
    },
    isValidBirthday(value) {
      const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
      if (!match) {
        return false
      }
      const year = Number(match[1])
      const month = Number(match[2])
      const day = Number(match[3])
      const date = new Date(year, month - 1, day)
      return (
        date.getFullYear() === year &&
        date.getMonth() === month - 1 &&
        date.getDate() === day
      )
    },
    async reload() {
      const _user = await this.$axios.get('/api/user/current')
      if (_user) {
        this.$store.commit('user/setCurrent', _user)
        this.user = _user
        this.form = { ..._user }
        this.originalBirthday = _user.birthday || ''
        this.birthdayInput = this.isValidBirthday(this.originalBirthday)
          ? this.originalBirthday
          : ''
        this.birthdayTouched = false
      }
    },
  },
}
</script>

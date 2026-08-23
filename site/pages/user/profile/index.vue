<template>
  <div class="widget no-margin">
    <div class="widget-header">
      <div>
        <i class="iconfont icon-setting" />
        <span>个人资料&设置</span>
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

      <div
        v-if="
          birthdayBlessingAvailable &&
          (hasBirthday || form.birthdayBlessingNotifyAvailable)
        "
        class="field is-horizontal"
      >
        <div class="field-label is-normal">
          <label class="label">生日祝福</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div v-if="hasBirthday" class="control">
              <label class="notification-checkbox">
                <input
                  v-model="form.birthdayBlessingEnabled"
                  class="notification-checkbox-input"
                  type="checkbox"
                />
                <span class="notification-checkbox-box" aria-hidden="true" />
                <span>生日祝福附带随机潮人的留言</span>
              </label>
            </div>
            <div
              v-if="form.birthdayBlessingNotifyAvailable"
              class="control birthday-blessing-control"
            >
              <label class="notification-checkbox">
                <input
                  v-model="form.birthdayBlessingNotifyEnabled"
                  class="notification-checkbox-input"
                  type="checkbox"
                />
                <span class="notification-checkbox-box" aria-hidden="true" />
                <span>
                  自身祝福被收到时接收通知
                  <small class="secondary-setting-note">
                    （仅在后台有对应数据时有效）
                  </small>
                </span>
              </label>
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
        <div class="field-label is-normal">
          <label class="label">钉钉通知</label>
        </div>
        <div class="field-body">
          <div class="field">
            <div class="control">
              <label class="notification-checkbox">
                <input
                  v-model="dingTalk.enabled"
                  class="notification-checkbox-input"
                  type="checkbox"
                />
                <span class="notification-checkbox-box" aria-hidden="true" />
                <span>将站内消息发送到钉钉群机器人</span>
              </label>
            </div>

            <div v-if="dingTalk.enabled" class="ding-talk-settings">
              <div class="control">
                <span
                  v-if="dingTalk.webhookConfigured"
                  class="tag is-success is-light"
                >
                  Webhook 已配置
                </span>
              </div>

              <div class="control ding-talk-control">
                <input
                  v-model.trim="dingTalk.webhook"
                  class="input"
                  type="password"
                  autocomplete="new-password"
                  :placeholder="
                    dingTalk.webhookConfigured
                      ? '留空则继续使用已保存的 Webhook'
                      : 'https://oapi.dingtalk.com/robot/send?access_token=...'
                  "
                />
                <p class="help">仅支持钉钉群自定义机器人生成的官方 Webhook。</p>
              </div>

              <div class="control ding-talk-control">
                <input
                  v-model.trim="dingTalk.secret"
                  class="input"
                  type="password"
                  autocomplete="new-password"
                  :disabled="dingTalk.clearSecret"
                  :placeholder="
                    dingTalk.secretConfigured
                      ? '留空则继续使用已保存的加签密钥'
                      : '加签密钥（可选，以 SEC 开头）'
                  "
                />
                <label
                  v-if="dingTalk.secretConfigured"
                  class="notification-checkbox help"
                >
                  <input
                    v-model="dingTalk.clearSecret"
                    class="notification-checkbox-input"
                    type="checkbox"
                  />
                  <span class="notification-checkbox-box" aria-hidden="true" />
                  <span>清除已保存的加签密钥</span>
                </label>
              </div>

              <div class="control ding-talk-control">
                <input
                  v-model.trim="dingTalk.keyword"
                  class="input"
                  type="text"
                  maxlength="64"
                  autocomplete="off"
                  placeholder="自定义关键词（机器人启用关键词校验时填写）"
                />
              </div>
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
              <button
                v-if="dingTalk.enabled && dingTalk.webhookConfigured"
                class="button is-light"
                :class="{ 'is-loading': testingDingTalk }"
                :disabled="saving || testingDingTalk"
                @click="testDingTalk"
              >
                发送测试通知
              </button>
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
  data() {
    return {
      user: {},
      birthdayInput: '',
      birthdayTouched: false,
      originalBirthday: '',
      today: '',
      saving: false,
      testingDingTalk: false,
      dingTalk: {
        enabled: false,
        webhook: '',
        secret: '',
        keyword: '',
        webhookConfigured: false,
        secretConfigured: false,
        clearSecret: false,
      },
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
    birthdayBlessingAvailable() {
      return !!(this.$store.state.config.config || {}).birthdayRandomBlessing
    },
    hasBirthday() {
      return !!(this.birthdayTouched
        ? this.birthdayInput
        : this.originalBirthday)
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
        await this.$axios.post('/api/user/dingtalk/settings', {
          enabled: this.dingTalk.enabled,
          webhook: this.dingTalk.webhook,
          secret: this.dingTalk.secret,
          keyword: this.dingTalk.keyword,
          clearSecret: this.dingTalk.clearSecret,
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
    async testDingTalk() {
      this.testingDingTalk = true
      try {
        await this.$axios.post('/api/user/dingtalk/test')
        this.$message.success('测试通知已发送')
      } catch (e) {
        console.error(e)
        this.$message.error('测试通知发送失败：' + (e.message || e))
      } finally {
        this.testingDingTalk = false
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
      const [_user, dingTalk] = await Promise.all([
        this.$axios.get('/api/user/current'),
        this.$axios.get('/api/user/dingtalk/settings'),
      ])
      if (_user) {
        const pattern = _user.roles[0].split('_')
        const role = pattern[0]
        _user.position = _user.department + '-' + role
        this.$store.commit('user/setCurrent', _user)
        this.user = _user
        this.form = { ..._user }
        this.originalBirthday = _user.birthday || ''
        this.birthdayInput = this.isValidBirthday(this.originalBirthday)
          ? this.originalBirthday
          : ''
        this.birthdayTouched = false
      }
      if (dingTalk) {
        this.dingTalk = {
          enabled: dingTalk.enabled,
          webhook: '',
          secret: '',
          keyword: dingTalk.keyword || '',
          webhookConfigured: dingTalk.webhookConfigured,
          secretConfigured: dingTalk.secretConfigured,
          clearSecret: false,
        }
      }
    },
  },
}
</script>

<style lang="scss" scoped>
.ding-talk-settings {
  margin-top: 0.75rem;
}

.control + .birthday-blessing-control {
  margin-top: 0.75rem;
}

.secondary-setting-note {
  color: #999;
  font-size: 12px;
}

.ding-talk-control {
  margin-top: 0.75rem;
}

.notification-checkbox {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  cursor: pointer;
  user-select: none;
}

.notification-checkbox-input {
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
}

.notification-checkbox-box {
  position: relative;
  flex: 0 0 18px;
  width: 18px;
  height: 18px;
  border: 1px solid #b5b5b5;
  border-radius: 3px;
  background: #fff;
}

.notification-checkbox-input:checked + .notification-checkbox-box {
  border-color: #48c78e;
  background: #48c78e;
}

.notification-checkbox-input:checked + .notification-checkbox-box::after {
  position: absolute;
  top: 2px;
  left: 5px;
  width: 5px;
  height: 9px;
  border: solid #fff;
  border-width: 0 2px 2px 0;
  content: '';
  transform: rotate(45deg);
}

.notification-checkbox-input:focus-visible + .notification-checkbox-box {
  box-shadow: 0 0 0 2px rgba(72, 199, 142, 0.25);
}
</style>

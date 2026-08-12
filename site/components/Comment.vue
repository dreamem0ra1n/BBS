<template>
  <div class="comment-component">
    <div class="comment-header">
      <span v-if="commentCount > 0">{{ commentCount }}条评论</span>
      <span v-else>评论</span>
      <span class="comment-order" @click="changeOrder">{{
        ascOrder === 1 ? '正序' : '倒序'
      }}</span>
    </div>

    <template v-if="isLogin">
      <comment-input
        v-if="!noComment"
        ref="input"
        :mode="mode"
        :entity-id="entityId"
        :entity-type="entityType"
        @created="commentCreated"
      />
    </template>
    <div v-else class="comment-not-login">
      <div class="comment-login-div">
        请
        <a style="font-weight: 700" @click="toLogin">登录</a>后发表观点
      </div>
    </div>

    <comment-list
      ref="list"
      :entity-id="entityId"
      :entity-type="entityType"
      :comments-page="commentsPage"
      @reply="reply"
      @deleted="$emit('deleted')"
    />
  </div>
</template>

<script>
export default {
  props: {
    mode: {
      type: String,
      default: 'markdown',
    },
    entityType: {
      type: String,
      default: '',
      required: true,
    },
    entityId: {
      type: Number || String,
      default: 0,
      required: true,
    },
    commentsPage: {
      type: Object,
      default() {
        return {}
      },
    },
    commentCount: {
      type: Number,
      default: 0,
    },
    reGain: {
      type: Function,
      default: (order) => {},
    },
    noComment: {
      type: Boolean,
      default: false,
    },
  },
  computed: {
    isLogin() {
      return this.$store.state.user.current != null
    },
    user() {
      return this.$store.state.user.current
    },
    config() {
      return this.$store.state.config.config
    },
    ascOrder() {
      return this.$store.state.env.ascOrder
    },
  },
  methods: {
    commentCreated(data) {
      if (this.ascOrder === 1) this.$emit('reGain', 0)
      else this.$refs.list.append(data)
      this.$emit('created')
    },
    reply(quote) {
      this.$refs.input.reply(quote)
    },
    toLogin() {
      this.$toSignin()
    },
    changeOrder() {
      let order
      if (this.ascOrder === 0) order = 1
      else order = 0
      this.$emit('reGain', order)
    },
  },
}
</script>
<style lang="scss" scoped>
.comment-component {
  background-color: var(--bg-color);
  border-radius: 3px;
  .comment-header {
    padding-top: 20px;
    margin: 0 10px;
    color: var(--text-color);
    font-size: 16px;
    font-weight: 500;
  }

  .comment-not-login {
    margin: 10px;
    border: 1px solid var(--border-color);
    border-radius: 0;
    overflow: hidden;
    position: relative;
    padding: 10px;
    box-sizing: border-box;

    .comment-login-div {
      color: var(--text-color4);
      cursor: pointer;
      border-radius: 3px;
      padding: 0 10px;

      a {
        margin-left: 10px;
        margin-right: 10px;
      }
    }
  }
  .comment-order {
    float: right;
    cursor: pointer;
  }
}
</style>

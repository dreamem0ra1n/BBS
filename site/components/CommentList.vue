<template>
  <div class="comments">
    <div
      v-for="comment in commentResults"
      :id="commentAnchorId(comment.commentId)"
      :key="comment.commentId"
      class="comment"
    >
      <div class="comment-item-left">
        <avatar :user="comment.user" size="40" round has-border />
      </div>
      <div class="comment-item-main">
        <div class="comment-meta">
          <nuxt-link :to="'/user/' + comment.user.id" class="comment-nickname">
            {{ comment.user.nickname }}
          </nuxt-link>
          <time
            class="comment-time"
            :datetime="
              (comment.isOldBBS
                ? comment.createTime * 1000
                : comment.createTime) | formatDate('yyyy-MM-ddTHH:mm:ss')
            "
            >{{
              (comment.isOldBBS
                ? comment.createTime * 1000
                : comment.createTime) | prettyDate
            }}</time
          >
        </div>
        <div
          v-viewer
          v-lazy-container="{ selector: 'img' }"
          class="comment-content-wrapper"
        >
          <div
            v-if="comment.content"
            class="comment-content content"
            v-html="comment.content.replace(/\n/gm, '<br>')"
          ></div>
          <div
            v-if="comment.imageList && comment.imageList.length"
            class="comment-image-list"
          >
            <img
              v-for="(image, imageIndex) in comment.imageList"
              :key="imageIndex"
              :data-src="image.url"
            />
          </div>
        </div>
        <div
          v-if="comment.lastEditUser && comment.lastEditTime"
          class="comment-edit-record"
        >
          该帖由
          <nuxt-link :to="'/user/' + comment.lastEditUser.id">
            {{ comment.lastEditUser.nickname }}
          </nuxt-link>
          于 {{ comment.lastEditTime | formatDate }} 编辑
        </div>
        <div class="comment-actions">
          <div
            v-if="comment.status === 0"
            class="comment-action-item"
            :class="{ active: comment.liked }"
            @click="like(comment)"
          >
            <i class="iconfont icon-like"></i>
            <span>{{ comment.liked ? '已赞' : '点赞' }}</span>
            <span v-if="comment.likeCount > 0">{{ comment.likeCount }}</span>
          </div>
          <div
            v-if="comment.status === 0"
            class="comment-action-item"
            :class="{ active: reply.commentId === comment.commentId }"
            @click="switchShowReply(comment)"
          >
            <i class="iconfont icon-comment"></i>
            <span>{{
              reply.commentId === comment.commentId ? '取消评论' : '评论'
            }}</span>
          </div>
          <div
            v-if="comment.canEdit"
            class="comment-action-item"
            @click="editComment(comment)"
          >
            <i class="iconfont icon-edit"></i>
            <span>编辑</span>
          </div>
          <div
            v-if="comment.canDelete"
            class="comment-action-item"
            @click="deleteComment(comment)"
          >
            <i class="iconfont icon-delete"></i>
            <span>删除</span>
          </div>
        </div>
        <div
          v-if="reply.commentId === comment.commentId"
          class="comment-reply-form"
        >
          <text-editor
            :ref="`editor${comment.commentId}`"
            v-model="reply.value"
            :height="100"
            @submit="submitReply(comment)"
          />
        </div>
        <sub-comment-list
          v-if="
            comment.replies &&
            comment.replies.results &&
            comment.replies.results.length
          "
          :key="getNewKey()"
          :comment-id="comment.commentId"
          :data="comment.replies"
          @reply="onReply(comment, $event)"
          @deleted="replyDeleted(comment)"
        />
      </div>
    </div>
    <comment-edit-dialog
      :visible.sync="editVisible"
      :comment="editingComment"
      @updated="commentUpdated"
    />
  </div>
</template>

<script>
import HTMLDecode from '../utils/HTMLDecode'
import SubCommentList from './SubCommentList.vue'
export default {
  components: { SubCommentList },
  props: {
    entityType: {
      type: String,
      default: '',
      required: true,
    },
    entityId: {
      type: String || Number,
      default: 0,
      required: true,
    },
    commentsPage: {
      type: Object,
      default() {
        return {}
      },
    },
  },
  data() {
    return {
      commentResults: [],
      showReplyCommentId: 0,
      editVisible: false,
      editingComment: null,
      reply: {
        commentId: 0,
        value: {
          content: '',
          imageList: [],
        },
      },
      hashScrolled: false,
    }
  },
  computed: {
    user() {
      return this.$store.state.user.current
    },
    isLogin() {
      return this.$store.state.user.current != null
    },
    ascOrder() {
      return this.$store.state.env.ascOrder
    },
  },
  watch: {
    commentsPage: {
      immediate: true,
      handler(value) {
        this.commentResults = ((value && value.results) || []).slice()
      },
    },
  },
  mounted() {
    this.scrollToHashComment()
  },
  methods: {
    HTMLDecode,
    commentAnchorId(commentId) {
      return `comment-${commentId}`
    },
    async scrollToHashComment() {
      if (typeof window === 'undefined' || this.hashScrolled) {
        return
      }
      const match = window.location.hash.match(/^#comment-(\d+)$/)
      if (!match) {
        return
      }

      const targetId = this.commentAnchorId(match[1])
      await this.$nextTick()
      const target = document.getElementById(targetId)
      if (target) {
        this.hashScrolled = true
        target.scrollIntoView({ block: 'center' })
      }
    },
    append(data) {
      if (data) {
        this.commentResults.unshift(data)
      }
    },
    editComment(comment) {
      this.editingComment = comment
      this.editVisible = true
    },
    commentUpdated(updated) {
      Object.assign(this.editingComment, updated)
    },
    async deleteComment(comment) {
      try {
        await this.$confirm('是否确认删除该评论？')
        await this.$axios.post(`/api/comment/delete/${comment.commentId}`)
        if (comment.commentCount > 0) {
          comment.status = 1
          comment.content = '内容已删除'
          comment.imageList = []
          comment.canEdit = false
          comment.canDelete = false
        } else {
          const index = this.commentResults.indexOf(comment)
          if (index !== -1) this.commentResults.splice(index, 1)
        }
        this.$emit('deleted')
        this.$message.success('删除成功')
      } catch (e) {
        if (e !== 'cancel' && e !== 'close') {
          this.$message.error('删除失败：' + (e.message || e))
        }
      }
    },
    replyDeleted(parent) {
      parent.commentCount = Math.max(0, parent.commentCount - 1)
    },
    async like(comment) {
      try {
        await this.$axios.post(`/api/comment/like/${comment.commentId}`)
        comment.liked = true
        comment.likeCount = comment.likeCount + 1
        this.$message.success('点赞成功')
      } catch (e) {
        if (e.errorCode === 1) {
          this.$msgSignIn()
        } else {
          this.$message.error(e.message || e)
        }
      }
    },
    switchShowReply(comment) {
      if (!this.user) {
        this.$msgSignIn()
        return
      }

      if (this.reply.commentId === comment.commentId) {
        this.hideReply(comment)
      } else {
        this.reply.commentId = comment.commentId
        setTimeout(() => {
          this.$refs[`editor${comment.commentId}`][0].focus()
        }, 0)
      }
    },
    hideReply(comment) {
      this.reply.commentId = 0
      this.reply.value.content = ''
      this.reply.value.imageList = []
    },
    async submitReply(parent) {
      try {
        const ret = await this.$axios.post('/api/comment/create', {
          entityType: 'comment',
          entityId: parent.commentId,
          content: this.reply.value.content,
          imageList:
            this.reply.value.imageList && this.reply.value.imageList.length
              ? JSON.stringify(this.reply.value.imageList)
              : '',
        })
        this.hideReply()
        this.appendReply(parent, ret)
        this.$message.success('发布成功')
      } catch (e) {
        if (e.errorCode === 1) {
          this.$msgSignIn()
        } else {
          this.$message.error(e.message || e)
        }
      }
    },
    onReply(parent, comment) {
      this.appendReply(parent, comment)
    },
    appendReply(parent, comment) {
      if (parent.replies && parent.replies.results) {
        parent.replies.results.push(comment)
      } else {
        parent.replies = {
          results: [comment],
        }
      }
    },
    getNewKey() {
      return new Date().toISOString()
    },
  },
}
</script>

<style scoped lang="scss">
.comments {
  padding: 10px;
  font-size: 14px;

  .comment {
    display: flex;
    padding: 10px 0;

    &:not(:last-child) {
      border-bottom: 1px solid var(--border-color);
    }

    .comment-item-main {
      flex: 1 1 auto;
      margin-left: 16px;

      .comment-meta {
        display: flex;
        justify-content: space-between;
        .comment-nickname {
          font-size: 14px;
          font-weight: 600;
          color: var(--text-color);

          &:hover {
            color: var(--text-link-color);
          }
        }

        .comment-time {
          font-size: 13px;
          color: var(--text-color3);
        }
      }

      .comment-content-wrapper {
        .comment-content {
          margin-top: 10px;
          margin-bottom: 0;
          color: var(--text-color);
          max-width: 100%;
          overflow-wrap: anywhere;
          word-break: break-word;

          pre {
            max-width: 100%;
            overflow-x: auto;
            white-space: pre;
          }

          code {
            overflow-wrap: normal;
            word-break: normal;
          }
        }
        .comment-image-list {
          margin-top: 10px;

          img {
            width: 72px;
            height: 72px;
            line-height: 72px;
            cursor: pointer;
            object-fit: cover;
            transition: all 0.5s ease-out 0.1s;

            &:not(:last-child) {
              margin-right: 8px;
            }

            &:hover {
              transform: matrix(1.04, 0, 0, 1.04, 0, 0);
              backface-visibility: hidden;
            }
          }
        }
      }

      .comment-actions {
        margin-top: 10px;
        display: flex;
        align-items: center;

        .comment-action-item {
          line-height: 22px;
          font-size: 13px;
          cursor: pointer;
          color: var(--text-color3);
          user-select: none;

          &:hover {
            color: var(--text-link-color);
          }

          &.active {
            color: var(--text-link-color);
            font-weight: 500;
          }

          &:not(:last-child) {
            margin-right: 16px;
          }

          .iconfont {
            font-size: 13px;
          }
        }
      }

      .comment-edit-record {
        margin-top: 8px;
        color: var(--text-color3);
        font-size: 12px;

        a {
          color: var(--text-link-color);
        }
      }

      .comment-reply-form {
        margin-top: 10px;
      }

      .comment-replies {
        margin-top: 10px;
        // padding: 10px;
        background-color: var(--bg-color2);
      }
    }
  }

  .reply {
    display: flex;
  }
}
</style>

<template>
  <ul class="topic-list">
    <li v-for="topic in topics" :key="topic.topicId" class="topic-item">
      <div
        class="topic-avatar"
        :href="'/user/' + topic.user.id"
        :title="topic.user.nickname"
      >
        <avatar :user="topic.user" />
      </div>
      <div class="topic-main-content">
        <div class="topic-top">
          <div class="topic-userinfo">
            <avatar class="topic-inline-avatar" :user="topic.user" size="20" />
            <my-link :to="'/user/' + topic.user.id">{{
              topic.user.nickname
            }}</my-link>
            <span v-if="showSticky && topic.sticky" class="topic-sticky-icon"
              >置顶</span
            >
          </div>
          <div class="topic-time">
            发布于{{ topic.createTime | prettyDate }}
          </div>
        </div>
        <div class="topic-content" :class="{ 'topic-tweet': topic.type === 1 }">
          <template v-if="topic.type === 0">
            <h1 class="topic-title">
              <my-link :to="'/topic/' + topic.topicId">{{
                topic.title
              }}</my-link>
            </h1>
            <my-link :to="'/topic/' + topic.topicId" class="topic-summary">{{
              topic.summary
            }}</my-link>
          </template>
          <template v-if="topic.type === 1">
            <my-link
              v-if="topic.content"
              :to="'/topic/' + topic.topicId"
              class="topic-summary"
              >{{ topic.content }}</my-link
            >
            <ul
              v-if="topic.imageList && topic.imageList.length"
              class="topic-image-list"
            >
              <li v-for="(image, index) in topic.imageList" :key="index">
                <my-link :to="'/topic/' + topic.topicId" class="image-item">
                  <img v-lazy="image.preview" />
                </my-link>
              </li>
            </ul>
          </template>
        </div>
        <div class="topic-bottom">
          <div class="topic-handlers">
            <div
              class="btn"
              :class="{ liked: topic.liked }"
              @click="like(topic)"
            >
              <i class="iconfont icon-like" />{{ topic.liked ? '已赞' : '赞' }}
              <span v-if="topic.likeCount > 0">{{ topic.likeCount }}</span>
            </div>
            <div class="btn" @click="toTopicDetail(topic.topicId)">
              <i class="iconfont icon-comment" />评论
              <span v-if="topic.commentCount > 0">{{
                topic.commentCount
              }}</span>
            </div>
            <div class="btn" @click="toTopicDetail(topic.topicId)">
              <i class="iconfont icon-read" />浏览
              <span v-if="topic.viewCount > 0">{{ topic.viewCount }}</span>
            </div>
          </div>
          <div class="topic-tags">
            <my-link
              v-if="topic.node"
              class="topic-tag"
              :to="'/topics/node/' + topic.node.nodeId"
              :alt="topic.node.name"
              >{{ topic.node.name }}</my-link
            >
          </div>
        </div>
      </div>
    </li>
  </ul>
</template>

<script>
export default {
  props: {
    topics: {
      type: Array,
      default() {
        return []
      },
      required: false,
    },
    showAvatar: {
      type: Boolean,
      default: true,
    },
    showSticky: {
      type: Boolean,
      default: false,
    },
  },
  methods: {
    async like(topic) {
      try {
        await this.$axios.post('/api/topic/like/' + topic.topicId)
        topic.liked = true
        topic.likeCount++
        this.$message.success('点赞成功')
      } catch (e) {
        if (e.errorCode === 1) {
          this.$msgSignIn()
        } else {
          this.$message.error(e.message || e)
        }
      }
    },
    toTopicDetail(topicId) {
      this.$linkTo(`/topic/${topicId}`)
    },
  },
}
</script>

<style lang="scss" scoped></style>

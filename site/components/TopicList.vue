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
            <nuxt-link :to="'/user/' + topic.user.id">{{
              topic.user.nickname
            }}</nuxt-link>
            <span v-if="showSticky && topic.sticky" class="topic-sticky-icon"
              >置顶</span
            >
          </div>
          <div class="topic-time">
            发布于{{
              (isOld ? topic.createTime * 1000 : topic.createTime) | prettyDate
            }}
          </div>
        </div>
        <div class="topic-content" :class="{ 'topic-tweet': topic.type === 1 }">
          <template v-if="topic.type === 0">
            <h1 class="topic-title">
              <nuxt-link
                :to="'/topic/' + (isOld ? 'OLD' : '') + topic.topicId"
                target="_blank"
                rel="noopener noreferrer"
                >{{ topic.title }}</nuxt-link
              >
            </h1>
            <nuxt-link
              :to="'/topic/' + (isOld ? 'OLD' : '') + topic.topicId"
              class="topic-summary"
              target="_blank"
              rel="noopener noreferrer"
              >{{ topic.summary }}</nuxt-link
            >
          </template>
          <template v-if="topic.type === 1">
            <nuxt-link
              v-if="isOld && topic.content"
              :to="'/topic/' + (isOld ? 'OLD' : '') + topic.topicId"
              class="topic-summary"
              target="_blank"
              rel="noopener noreferrer"
              >{{ topic.content }}</nuxt-link
            >
            <div
              v-if="!isOld && topic.content"
              class="topic-summary"
              @click="onTweetContentClick($event, topic.topicId)"
              v-html="topic.content"
            ></div>
            <ul
              v-if="topic.imageList && topic.imageList.length"
              class="topic-image-list"
            >
              <li v-for="(image, index) in topic.imageList" :key="index">
                <nuxt-link
                  :to="'/topic/' + (isOld ? 'OLD' : '') + topic.topicId"
                  class="image-item"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <img v-lazy="image.preview" />
                </nuxt-link>
              </li>
            </ul>
          </template>
        </div>
        <div class="topic-bottom">
          <div class="topic-handlers" :v-if="!isOld">
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
            <nuxt-link
              v-if="topic.node"
              class="topic-tag"
              :to="'/topics/node/' + topic.node.nodeId"
              :alt="topic.node.name"
              >{{ topic.node.name }}</nuxt-link
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
    isOld: {
      type: Boolean,
      default: false,
    },
  },
  methods: {
    onTweetContentClick(event, topicId) {
      if (!event.target.closest('a')) {
        this.toTopicDetail(topicId)
      }
    },
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
      const path = '/topic/' + (this.isOld ? 'OLD' : '') + topicId
      const { href } = this.$router.resolve(path)
      window.open(href, '_blank', 'noopener,noreferrer')
    },
  },
}
</script>

<style lang="scss" scoped></style>

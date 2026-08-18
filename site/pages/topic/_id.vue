<template>
  <div>
    <section class="main">
      <div class="container main-container left-main size-360">
        <div class="left-container">
          <div class="main-content no-padding no-bg">
            <article
              class="topic-detail"
              itemscope
              itemtype="http://schema.org/BlogPosting"
            >
              <div class="topic-header">
                <div class="topic-header-left">
                  <avatar :user="topic.user" size="45" />
                </div>
                <div class="topic-header-center">
                  <div class="topic-nickname" itemprop="headline">
                    <nuxt-link
                      itemprop="author"
                      itemscope
                      itemtype="http://schema.org/Person"
                      :to="'/user/' + topic.user.id"
                      >{{ topic.user.nickname }}</nuxt-link
                    >
                  </div>
                  <div class="topic-meta">
                    <span v-if="!isOld" class="meta-item">
                      浏览 {{ topic.viewCount }}
                    </span>
                    <span class="meta-item">
                      发布于
                      <time
                        :datetime="
                          topic.isOldBBS
                            ? topic.createTime * 1000
                            : topic.createTime
                              | formatDate('yyyy-MM-ddTHH:mm:ss')
                        "
                        itemprop="datePublished"
                        >{{
                          topic.isOldBBS
                            ? topic.createTime * 1000
                            : topic.createTime | prettyDate
                        }}</time
                      >
                    </span>
                  </div>
                </div>
                <div class="topic-header-right">
                  <topic-manage-menu v-if="!isOld" v-model="topic" />
                </div>
              </div>

              <!--内容-->
              <div
                class="topic-content content"
                :class="{
                  'topic-tweet': topic.type === 1,
                }"
                itemprop="articleBody"
              >
                <h1 v-if="topic.title" class="topic-title" itemprop="headline">
                  {{ topic.title }}
                </h1>
                <div
                  ref="topicContent"
                  v-lazy-container="{ selector: 'img' }"
                  class="topic-content-detail"
                  v-html="topic.content"
                ></div>
                <ul
                  v-if="topic.imageList && topic.imageList.length"
                  v-viewer
                  class="topic-image-list"
                >
                  <li v-for="(image, index) in topic.imageList" :key="index">
                    <div class="image-item">
                      <img :src="image.preview" :data-src="image.url" />
                    </div>
                  </li>
                </ul>
                <div
                  v-if="hideContent && hideContent.exists"
                  class="topic-content-detail hide-content"
                >
                  <div v-if="hideContent.show" class="widget has-border">
                    <div class="widget-header">
                      <span>
                        <i class="iconfont icon-lock" />
                        <span>隐藏内容</span>
                      </span>
                    </div>
                    <div class="widget-content" v-html="hideContent.content" />
                  </div>
                  <div v-else class="hide-content-tip">
                    <i class="iconfont icon-lock" />
                    <span>隐藏内容，请回复后查看</span>
                  </div>
                </div>
                <div
                  v-if="topic.gifts && topic.gifts.length"
                  class="topic-gift-records"
                >
                  <div class="topic-gift-title">
                    <i class="iconfont icon-score" /> 赠米详情
                  </div>
                  <div
                    v-for="gift in topic.gifts"
                    :key="gift.giftId"
                    class="topic-gift-record"
                  >
                    <nuxt-link :to="'/user/' + gift.user.id">
                      {{ gift.user.nickname }}
                    </nuxt-link>
                    赠米 <strong>{{ gift.score }}</strong>
                    <span class="topic-gift-reason">“{{ gift.reason }}”</span>
                    <time>{{ gift.createTime | formatDate }}</time>
                  </div>
                </div>
                <div
                  v-if="topic.lastEditUser && topic.lastEditTime"
                  class="topic-edit-record"
                >
                  本帖由
                  <nuxt-link :to="'/user/' + topic.lastEditUser.id">
                    {{ topic.lastEditUser.nickname }}
                  </nuxt-link>
                  于
                  <time
                    :datetime="
                      topic.lastEditTime | formatDate('yyyy-MM-ddTHH:mm:ss')
                    "
                    >{{ topic.lastEditTime | formatDate }}</time
                  >
                  编辑
                </div>
              </div>

              <!--节点、标签-->
              <div class="topic-tags">
                <nuxt-link
                  v-if="topic.node"
                  :to="'/topics/node/' + topic.node.nodeId"
                  class="topic-tag"
                  >{{ topic.node.name }}</nuxt-link
                >
                <nuxt-link
                  v-for="tag in topic.tags"
                  :key="tag.tagId"
                  :to="'/topics/tag/' + tag.tagId"
                  class="topic-tag"
                  >#{{ tag.tagName }}</nuxt-link
                >
              </div>

              <!-- 点赞用户列表 -->
              <div
                v-if="likeUsers && likeUsers.length"
                class="topic-like-users"
              >
                <avatar
                  v-for="likeUser in likeUsers"
                  :key="likeUser.id"
                  :user="likeUser"
                  :round="true"
                  :has-border="true"
                  size="24"
                />
                <span class="like-count">{{ topic.likeCount }}</span>
              </div>

              <!-- 功能按钮 -->
              <div v-if="!isOld" class="topic-actions">
                <div
                  class="action"
                  :class="{ disabled: ownTopic }"
                  @click="openGiftDialog"
                >
                  <i class="action-icon iconfont icon-score" />
                  <div class="action-text">
                    <span>赠米</span>
                  </div>
                </div>
                <div
                  class="action"
                  :class="{ disabled: liked }"
                  @click="like(topic)"
                >
                  <i
                    class="action-icon iconfont icon-like"
                    :class="{ 'checked-icon': liked }"
                  />
                  <div class="action-text">
                    <span>点赞</span>
                    <span v-if="topic.likeCount > 0">
                      ({{ topic.likeCount }})
                    </span>
                  </div>
                </div>
                <div class="action" @click="addFavorite(topic.topicId)">
                  <i
                    class="action-icon iconfont"
                    :class="{
                      'icon-has-favorite': favorited,
                      'icon-favorite': !favorited,
                      'checked-icon': favorited,
                    }"
                  />
                  <div class="action-text">
                    <span>收藏</span>
                  </div>
                </div>
                <div class="action" @click="copyShareLink">
                  <i class="action-icon iconfont icon-share" />
                  <div class="action-text">
                    <span>复制分享链接</span>
                  </div>
                </div>
              </div>
            </article>

            <el-dialog
              title="给本帖赠米"
              :visible.sync="giftDialogVisible"
              width="min(440px, 90%)"
              custom-class="topic-gift-dialog"
              append-to-body
            >
              <div class="topic-gift-form">
                <label>赠米数量</label>
                <el-input-number
                  v-model="giftForm.score"
                  :min="1"
                  :max="giftScoreMax"
                  :step="1"
                  step-strictly
                />
                <div class="topic-gift-tip">
                  单次可赠 1-{{ giftScoreMax }} 米，当前可用
                  {{ giftBalance }} 米
                </div>
                <label>赠米理由</label>
                <el-input
                  v-model.trim="giftForm.reason"
                  maxlength="15"
                  show-word-limit
                  placeholder="请输入赠米理由"
                />
              </div>
              <span slot="footer">
                <el-button @click="giftDialogVisible = false">取消</el-button>
                <el-button
                  type="primary"
                  :loading="gifting"
                  @click="submitGift"
                >
                  确认赠米
                </el-button>
              </span>
            </el-dialog>

            <!-- 评论 -->
            <comment
              :entity-id="entityId"
              :comments-page="commentsPage"
              :comment-count="topic.commentCount"
              :mode="topic.type === 1 ? 'text' : 'markdown'"
              :no-comment="isOld"
              entity-type="topic"
              @created="commentCreated"
              @deleted="commentDeleted"
              @reGain="reGain"
            />
          </div>
        </div>
        <div class="right-container">
          <user-info :user="topic.user" />
        </div>
      </div>
    </section>
  </div>
</template>

<script>
import { Loading } from 'element-ui'
import XBBCODE from '~/utils/xbbcode'
import CommonHelper from '~/common/CommonHelper'
export default {
  async asyncData({ $axios, params, error, store }) {
    let topic
    let liked = null
    let favorited = null
    let likeUsers
    try {
      topic = await $axios.get('/api/topic/' + params.id)
    } catch (e) {
      error({
        statusCode: 404,
        message: '话题不存在',
      })
      return
    }
    const commentsPage = await $axios.get('/api/comment/comments', {
      params: {
        entityType: 'topic',
        entityId: params.id,
        asc_order: store.state.env.ascOrder,
      },
    })
    if (!topic.isOldBBS) {
      ;[liked, favorited, likeUsers] = await Promise.all([
        $axios.get('/api/like/liked', {
          params: {
            entityType: 'topic',
            entityId: params.id,
          },
        }),
        $axios.get('/api/favorite/favorited', {
          params: {
            entityType: 'topic',
            entityId: params.id,
          },
        }),

        $axios.get('/api/topic/recentlikes/' + params.id),
      ])
    }
    if (topic.isOldBBS) {
      topic.content = XBBCODE.process({
        text: topic.content,
      }).html
      if (commentsPage?.results) {
        commentsPage.results.forEach((comment) => {
          comment.content = XBBCODE.process({
            text: comment.content,
          }).html
        })
      }
    }
    return {
      topic,
      commentsPage,
      favorited: favorited?.favorited,
      liked: liked?.liked,
      likeUsers,
      entityId: params.id,
    }
  },
  data() {
    return {
      hideContent: null,
      giftDialogVisible: false,
      gifting: false,
      giftBalance: null,
      giftForm: {
        score: 1,
        reason: '',
      },
    }
  },
  head() {
    return {
      title: this.$topicSiteTitle(this.topic),
      link: [
        {
          rel: 'stylesheet',
          href: CommonHelper.highlightCss,
        },
      ],
      script: [
        {
          type: 'text/javascript',
          src: CommonHelper.highlightScript,
          callback: () => {
            // 客户端渲染的时候执行这里进行代码高亮
            CommonHelper.initHighlight()
          },
        },
      ],
    }
  },
  computed: {
    isOld() {
      return this.topic.isOldBBS
    },
    user() {
      return this.$store.state.user.current
    },
    ascOrder() {
      return this.$store.state.env.ascOrder
    },
    ownTopic() {
      return this.user && this.user.id === this.topic.user.id
    },
    giftScoreMax() {
      return this.$store.state.config.config.scoreConfig?.giftScoreMax || 50
    },
  },
  mounted() {
    // 加载隐藏内容
    this.getHideContent()
    this.$store.commit('env/setCurrentTag', -1919810)
    this.$store.commit('env/setAscOrder', 0)
    // 为了解决服务端渲染时，没有刷新meta中的script，callback没执行，导致代码高亮失败的问题
    // 所以服务端渲染时会调用这里的方法进行代码高亮
    CommonHelper.initHighlight(this)
  },
  methods: {
    commentCreated() {
      this.getHideContent()
    },
    commentDeleted() {
      this.topic.commentCount = Math.max(0, this.topic.commentCount - 1)
    },
    async addFavorite(topicId) {
      if (this.topic.isOldBBS) {
        return
      }
      try {
        if (this.favorited) {
          await this.$axios.get('/api/favorite/delete', {
            params: {
              entityType: 'topic',
              entityId: topicId,
            },
          })
          this.favorited = false
          this.$message.success('已取消收藏')
        } else {
          await this.$axios.get('/api/topic/favorite/' + topicId)
          this.favorited = true
          this.$message.success('收藏成功')
        }
      } catch (e) {
        console.error(e)
        this.$message.error('收藏失败：' + (e.message || e))
      }
    },
    openGiftDialog() {
      if (!this.user) {
        this.$msgSignIn()
        return
      }
      if (this.ownTopic) {
        this.$message.warning('不能给自己创建的话题赠米')
        return
      }
      if (this.giftBalance === null) {
        this.giftBalance = this.user.score
      }
      this.giftForm = {
        score: Math.min(Math.max(1, this.giftForm.score), this.giftScoreMax),
        reason: '',
      }
      this.giftDialogVisible = true
    },
    async submitGift() {
      const reason = this.giftForm.reason.trim()
      if (!reason) {
        this.$message.warning('请输入赠米理由')
        return
      }
      if ([...reason].length > 15) {
        this.$message.warning('赠米理由不能超过15个字')
        return
      }
      if (this.gifting) {
        return
      }
      this.gifting = true
      try {
        const gift = await this.$axios.post(
          `/api/topic/gift/${this.topic.topicId}`,
          {
            score: this.giftForm.score,
            reason,
          }
        )
        this.topic.gifts = this.topic.gifts || []
        this.topic.gifts.push(gift)
        this.giftBalance -= gift.score
        this.giftDialogVisible = false
        this.$message.success('赠米成功')
      } catch (e) {
        if (e.errorCode === 1) {
          this.$msgSignIn()
        } else {
          this.$message.error(e.message || e)
        }
      } finally {
        this.gifting = false
      }
    },
    async like(topic) {
      if (this.topic.isOldBBS) {
        return
      }
      try {
        if (this.liked) {
          return
        }
        await this.$axios.post('/api/topic/like/' + topic.topicId)
        this.liked = true
        topic.likeCount++
        this.likeUsers = this.likeUsers || []
        this.likeUsers.unshift(this.$store.state.user.current)
      } catch (e) {
        if (e.errorCode === 1) {
          this.$msgSignIn()
        } else {
          this.liked = true
          this.$message.error(e.message || e)
        }
      }
    },
    async copyShareLink() {
      const nodeName = this.topic.node?.name || '无节点'
      const topicTitle = this.topic.title || '无标题'
      const topicUrl = `https://www.qsc.zju.edu.cn/bbs2/topic/${this.topic.topicId}`
      const shareText = `【${nodeName}】${topicTitle} ${topicUrl} 复制本链接，打开【QSCBBS】网页端，直接查看本帖！`

      try {
        if (navigator.clipboard && window.isSecureContext) {
          await navigator.clipboard.writeText(shareText)
        } else {
          const textarea = document.createElement('textarea')
          textarea.value = shareText
          textarea.setAttribute('readonly', '')
          textarea.style.position = 'fixed'
          textarea.style.opacity = '0'
          document.body.appendChild(textarea)
          textarea.select()
          const copied = document.execCommand('copy')
          document.body.removeChild(textarea)
          if (!copied) {
            throw new Error('浏览器不支持复制到剪贴板')
          }
        }
        this.$message.success('分享链接已复制')
      } catch (e) {
        console.error(e)
        this.$message.error('复制分享链接失败，请稍后重试')
      }
    },
    async getHideContent() {
      try {
        this.hideContent = await this.$axios.get('/api/topic/hide_content', {
          params: {
            topicId: this.topic.topicId,
          },
        })
      } catch (e) {
        console.log(e)
      }
    },
    reGain(order) {
      const me = this
      const load = Loading.service({
        target: '.comment-component',
        background: 'var(--bg-color) ',
      })
      this.$store.commit('env/setAscOrder', order)
      this.$axios
        .get('/api/comment/comments', {
          params: {
            entityType: 'topic',
            entityId: this.entityId,
            asc_order: order,
          },
        })
        .then((res) => {
          me.commentsPage = res
          if (me.topic.isOldBBS) {
            me?.commentsPage?.results.forEach((comment) => {
              comment.content = XBBCODE.process({
                text: comment.content,
              }).html
            })
          }
          load.close()
        })
        .catch((e) => {
          console.log(e)
          load.close()
        })
    },
  },
}
</script>

<style lang="scss" scoped>
.el-loading-mask {
  background-color: var(--bg-color) !important;
}
</style>

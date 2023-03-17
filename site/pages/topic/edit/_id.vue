<template>
  <section class="main">
    <div class="container">
      <div class="topic-create-form">
        <h1 class="title">修改帖子</h1>

        <div class="field">
          <div class="control">
            <div
              v-for="node in nodes"
              :key="node.nodeId"
              class="topic-tag"
              :class="{ selected: postForm.nodeId === node.nodeId }"
              @click="postForm.nodeId = node.nodeId"
            >
              <span>{{ node.name }}</span>
            </div>
          </div>
        </div>

        <div class="field">
          <div class="control">
            <input
              v-model="postForm.title"
              class="input topic-title"
              type="text"
              placeholder="请输入帖子标题"
            />
          </div>
        </div>

        <div class="field">
          <div class="control">
            <markdown-editor
              v-model="postForm.content"
              placeholder="可空，将图片复制或拖入编辑器可上传"
            />
          </div>
        </div>

        <div v-if="isEnableHideContent || topic.hideContent" class="field">
          <div class="control">
            <markdown-editor
              ref="mdEditor"
              v-model="postForm.hideContent"
              height="200px"
              placeholder="隐藏内容，评论后可见"
            />
          </div>
        </div>

        <div class="field">
          <div class="control">
            <tag-input
              :nodeId="postForm.nodeId"
              @setTag="setTag"
              :tags="postForm.tags"
            />
          </div>
        </div>
        <div class="field">
          <div class="control">
            <el-select v-model="postForm.access_lv" placeholder="请选择可见性">
              <el-option
                v-for="lv in levelArray"
                :key="lv.level"
                :value="lv.level"
                :label="lv.description"
              ></el-option>
            </el-select>
          </div>
        </div>
      </div>
      <div class="field is-grouped">
        <div class="control">
          <a
            :class="{ 'is-loading': publishing }"
            :disabled="publishing"
            class="button is-success"
            @click="submitCreate"
            >提交更改</a
          >
        </div>
      </div>
    </div>
  </section>
</template>

<script>
export default {
  middleware: 'authenticated',
  async asyncData({ $axios, params }) {
    const [topic, nodes] = await Promise.all([
      $axios.get('/api/topic/edit/' + params.id),
      $axios.get('/api/topic/nodes'),
    ])
    return {
      topic,
      nodes,
      postForm: {
        nodeId: topic.nodeId,
        title: topic.title,
        tags: topic.tags,
        content: topic.content,
        hideContent: topic.hideContent,
        access_lv: topic.access_lv,
      },
    }
  },
  data() {
    return {
      publishing: false, // 当前是否正处于发布中...
      levelArray: [
        {
          level: 1,
          description: '全员可见',
        },
        {
          level: 2,
          description: '正式成员及以上可见',
        },
        {
          level: 3,
          description: '顾问及以上可见',
        },
        {
          level: 4,
          description: '管理层可见',
        },
      ],
    }
  },
  head() {
    return {
      title: this.$siteTitle('修改话题'),
    }
  },
  computed: {
    currentUser() {
      return this.$store.state.user.current
    },
    isEnableHideContent() {
      return this.$store.state.config.config.enableHideContent
    },
  },
  methods: {
    async submitCreate() {
      const me = this
      if (me.publishing) {
        return
      }
      me.publishing = true
      if (this.postForm.content.length > 5000) {
        this.$message.error('字数超过5000上限')
        return
      }
      try {
        const topic = await this.$axios.post(
          '/api/topic/edit/' + this.topic.topicId,
          {
            nodeId: this.postForm.nodeId,
            title: this.postForm.title,
            content: this.postForm.content,
            hideContent: this.postForm.hideContent,
            tags: this.postForm.tags ? this.postForm.tags.join(',') : '',
            access_lv: this.postForm.access_lv,
          }
        )
        this.$msg({
          message: '修改成功',
          onClose() {
            me.$linkTo('/topic/' + topic.topicId)
          },
        })
      } catch (e) {
        console.error(e)
        me.publishing = false
        this.$message.error('提交失败：' + (e.message || e))
      }
    },
    setTag(tags) {
      this.postForm.tags = tags
    },
  },
}
</script>

<style lang="scss" scoped></style>

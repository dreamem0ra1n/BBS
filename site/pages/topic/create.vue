<template>
  <section class="main">
    <div class="container">
      <div class="topic-create-form">
        <h1 class="title">发帖子</h1>

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
              ref="mdEditor"
              v-model="postForm.content"
              placeholder="请输入你要发表的内容..."
            />
          </div>
        </div>

        <div v-if="isEnableHideContent" class="field">
          <div class="control">
            <markdown-editor
              ref="mdEditor"
              v-model="postForm.hideContent"
              height="200px"
              placeholder="隐藏内容，评论后可见"
            />
          </div>
        </div>

        <!--<div v-if="postForm.type === 0" class="field">
          <div class="control">
            <simple-editor
              ref="simpleEditor"
              @input="onSimpleEditorInput"
              @submit="submitCreate"
            />
          </div>
        </div>-->

        <div class="field">
          <div class="control">
            <tag-input :nodeId="postForm.nodeId" @setTag="setTag" />
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

        <div class="field is-grouped">
          <div class="control">
            <a
              :class="{ 'is-loading': publishing }"
              :disabled="publishing"
              class="button is-success"
              @click="submitCreate"
              >{{ '发表帖子' }}</a
            >
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
export default {
  middleware: 'authenticated',
  async asyncData({ $axios, query, store }) {
    // 节点
    const nodes = await $axios.get('/api/topic/nodes')

    // 发帖标签
    const config = store.state.config.config || {}
    const nodeId =
      (store.state.currentNodeId !== 0 ? store.state.currentNodeId : null) ??
      config.defaultNodeId
    let currentNode = null
    if (nodeId) {
      try {
        currentNode = await $axios.get('/api/topic/node?nodeId=' + nodeId)
      } catch (e) {
        console.error(e)
      }
    }

    const type = parseInt(query.type || 0) || 0

    return {
      nodes,
      postForm: {
        type,
        nodeId,
        title: '',
        tags: [],
        content: '',
        hideContent: '',
        imageList: [],
        access_lv: '',
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
      title: this.$siteTitle(this.postForm.type === 1 ? '发动态' : '发帖子'),
    }
  },
  computed: {
    user() {
      return this.$store.state.user.current
    },
    config() {
      return this.$store.state.config.config
    },
    isEnableHideContent() {
      return this.config.enableHideContent
    },
  },
  watchQuery: ['type', 'nodeId'],
  mounted() {
    this.postForm.nodeId = this.$store.state.env.currentNodeId
  },
  methods: {
    setTag(tags) {
      this.postForm.tags = tags
    },
    async submitCreate() {
      if (this.publishing) {
        return
      }
      if (this.uploading) {
        this.$message.error('正在上传中...请上传完成后提交')
        return
      }
      if (!this.postForm.nodeId && this.postForm.nodeId !== 0) {
        this.$message.error('请选择节点')
        return
      }
      if (!this.postForm.content) {
        this.$message.error('请输入帖子内容')
        return
      }
      if (!this.postForm.access_lv || this.postForm.access_lv === '') {
        this.$message.error('请选择可见性')
        return
      }
      if (!this.postForm.tags || this.postForm.tags.length === 0) {
        this.$message.error('请选择标签')
        return
      }
      if (this.postForm.content.length > 5000) {
        this.$message.error('字数超过5000上限')
        return
      }
      this.publishing = true

      if (this.$refs.simpleEditor && this.$refs.simpleEditor.isOnUpload()) {
        this.$message.warning('正在上传中...请上传完成后提交')
        return
      }
      const me = this
      try {
        const topic = await this.$axios.post('/api/topic/create', {
          type: this.postForm.type,
          nodeId: this.postForm.nodeId,
          title: this.postForm.title,
          content: this.postForm.content,
          hideContent: this.postForm.hideContent,
          access_lv: this.postForm.access_lv,
          imageList:
            this.postForm.imageList && this.postForm.imageList.length
              ? JSON.stringify(this.postForm.imageList)
              : '',
          tags: this.postForm.tags ? this.postForm.tags.join(',') : '',
        })
        if (this.$refs.mdEditor) {
          this.$refs.mdEditor.clearCache()
        }
        this.$msg({
          message: '提交成功',
          onClose() {
            me.$linkTo('/topic/' + topic.topicId)
          },
        })
      } catch (e) {
        this.publishing = false
        this.$message.error(e.message || e)
      }
    },

    onSimpleEditorInput(value) {
      this.postForm.content = value.content
      this.postForm.imageList = value.imageList
    },
  },
}
</script>

<style>
input {
  display: none;
}
.topic-title {
  border-color: var(--border-color2);
}
.el-select .el-input__inner,
.el-popper,
.el-zoom-in-top-leave-active,
.el-zoom-in-top-leave-to {
  background-color: var(--bg-color);
  border-color: var(--border-color2);
}
.el-select-dropdown {
  background-color: var(--bg-color);
  border-color: var(--border-color2);
}
.el-select-dropdown__item.hover,
.el-select-dropdown__item:hover {
  background-color: var(--bg-color3);
}
.el-select-dropdown__item {
  background-color: var(--bg-color);
  border-color: var(--border-color2);
}
</style>

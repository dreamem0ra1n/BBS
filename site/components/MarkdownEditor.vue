<template>
  <div class="bbsgoEditor">
    <v-md-editor
      ref="editor"
      v-model="content"
      :left-toolbar="toolbars"
      :toolbar="myToolbar"
      :right-toolbar="rightToolbar"
      :height="height"
      :placeholder="placeholder"
      :disabled-menus="[]"
      mode="edit"
      @input="handleInput"
      @upload-image="uploadImage"
      @keydown.ctrl.enter.native="submit"
      @keydown.meta.enter.native="submit"
    ></v-md-editor>
    <div class="editor-counter">{{ contentLength }} / {{ maxLength }}</div>
    <input ref="upload" type="file" @input="uploadFile" />
  </div>
</template>

<script>
import Vue from 'vue'
import VMdEditor from '@kangc/v-md-editor'
import '@kangc/v-md-editor/lib/style/base-editor.css'
import githubTheme from '@kangc/v-md-editor/lib/theme/github.js'
import '@kangc/v-md-editor/lib/theme/style/github.css'
// highlightjs
import hljs from 'highlight.js'
import {
  CONTENT_MAX_LENGTH,
  contentLength,
  limitContent,
} from '../utils/content'

VMdEditor.use(githubTheme, {
  Hljs: hljs,
})
let _this
Vue.use(VMdEditor)

export default {
  props: {
    value: {
      type: String,
      default: '',
    },
    height: {
      type: String,
      default: '400px', // normal、mini
    },
    placeholder: {
      type: String,
      default: '请输入...',
    },
    maxLength: {
      type: Number,
      default: CONTENT_MAX_LENGTH,
    },
  },
  data() {
    return {
      width: '100%',
      content: this.value,
      editor: {},
      myToolbar: {
        customToolBar1: {
          title: '上传文件',
          icon: 'v-md-icon-toc',
          action(editor) {
            _this.editor = editor
            _this.$refs.upload.dispatchEvent(new MouseEvent('click'))
          },
        },
        myLink: {
          title: '添加链接',
          icon: 'v-md-icon-link',
          action(editor) {
            editor.insert(function (selected) {
              const prefix = '['
              const suffix = '](链接地址)'
              const placeholder = '链接名称'
              const content = selected || placeholder
              return {
                text: `${prefix}${content}${suffix}`,
                selected: content,
              }
            })
          },
        },
      },
    }
  },
  computed: {
    contentLength() {
      return contentLength(this.content)
    },
    isMobile() {
      return this.$store.state.env.isMobile
    },
    toolbars() {
      if (this.isMobile) {
        return 'h bold italic strikethrough image customToolBar1'
      } else {
        return 'undo redo clear | h bold italic strikethrough quote | ul ol table hr | myLink image code | customToolBar1'
      }
    },
    rightToolbar() {
      if (this.$store.state.env.isMobile) {
        return 'fullscreen'
      }
      return 'preview sync-scroll fullscreen'
    },
  },
  watch: {
    value(value) {
      this.content = limitContent(value, this.maxLength)
    },
  },
  mounted() {
    _this = this
    const limited = limitContent(this.value, this.maxLength)
    this.content = limited
    if (limited !== this.value) {
      this.$emit('input', limited)
    }
    this.setTextareaMaxLength()
  },
  updated() {
    this.setTextareaMaxLength()
  },
  methods: {
    setTextareaMaxLength() {
      const textarea = this.$el.querySelector('textarea')
      if (textarea) {
        const nativeMaxLength = this.maxLength * 2
        textarea.maxLength = nativeMaxLength
        textarea.setAttribute('maxlength', nativeMaxLength)
      }
    },
    handleInput(value) {
      const limited = limitContent(value, this.maxLength)
      if (limited !== value) {
        this.content = limited
      }
      this.$emit('input', limited)
    },
    uploadFile(e) {
      const files = e.target.files
      if (files.length <= 0) {
        return
      }
      const file = files[0]
      const fileName = file.name
      const formData = new FormData()
      formData.append('file', file)
      const that = this
      this.$axios
        .post('/api/file/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        })
        .then((ret) => {
          that.editor.insert(() => {
            return {
              text:
                '<a href="' +
                that.$store.state.env.currentURL +
                '/bbs2' +
                '/api/file/download/' +
                ret.file_id +
                '" download="' +
                fileName +
                '">点击下载附件</a>',
            }
          })
        })
        .catch((err) => {
          alert(err.message)
        })
    },
    submit() {
      this.$emit('submit', this.content)
    },
    /**
     * 上传图片
     */
    uploadImage(event, insertImage, files) {
      if (!files || !files.length) {
        return
      }
      for (let i = 0; i < files.length; i++) {
        const file = files[i]
        const formData = new FormData()

        formData.append('image', file, file.name)

        this.$axios
          .post('/api/file/upload/img', formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
          })
          .then((ret) => {
            insertImage({
              url: ret.url,
              desc: ' ',
            })
          })
          .catch((err) => {
            alert(err.message)
          })
      }
    },
    /**
     * 清空编辑器内容
     */
    clear() {
      this.content = ''
      this.$emit('input', this.content)
    },
    /**
     * 清理缓存
     */
    clearCache() {},
  },
}
</script>

<style lang="scss">
input {
  display: none;
}
.bbsgoEditor {
  .editor-counter {
    padding: 3px 8px;
    color: var(--text-color3);
    font-size: 12px;
    text-align: right;
    background-color: var(--markdown-background);
  }

  .v-md-editor {
    box-shadow: none !important;
    border: 1px solid var(--border-color2);
    background-color: var(--markdown-background);
    textarea {
      background-color: var(--markdown-background);
      color: var(--markdown-text);
    }
    .v-md-editor__toolbar {
      background-color: var(--markdown-background);
      border: 1px solid var(--border-color2);
      padding: 3px;

      .v-md-editor__toolbar-item {
        font-size: 14px !important;
        color: var(--markdown-text);
        background-color: var(--markdown-background);
      }
      .v-md-editor__toolbar-divider:before {
        border-left-color: var(--border-color2);
      }
    }

    .v-md-editor__editor-wrapper {
      border: 1px solid var(--border-color2);
      background-color: var(--markdown-background) !important;
    }

    .v-md-editor__preview-wrapper {
      border: 1px solid var(--border-color2);
      background-color: var(--markdown-background) !important;
    }
  }
}
</style>

<template>
  <div v-viewer="viewerOptions" class="image-uploads">
    <div
      v-for="image in previewFiles"
      :key="image.url"
      class="preview-item"
      :class="{ deleted: image.deleted }"
      :style="{ width: size, height: size, margin: previewItemMargin }"
    >
      <img :src="image.url" class="image-item" />
      <el-progress
        v-show="image.progress < 100"
        :percentage="image.progress"
        color="#25A9F6"
        :show-text="false"
        class="progress"
      />
      <div v-show="image.progress < 100" class="cover">上传中...</div>
      <div
        :class="{
          'upload-delete': true,
          'show-delete': image.progress === 100,
        }"
        @click="removeItem(image)"
      >
        <i class="iconfont icon-delete" />
      </div>
    </div>
    <div
      v-show="previewFiles.length < limit"
      class="add-image-btn"
      :style="{ width: size, height: size }"
      @click="onClick($event)"
    >
      <input
        ref="uploadImage"
        :accept="accept"
        type="file"
        multiple
        @click.stop
        @change="onInput"
      />
      <slot name="add-image-button">
        <i class="iconfont icon-add" />
      </slot>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    value: {
      type: Array,
      default() {
        return []
      },
    },
    accept: {
      type: String,
      default: 'image/*',
    },
    limit: {
      type: Number,
      default: 9,
    },
    sizeLimit: {
      type: Number,
      default: 1024 * 1024 * 20,
    },
    onUpload: {
      type: Boolean,
      default: false,
    },
    size: {
      type: String,
      default: '94px',
    },
  },
  data() {
    return {
      fileList: [],
      previewFiles: [],
      viewerOptions: {
        zIndex: 9999,
        navbar: false,
        title: false,
        tooltip: false,
        movable: false,
        scalable: false,
        url: 'src',
        filter(image) {
          return (
            image.classList &&
            image.classList.length &&
            image.classList.contains('image-item')
          )
        },
      },
    }
  },
  computed: {
    uploading() {
      return this.previewFiles.some((image) => image.progress < 100)
    },
    previewItemMargin() {
      if (this.previewFiles.length > 1 || this.limit > 1) {
        // margin-right: 10px;
        // margin-bottom: 10px;
        return '0 10px 10px 0'
      }
      return '0'
    },
  },
  watch: {
    uploading(value) {
      this.$emit('update:onUpload', value)
    },
    value: {
      handler(value) {
        if (value !== this.fileList) {
          this.releasePreviews()
          this.fileList = value ? [...value] : []
          this.previewFiles = this.fileList.map((image) => ({
            ...image,
            progress: 100,
            deleted: false,
            result: image,
            objectUrl: '',
          }))
        }
      },
      deep: true,
      immediate: true,
    },
  },
  beforeDestroy() {
    this.clear()
  },
  methods: {
    onClick() {
      this.$refs.uploadImage.click()
    },
    onInput(e) {
      const files = Array.from(e.target.files || [])
      e.target.value = ''
      this.addFiles(files)
    },
    addFiles(files) {
      if (!files || !files.length) return // 没有文件
      if (!this.checkSizeLimit(files)) return // 文件大小检查
      if (!this.checkLengthLimit(files)) return // 文件数量检查

      for (const file of files) {
        const url = this.getObjectURL(file)
        const image = {
          name: file.name,
          url,
          progress: 0,
          deleted: false,
          size: file.size,
          result: null,
          objectUrl: url,
        }
        this.previewFiles.push(image)
        this.uploadFile(file, image)
      }
    },
    async uploadFile(file, image) {
      const config = {
        onUploadProgress: (progressEvent) => {
          if (!this.previewFiles.includes(image) || image.progress === 100) {
            return
          }
          if (progressEvent.lengthComputable && progressEvent.total > 0) {
            image.progress = Math.min(
              90,
              Math.round((progressEvent.loaded / progressEvent.total) * 90)
            )
          }
        },
      }
      try {
        const formData = new FormData()
        formData.append('image', file, file.name)
        const result = await this.$axios.post(
          '/api/file/upload/img',
          formData,
          config
        )
        if (!this.previewFiles.includes(image)) {
          return
        }
        if (!result || !result.url) {
          throw new Error('服务器未返回图片地址，请重试')
        }
        image.result = result
        image.progress = 100
        this.emitFiles()
      } catch (error) {
        const index = this.previewFiles.indexOf(image)
        if (index < 0) {
          return
        }
        this.previewFiles.splice(index, 1)
        this.releasePreview(image)
        this.$message.error(
          `${file.name} 上传失败：${this.uploadErrorMessage(error)}`
        )
        this.$emit('error', error)
      }
    },
    uploadErrorMessage(error) {
      const response = error && error.response
      const data = response && response.data
      if (response && response.status === 413) {
        return '图片超过服务器允许的大小，请压缩后重试'
      }
      const message =
        (data && (data.message || data.msg)) ||
        (error && (error.message || error.msg))
      if (
        typeof message === 'string' &&
        message.includes(
          'Storage backend has reached its minimum free drive threshold'
        )
      ) {
        return '服务器存储空间不足，暂时无法上传，请联系管理员清理或扩容'
      }
      if (error && error.message === 'Network Error') {
        return '网络连接失败，请检查网络后重试'
      }
      if (error && error.code === 'ECONNABORTED') {
        return '上传超时，请稍后重试'
      }
      return message || '请稍后重试'
    },
    emitFiles() {
      this.fileList = this.previewFiles
        .filter((image) => image.result && !image.deleted)
        .map((image) => image.result)
      this.$emit('input', this.fileList)
    },
    removeItem(image) {
      if (image.progress < 100 || image.deleted) {
        return
      }
      this.$confirm('确定删除此内容吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(
        () => {
          if (!this.previewFiles.includes(image)) {
            return
          }
          image.deleted = true
          this.emitFiles()
          setTimeout(() => {
            const index = this.previewFiles.indexOf(image)
            if (index >= 0) {
              this.previewFiles.splice(index, 1)
              this.releasePreview(image)
              this.$message.success('删除成功')
            }
          }, 900)
        },
        () => console.log('取消删除')
      )
    },
    checkSizeLimit(files) {
      let pass = true
      for (let i = 0; i < files.length; i++) {
        if (files[i].size > this.sizeLimit) {
          pass = false
        }
      }
      if (!pass)
        this.$message.error(
          `图片大小不可超过 ${this.sizeLimit / 1024 / 1024} MB`
        )
      return pass
    },
    checkLengthLimit(files) {
      if (this.previewFiles.length + files.length > this.limit) {
        this.$message.warning(`图片最多上传${this.limit}张`)
        this.$emit('exceed', files)
        return false
      } else {
        return true
      }
    },
    getObjectURL(file) {
      let url = null
      if (window.createObjectURL) {
        // basic
        url = window.createObjectURL(file)
      } else if (window.URL) {
        // mozilla(firefox)
        url = window.URL.createObjectURL(file)
      } else if (window.webkitURL) {
        // webkit or chrome
        url = window.webkitURL.createObjectURL(file)
      }
      return url
    },
    clear() {
      this.releasePreviews()
      this.fileList = []
      this.previewFiles = []
    },
    releasePreview(image) {
      if (image.objectUrl) {
        URL.revokeObjectURL(image.objectUrl)
        image.objectUrl = ''
      }
    },
    releasePreviews() {
      this.previewFiles.forEach(this.releasePreview)
    },
  },
}
</script>

<style lang="scss" scoped>
.image-uploads {
  display: flex;
  flex-wrap: wrap;

  .preview-item {
    position: relative;
    border: 2px dashed var(--border-color);

    &.deleted {
      transition: 1s all;
      transform: translateY(-100%);
      opacity: 0;
    }

    .image-item {
      cursor: pointer;
      // border-radius: 5px;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .progress {
      position: absolute;
      top: 80px;
      width: 100%;
      height: 6px;
      padding: 0 10px;
    }

    .cover {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      color: var(--text-color2);
      background: rgba(255, 255, 255, 0.5);
      font-size: 12px;
      display: flex;
      justify-content: center;
      align-items: center;
    }

    .upload-delete {
      cursor: pointer;
      position: absolute;
      left: 0;
      bottom: 0;
      height: 20px;
      width: 100%;
      display: none;
      justify-content: center;
      align-items: center;
      background: rgba(0, 0, 0, 0.3);
      text-align: center;
      vertical-align: middle;
      line-height: 20px;

      i.iconfont {
        font-size: 14px;
        fill: white;
        color: var(--text-color5);
        font-weight: 700;
      }
    }

    &:hover {
      .upload-delete.show-delete {
        display: flex;
      }
    }
  }

  .add-image-btn {
    cursor: pointer;
    border: 2px dashed var(--border-color);
    position: relative;

    input[type='file'] {
      cursor: pointer;
      display: none;
    }

    .icon-add {
      font-size: 30px;
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      color: #1878f3; // TODO
    }
  }
}
</style>

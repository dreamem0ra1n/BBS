<template>
  <div class="avatar-edit">
    <div class="avatar-view" :style="{ backgroundImage: 'url(' + value + ')' }">
      <div class="upload-view" @click="pickImage">
        <i class="iconfont icon-upload" />
        <span>点击修改</span>
      </div>
    </div>
    <input
      ref="uploadImage"
      accept="image/*"
      type="file"
      @change="selectImage"
      @input="selectImage"
    />

    <div
      v-if="cropDialogVisible"
      class="crop-modal"
      role="dialog"
      aria-modal="true"
      @click.self="cancelCrop"
    >
      <div class="crop-dialog">
        <div class="crop-dialog-header">
          <span>裁剪头像</span>
          <button
            type="button"
            aria-label="关闭"
            :disabled="uploading"
            @click="cancelCrop"
          >
            ×
          </button>
        </div>
        <div class="crop-editor">
          <div
            ref="cropStage"
            class="crop-stage"
            :style="cropStageStyle"
            @mousedown="startCropDrag"
            @touchstart="startCropDrag"
          >
            <img
              v-if="cropImageUrl"
              ref="cropImage"
              :src="cropImageUrl"
              class="crop-image"
              :style="cropImageStyle"
              draggable="false"
              @load="onCropImageLoad"
            />
            <div class="crop-frame" />
          </div>
          <div class="crop-zoom">
            <span>缩放</span>
            <input
              v-model.number="cropZoom"
              type="range"
              min="1"
              max="3"
              step="0.01"
              :disabled="!cropImageLoaded"
              @input="updateCropScale"
            />
          </div>
          <p class="crop-tip">拖动图片调整裁剪区域</p>
        </div>
        <div class="crop-dialog-footer">
          <button type="button" :disabled="uploading" @click="cancelCrop">
            取消
          </button>
          <button
            type="button"
            class="confirm-button"
            :disabled="!cropImageLoaded || uploading"
            @click="confirmCrop"
          >
            {{ uploading ? '上传中…' : '确定' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    value: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      cropDialogVisible: false,
      cropImageUrl: '',
      cropImageLoaded: false,
      cropImageWidth: 0,
      cropImageHeight: 0,
      cropStageSize: 280,
      cropScale: 1,
      cropBaseScale: 1,
      cropZoom: 1,
      cropOffset: { x: 0, y: 0 },
      dragging: false,
      dragStart: null,
      uploading: false,
      processingFileSelection: false,
    }
  },
  computed: {
    cropStageStyle() {
      return {
        width: `${this.cropStageSize}px`,
        height: `${this.cropStageSize}px`,
      }
    },
    cropImageStyle() {
      return {
        width: `${this.cropImageWidth * this.cropScale}px`,
        height: `${this.cropImageHeight * this.cropScale}px`,
        transform: `translate(${this.cropOffset.x}px, ${this.cropOffset.y}px)`,
      }
    },
  },
  beforeDestroy() {
    this.stopCropDrag()
    this.revokeCropImageUrl()
  },
  methods: {
    pickImage() {
      this.$refs.uploadImage.dispatchEvent(new MouseEvent('click'))
    },
    async selectImage(e) {
      if (this.processingFileSelection) {
        return
      }
      this.processingFileSelection = true
      this.$nextTick(() => {
        this.processingFileSelection = false
      })
      const file = e.target.files && e.target.files[0]
      e.target.value = ''
      if (!file) {
        return
      }
      if (!file.type || !file.type.startsWith('image/')) {
        this.$message.error('请选择图片文件')
        return
      }
      if (this.isGifFile(file)) {
        this.$message.info('GIF 将保留动画并直接上传')
        await this.uploadOriginalAvatar(file)
        return
      }
      this.revokeCropImageUrl()
      this.cropImageUrl = URL.createObjectURL(file)
      this.cropImageLoaded = false
      this.cropDialogVisible = true
      this.$nextTick(() => this.updateCropStageSize())
    },
    isGifFile(file) {
      return (
        file.type.toLowerCase() === 'image/gif' ||
        file.name.toLowerCase().endsWith('.gif')
      )
    },
    async uploadOriginalAvatar(file) {
      if (this.uploading) {
        return
      }
      this.uploading = true
      try {
        await this.uploadAvatarFile(file)
      } catch (error) {
        this.handleUploadError(error)
      } finally {
        this.uploading = false
      }
    },
    updateCropStageSize() {
      if (!this.$refs.cropStage) {
        return
      }
      const stageWidth = this.$refs.cropStage.clientWidth
      if (stageWidth > 0 && stageWidth !== this.cropStageSize) {
        this.cropStageSize = stageWidth
        if (this.cropImageLoaded) {
          this.initializeCropPosition()
        }
      }
    },
    onCropImageLoad() {
      const image = this.$refs.cropImage
      if (!image || !image.naturalWidth || !image.naturalHeight) {
        return
      }
      this.cropImageWidth = image.naturalWidth
      this.cropImageHeight = image.naturalHeight
      this.cropImageLoaded = true
      this.initializeCropPosition()
    },
    initializeCropPosition() {
      this.cropBaseScale = Math.max(
        this.cropStageSize / this.cropImageWidth,
        this.cropStageSize / this.cropImageHeight
      )
      this.cropZoom = 1
      this.cropScale = this.cropBaseScale
      this.cropOffset = {
        x: (this.cropStageSize - this.cropImageWidth * this.cropScale) / 2,
        y: (this.cropStageSize - this.cropImageHeight * this.cropScale) / 2,
      }
    },
    updateCropScale() {
      if (!this.cropImageLoaded) {
        return
      }
      const previousScale = this.cropScale
      const nextScale = this.cropBaseScale * this.cropZoom
      const centerX =
        (this.cropStageSize / 2 - this.cropOffset.x) / previousScale
      const centerY =
        (this.cropStageSize / 2 - this.cropOffset.y) / previousScale
      this.cropScale = nextScale
      this.cropOffset = this.clampCropOffset({
        x: this.cropStageSize / 2 - centerX * nextScale,
        y: this.cropStageSize / 2 - centerY * nextScale,
      })
    },
    startCropDrag(event) {
      if (!this.cropImageLoaded || this.uploading) {
        return
      }
      const point = this.getPointerPosition(event)
      if (!point) {
        return
      }
      event.preventDefault()
      this.dragging = true
      this.dragStart = {
        x: point.x,
        y: point.y,
        offsetX: this.cropOffset.x,
        offsetY: this.cropOffset.y,
      }
      window.addEventListener('mousemove', this.onCropDrag)
      window.addEventListener('mouseup', this.stopCropDrag)
      window.addEventListener('touchmove', this.onCropDrag, { passive: false })
      window.addEventListener('touchend', this.stopCropDrag)
    },
    onCropDrag(event) {
      if (!this.dragging || !this.dragStart) {
        return
      }
      const point = this.getPointerPosition(event)
      if (!point) {
        return
      }
      event.preventDefault()
      this.cropOffset = this.clampCropOffset({
        x: this.dragStart.offsetX + point.x - this.dragStart.x,
        y: this.dragStart.offsetY + point.y - this.dragStart.y,
      })
    },
    stopCropDrag() {
      this.dragging = false
      this.dragStart = null
      if (typeof window !== 'undefined') {
        window.removeEventListener('mousemove', this.onCropDrag)
        window.removeEventListener('mouseup', this.stopCropDrag)
        window.removeEventListener('touchmove', this.onCropDrag)
        window.removeEventListener('touchend', this.stopCropDrag)
      }
    },
    getPointerPosition(event) {
      const point =
        event.touches && event.touches.length ? event.touches[0] : event
      if (!point || typeof point.clientX !== 'number') {
        return null
      }
      return { x: point.clientX, y: point.clientY }
    },
    clampCropOffset(offset) {
      const imageWidth = this.cropImageWidth * this.cropScale
      const imageHeight = this.cropImageHeight * this.cropScale
      return {
        x: Math.min(0, Math.max(this.cropStageSize - imageWidth, offset.x)),
        y: Math.min(0, Math.max(this.cropStageSize - imageHeight, offset.y)),
      }
    },
    cancelCrop() {
      if (!this.uploading) {
        this.cropDialogVisible = false
      }
    },
    async confirmCrop() {
      if (!this.cropImageLoaded || this.uploading) {
        return
      }
      this.uploading = true
      try {
        const file = await this.createCroppedFile()
        await this.uploadAvatarFile(file)
      } catch (error) {
        this.handleUploadError(error)
      } finally {
        this.uploading = false
      }
    },
    async uploadAvatarFile(file) {
      const formData = new FormData()
      formData.append('image', file, file.name)
      formData.append('source', 'avatar')
      const ret = await this.$axios.post('/api/file/upload/img', formData)
      await this.$axios.post('/api/user/update/avatar', { avatar: ret.url })
      this.cropDialogVisible = false
      this.$emit('input', ret.url)
      this.$emit('success', ret.url)
      this.resetCrop()
    },
    handleUploadError(error) {
      const message =
        (error && error.message) ||
        (error && error.msg) ||
        (error && error.data && (error.data.message || error.data.msg)) ||
        '头像上传失败'
      this.$message.error(message)
      this.$emit('error', error)
    },
    createCroppedFile() {
      return new Promise((resolve, reject) => {
        const image = this.$refs.cropImage
        if (!image) {
          reject(new Error('图片加载失败'))
          return
        }
        const canvas = document.createElement('canvas')
        const outputSize = 512
        const sourceSize = this.cropStageSize / this.cropScale
        const sourceX = Math.max(0, -this.cropOffset.x / this.cropScale)
        const sourceY = Math.max(0, -this.cropOffset.y / this.cropScale)
        canvas.width = outputSize
        canvas.height = outputSize
        const context = canvas.getContext('2d')
        context.drawImage(
          image,
          sourceX,
          sourceY,
          sourceSize,
          sourceSize,
          0,
          0,
          outputSize,
          outputSize
        )
        canvas.toBlob(
          (blob) => {
            if (!blob) {
              reject(new Error('头像裁剪失败'))
              return
            }
            resolve(
              new File([blob], 'avatar.jpg', {
                type: 'image/jpeg',
                lastModified: Date.now(),
              })
            )
          },
          'image/jpeg',
          0.9
        )
      })
    },
    revokeCropImageUrl() {
      if (this.cropImageUrl && typeof URL !== 'undefined') {
        URL.revokeObjectURL(this.cropImageUrl)
      }
    },
    resetCrop() {
      this.stopCropDrag()
      this.revokeCropImageUrl()
      this.cropImageUrl = ''
      this.cropImageLoaded = false
      this.cropImageWidth = 0
      this.cropImageHeight = 0
      this.cropScale = 1
      this.cropZoom = 1
    },
  },
}
</script>

<style lang="scss" scoped>
.avatar-edit {
  .avatar-view {
    width: 120px;
    height: 120px;
    background-size: cover;
    background-color: var(--bg-color2);
    border-radius: 50%;
    position: relative;

    &:hover {
      .upload-view {
        visibility: visible;
      }
    }

    .upload-view {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      color: var(--text-color);
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
      border-radius: 50%;
      background-color: var(--bg-color-alpha);
      visibility: hidden;
      cursor: pointer;

      span {
        font-size: 13px;
        font-weight: 500;
      }
    }
  }

  input[type='file'] {
    display: none;
  }
}

.crop-modal {
  position: fixed;
  z-index: 3000;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(0, 0, 0, 0.55);

  .crop-dialog {
    width: 420px;
    max-width: 100%;
    max-height: calc(100vh - 48px);
    overflow: auto;
    border-radius: 6px;
    background: var(--bg-color);
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.25);
  }

  .crop-dialog-header,
  .crop-dialog-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 18px;
    color: var(--text-color);
  }

  .crop-dialog-header {
    border-bottom: 1px solid var(--border-color);
    font-size: 16px;
    font-weight: 600;

    button {
      padding: 0;
      border: 0;
      background: transparent;
      color: var(--text-color-secondary);
      cursor: pointer;
      font-size: 24px;
      line-height: 20px;
    }
  }

  .crop-dialog-footer {
    justify-content: flex-end;
    border-top: 1px solid var(--border-color);

    button {
      margin-left: 10px;
      padding: 7px 15px;
      border: 1px solid var(--border-color);
      border-radius: 4px;
      background: var(--bg-color);
      color: var(--text-color);
      cursor: pointer;

      &:disabled {
        cursor: not-allowed;
        opacity: 0.6;
      }
    }

    .confirm-button {
      border-color: var(--color2);
      background: var(--color2);
      color: #fff;
    }
  }
}

.crop-editor {
  display: flex;
  align-items: center;
  flex-direction: column;

  .crop-stage {
    max-width: calc(100vw - 48px);
    max-height: calc(100vw - 48px);
    overflow: hidden;
    position: relative;
    background: #202020;
    cursor: move;
    user-select: none;
    touch-action: none;

    .crop-image {
      position: absolute;
      top: 0;
      left: 0;
      max-width: none;
      pointer-events: none;
      transform-origin: 0 0;
    }

    .crop-frame {
      position: absolute;
      z-index: 1;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      border: 2px solid #fff;
      box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.45);
      pointer-events: none;
    }
  }

  .crop-zoom {
    display: flex;
    align-items: center;
    width: 100%;
    max-width: 280px;
    margin-top: 18px;
    color: var(--text-color);
    font-size: 13px;

    span {
      margin-right: 10px;
    }

    input {
      flex: 1;
    }
  }

  .crop-tip {
    margin-top: 8px;
    color: var(--text-color-secondary);
    font-size: 12px;
  }
}
</style>

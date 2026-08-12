<template>
  <el-dialog
    title="编辑评论"
    :visible.sync="dialogVisible"
    width="min(600px, 90%)"
    custom-class="comment-edit-dialog"
    append-to-body
    @closed="close"
  >
    <text-editor
      v-if="dialogVisible"
      v-model="form"
      :height="120"
      submit-text="保存"
      @submit="submit"
    />
  </el-dialog>
</template>

<script>
export default {
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    comment: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      dialogVisible: this.visible,
      submitting: false,
      form: {
        content: '',
        imageList: [],
      },
    }
  },
  watch: {
    visible(value) {
      this.dialogVisible = value
      if (value) {
        this.form = {
          content: this.comment.rawContent || '',
          imageList: (this.comment.imageList || []).map((image) => ({
            url: image.url,
            preview: image.preview,
          })),
        }
      }
    },
    dialogVisible(value) {
      this.$emit('update:visible', value)
    },
  },
  methods: {
    close() {
      this.dialogVisible = false
    },
    async submit() {
      if (this.submitting || !this.comment) {
        return
      }
      this.submitting = true
      try {
        const updated = await this.$axios.post(
          `/api/comment/edit/${this.comment.commentId}`,
          {
            content: this.form.content,
            imageList: this.form.imageList.length
              ? JSON.stringify(this.form.imageList)
              : '',
          }
        )
        this.$emit('updated', updated)
        this.$message.success('编辑成功')
        this.close()
      } catch (e) {
        this.$message.error('编辑失败：' + (e.message || e))
      } finally {
        this.submitting = false
      }
    },
  },
}
</script>

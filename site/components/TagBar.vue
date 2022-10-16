<template>
  <div class="select-tags">
    <div class="tags-container">
      <div
        v-for="tag in tags"
        @click="chooseTag(tag.id)"
        v-bind:key="tag.id + 'tag'"
      >
        <div :class="tagClass(tag.id)">{{ tag.name }}</div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  props: {
    nodeId: {
      type: Number,
      required: true,
    },
  },
  async mounted() {
    const result = await this.$axios.get('/api/tag/list/' + this.nodeId)
    this.tags = result
    this.currTag = this.$store.state.env.currentTag
  },
  data() {
    return {
      tags: [],
    }
  },
  methods: {
    chooseTag(id) {
      this.$linkTo('/topic/' + this.nodeId.toString() + '/' + id.toString())
    },
    tagClass(id) {
      if (this.currTag === id) {
        return 'selected-item'
      }
      return 'tag-item'
    },
  },
}
</script>
<style lang="scss" scoped>
.select-tags {
  display: flex;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color2);
  border-radius: 4px;
  box-shadow: inset 0 1px 2px rgba(10, 10, 10, 0.1);
  color: var(--text-color);
  padding: 0 8px;

  .input {
    border: none;
    box-shadow: none;
    margin: 0;
    padding: 0;
  }

  .tags-container {
    display: flex;

    .tag-item,
    .selected-item {
      margin: 5px;
      padding: 0 25px;
      background: var(--tag-back-color);
      color: var(--text-color);
      line-height: 30px;
      border-radius: 20px;
      border: solid;
      border-width: 1px;
      text-align: center;
      font-size: 12px;
      white-space: nowrap;
      cursor: pointer;
    }
    .selected-item {
      color: var(--text-link-color);
    }
  }
}
</style>

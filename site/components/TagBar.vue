<template>
  <div class="select-tags">
    <div class="tags-container">
      <div
        v-for="tag in tags"
        :class="tagClass(tag.tagId)"
        @click="chooseTag(tag.tagId)"
        v-bind:key="tag.tagId"
      >
        <div class="tag-item">{{ tag.tagName }}</div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  async asyncData({ params, $axios }) {
    console.log(params)
    const result = await $axios.get('/api/tag/tags/' + params.nodeId)
    return {
      tags: result.data,
      nodeId: params.nodeId || 0,
      currTag: params.tagId || null,
    }
  },
  data() {
    return {
      tags: [
        {
          tagId: 111,
          tagName: '111',
        },
      ],
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

  .tags-selected {
    display: flex;

    .tag-item {
      margin: 5px;
      padding: 0 25px;
      background: #ffffff;
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

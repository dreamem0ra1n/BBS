<template>
  <div class="select-tags">
    <div class="tags-container">
      <div class="tags-selected">
        <div
          v-for="(item, index) in allTag[currentDepId]"
          :key="item.id + index"
          :class="tagClass(item.name)"
          @click="selectTag(item.name)"
        >
          {{ item.name }}
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    nodeId: {
      type: Number,
      default: 0,
    },
    setTags: {
      type: Function,
      default: (id) => {},
    },
    tags: {
      type: Array,
      default: () => [],
    },
  },
  async mounted() {
    try {
      this.deps = await this.$axios.get('/api/topic/nodes')
      await Promise.all(
        this.deps.map((dep) => {
          return this.$axios
            .get('/api/tag/list/' + dep.nodeId)
            .then((depTag) => {
              this.allTag[dep.nodeId] = depTag
            })
        })
      )
      this.allTag.push()
      if (this.tags) this.selectIndex = this.tags
    } catch (e) {
      console.log(e)
    }
  },
  data() {
    return {
      allTag: [],
      deps: [],
      selectIndex: [],
    }
  },
  watch: {
    nodeId() {
      this.selectIndex = []
    },
  },
  computed: {
    currentDepId() {
      return this.nodeId
    },
  },
  methods: {
    /**
     * 手动点击选择标签
     * @param name
     */
    selectTag(name) {
      if (!this.selectIndex.includes(name)) {
        this.addTag(name)
      } else this.removeTag(name)
      this.$emit('setTag', this.selectIndex)
    },

    /**
     * 添加标签
     * @param event
     */
    addTag(name) {
      this.selectIndex.push(name)
    },
    removeTag(name) {
      this.selectIndex = this.selectIndex.filter((item) => {
        return item !== name
      })
    },
    tagClass(name) {
      if (this.selectIndex.includes(name)) {
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

  .tags-selected {
    display: flex;

    .tag-item,
    .selected-item {
      margin: 5px;
      padding: 0 25px;
      background: var(--bg-color);
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

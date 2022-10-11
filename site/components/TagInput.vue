<template>
  <div class="select-tags">
    <div class="tags-container">
      <div class="tags-selected">
        <div
          v-for="(item, index) in allTag[currentDepId]"
          :key="item.id + index"
          :class="tagClass(item.id)"
          @click="selectTag(item.id)"
        >
          {{ item.name }}
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { all } from 'q'

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
    } catch (e) {
      console.log(e)
    }
  },
  data() {
    return {
      tags: [],
      allTag: [],
      deps: [],
      maxTagCount: 1, // 最多可以选择的标签数量
      maxWordCount: 15, // 每个标签最大字数
      showRecommendTags: false, // 是否显示推荐
      inputTag: '',
      presetTags: [],
      autocompleteTags: [],
      selectIndex: [],
    }
  },
  watch: {
    nodeId() {
      this.tags = []
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
     * @param index
     */
    selectTag(index) {
      if (!this.selectIndex.includes(index)) {
        this.addTag(index)
      } else this.removeTag(index)
      this.$emit(
        'setTag',
        this.tags.map((tag) => tag.name)
      )
    },

    /**
     * 添加标签
     * @param event
     */
    addTag(index) {
      this.selectIndex.push(index)
      this.tags = this.allTag[this.currentDepId].filter(
        (tag) => tag.id === index
      )
    },
    removeTag(index) {
      this.selectIndex = this.selectIndex.filter((item) => {
        return item !== index
      })
      this.tags = this.tags.filter((tag) => {
        return tag.id !== index
      })
    },
    tagClass(id) {
      if (this.selectIndex.includes(id)) {
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

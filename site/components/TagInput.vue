<template>
  <div class="select-tags">
    <input id="tags" v-model="tags" name="tags" type="hidden" />
    <div class="tags-selected">
      <div v-for="tag in tags" :key="tag.tagId" class="tag-item">
        <span>{{ tag.tagName }}</span>
      </div>
    </div>
    <transition name="el-zoom-in-bottom">
      <div class="autocomplete-tags">
        <div class="tags-container">
          <section class="tag-section">
            <div
              v-for="(item, index) in allTag[currentDepId]"
              :key="index"
              class="tag-item"
              @click="selectTag(item.tagId)"
              v-text="item.tagName"
            />
          </section>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
export default {
  async asyncData({ $axios, params, store }) {
    try {
      const _allTag = []
      const _deps = await $axios.get('/api/topic/nodes')
      await Promise.all(
        _deps.map((dep) => {
          return $axios.get('/api/tag/tags/' + dep.nodeId).then((depTag) => {
            _allTag[dep.id] = depTag
          })
        })
      )
      return {
        allTag: _allTag || [[{ tagId: 114514, tagName: '666' }]],
        deps: _deps,
      }
    } catch (e) {
      console.log(e)
    }
  },
  props: {
    value: {
      type: Object,
      default() {
        return {}
      },
    },
  },
  data() {
    return {
      tags: this.value.tags || [], // 与上级组件中的 postForm.tags 双向绑定
      allTag: [[{ tagId: 114514, tagName: '666' }]],
      deps: [],
      maxTagCount: 1, // 最多可以选择的标签数量
      maxWordCount: 15, // 每个标签最大字数
      showRecommendTags: false, // 是否显示推荐
      inputTag: '',
      presetTags: [],
      autocompleteTags: [],
      selectIndex: [],
      currentDepId: this.value.nodeId || 0,
    }
  },
  computed: {
    showTags() {
      return this.allTag[this.currentDepId]
    },
  },
  methods: {
    /**
     * 手动点击选择标签
     * @param index
     */
    selectTag(index) {
      if (!this.selectIndex.includes(index)) this.addTag(index)
      else this.removeTag(index)
    },

    /**
     * 添加标签
     * @param event
     */
    addTag(index) {
      this.selectIndex.push(index)
      this.tags.push(
        this.allTag[this.currentDepId].filter((tag) => tag.tagId === index)
      )
    },
    removeTag(index) {
      this.selectIndex.filter((item) => {
        return item !== index
      })
      this.tags.filter((tag) => {
        return tag.tagId !== index
      })
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
      padding: 0 10px;
      background: var(--bg-color3);
      color: var(--text-color);
      line-height: 30px;
      border-radius: 5px;

      text-align: center;
      font-size: 12px;
      white-space: nowrap;

      i {
        font-size: 12px;
        margin-left: 4px;
      }

      i:hover {
        color: red;
        cursor: pointer;
      }
    }
  }

  .autocomplete-tags {
    z-index: 2000;
    left: 0;
    right: 0;
    top: 42px;
    bottom: 0;
    position: absolute;

    .tags-container {
      scroll-behavior: smooth;
      position: relative;
      // background: #f7f7f7;
      background-color: var(--bg-color);
      border-left: 1px solid var(--border-color2);
      border-right: 1px solid var(--border-color2);
      border-bottom: 1px solid var(--border-color2);

      .tag-section {
        font-size: 14px;
        line-height: 16px;

        .tag-item {
          padding: 8px 15px;
          cursor: pointer;

          &.active,
          &:hover {
            color: var(--text-color5);
            // background: #006bde;
            background-color: var(--bg-color2);
          }
        }
      }
    }
  }

  .recommend-tags {
    z-index: 2000;
    left: 0;
    right: 0;
    top: 42px;
    bottom: 0;
    position: absolute;

    .tags-container {
      scroll-behavior: smooth;
      position: relative;
      background: #f7f7f7;
      border-left: 1px solid var(--border-color2);
      border-right: 1px solid var(--border-color2);
      border-bottom: 1px solid var(--border-color2);
      padding: 0 10px 10px 10px;

      .header {
        font-weight: bold;
        font-size: 15px;
        color: #017e66;
        border-bottom: 1px solid var(--border-color2);
        margin-bottom: 5px;
        padding-top: 5px;
        padding-bottom: 5px;

        .close-recommend {
          float: right;
          cursor: pointer;
          &:hover {
            color: red;
          }
        }
      }

      .tag-item {
        padding: 0 11px;
        border-radius: 5px;
        display: inline-block;
        color: #017e66;
        background-color: rgba(1, 126, 102, 0.08);
        height: 22px;
        line-height: 22px;
        font-weight: normal;
        font-size: 13px;
        text-align: center;

        &:not(:last-child) {
          margin-right: 5px;
        }

        img {
          width: 16px;
          height: 16px;
          margin-right: 5px;
          margin-top: -1px;
          vertical-align: middle;
        }
      }

      .tag-item:hover,
      .tag-item:focus {
        background-color: #017e66;
        color: var(--text-color5);
        text-decoration: none;
      }
    }
  }
}
</style>

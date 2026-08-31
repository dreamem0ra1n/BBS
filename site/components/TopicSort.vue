<template>
  <div class="topic-sort" aria-label="帖子排序方式">
    <span class="topic-sort-label">排序</span>
    <nuxt-link
      :class="{ active: normalizedValue === 'comment' }"
      :to="sortLink('comment')"
    >
      最新评论
    </nuxt-link>
    <nuxt-link
      :class="{ active: normalizedValue === 'create' }"
      :to="sortLink('create')"
    >
      最新发帖
    </nuxt-link>
  </div>
</template>

<script>
export default {
  props: {
    value: {
      type: String,
      default: 'comment',
    },
  },
  computed: {
    normalizedValue() {
      return this.value === 'create' ? 'create' : 'comment'
    },
  },
  methods: {
    sortLink(sort) {
      const query = { ...this.$route.query }
      if (sort === 'create') {
        query.sort = sort
      } else {
        delete query.sort
      }
      return { path: this.$route.path, query }
    },
  },
}
</script>

<style lang="scss" scoped>
.topic-sort {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  margin: 10px 0;
  padding: 8px 12px;
  border-radius: 3px;
  background: var(--bg-color);
  color: var(--text-color3);
  font-size: 13px;

  .topic-sort-label {
    margin-right: 2px;
  }

  a {
    padding: 4px 10px;
    border-radius: 14px;
    color: var(--text-color3);

    &:hover,
    &.active {
      color: var(--text-link-color);
      background: var(--bg-color2);
    }
  }
}
</style>

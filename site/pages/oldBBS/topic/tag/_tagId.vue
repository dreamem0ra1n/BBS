<template>
  <div class="topics-main">
    <div class="container main-container left-main size-320">
      <div class="left-container">
        <div class="main-content no-padding no-bg topics-wrapper">
          <div class="topics-nav">
            <old-nav :currentTag="tagId" :data="data" />
          </div>
          <div class="topics-main">
            <old-tag-bar :tags="tags" :currTag="tagId" />
            <load-more
              v-if="topicsPage"
              v-slot="{ results }"
              :init-data="topicsPage"
              :url="'/api/topic/tag/topics?isOld=true&tagId=' + tagId"
            >
              <topic-list :topics="results" :isOld="true" />
            </load-more>
          </div>
        </div>
      </div>
      <div class="right-container">
        <friend-links :links="links" />
      </div>
    </div>
  </div>
</template>
<script>
export default {
  async asyncData({ $axios, params, error, store }) {
    const json = require('~/oldBBSgit.json')
    const data = await $axios.get('/api/topic/tag/topics', {
      params: {
        isOld: true,
        tagId: parseInt(params.tagId),
      },
    })
    let tags = []
    json.forEach((group) => {
      group.fdn.forEach((forum) => {
        forum.fdn.forEach((sub) => {
          if (
            sub.fid === parseInt(params.tagId) ||
            forum.fid === parseInt(params.tagId)
          ) {
            tags = Object.assign(forum.fdn)
          }
        })
      })
    })
    const links = await $axios.get('/api/link/toplinks')
    return {
      tagId: parseInt(params.tagId),
      topicsPage: data,
      links,
      data: json,
      tags,
    }
  },
  data() {
    return {}
  },
  computed: {},
  methods: {},
}
</script>

<style lang="scss" scoped></style>

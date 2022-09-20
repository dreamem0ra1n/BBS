export const state = () => ({
  keyword: '',
  nodeId: 0,
  timeRange: 0,
  page: 1,
  searchPage: null,
  loading: false,
  tags: [],
})

export const mutations = {
  setKeyword(state, keyword) {
    state.keyword = keyword
  },
  setNodeId(state, nodeId) {
    state.nodeId = nodeId
  },
  setTimeRange(state, timeRange) {
    state.timeRange = timeRange
  },
  setPage(state, page) {
    state.page = page
  },
  setSearchPage(state, searchPage) {
    state.searchPage = searchPage
  },
  setLoading(state, loading) {
    state.loading = loading
  },
  setTags(state, tags) {
    state.tags = tags
  },
}

export const actions = {
  initParams(context, { keyword, nodeId, page }) {
    context.commit('setKeyword', keyword || '')
    context.commit('setNodeId', nodeId || 0)
    context.commit('setPage', page || 1)
  },
  initTags(context) {
    // this.$axios.
    // context.commit("setTags",)
  },
  changeNodeId(context, nodeId) {
    context.commit('setNodeId', nodeId || 0)
    context.dispatch('searchTopic')
  },
  changeTimeRange(context, timeRange) {
    context.commit('setTimeRange', timeRange || 0)
    context.dispatch('searchTopic')
  },
  async searchTopic({ state, commit }) {
    commit('setLoading', true)
    try {
      const result = await this.$axios.post('/api/search/topic', {
        keyword: state.keyword,
        nodeId: state.nodeId,
        timeRange: state.timeRange,
        page: state.page,
      })
      commit('setSearchPage', result)
    } finally {
      commit('setLoading', false)
    }
  },
}

export const getters = {}

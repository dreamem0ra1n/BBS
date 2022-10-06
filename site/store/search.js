export const state = () => ({
  keyword: '',
  nodeId: 0,
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
  changeNodeId(context, nodeId) {
    context.commit('setNodeId', nodeId || 0)
    context.dispatch('searchTopic')
  },
  async searchTopic({ state, commit }) {
    commit('setLoading', true)
    try {
      const result = await this.$axios.post('/api/search/topic', {
        keyword: state.keyword,
        nodeId: state.nodeId,
        page: state.page,
      })
      commit('setSearchPage', result)
    } finally {
      commit('setLoading', false)
    }
  },
}

export const getters = {}

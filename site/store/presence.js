export const state = () => ({ users: [], total: 0, enabled: false })

export const mutations = {
  setUsers(state, users) {
    state.users = users
  },
  setTotal(state, total) {
    state.total = total
  },
  setEnabled(state, enabled) {
    state.enabled = enabled
  },
}

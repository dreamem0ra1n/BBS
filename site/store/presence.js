export const state = () => ({ users: [], enabled: false })

export const mutations = {
  setUsers(state, users) {
    state.users = users
  },
  setEnabled(state, enabled) {
    state.enabled = enabled
  },
}

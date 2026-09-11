export default function ({ $axios, app, store }) {
  let socket = null
  let retryAttempt = 0
  let stopped = false

  const retry = () => {
    if (stopped || !store.state.user.current) return
    store.commit('presence/setEnabled', false)
    const base = Math.min(30000, 1000 * 2 ** retryAttempt++)
    const jitter = Math.floor(Math.random() * Math.max(250, base * 0.25))
    setTimeout(connect, base + jitter)
  }

  const connect = async () => {
    if (stopped || !store.state.user.current || socket) return
    try {
      const { ticket } = await $axios.post('/api/presence/ticket')
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      socket = new WebSocket(
        `${protocol}//${
          window.location.host
        }/bbs2/api/ws/presence?ticket=${encodeURIComponent(ticket)}`
      )
      socket.onopen = () => {
        retryAttempt = 0
        store.commit('presence/setEnabled', true)
      }
      socket.onclose = () => {
        socket = null
        retry()
      }
      socket.onerror = () => {
        if (socket) socket.close()
      }
      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          if (message.type === 'snapshot') {
            const users = message.users || []
            store.commit('presence/setUsers', users)
            store.commit(
              'presence/setTotal',
              typeof message.total === 'number' ? message.total : users.length
            )
          }
          if (message.type === 'unavailable') {
            store.commit('presence/setEnabled', false)
          }
        } catch (e) {
          /* ignore malformed server messages */
        }
      }
    } catch (e) {
      retry()
    }
  }

  store.watch(
    (state) => state.user.current,
    (user) => {
      stopped = !user
      if (user) {
        stopped = false
        retryAttempt = 0
        connect()
      } else if (socket) {
        socket.close()
        socket = null
        store.commit('presence/setUsers', [])
        store.commit('presence/setTotal', 0)
        store.commit('presence/setEnabled', false)
      }
    },
    { immediate: true }
  )
}

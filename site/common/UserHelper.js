class UserHelper {
  getPosition(user) {
    if (!user) {
      return ''
    }
    if (typeof user.position === 'string' && user.position.trim()) {
      return user.position
    }
    const department =
      typeof user.department === 'string' ? user.department.trim() : ''
    const firstRole =
      Array.isArray(user.roles) && typeof user.roles[0] === 'string'
        ? user.roles[0]
        : ''
    const role = firstRole.split('_')[0].trim()
    return [department, role].filter(Boolean).join('-')
  }

  hasRole(user, role) {
    if (!user || !user.roles || !user.roles.length) {
      return false
    }
    for (let i = 0; i < user.roles.length; i++) {
      if (user.roles[i].includes(role)) {
        return true
      }
    }
    return false
  }

  hasAnyRole(user, ...roles) {
    if (!roles || !roles.length) {
      return false
    }
    for (let i = 0; i < roles.length; i++) {
      if (this.hasRole(user, roles[i])) {
        return true
      }
    }
    return false
  }

  isOwner(user) {
    return this.hasRole(user, 'owner')
  }

  isAdmin(user) {
    return this.hasRole(user, '中管') || this.hasRole(user, '高管')
  }
}

export default new UserHelper()

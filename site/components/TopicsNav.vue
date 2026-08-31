<template>
  <nav class="dock-nav">
    <ul>
      <li :class="{ active: currentNodeId === 1 }">
        <nuxt-link to="/topics/node/1">
          <span class="node-name"
            ><img
              class="node-logo nav-logo"
              src="~/assets/images/icon/gonggao.svg"
            />公告</span
          >
        </nuxt-link>
      </li>
      <li :class="{ active: currentNodeId === 0 }">
        <nuxt-link to="/topics/node/newest">
          <span class="node-name"
            ><img
              class="node-logo nav-logo"
              src="~/assets/images/icon/zuixinnew3.svg"
            />最新</span
          >
        </nuxt-link>
      </li>
      <li :class="{ active: currentNodeId === -1 }">
        <nuxt-link to="/topics/node/recommend">
          <span class="node-name">
            <img
              class="node-logo nav-logo"
              src="~/assets/images/icon/tuijian.svg"
            />推荐</span
          >
        </nuxt-link>
      </li>
      <li :class="{ active: currentNodeId === -2 }">
        <nuxt-link to="/topics/node/feed">
          <span class="node-name">
            <img
              class="node-logo nav-logo"
              src="~/assets/images/feed.png"
            />关注</span
          >
        </nuxt-link>
      </li>
      <li class="dock-nav-divider"></li>
      <li
        v-for="node in nodes.filter((node) => node.nodeId !== 1)"
        :key="node.nodeId"
        :class="{ active: currentNodeId === node.nodeId }"
      >
        <div @click="goToTag(node.nodeId)">
          <img v-if="node.logo" class="node-logo" :src="node.logo" />
          <img v-else class="node-logo" src="~/assets/images/node.png" />
          <span class="node-name">{{ node.name }}</span>
        </div>
      </li>
    </ul>
  </nav>
</template>

<script>
export default {
  props: {
    nodes: {
      type: Array,
      default() {
        return []
      },
    },
  },
  computed: {
    currentNodeId() {
      return this.$store.state.env.currentNodeId
    },
  },
  methods: {
    goToTag(id) {
      this.$store.commit('env/setCurrentTag', -1919810)
      this.$linkTo('/topics/node/' + id)
    },
  },
}
</script>

<style lang="scss" scoped>
.dock-nav {
  display: block;
  position: -webkit-sticky;
  position: sticky;
  top: 10px;

  width: 200px;
  border-radius: 2px;
  background-color: var(--bg-color);
  transition: all 0.2s linear;

  ul {
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 16px 12px;

    li:not(.dock-nav-divider) {
      position: relative;
      cursor: pointer;
      height: fit-content;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
      color: var(--text-color);
      //padding: 0 12px;
      border-radius: 3px;
      transition: background-color 0.2s, color 0.2s;
      font-weight: 500;
      height: 40px;

      &:not(:first-child) {
        margin-top: 10px;
      }

      &.active {
        background-color: var(--qsc-color);
        color: var(--text-color5);

        a {
          color: var(--text-color5);
        }
      }

      &:not(.active):hover {
        background-color: hsla(0, 0%, 94.9%, 0.6);
      }

      div {
        text-decoration: none;
        cursor: pointer;
        color: var(--text-color3);
        width: 100%;
        height: 100%;
        text-align: center;
        line-height: 30px;
        padding-left: 10px;

        display: flex;
        align-items: center;
        //justify-content: center;
        .node-logo {
          width: 24px;
          height: 24px;
          border-radius: 4px;
          margin-right: 10px;
          background-color: var(--bg-color);
        }
      }
    }

    li.dock-nav-divider {
      height: 15px;
      border-bottom: 1px solid var(--border-color);
    }
  }
  .node-logo {
    width: 24px;
    height: 24px;
    border-radius: 4px;
    margin-right: 10px;
  }
  .nav-logo {
    margin-bottom: -5px;
  }
}
</style>

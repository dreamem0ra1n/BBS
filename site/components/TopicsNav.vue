<template>
  <nav class="dock-nav">
    <ul>
      <li
        :class="{
          active: currentNodeId === 1,
          'has-description': nodeDescription(1),
        }"
        :title="nodeDescription(1) || null"
      >
        <nuxt-link to="/topics/node/1">
          <span class="node-name"
            ><img
              class="node-logo nav-logo"
              src="~/assets/images/icon/gonggao.svg"
            />公告</span
          >
        </nuxt-link>
        <span v-if="nodeDescription(1)" class="node-description" role="tooltip">
          {{ nodeDescription(1) }}
        </span>
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
        :class="{
          active: currentNodeId === node.nodeId,
          'has-description': node.description,
        }"
        :title="node.description || null"
      >
        <div @click="goToTag(node.nodeId)">
          <img v-if="node.logo" class="node-logo" :src="node.logo" />
          <img v-else class="node-logo" src="~/assets/images/node.png" />
          <span class="node-name">{{ node.name }}</span>
        </div>
        <span v-if="node.description" class="node-description" role="tooltip">
          {{ node.description }}
        </span>
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
    nodeDescription(nodeId) {
      const node = this.nodes.find((item) => item.nodeId === nodeId)
      return node ? node.description : ''
    },
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
  z-index: 20;

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

      .node-description {
        position: absolute;
        z-index: 10;
        top: 50%;
        left: calc(100% + 8px);
        width: 260px;
        padding: 8px 10px;
        border: 1px solid var(--border-color);
        border-radius: 4px;
        background-color: var(--bg-color);
        color: var(--text-color3);
        font-size: 13px;
        font-weight: 400;
        line-height: 1.5;
        text-align: left;
        white-space: normal;
        box-shadow: 0 2px 8px var(--bg-color-alpha);
        opacity: 0;
        visibility: hidden;
        pointer-events: none;
        transform: translate(-4px, -50%);
        transition: opacity 0.15s, transform 0.15s, visibility 0.15s;
      }

      &.has-description:hover .node-description,
      &.has-description:focus-within .node-description {
        opacity: 1;
        visibility: visible;
        transform: translate(0, -50%);
      }

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

<template>
  <nav class="dock-nav">
    <ul>
      <li
        v-for="node in data"
        :key="node.fid"
        :class="{
          active: isActive(node),
        }"
      >
        <div class="wrapper">
          <div @click="goToTag(node.fid)">
            <img class="node-logo" src="~/assets/images/node.png" />
            <span class="node-name">{{ node.name }}</span>
          </div>
          <ul v-if="isActive(node)">
            <li
              v-for="forum in node.fdn"
              :key="forum.fid"
              :class="{
                active: isActive(forum),
              }"
            >
              <div @click="goToTag(forum.fid)">
                <span class="node-name">{{ forum.fname }}</span>
              </div>
            </li>
          </ul>
        </div>
      </li>
    </ul>
  </nav>
</template>
<script>
export default {
  props: {
    currentTag: {
      type: Number,
      required: false,
      default() {
        return -1
      },
    },
    data: {
      type: Array,
      required: true,
    },
  },
  methods: {
    goToTag(id) {
      // this.$store.commit('env/setCurrentTag', 'old' + id)
      this.$linkTo('/oldBBS/topic/tag/' + id)
    },
    isActive(group) {
      if (this.currentTag === group.fid) return true
      if (group.fdn) {
        for (const forum of group.fdn) {
          if (forum.fid === this.currentTag) return true
          if (forum.fdn)
            for (const sub of forum.fdn) {
              if (sub.fid === this.currentTag) return true
            }
        }
      }
      return false
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
      min-height: 40px;

      &:not(:first-child) {
        margin-top: 10px;
      }

      &.active {
        background-color: var(--navbar-box-shadow-color);
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
      li.active {
        background-color: var(--qsc-color);
        color: var(--text-color5);
      }
      &:not(.active):hover {
        background-color: var(--bg-color);
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
.wrapper {
  display: block !important;
}
</style>

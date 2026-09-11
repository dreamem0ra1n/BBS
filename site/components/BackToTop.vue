<template>
  <transition name="back-to-top">
    <div
      v-show="visible"
      class="back-to-top"
      role="button"
      tabindex="0"
      aria-label="回到顶部"
      title="回到顶部"
      @click="scrollToTop"
      @keydown.enter.prevent="scrollToTop"
      @keydown.space.prevent="scrollToTop"
    >
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M12 20.5C11.5858 20.5 11.25 20.1642 11.25 19.75L11.25 5.56066L5.03033 11.7803C4.73744 12.0732 4.26256 12.0732 3.96967 11.7803C3.67678 11.4874 3.67678 11.0126 3.96967 10.7197L11.4697 3.21967C11.7626 2.92678 12.2374 2.92678 12.5303 3.21967L20.0303 10.7197C20.3232 11.0126 20.3232 11.4874 20.0303 11.7803C19.7374 12.0732 19.2626 12.0732 18.9697 11.7803L12.75 5.56066L12.75 19.75C12.75 20.1642 12.4142 20.5 12 20.5Z"
          fill-rule="evenodd"
          clip-rule="evenodd"
        />
      </svg>
    </div>
  </transition>
</template>

<script>
// 回到顶部按钮：滚动超过一定距离后出现在页面右下角
export default {
  data() {
    return {
      visible: false,
      // 滚动多少像素后显示按钮
      showThreshold: 300,
      ticking: false,
    }
  },
  mounted() {
    // passive 监听，避免阻塞页面滚动
    window.addEventListener('scroll', this.onScroll, { passive: true })
    this.onScroll()
  },
  beforeDestroy() {
    window.removeEventListener('scroll', this.onScroll)
  },
  methods: {
    onScroll() {
      if (this.ticking) {
        return
      }
      this.ticking = true
      window.requestAnimationFrame(() => {
        const scrollTop =
          window.pageYOffset || document.documentElement.scrollTop || 0
        this.visible = scrollTop > this.showThreshold
        this.ticking = false
      })
    },
    scrollToTop() {
      const prefersReducedMotion =
        typeof window.matchMedia === 'function' &&
        window.matchMedia('(prefers-reduced-motion: reduce)').matches
      if (
        !prefersReducedMotion &&
        'scrollBehavior' in document.documentElement.style
      ) {
        window.scrollTo({ top: 0, behavior: 'smooth' })
      } else {
        window.scrollTo(0, 0)
      }
    },
  },
}
</script>

<style lang="scss" scoped>
.back-to-top {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 40;

  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;

  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: 50%;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  cursor: pointer;
  transition: box-shadow 0.2s linear, transform 0.2s linear,
    background-color 0.2s linear;

  svg path {
    // 注意：这里不能用 --text-color3，它在暗色主题下未被求值（findColorInvert），
    // 属于非法值，会让 fill 回退成黑色，导致深色背景上看不见图标
    fill: var(--text);
    transition: fill 0.2s linear;
  }

  &:hover,
  &:focus {
    outline: none;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.18);

    svg path {
      fill: var(--text-link-color);
    }
  }

  @media screen and (max-width: 768px) {
    right: 12px;
    bottom: 16px;
    width: 40px;
    height: 40px;
  }
}

.back-to-top-enter-active,
.back-to-top-leave-active {
  transition: opacity 0.2s linear, transform 0.2s linear;
}

.back-to-top-enter,
.back-to-top-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>

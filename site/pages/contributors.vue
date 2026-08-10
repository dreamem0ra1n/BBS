<template>
  <section class="main contributors-page">
    <div class="container">
      <div class="widget page-intro">
        <div class="widget-content">
          <h1 class="page-title">开发者与贡献者</h1>
          <p class="page-description">
            感谢每一位参与维护、开发和完善求是潮 BBS 的潮人。
          </p>
        </div>
      </div>

      <div
        v-for="group in contributorGroups"
        :key="group.title"
        class="widget contributor-section"
      >
        <div class="widget-header">
          <span>{{ group.title }}</span>
          <span class="count">{{ group.members.length }}</span>
        </div>
        <div class="widget-content contributor-grid">
          <a
            v-for="member in group.members"
            :key="member.github"
            :href="`https://github.com/${member.github}`"
            class="contributor-card"
            target="_blank"
            rel="noopener noreferrer"
          >
            <span class="avatar-frame">
              <img :src="member.avatar" :alt="`${member.name} 的头像`" />
            </span>
            <span class="contributor-info">
              <strong class="contributor-name">{{ member.name }}</strong>
              <span class="github-name">@{{ member.github }}</span>
            </span>
          </a>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
export default {
  data() {
    return {
      contributorGroups: [
        {
          title: '现任维护/开发',
          members: [
            {
              name: 'dreamem0ra1n',
              github: 'dreamem0ra1n',
              avatar: require('~/assets/images/contributors/晚自习.jpg'),
            },
          ],
        },
        {
          title: '历史贡献者',
          members: [
            {
              name: '晚自习',
              github: 'dreamem0ra1n',
              avatar: require('~/assets/images/contributors/晚自习.jpg')
            },
            {
              name: 'token',
              github: 'Ag2SO4pro',
              avatar: require('~/assets/images/contributors/token.jpg'),
            },
            {
              name: 'pacman',
              github: 'IsshikiHugh',
              avatar: require('~/assets/images/contributors/pacman.png'),
            },
            {
              name: 'PTA',
              github: 'chenyanzz',
              avatar: require('~/assets/images/contributors/PTA.png'),
            },
            {
              name: 'sync',
              github: 'sync-99',
              avatar: require('~/assets/images/contributors/sync.png'),
            },
            {
              name: '墨涵',
              github: 'mohan-zeyu',
              avatar: require('~/assets/images/contributors/墨涵.png'),
            },
            {
              name: 'gingo',
              github: 'jzl-66666a',
              avatar: require('~/assets/images/contributors/gingo.jpg'),
            },
          ],
        },
      ],
    }
  },
  head() {
    return {
      title: this.$siteTitle('开发者与贡献者'),
      meta: [
        {
          hid: 'description',
          name: 'description',
          content: '求是潮 BBS 现任维护开发组与历史贡献者名单',
        },
      ],
    }
  },
}
</script>

<style lang="scss" scoped>
.contributors-page {
  .container {
    max-width: 960px;
  }
}

.page-intro {
  .widget-content {
    padding: 18px 4px;
  }
}

.page-title {
  color: var(--text-color);
  font-size: 24px;
  font-weight: 700;
  line-height: 1.4;
  margin: 0 0 6px;
}

.page-description {
  color: var(--text-color3);
  font-size: 14px;
  margin: 0;
}

.contributor-section {
  margin-top: 12px;

  .widget-header {
    justify-content: flex-start;
  }
}

.contributor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.contributor-card {
  display: flex;
  align-items: center;
  min-width: 0;
  padding: 12px;
  color: var(--text-color);
  background: var(--bg-color2);
  border: 1px solid var(--border-color);
  border-radius: calc(0.25rem - 1px);
  transition: border-color 0.2s, transform 0.2s;

  &:hover {
    color: var(--text-link-color);
    border-color: var(--border-hover-color);
    transform: translateY(-1px);
  }
}

.avatar-frame {
  width: 58px;
  height: 58px;
  min-width: 58px;
  padding: 3px;
  overflow: hidden;
  background: var(--bg-color);
  border: 1px solid var(--border-color2);
  border-radius: 50%;

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
  }
}

.contributor-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
  margin-left: 14px;
}

.contributor-name,
.github-name {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.contributor-name {
  color: var(--text-color);
  font-size: 16px;
  line-height: 1.5;
}

.github-name {
  color: var(--text-color3);
  font-size: 14px;
  line-height: 1.5;
}

@media screen and (max-width: 600px) {
  .contributor-grid {
    grid-template-columns: 1fr;
  }

  .page-title {
    font-size: 21px;
  }
}
</style>

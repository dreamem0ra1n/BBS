<template>
  <view class="page">
    <view v-if="isDev" class="dev-banner">
      DEV · {{ apiBaseUrl }} · 服务端登录态：{{ serverAuth || '检测中…' }}
    </view>

    <view class="card page-header">
      <view class="title">BBS</view>
      <view class="muted">{{ currentUser ? `欢迎回来，${currentUser.nickname}` : '浏览最新话题' }}</view>
      <button v-if="!currentUser" class="login-button" size="mini" @click="goLogin">登录</button>
    </view>

    <view v-for="topic in topics" :key="topic.topicId" class="card topic" @click="openTopic(topic.topicId)">
      <view class="topic-title">{{ topic.title || topic.content }}</view>
      <view class="muted">{{ topic.user?.nickname }} · {{ topic.commentCount }} 条评论 · {{ topic.viewCount }} 次浏览</view>
    </view>

    <view v-if="!loading && !topics.length" class="empty">暂无话题</view>
    <view v-if="loading" class="empty">加载中…</view>
    <view v-if="!hasMore && topics.length" class="empty">已经到底了</view>
  </view>
</template>

<script setup lang="ts">
import { onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import { onMounted, ref } from 'vue'
import { env } from '../../config/env'
import { getCurrentUser } from '../../utils/auth'
import { get } from '../../utils/request'

interface Topic {
  topicId: number
  title?: string
  content?: string
  commentCount?: number
  viewCount?: number
  user?: { nickname?: string }
}

interface CurrentUser {
  nickname?: string
}

interface CursorResult<T> {
  results: T[]
  cursor: string | number
  hasMore: boolean
}

const topics = ref<Topic[]>([])
const cursor = ref<string | number>(0)
const hasMore = ref(true)
const loading = ref(false)
const currentUser = ref<CurrentUser | null>(getCurrentUser<CurrentUser>())
// 开发包里显示当前接口地址，避免误以为连的是本地后端
const isDev = process.env.NODE_ENV === 'development'
const apiBaseUrl = env.apiBaseUrl
// 开发包里显示后端是否认得当前 token（后端会给匿名访客屏蔽帖子标题）
const serverAuth = ref('')
// 上一次拉列表时的登录态，用于识别登录/退出
let lastToken: string = uni.getStorageSync('userToken') || ''

async function loadTopics(reset: boolean): Promise<void> {
  if (loading.value) return
  loading.value = true
  try {
    const result = await get<CursorResult<Topic>>('/api/topic/topics', {
      cursor: reset ? 0 : cursor.value,
    })
    topics.value = reset ? result.results || [] : topics.value.concat(result.results || [])
    cursor.value = result.cursor || 0
    hasMore.value = !!result.hasMore
  } catch (error) {
    uni.showToast({ title: error instanceof Error ? error.message : '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function goLogin(): void {
  uni.navigateTo({ url: '/pages/login/login' })
}

function openTopic(id: number): void {
  uni.navigateTo({ url: `/pages/topic-detail/topic-detail?id=${id}` })
}

async function probeServerAuth(): Promise<void> {
  if (!isDev) return
  try {
    const me = await get<{ id?: number; nickname?: string }>('/api/user/current')
    serverAuth.value = me && me.id ? `已登录 ${me.nickname || me.id}` : '未登录'
  } catch (error) {
    serverAuth.value = '检测失败'
  }
}

onMounted(() => {
  loadTopics(true)
  probeServerAuth()
})
onShow(() => {
  currentUser.value = getCurrentUser<CurrentUser>()
  const token: string = uni.getStorageSync('userToken') || ''
  // 登录/退出后必须重新拉列表：后端对匿名访客会把标题屏蔽成“无权访问”
  if (token !== lastToken) {
    lastToken = token
    loadTopics(true)
  }
  probeServerAuth()
})
onPullDownRefresh(async () => {
  await loadTopics(true)
  uni.stopPullDownRefresh()
})
onReachBottom(() => {
  if (hasMore.value) loadTopics(false)
})
</script>

<style lang="scss" scoped>
.page-header {
  position: relative;
}

.dev-banner {
  margin-bottom: 16rpx;
  padding: 8rpx 16rpx;
  color: #92400e;
  background: #fef3c7;
  border-radius: 8rpx;
  font-size: 22rpx;
  word-break: break-all;
}

.title {
  margin-bottom: 8rpx;
  font-size: 40rpx;
  font-weight: 600;
}

.login-button {
  position: absolute;
  top: 28rpx;
  right: 24rpx;
  color: #ffffff;
  background: #2563eb;
}

.topic-title {
  margin-bottom: 14rpx;
  font-size: 32rpx;
  font-weight: 500;
  line-height: 1.45;
}
</style>

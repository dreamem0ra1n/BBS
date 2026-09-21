<template>
  <view class="page">
    <view v-if="topic" class="card">
      <view class="title">{{ topic.title }}</view>
      <view class="muted">{{ topic.user?.nickname }} · {{ topic.viewCount }} 次浏览</view>
      <rich-text class="content" :nodes="topic.content" />
    </view>
    <view v-else class="empty">
      <view>{{ loading ? '加载中…' : loadError || '话题不存在' }}</view>
      <view v-if="!loading && loadError" class="muted">
        后端拒绝时会返回“无权限”，通常是没有登录态或该部门的阅读权限不足。
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { get } from '../../utils/request'

interface Topic {
  title?: string
  content?: string
  viewCount?: number
  user?: { nickname?: string }
}

const topic = ref<Topic | null>(null)
const loading = ref(true)
const loadError = ref('')

onLoad(async (options) => {
  if (!options?.id) {
    loading.value = false
    return
  }
  try {
    topic.value = await get<Topic>(`/api/topic/${options.id}`)
  } catch (e) {
    loadError.value = e instanceof Error ? e.message : '加载失败'
    uni.showToast({ title: loadError.value, icon: 'none' })
  } finally {
    loading.value = false
  }
})
</script>

<style lang="scss" scoped>
.title {
  margin-bottom: 12rpx;
  font-size: 38rpx;
  font-weight: 600;
  line-height: 1.4;
}

.content {
  display: block;
  margin-top: 32rpx;
  line-height: 1.7;
}
</style>

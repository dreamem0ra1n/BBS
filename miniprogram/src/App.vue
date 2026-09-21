<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'

onLaunch(() => {
  console.info('BBS mini program launched')
})
</script>

<style lang="scss">
page {
  background: #f6f7f9;
  color: #1f2937;
  font-size: 28rpx;
}

.page {
  box-sizing: border-box;
  min-height: 100vh;
  padding: 24rpx;
}

.card {
  box-sizing: border-box;
  margin-bottom: 20rpx;
  padding: 24rpx;
  background: #ffffff;
  border-radius: 12rpx;
  box-shadow: 0 2rpx 12rpx rgba(31, 41, 55, 0.06);
}

.muted {
  color: #6b7280;
  font-size: 24rpx;
}

.empty {
  padding: 80rpx 24rpx;
  color: #6b7280;
  text-align: center;
}
</style>

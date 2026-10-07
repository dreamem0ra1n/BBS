<template>
  <section class="main checkin-page">
    <div class="container checkin-layout">
      <div class="checkin-main">
        <div class="widget summary-card">
          <div class="widget-header">签到中心</div>
          <div class="widget-content summary-content">
            <div class="summary-item">
              <strong>{{ overview.todayTotal }}</strong
              ><span>当天签到总人数</span>
            </div>
            <div class="summary-item">
              <strong>{{ overview.consecutiveDays }}</strong
              ><span>连续签到天数</span>
            </div>
            <div class="summary-item">
              <strong>{{ overview.totalCheckInDays }}</strong
              ><span>累计签到天数</span>
            </div>
            <div class="summary-action">
              <button
                class="button is-primary"
                :disabled="overview.checkedInToday || checkInLoading"
                @click="checkIn"
              >
                {{
                  overview.checkedInToday
                    ? '今日已签到'
                    : checkInLoading
                    ? '签到中…'
                    : '签到'
                }}
              </button>
              <p v-if="message" class="help" :class="messageType">
                {{ message }}
              </p>
            </div>
            <p class="streak-tip">
              你已经连续签到 {{ overview.consecutiveDays }} 天啦！
            </p>
          </div>
        </div>

        <div class="widget calendar-card">
          <div class="widget-header calendar-header">
            <button
              class="button is-small"
              :disabled="!previousMonth"
              @click="changeMonth(-1)"
            >
              上月
            </button>
            <span>{{ overview.month }}</span>
            <button
              class="button is-small"
              :disabled="!nextMonth"
              @click="changeMonth(1)"
            >
              下月
            </button>
          </div>
          <div class="widget-content">
            <div v-if="loading" class="empty-state">日历加载中…</div>
            <div v-else-if="!overview.calendar.length" class="empty-state">
              本月暂无签到数据
            </div>
            <div v-else class="calendar">
              <div
                v-for="label in weekLabels"
                :key="label"
                class="calendar-label"
              >
                {{ label }}
              </div>
              <div
                v-for="blank in firstDay"
                :key="'blank-' + blank"
                class="calendar-day blank"
              />
              <div
                v-for="item in overview.calendar"
                :key="item.date"
                class="calendar-day"
                :class="item.status"
              >
                <span>{{ item.date.slice(-2) }}</span>
                <b v-if="item.status === 'checked'">✓</b>
                <b v-else-if="item.status === 'makeup'">补</b>
                <b v-else-if="item.status === 'missed'">未</b>
              </div>
            </div>
            <div class="calendar-legend">
              <span class="checked">✓ 正常签到</span
              ><span class="makeup">补 补签</span><span>未 漏签</span>
            </div>
          </div>
        </div>
      </div>

      <div class="checkin-side">
        <div class="widget makeup-card">
          <div class="widget-header">补签卡</div>
          <div class="widget-content">
            <p class="makeup-count">
              {{ overview.makeupCards }} <small>张</small>
            </p>
            <button
              class="button is-warning is-fullwidth"
              :disabled="makeupLoading || !overview.makeupCards"
              @click="makeup"
            >
              {{ makeupLoading ? '补签中…' : '自动补签最近漏签日' }}
            </button>
            <p class="help">每累计 7 个有效签到日获得 1 张。</p>
          </div>
        </div>
        <rank-list
          ref="todayRank"
          title="今日排行"
          endpoint="/api/checkin/rank/today"
          today
        />
        <rank-list
          ref="consecutiveRank"
          title="连续签到排行"
          endpoint="/api/checkin/rank/consecutive"
        />
      </div>
    </div>
  </section>
</template>

<script>
import CheckInRank from '~/components/CheckInRank.vue'

const emptyOverview = () => ({
  todayTotal: 0,
  consecutiveDays: 0,
  totalCheckInDays: 0,
  makeupCards: 0,
  checkedInToday: false,
  month: '',
  calendar: [],
})

export default {
  components: {
    RankList: CheckInRank,
  },
  middleware: 'authenticated',
  data() {
    return {
      overview: emptyOverview(),
      loading: false,
      checkInLoading: false,
      makeupLoading: false,
      message: '',
      messageType: '',
      monthDate: null,
    }
  },
  head() {
    return { title: this.$siteTitle('签到中心') }
  },
  computed: {
    weekLabels() {
      return ['日', '一', '二', '三', '四', '五', '六']
    },
    firstDay() {
      return this.overview.month
        ? new Date(this.overview.month + '-01T00:00:00').getDay()
        : 0
    },
    previousMonth() {
      return Boolean(this.monthDate)
    },
    nextMonth() {
      return (
        this.monthDate &&
        this.monthDate <
          new Date(new Date().getFullYear(), new Date().getMonth(), 1)
      )
    },
  },
  mounted() {
    this.loadOverview()
  },
  methods: {
    monthParam() {
      const year = this.monthDate.getFullYear()
      const month = String(this.monthDate.getMonth() + 1).padStart(2, '0')
      return year + '-' + month
    },
    async loadOverview(month) {
      this.loading = true
      try {
        this.overview = {
          ...emptyOverview(),
          ...(await this.$axios.get('/api/checkin/overview', {
            params: month ? { month } : {},
          })),
        }
        this.monthDate = new Date(this.overview.month + '-01T00:00:00')
      } catch (error) {
        this.showMessage(error.message || '签到信息加载失败', 'is-danger')
      } finally {
        this.loading = false
      }
    },
    changeMonth(offset) {
      const next = new Date(this.monthDate)
      next.setMonth(next.getMonth() + offset)
      this.monthDate = next
      this.loadOverview(this.monthParam())
    },
    async checkIn() {
      if (this.checkInLoading || this.overview.checkedInToday) return
      this.checkInLoading = true
      try {
        await this.$axios.post('/api/checkin/checkin')
        this.showMessage('签到成功', 'is-success')
        await this.loadOverview(this.monthParam())
        this.refreshRanks()
      } catch (error) {
        this.showMessage(error.message || '签到失败，请稍后重试', 'is-danger')
      } finally {
        this.checkInLoading = false
      }
    },
    async makeup() {
      if (this.makeupLoading || !this.overview.makeupCards) return
      this.makeupLoading = true
      try {
        const result = await this.$axios.post('/api/checkin/makeup')
        this.showMessage('已补签 ' + result.date, 'is-success')
        await this.loadOverview(this.monthParam())
        this.refreshRanks()
      } catch (error) {
        this.showMessage(error.message || '补签失败，请稍后重试', 'is-danger')
      } finally {
        this.makeupLoading = false
      }
    },
    showMessage(text, type) {
      this.message = text
      this.messageType = type
    },
    refreshRanks() {
      this.$refs.todayRank.refresh()
      this.$refs.consecutiveRank.refresh()
    },
  },
}
</script>

<style lang="scss" scoped>
.checkin-layout {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}
.checkin-main {
  flex: 1;
  min-width: 0;
}
.checkin-side {
  width: 320px;
  min-width: 0;
}
.summary-content {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 20px;
}
.summary-item {
  display: flex;
  flex-direction: column;
  min-width: 90px;
  color: var(--text-color3);
}
.summary-item strong {
  font-size: 1.5rem;
  color: var(--text-color);
}
.summary-action {
  margin-left: auto;
  text-align: center;
}
.streak-tip {
  width: 100%;
  margin: 0;
  color: var(--text-color3);
}
.calendar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.calendar {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 5px;
}
.calendar-label,
.calendar-day {
  min-height: 42px;
  text-align: center;
}
.calendar-label {
  color: var(--text-color3);
  font-size: 0.85rem;
}
.calendar-day {
  display: flex;
  flex-direction: column;
  justify-content: center;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-color3);
}
.calendar-day b {
  color: var(--text-color3);
  font-size: 0.75rem;
}
.calendar-day.checked {
  background: rgba(72, 199, 142, 0.18);
  border-color: #48c78e;
}
.calendar-day.checked b {
  color: #259b6c;
}
.calendar-day.makeup {
  background: rgba(255, 221, 87, 0.2);
  border-color: #ffdd57;
}
.calendar-day.makeup b {
  color: #9c7b00;
}
.calendar-day.future,
.calendar-day.before_register,
.calendar-day.blank {
  opacity: 0.35;
}
.calendar-day.future b,
.calendar-day.before_register b {
  display: none;
}
.calendar-legend {
  display: flex;
  gap: 15px;
  margin-top: 12px;
  font-size: 0.8rem;
  color: var(--text-color3);
}
.calendar-legend .checked {
  color: #259b6c;
}
.calendar-legend .makeup {
  color: #9c7b00;
}
.makeup-count {
  font-size: 2rem;
  margin-bottom: 12px;
}
.makeup-count small {
  font-size: 1rem;
}
.empty-state {
  color: var(--text-color3);
  text-align: center;
  padding: 12px 0;
  font-size: 0.85rem;
}
.help.is-danger {
  color: #f14668;
}
.help.is-success {
  color: #48c78e;
}
@media screen and (max-width: 768px) {
  .checkin-layout {
    display: block;
  }
  .checkin-side {
    width: auto;
  }
  .summary-content {
    gap: 12px;
  }
  .summary-action {
    width: 100%;
    margin-left: 0;
  }
  .calendar-day {
    min-height: 38px;
  }
}
</style>

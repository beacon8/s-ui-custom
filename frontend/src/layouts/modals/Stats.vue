<template>
  <a-modal
    :open="visible"
    :width="800"
    :footer="null"
    @cancel="$emit('close')"
  >
    <template #title>
      <a-flex
        justify="space-between"
        align="center"
        style="padding-inline-end: 24px;"
      >
        <div>
          <div>{{ $t('stats.graphTitle') }}</div>
          <div style="font-size: 12px; font-weight: normal; opacity: 0.65;">
            {{ $t('objects.' + resource) + " : " + tag }}
          </div>
        </div>
        <a-space>
          <a-space size="small">
            <a-switch
              v-model:checked="autoRefresh"
              size="small"
            />
            <span style="font-size: 13px; font-weight: normal;">{{ $t('stats.autoRefresh') }}</span>
          </a-space>
          <a-tooltip :title="$t('actions.update')">
            <a-button
              type="text"
              shape="circle"
              :aria-label="$t('actions.update')"
              @click="loadData"
            >
              <template #icon>
                <AppIcon
                  name="refresh"
                  :spin="loading"
                />
              </template>
            </a-button>
          </a-tooltip>
        </a-space>
      </a-flex>
    </template>

    <div style="text-align: center;">
      <a-radio-group
        v-model:value="limit"
        option-type="button"
        button-style="solid"
        size="small"
        :disabled="loading"
        @change="selectPreset(limit)"
      >
        <a-radio-button
          v-for="p in periods"
          :key="p.value"
          :value="p.value"
        >
          {{ p.title }}
        </a-radio-button>
        <a-radio-button :value="0">
          <AppIcon name="calendar-range" />
        </a-radio-button>
      </a-radio-group>

      <a-row
        v-if="limit === 0"
        :gutter="[12, 12]"
        align="middle"
        justify="center"
        style="margin-top: 12px; margin-bottom: 4px;"
      >
        <a-col
          :xs="24"
          :sm="10"
        >
          <DatePick
            :expiry="rangeStart"
            :label="$t('stats.from')"
            input-id="statsFrom"
            @submit="setRangeStart"
          />
        </a-col>
        <a-col
          :xs="24"
          :sm="10"
        >
          <DatePick
            :expiry="rangeEnd"
            :label="$t('stats.to')"
            input-id="statsTo"
            @submit="setRangeEnd"
          />
        </a-col>
      </a-row>

      <a-space
        v-if="loaded && !loading"
        wrap
        style="margin-top: 10px; justify-content: center; width: 100%;"
      >
        <a-tag color="orange">
          {{ $t('stats.upload') }}: {{ fmt(totalUp) }}
        </a-tag>
        <a-tag color="green">
          {{ $t('stats.download') }}: {{ fmt(totalDown) }}
        </a-tag>
        <a-tag color="blue">
          {{ $t('main.stats.totalUsage') }}: {{ fmt(totalUp + totalDown) }}
        </a-tag>
      </a-space>

      <div
        id="container"
        style="height: 400px; margin-top: 12px;"
      >
        <a-skeleton
          v-if="loading"
          active
          :paragraph="{ rows: 8 }"
        />
        <template v-else>
          <a-alert
            v-if="alert"
            :message="$t('noData')"
            type="warning"
            show-icon
          />
          <LineChart
            v-if="loaded"
            :key="themeName"
            :data="usage"
            :options="<any>options"
          />
        </template>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts">
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'
import { HumanReadable } from '@/plugins/utils'
import { isDarkTheme, currentThemeName } from '@/composables/useThemeSwitcher'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import type { ChartData } from 'chart.js'
import { ref } from 'vue'
import { Line } from 'vue-chartjs'
import DatePick from '@/components/DateTime.vue'
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)
ChartJS.defaults.font.family = 'Vazirmatn'

// api/stats answers with the window it covers and a sparse map of bucket index
// to [upload, download]; buckets with no traffic are simply absent.
interface StatsResponse {
  stats: Record<string, [number, number]>
  startTime: number
  bucketSpan: number
  numBuckets?: number
}

// The query api/stats takes: the resource to chart, plus either a preset span
// or an explicit range.
interface StatsParams {
  resource: string
  tag: string
  limit?: number
  start?: number
  end?: number
}

export default {
  components: {
    LineChart: Line,
    DatePick
  },
  props: {
    visible: { type: Boolean, required: true },
    resource: { type: String, required: true },
    tag: { type: String, required: true },
  },
  emits: ['close'],
  data() {
    return {
      loading: false,
      loaded: false,
      alert: false,
      intervalId: <ReturnType<typeof setInterval> | 0>0,
      autoRefresh: localStorage.getItem('statsAutoRefresh') === 'true',
      // Sum of the loaded window, shown as chips above the chart (#1219)
      totalUp: 0,
      totalDown: 0,
      limit: 1,
      rangeStart: 0,
      rangeEnd: 0,
      periods: [
        { value: 1, title: i18n.global.n(1) + i18n.global.t('date.h')},
        { value: 6, title: i18n.global.n(6) + i18n.global.t('date.h')},
        { value: 12, title: i18n.global.n(12) + i18n.global.t('date.h')},
        { value: 24, title: i18n.global.n(1) + i18n.global.t('date.d')},
        { value: 48, title: i18n.global.n(2) + i18n.global.t('date.d')},
        { value: 240, title: i18n.global.n(10) + i18n.global.t('date.d')},
        { value: 480, title: i18n.global.n(20) + i18n.global.t('date.d')},
        { value: 720, title: i18n.global.n(30) + i18n.global.t('date.d')},
      ],
      baseOptions: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          intersect: false,
          mode: 'index',
        },
        elements: {
          point: { pointStyle: 'circle', radius: 0, hitRadius: 10 },
          line: { tension: 0.3, borderWidth: 2 },
        },
        plugins: {
          tooltip: {
            callbacks: {
              footer: (items: { raw: number | null }[]) => {
                return HumanReadable.sizeFormat(items.reduce((acc, c) => acc + (c.raw ?? 0), 0))
              }
            }
          }
        },
        scales: {
          y: {
            grid: {
              color: '#777777',
            },
            beginAtZero: true,
            ticks: {
              callback: function(label: string | number) {
                return label == 0 ? 0 : HumanReadable.sizeFormat(Number(label),0)
              },
              count: 10
            }
          }
        }
      },
      usage: ref(<ChartData<'line'>>{}),
    }
  },
  computed: {
    themeName(): string {
      return currentThemeName.value
    },
    options() {
      const dark = isDarkTheme.value
      const text = dark ? 'rgba(255, 255, 255, 0.85)' : 'rgba(0, 0, 0, 0.88)'
      const gridY = dark ? '#333333' : '#88888850'
      const gridX = dark ? '#333333' : '#88888850'
      return {
        ...this.baseOptions,
        plugins: {
          ...this.baseOptions.plugins,
          legend: { labels: { color: text } },
        },
        scales: {
          y: {
            ...this.baseOptions.scales.y,
            grid: { color: gridY },
            ticks: { ...this.baseOptions.scales.y.ticks, color: text },
          },
          x: {
            grid: { color: gridX },
            ticks: { color: text },
          },
        },
      }
    },
  },
  watch: {
    visible(v) {
      if (v) {
        this.limit = 1
        this.rangeStart = 0
        this.rangeEnd = 0
        this.loadData()
        if (this.autoRefresh) this.startAutoRefresh()
      } else {
        this.loaded = false
        this.alert = false
        this.usage.labels = []
        if (this.usage.datasets) {
          this.usage.datasets[0].data = []
          this.usage.datasets[1].data = []
        }
        this.stopAutoRefresh()
      }
    },
    autoRefresh(v) {
      localStorage.setItem('statsAutoRefresh', v ? 'true' : 'false')
      // Live refresh only makes sense on presets, not a fixed custom range
      if (v && this.limit !== 0) this.startAutoRefresh()
      else this.stopAutoRefresh()
    }
  },
  methods: {
    async loadData() {
      this.loading = true
      let params: StatsParams = { resource: this.resource, tag: this.tag }
      let span = this.limit
      if (this.limit === 0) {
        if (!this.rangeStart || !this.rangeEnd || this.rangeEnd <= this.rangeStart) {
          this.alert = true
          this.loaded = false
          this.loading = false
          return
        }
        params.start = this.rangeStart
        params.end = this.rangeEnd
        span = Math.max(1, Math.round((this.rangeEnd - this.rangeStart) / 3600))
      } else {
        params.limit = this.limit
      }
      const data = await HttpUtils.get<StatsResponse>('api/stats', params)
      if (data.success && data.obj.stats) {
        const {stats, bucketSpan, startTime} = data.obj
        const count = data.obj.numBuckets ?? 360
        const l = String(i18n.global.locale) == 'fa' ? "fa-IR" : "en-US"
        const labels = <string[]>[]
        const uplinkData = <(number|null)[]>[]
        const downlinkData = <(number|null)[]>[]
        this.totalUp = 0
        this.totalDown = 0
        for (let i = 0; i<count; i++) {
          const step = startTime + (i*bucketSpan)
          labels.push(this.genLable(step*1000,l,span))
          if (!stats[i]) {
            uplinkData.push(null)
            downlinkData.push(null)
          } else {
            uplinkData.push(stats[i][0])
            downlinkData.push(stats[i][1])
            this.totalUp += stats[i][0] ?? 0
            this.totalDown += stats[i][1] ?? 0
          }
        }
        this.usage = {
          labels: labels,
          datasets: [
            {
              label: i18n.global.t('stats.upload'),
              backgroundColor: 'rgba(255, 165, 0, 0.4)',
              borderColor: 'rgba(255, 165, 0)',
              fill: true,
              data: uplinkData
            },
            {
              label: i18n.global.t('stats.download'),
              backgroundColor: 'rgba(0, 128, 0, 0.2)',
              borderColor: 'rgba(0, 128, 0)',
              fill: true,
              data: downlinkData
            }
          ],
        }
        this.loaded = true
        this.alert = false
      } else {
        this.alert = true
        this.loaded = false
      }
      this.loading = false
    },
    startAutoRefresh() {
      if (!this.intervalId) {
        this.intervalId = setInterval(() => { this.loadData() }, 10000)
      }
    },
    stopAutoRefresh() {
      if (this.intervalId) {
        clearInterval(this.intervalId)
        this.intervalId = 0
      }
    },
    selectPreset(v:number) {
      if (v === 0) {
        // Custom range mode: default to the last 24h, never live-refresh
        if (!this.rangeStart || !this.rangeEnd) {
          const now = Math.floor(Date.now() / 1000)
          this.rangeEnd = now
          this.rangeStart = now - 86400
        }
        this.stopAutoRefresh()
      } else if (this.autoRefresh) {
        // Back to a preset: resume live mode only when the user opted in
        this.startAutoRefresh()
      }
      this.loadData()
    },
    setRangeStart(v:number) {
      this.rangeStart = v
      this.loadData()
    },
    setRangeEnd(v:number) {
      this.rangeEnd = v
      this.loadData()
    },
    fmt(v: number) {
      return HumanReadable.sizeFormat(v)
    },
    genLable(step:number, locale: string, limit: number) {
      return new Date(step).toLocaleString(locale,{
        month: limit < 480 ? undefined : '2-digit',
        day: limit < 24 ? undefined : '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
      })
    },
  }
}
</script>

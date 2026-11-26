<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { HumanReadable } from '@/plugins/utils'
import { isDarkTheme } from '@/composables/useThemeSwitcher'

// One resource reading. Memory, disk and swap report a pair; cpu is a percent.
interface Usage {
  current: number
  total: number
}

interface TilesData {
  cpu?: number
  mem?: Usage
  dsk?: Usage
  swp?: Usage
}

type GaugeState = 'usage' | 'empty' | 'disabled'

interface GaugeView {
  state: GaugeState
  percent: number
  value?: string
  capacity?: string
  color: string
}

const props = defineProps<{
  tilesData: TilesData
  type: string
}>()

const { t } = useI18n()

const isNumber = (value: unknown): value is number =>
  typeof value === 'number' && Number.isFinite(value)

const percentOf = (current: number, total: number): number => {
  if (!Number.isFinite(current) || !Number.isFinite(total) || total <= 0) return 0
  return Math.min(100, Math.max(0, Math.ceil(current * 100 / total)))
}

const colorOf = (percent: number): string => {
  if (percent > 90) return '#ff4d4f'
  if (percent > 70) return '#faad14'
  return '#1677ff'
}

const usageFor = (usage?: Usage): GaugeView => {
  if (!usage || !isNumber(usage.current) || !isNumber(usage.total)) {
    return { state: 'empty', percent: 0, color: colorOf(0) }
  }
  if (usage.total <= 0) {
    if (props.type === 'g-swp') {
      return { state: 'disabled', percent: 0, color: colorOf(0) }
    }
    return { state: 'empty', percent: 0, color: colorOf(0) }
  }
  const percent = percentOf(usage.current, usage.total)
  return {
    state: 'usage',
    percent,
    value: `${percent}%`,
    capacity: `${HumanReadable.sizeFormat(usage.current, 0)} / ${HumanReadable.sizeFormat(usage.total, 0)}`,
    color: colorOf(percent),
  }
}

const data = computed((): GaugeView => {
  switch (props.type) {
    case 'g-cpu': {
      if (!isNumber(props.tilesData.cpu)) {
        return { state: 'empty', percent: 0, color: colorOf(0) }
      }
      const percent = Math.min(100, Math.max(0, Math.ceil(props.tilesData.cpu)))
      return {
        state: 'usage',
        percent,
        value: `${percent}%`,
        color: colorOf(percent),
      }
    }
    case 'g-mem':
      return usageFor(props.tilesData.mem)
    case 'g-dsk':
      return usageFor(props.tilesData.dsk)
    case 'g-swp':
      return usageFor(props.tilesData.swp)
    default:
      return { state: 'empty', percent: 0, color: colorOf(0) }
  }
})

const trackColor = computed(() => isDarkTheme.value ? 'rgba(255, 255, 255, 0.16)' : 'rgba(0, 0, 0, 0.06)')
const emptyText = computed(() => data.value.state === 'disabled' ? t('main.gauge.disabled') : t('main.gauge.noData'))
</script>

<template>
  <div
    class="gauge"
    :class="{ 'gauge--empty': data.state !== 'usage' }"
  >
    <div class="gauge__row">
      <span class="gauge__value">
        {{ data.state === 'usage' ? data.value : emptyText }}
      </span>
      <span
        v-if="data.capacity"
        class="gauge__capacity"
      >
        {{ data.capacity }}
      </span>
    </div>
    <a-progress
      :percent="data.percent"
      :show-info="false"
      :stroke-color="data.color"
      :trail-color="trackColor"
      :stroke-width="8"
      class="gauge__progress"
    />
  </div>
</template>

<style scoped>
.gauge {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 8px;
  min-height: 64px;
  width: 100%;
  min-width: 0;
}

.gauge__row {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 2px 12px;
  min-width: 0;
}

.gauge__value {
  flex: none;
  font-size: 24px;
  font-weight: 600;
  line-height: 1.1;
  letter-spacing: -0.02em;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.gauge__capacity {
  margin-left: auto;
  min-width: 0;
  font-size: 12px;
  line-height: 1.3;
  opacity: 0.65;
  text-align: right;
  overflow-wrap: anywhere;
  font-variant-numeric: tabular-nums;
}

.gauge--empty .gauge__value {
  font-size: 14px;
  font-weight: 500;
  letter-spacing: 0;
  opacity: 0.55;
}

.gauge__progress {
  margin: 0;
}

.gauge__progress :deep(.ant-progress-outer) {
  margin-inline-end: 0;
}
</style>

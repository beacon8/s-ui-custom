<template>
  <a-modal
    :open="visible"
    :width="750"
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
          <div>{{ $t('sessions.title') }}</div>
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

    <div style="max-height: 65vh; overflow-y: auto;">
      <a-alert
        v-if="!loading && sessions.length == 0"
        :message="$t('sessions.empty')"
        type="info"
        show-icon
      />
      <a-table
        v-else
        :data-source="sessions"
        :columns="headers"
        :loading="loading"
        :pagination="sessions.length > 10 ? { pageSize: 10, showSizeChanger: false } : false"
        size="small"
        row-key="id"
        :scroll="{ x: 'max-content' }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'user'">
            <a-tooltip :title="record.user">
              <span class="session-truncate session-tag">{{ record.user }}</span>
            </a-tooltip>
          </template>
          <template v-else-if="column.key === 'inbound'">
            <a-tooltip :title="record.inbound">
              <span class="session-truncate session-tag">{{ record.inbound }}</span>
            </a-tooltip>
          </template>
          <template v-else-if="column.key === 'source'">
            <a-tooltip :title="record.source">
              <span class="session-truncate session-addr">{{ record.source }}</span>
            </a-tooltip>
          </template>
          <template v-else-if="column.key === 'destination'">
            <a-tooltip>
              <template #title>
                <template v-if="record.domain && record.domain != record.destination.split(':')[0]">
                  {{ record.domain }}<br>
                </template>
                {{ record.destination }}<br>
                {{ record.network }}
              </template>
              <span class="session-truncate session-addr">{{ record.domain || record.destination }}</span>
            </a-tooltip>
          </template>
          <template v-else-if="column.key === 'outbound'">
            <a-tooltip
              v-if="record.rule"
              :title="$t('objects.rule') + ': ' + record.rule"
            >
              <span class="session-truncate session-tag">{{ record.outbound }}</span>
            </a-tooltip>
            <span
              v-else
              class="session-truncate session-tag"
            >{{ record.outbound }}</span>
          </template>
          <template v-else-if="column.key === 'up'">
            <a-tooltip>
              <template #title>
                {{ $t('stats.upload') }}: {{ fmt(record.up) }}<br>
                {{ $t('stats.download') }}: {{ fmt(record.down) }}<br>
                {{ $t('sessions.duration') }}: {{ duration(record.createdAt) }}
              </template>
              <span>{{ fmt(record.up + record.down) }}</span>
            </a-tooltip>
          </template>
        </template>
      </a-table>
    </div>

    <template v-if="resource == 'user'">
      <a-divider style="margin: 12px 0;" />
      <a-flex justify="start">
        <a-button
          danger
          :loading="closing"
          :disabled="sessions.length == 0"
          @click="closeSessions"
        >
          <template #icon>
            <AppIcon name="close-network" />
          </template>
          {{ $t('sessions.closeAll') }}
        </a-button>
      </a-flex>
    </template>
  </a-modal>
</template>

<script lang="ts" setup>
import HttpUtils from '@/plugins/httputil'
import { HumanReadable } from '@/plugins/utils'
import { i18n } from '@/locales'
import { computed, ref, watch } from 'vue'

// One live connection as api/sessions reports it, mirroring the core's
// SessionInfo snapshot. Everything but the id and the counters is omitted by
// the backend when empty.
interface Session {
  id: string
  inbound?: string
  user?: string
  outbound?: string
  network?: string
  source?: string
  // Omitted by the backend only when empty, which a live session never is.
  destination: string
  domain?: string
  rule?: string
  createdAt: number
  up: number
  down: number
}

const props = defineProps<{ visible: boolean, resource: string, tag: string }>()
defineEmits(['close'])

const loading = ref(false)
const closing = ref(false)
const sessions = ref(<Session[]>[])
const autoRefresh = ref(localStorage.getItem('sessionsAutoRefresh') === 'true')
let intervalId: ReturnType<typeof setInterval> | 0 = 0

const headers = computed(() => {
  // The resource the list was opened from is already in the subtitle, so its
  // column is dropped and the client takes that place.
  const columns = [
    { title: i18n.global.t('sessions.source'), dataIndex: 'source', key: 'source' },
    { title: i18n.global.t('sessions.destination'), dataIndex: 'destination', key: 'destination' },
    {
      title: i18n.global.t('stats.usage'),
      dataIndex: 'up',
      key: 'up',
      width: 90,
      sorter: (a: Session, b: Session) => (a.up + a.down) - (b.up + b.down),
    },
  ]
  const inbound = { title: i18n.global.t('objects.inbound'), dataIndex: 'inbound', key: 'inbound' }
  const outbound = { title: i18n.global.t('objects.outbound'), dataIndex: 'outbound', key: 'outbound' }
  const client = { title: i18n.global.t('objects.client'), dataIndex: 'user', key: 'user' }
  switch (props.resource) {
    case 'inbound':
      return [client, ...columns.slice(0, 2), outbound, columns[2]]
    case 'outbound':
      return [client, ...columns.slice(0, 2), inbound, columns[2]]
    case 'endpoint':
      // An endpoint can be on either side of a connection, so both tags stay:
      // they are what tells the two cases apart.
      return [client, inbound, ...columns.slice(0, 2), outbound, columns[2]]
    default:
      return [inbound, ...columns.slice(0, 2), outbound, columns[2]]
  }
})

const loadData = async () => {
  loading.value = true
  const msg = await HttpUtils.get<Session[]>('api/sessions', { resource: props.resource, tag: props.tag })
  sessions.value = msg.success ? (msg.obj ?? []) : []
  loading.value = false
}

const closeSessions = async () => {
  closing.value = true
  const msg = await HttpUtils.post('api/closeSessions', { u: props.tag })
  // Empty the list now rather than waiting for the next poll, so it matches
  // what the operator just did even with auto refresh off.
  if (msg.success) sessions.value = []
  closing.value = false
}

const fmt = (v: number) => HumanReadable.sizeFormat(v)

const duration = (createdAt: number) => {
  const seconds = Math.max(0, Math.floor(Date.now() / 1000) - createdAt)
  if (seconds < 60) return i18n.global.n(seconds) + i18n.global.t('date.s')
  if (seconds < 3600) return i18n.global.n(Math.floor(seconds / 60)) + i18n.global.t('date.m')
  return i18n.global.n(Math.floor(seconds / 3600)) + i18n.global.t('date.h')
}

const startAutoRefresh = () => {
  if (!intervalId) intervalId = setInterval(() => { loadData() }, 5000)
}
const stopAutoRefresh = () => {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = 0
  }
}

watch(() => props.visible, (v) => {
  if (v) {
    loadData()
    if (autoRefresh.value) startAutoRefresh()
  } else {
    sessions.value = []
    stopAutoRefresh()
  }
})

watch(autoRefresh, (v) => {
  localStorage.setItem('sessionsAutoRefresh', v ? 'true' : 'false')
  if (v && props.visible) startAutoRefresh()
  else stopAutoRefresh()
})
</script>

<style scoped>
/* Addresses and tags are cut to keep the row from growing; the full value is
   one hover away. */
.session-truncate {
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}
.session-addr {
  max-width: 130px;
}
.session-tag {
  max-width: 90px;
}
</style>

<template>
  <a-modal
    :open="visible"
    :title="$t('client.ipListTitle') + (user ? ' · ' + user : '')"
    :width="640"
    :footer="null"
    @cancel="$emit('close')"
  >
    <a-table
      :columns="headers"
      :data-source="entries"
      :loading="loading"
      :pagination="false"
      row-key="ip"
      size="small"
      :scroll="{ x: 'max-content' }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'ip'">
          <span style="font-variant-numeric: tabular-nums;">{{ record.ip }}</span>
        </template>
        <template v-else-if="column.key === 'since'">
          <a-tooltip :title="dateText(record.since)">
            <span>{{ dateText(record.since) }}</span>
          </a-tooltip>
        </template>
      </template>
    </a-table>
    <div class="ip-list-hint">
      {{ $t('client.ipListHint', { limit: limitText }) }}
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'

interface IpEntry {
  ip: string
  sessions: number
  since: number
}

interface IpSummary {
  count: number
  limit: number
  ips: IpEntry[]
}

const props = defineProps<{ visible: boolean, user: string }>()
defineEmits(['close'])

const entries = ref<IpEntry[]>([])
const limit = ref(0)
const loading = ref(false)

const limitText = computed(() => limit.value > 0 ? String(limit.value) : '∞')

const headers = computed(() => [
  { title: i18n.global.t('client.ips'), dataIndex: 'ip', key: 'ip' },
  { title: i18n.global.t('client.ipSessions'), dataIndex: 'sessions', key: 'sessions' },
  { title: i18n.global.t('client.ipSince'), dataIndex: 'since', key: 'since' },
])

const dateText = (since: number): string => since > 0 ? new Date(since * 1000).toLocaleString() : '-'

const load = async () => {
  loading.value = true
  const msg = await HttpUtils.get<IpSummary>('api/userIps', { user: props.user })
  loading.value = false
  if (msg.success && msg.obj) {
    entries.value = msg.obj.ips ?? []
    limit.value = msg.obj.limit ?? 0
  } else {
    entries.value = []
    limit.value = 0
  }
}

watch(() => props.visible, (visible) => {
  if (visible) {
    void load()
  }
})
</script>

<style scoped>
.ip-list-hint {
  margin-top: 12px;
  opacity: 0.65;
  font-size: 12px;
}
</style>

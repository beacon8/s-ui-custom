<template>
  <a-modal
    :open="visible"
    :width="400"
    :footer="null"
    @cancel="$emit('update:visible', false)"
  >
    <template #title>
      <a-flex
        justify="space-between"
        align="center"
        style="padding-inline-end: 24px;"
      >
        <span>{{ $t('main.stats.title') }}</span>
        <a-tooltip :title="$t('actions.update')">
          <a-button
            type="text"
            shape="circle"
            :aria-label="$t('actions.update')"
            @click="refresh"
          >
            <template #icon>
              <AppIcon
                name="refresh"
                :spin="loading"
              />
            </template>
          </a-button>
        </a-tooltip>
      </a-flex>
    </template>

    <a-table
      :data-source="tableRows"
      :columns="columns"
      :loading="loading"
      :pagination="false"
      :show-header="false"
      size="small"
      row-key="key"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'icon'">
          <AppIcon
            :name="record.icon"
            :color="record.color || undefined"
          />
        </template>
        <template v-else-if="column.key === 'value'">
          <span style="direction: ltr; display: inline-block;">
            {{ record.value }}
          </span>
        </template>
      </template>
    </a-table>
  </a-modal>
</template>

<script lang="ts">
import { computed, ref, watch } from 'vue'
import HttpUtils from '@/plugins/httputil'
import { HumanReadable } from '@/plugins/utils'
import { i18n } from '@/locales'

export default {
  props: {
    visible: { type: Boolean, default: false },
  },
  emits: ['update:visible'],
  setup(props) {
    const loading = ref(false)
    // api/status?r=db answers either with the counts directly or wrapped in a
    // `db` key, so the response type carries both shapes.
    interface DbUsage {
      clients?: number
      inbounds?: number
      outbounds?: number
      services?: number
      endpoints?: number
      clientUp?: number
      clientDown?: number
    }

    const info = ref<DbUsage>({})

    const clientUp = computed(() => HumanReadable.sizeFormat(info.value.clientUp ?? 0))
    const clientDown = computed(() => HumanReadable.sizeFormat(info.value.clientDown ?? 0))
    const totalUsage = computed(() => {
      const up = info.value.clientUp ?? 0
      const down = info.value.clientDown ?? 0
      return HumanReadable.sizeFormat(up + down)
    })

    const columns = [
      { key: 'icon', width: 40 },
      { key: 'label', dataIndex: 'label' },
      { key: 'value', dataIndex: 'value', align: 'right' as const },
    ]

    const tableRows = computed(() => {
      const t = (key: string) => i18n.global.t(key)
      return [
        { key: 'clients', icon: 'account-multiple', label: t('pages.clients'), value: info.value.clients ?? 0, color: undefined },
        { key: 'inbounds', icon: 'cloud-download', label: t('pages.inbounds'), value: info.value.inbounds ?? 0, color: undefined },
        { key: 'outbounds', icon: 'cloud-upload', label: t('pages.outbounds'), value: info.value.outbounds ?? 0, color: undefined },
        { key: 'services', icon: 'server', label: t('pages.services'), value: info.value.services ?? 0, color: undefined },
        { key: 'endpoints', icon: 'cloud-tags', label: t('pages.endpoints'), value: info.value.endpoints ?? 0, color: undefined },
        { key: 'clientUp', icon: 'cloud-upload', label: t('stats.upload'), value: clientUp.value, color: 'orange' },
        { key: 'clientDown', icon: 'cloud-download', label: t('stats.download'), value: clientDown.value, color: 'success' },
        { key: 'totalUsage', icon: 'chart-box', label: t('main.stats.totalUsage'), value: totalUsage.value, color: 'primary' },
      ]
    })

    const refresh = async () => {
      loading.value = true
      const data = await HttpUtils.get<DbUsage & { db?: DbUsage }>('api/status', { r: 'db' })
      if (data.success && data.obj) {
        info.value = data.obj.db ?? data.obj
      }
      loading.value = false
    }

    watch(() => props.visible, (v) => {
      if (v) refresh()
    })

    return {
      loading,
      info,
      clientUp,
      clientDown,
      totalUsage,
      columns,
      tableRows,
      refresh,
    }
  },
}
</script>

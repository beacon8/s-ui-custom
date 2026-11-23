<template>
  <a-modal
    :open="visible"
    :width="800"
    :title="$t('admin.changes')"
    :footer="null"
    @cancel="$emit('close')"
  >
    <a-form layout="vertical">
      <a-row
        :gutter="[12, 12]"
        align="bottom"
        style="margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="8"
          :md="6"
        >
          <a-form-item
            :label="$t('admin.actor')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="user"
              style="width: 100%"
              @change="loadData"
            >
              <a-select-option
                v-for="item in ['', 'DepleteJob', ...admins]"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="8"
          :md="6"
        >
          <a-form-item
            :label="$t('admin.key')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="key"
              style="width: 100%"
              @change="loadData"
            >
              <a-select-option
                v-for="item in ['', 'inbounds', 'outbounds', 'clients', 'route', 'tls', 'experimental']"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          :xs="12"
          :sm="8"
          :md="6"
        >
          <a-form-item
            :label="$t('count')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="chngCount"
              style="width: 100%"
              @change="loadData"
            >
              <a-select-option
                v-for="item in [10, 20, 30, 50, 100]"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col flex="none">
          <a-button
            :aria-label="$t('actions.update')"
            type="default"
            :loading="loading"
            @click="loadData"
          >
            <template #icon>
              <AppIcon name="refresh" />
            </template>
          </a-button>
        </a-col>
      </a-row>
    </a-form>
    <a-table
      :columns="changesHeaders"
      :data-source="changes"
      row-key="id"
      size="small"
      :loading="loading"
      :pagination="{ pageSize: 10 }"
    >
      <template #bodyCell="{ column, text }">
        <template v-if="column.key === 'dateTime'">
          <a-tag
            :bordered="false"
            dir="ltr"
          >
            {{ dateFormatted(text) }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-tag>
            {{ $t('actions.' + text) }}
          </a-tag>
        </template>
      </template>
      <template #expandedRowRender="{ record }">
        <a-card
          v-if="(record.index ?? 0) > 0"
          size="small"
          dir="ltr"
          style="margin-bottom: 8px;"
        >
          Index: {{ record.index }}
        </a-card>
        <a-card
          size="small"
          dir="ltr"
        >
          <pre style="margin: 0; overflow-x: auto;">{{ record.obj }}</pre>
        </a-card>
      </template>
    </a-table>
  </a-modal>
</template>

<script lang="ts">
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'
import type { PropType } from 'vue'

// One row of api/changes, mirroring the backend's Changes model. The expanded
// row also shows a position inside the changed list, which only some records
// carry.
interface Change {
  id: number
  dateTime: number
  actor: string
  key: string
  action: string
  obj?: unknown
  index?: number
}

export default {
  props: {
    admins: { type: Array as PropType<string[]>, required: true },
    actor: { type: String, required: true },
    visible: { type: Boolean, required: true },
  },
  emits: ['close'],
  data() {
    return {
      loading: false,
      changes: <Change[]>[],
      user: '',
      key: '',
      chngCount: 10,
      expanded: [],
      changesHeaders: [
        { title: 'ID', key: 'id', dataIndex: 'id' },
        { title: i18n.global.t('admin.date') + '-' + i18n.global.t('admin.time'), key: 'dateTime', dataIndex: 'dateTime' },
        { title: i18n.global.t('admin.actor'), key: 'Actor', dataIndex: 'Actor' },
        { title: i18n.global.t('admin.key'), key: 'key', dataIndex: 'key' },
        { title: i18n.global.t('admin.action'), key: 'action', dataIndex: 'action' },
      ],
    }
  },
  computed: {
    locale() {
      const l = i18n.global.locale.value
      switch (l) {
        case "zhHans":
          return "zh-cn"
        case "zhHant":
          return "zh-tw"
        default:
          return l
      }
    },
  },
  watch: {
    visible(newValue) {
      this.changes = []
      this.user = this.$props.actor
      this.key = ''
      this.chngCount = 10
      if (newValue) {
        this.loadData()
      }
    },
  },
  methods: {
    async loadData() {
      this.loading = true
      const data = await HttpUtils.get<Change[]>('api/changes',{ a: this.user, k: this.key, c: this.chngCount })
      if (data.success) {
        this.changes = data.obj?? []
        this.loading = false
      }
    },
    dateFormatted(dt: number): string {
      const date = new Date(dt*1000)
      return date.toLocaleString(this.locale)
    },
  },
}
</script>

<template>
  <OutboundVue
    :id="modal.id"
    v-model="modal.visible"
    :visible="modal.visible"
    :data="modal.data"
    :tags="outboundTags"
    @close="closeModal"
  />
  <OutboundReplaceIp
    :visible="replaceModal"
    :targets="replaceTargets"
    @close="closeReplace"
    @done="replaceDone = true"
  />
  <OutboundBulk
    v-model="bulkModal.visible"
    :visible="bulkModal.visible"
    :outbound-tags="outboundTags"
    @close="closeBulkModal"
  />
  <Stats
    v-model="stats.visible"
    :visible="stats.visible"
    :resource="stats.resource"
    :tag="stats.tag"
    @close="closeStats"
  />
  <Sessions
    v-model="sessions.visible"
    :visible="sessions.visible"
    resource="outbound"
    :tag="sessions.tag"
    @close="closeSessions"
  />
  <a-row
    :gutter="[12, 12]"
    justify="start"
    align="middle"
    style="margin-bottom: 12px;"
  >
    <a-col flex="none">
      <a-button
        type="primary"
        @click="showModal(0)"
      >
        {{ $t('actions.add') }}
      </a-button>
    </a-col>
    <a-col flex="none">
      <a-button
        type="primary"
        @click="showBulkModal"
      >
        {{ $t('actions.addbulk') }}
      </a-button>
    </a-col>
    <a-col flex="none">
      <a-button
        danger
        :disabled="selectedTags.length === 0"
        @click="replaceModal = true"
      >
        {{ $t('outboundReplace.button') }}<template v-if="selectedTags.length"> ({{ selectedTags.length }})</template>
      </a-button>
    </a-col>
    <a-col flex="none">
      <a-button
        type="default"
        :loading="testingAll"
        :disabled="testingAll || outbounds.length === 0"
        @click="checkAllOutbounds"
      >
        <template #icon>
          <AppIcon name="speedometer" />
        </template>
        {{ $t('actions.testAll') || 'Test all' }}
      </a-button>
    </a-col>
    <UrlTestTarget />
  </a-row>
  <a-row
    :gutter="[12, 12]"
    align="middle"
    style="margin-bottom: 12px;"
  >
    <a-col
      :xs="24"
      :sm="16"
      :md="10"
      :lg="8"
    >
      <a-input
        v-model:value="searchQuery"
        allow-clear
        :placeholder="$t('out.searchPlaceholder')"
      >
        <template #prefix>
          <AppIcon name="magnify" />
        </template>
      </a-input>
    </a-col>
    <a-col
      :xs="24"
      :sm="8"
      :md="6"
      :lg="4"
    >
      <span class="search-count">{{ filteredOutbounds.length }} / {{ outbounds.length }}</span>
    </a-col>
  </a-row>
  <a-row :gutter="[12, 12]">
    <a-col :span="24">
      <a-table
        :columns="headers"
        :data-source="filteredOutbounds"
        :pagination="paginationConfig"
        :row-selection="rowSelection"
        :scroll="{ x: 'max-content' }"
        row-key="tag"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'tag'">
            {{ display(record.tag) }}
          </template>
          <template v-else-if="column.key === 'type'">
            {{ display(record.type) }}
          </template>
          <template v-else-if="column.key === 'server'">
            {{ display(record.server) }}
          </template>
          <template v-else-if="column.key === 'server_port'">
            {{ display(record.server_port) }}
          </template>
          <template v-else-if="column.key === 'tls'">
            {{ tlsText(record) }}
          </template>
          <template v-else-if="column.key === 'online'">
            <a-tooltip
              v-if="onlines.includes(record.tag)"
              :title="$t('sessions.title')"
            >
              <a-tag
                color="green"
                style="cursor: pointer; margin: 0;"
                @click="showSessions(record.tag)"
              >
                {{ $t('online') }}
              </a-tag>
            </a-tooltip>
            <span v-else>—</span>
          </template>
          <template v-else-if="column.key === 'delay'">
            <a-spin
              v-if="checkResults[record.tag]?.loading"
              size="small"
            />
            <template v-else-if="checkResults[record.tag]">
              <a-tag
                v-if="checkResults[record.tag].success"
                color="green"
                style="margin: 0;"
              >
                {{ checkResults[record.tag].data?.Delay + $t('date.ms') }}
              </a-tag>
              <a-tooltip
                v-else
                :title="checkResults[record.tag].errorMessage || $t('failed')"
              >
                <AppIcon
                  name="close-circle"
                  color="error"
                  role="img"
                  :aria-hidden="false"
                  :aria-label="checkResults[record.tag].errorMessage || $t('failed')"
                />
              </a-tooltip>
            </template>
            <span v-else>—</span>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space size="small">
              <a-tooltip :title="$t('sessions.title')">
                <a-button
                  type="text"
                  size="small"
                  shape="circle"
                  :aria-label="$t('sessions.title')"
                  @click="showSessions(record.tag)"
                >
                  <template #icon>
                    <AppIcon name="account-multiple" />
                  </template>
                </a-button>
              </a-tooltip>
              <a-tooltip :title="$t('actions.test')">
                <a-button
                  type="text"
                  size="small"
                  shape="circle"
                  :loading="Boolean(checkResults[record.tag]?.loading)"
                  :aria-label="$t('actions.test')"
                  @click="checkOutbound(record.tag)"
                >
                  <template #icon>
                    <AppIcon name="speedometer" />
                  </template>
                </a-button>
              </a-tooltip>
              <a-tooltip :title="$t('actions.edit')">
                <a-button
                  type="text"
                  size="small"
                  shape="circle"
                  :aria-label="$t('actions.edit')"
                  @click="showModal(record.id)"
                >
                  <template #icon>
                    <AppIcon name="file-edit" />
                  </template>
                </a-button>
              </a-tooltip>
              <a-popconfirm
                :title="$t('confirm')"
                :ok-text="$t('yes')"
                :cancel-text="$t('no')"
                ok-type="danger"
                @confirm="delOutbound(record.tag)"
              >
                <a-tooltip :title="$t('actions.del')">
                  <a-button
                    type="text"
                    size="small"
                    danger
                    shape="circle"
                    :aria-label="$t('actions.del')"
                  >
                    <template #icon>
                      <AppIcon name="file-remove" />
                    </template>
                  </a-button>
                </a-tooltip>
              </a-popconfirm>
              <a-tooltip
                v-if="Data().enableTraffic"
                :title="$t('stats.graphTitle')"
              >
                <a-button
                  type="text"
                  size="small"
                  shape="circle"
                  :aria-label="$t('stats.graphTitle')"
                  @click="showStats(record.tag)"
                >
                  <template #icon>
                    <AppIcon name="chart-line" />
                  </template>
                </a-button>
              </a-tooltip>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-col>
  </a-row>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import HttpUtils from '@/plugins/httputil'
import OutboundVue from '@/layouts/modals/Outbound.vue'
import OutboundBulk from '@/layouts/modals/OutboundBulk.vue'
import OutboundReplaceIp from '@/layouts/modals/OutboundReplaceIp.vue'
import Stats from '@/layouts/modals/Stats.vue'
import Sessions from '@/layouts/modals/Sessions.vue'
import { Outbound } from '@/types/outbounds'
import { computed, ref, watch } from 'vue'
import { i18n } from '@/locales'
import UrlTestTarget from '@/components/UrlTestTarget.vue'
import { useUrlTestTarget } from '@/composables/useUrlTestTarget'
import { clampPage as clampPageValue } from '@/utils/pagination'
import { matchesOutboundSearch } from '@/utils/outboundSearch'

// What api/checkOutbound answers with.
interface CheckReply {
  OK: boolean
  Delay?: number
  Error?: string
}

interface CheckResult {
  loading?: boolean
  success: boolean
  data?: CheckReply | null
  errorMessage?: string
}

const checkResults = ref<Record<string, CheckResult>>({})
const { link: testLink } = useUrlTestTarget()

const checkOutbound = async (tag: string) => {
  checkResults.value = { ...checkResults.value, [tag]: { loading: true, success: false } }
  const msg = await HttpUtils.get<CheckReply>('api/checkOutbound', { tag, link: testLink.value })
  const success = msg.success && msg.obj?.OK
  const errorMessage = success ? undefined : (msg.obj?.Error ?? msg.msg ?? '')
  checkResults.value = {
    ...checkResults.value,
    [tag]: { loading: false, success, data: msg.obj ?? null, errorMessage }
  }
}

const testingAll = ref(false)

const checkAllOutbounds = async () => {
  const list = [...outbounds.value]
  if (list.length === 0) return
  testingAll.value = true
  try {
    await Promise.all(list.map((o) => checkOutbound(o.tag)))
  } finally {
    testingAll.value = false
  }
}

const outbounds = computed((): Outbound[] => {
  return <Outbound[]> Data().outbounds
})

const searchQuery = ref('')

const filteredOutbounds = computed((): Outbound[] => {
  return outbounds.value.filter((item) => matchesOutboundSearch(item, searchQuery.value))
})

const selectedTags = ref<string[]>([])
const replaceModal = ref(false)
const replaceDone = ref(false)
const selectable = (item: Outbound): boolean => ['socks', 'http'].includes(item.type)
const closeReplace = () => {
  replaceModal.value = false
  if (replaceDone.value) {
    selectedTags.value = []
    replaceDone.value = false
  }
}

const replaceTargets = computed((): Outbound[] => {
  return selectedTags.value
    .map(tag => outbounds.value.find(item => item.tag === tag))
    .filter((item): item is Outbound => item !== undefined && selectable(item))
})

// Ant reports the selected keys in table order. Keep the existing order and
// append only newly selected keys, so C -> B -> A is preserved.
const onSelectionChange = (keys: (string | number)[]) => {
  const next = keys.map(String)
  const nextSet = new Set(next)
  const kept = selectedTags.value.filter(tag => nextSet.has(tag))
  const added = next.filter(tag => !selectedTags.value.includes(tag))
  selectedTags.value = [...kept, ...added]
}

const rowSelection = computed(() => ({
  selectedRowKeys: selectedTags.value,
  preserveSelectedRowKeys: true,
  getCheckboxProps: (item: Outbound) => ({ disabled: !selectable(item) }),
  onChange: onSelectionChange,
}))

watch(outbounds, (list) => {
  selectedTags.value = selectedTags.value.filter(tag => {
    const item = list.find(outbound => outbound.tag === tag)
    return item !== undefined && selectable(item)
  })
})

const outboundTags = computed((): string[] => {
  return [...(Data().outbounds?.map((o:Outbound) => o.tag) ?? []), ...(Data().endpoints?.map(e => e.tag) ?? [])]
})

const onlines = computed(() => {
  return Data().onlines.outbound?? []
})

const display = (value: unknown): string => {
  if (value === undefined || value === null || value === '') return '—'
  return String(value)
}

const tlsText = (item: Outbound): string => {
  if (!Object.hasOwn(item, 'tls')) return '—'
  return i18n.global.t(item.tls?.enabled ? 'enable' : 'disable')
}

// The same localStorage setting as Clients.vue: changing the page size on
// either page is intentionally shared.
const itemPerPage = ref(localStorage.getItem('items-per-page') || '10')
const pageSize = computed(() => {
  const n = Number(itemPerPage.value)
  return n > 0 ? n : 10
})
const currentPage = ref(1)

const setItemPerPage = (items: number) => {
  itemPerPage.value = items.toString()
  localStorage.setItem('items-per-page', items.toString())
}

const clampPage = () => {
  currentPage.value = clampPageValue(currentPage.value, filteredOutbounds.value.length, pageSize.value)
}

watch([() => filteredOutbounds.value.length, pageSize], clampPage, { immediate: true })
watch(searchQuery, () => {
  currentPage.value = 1
})

const paginationConfig = computed(() => {
  if (filteredOutbounds.value.length <= 10) return false
  return {
    current: currentPage.value,
    pageSize: pageSize.value,
    total: filteredOutbounds.value.length,
    showSizeChanger: true,
    pageSizeOptions: ['10', '25', '50', '100'],
    onChange: (page: number) => {
      currentPage.value = page
    },
    onShowSizeChange: (current: number, size: number) => {
      setItemPerPage(size)
      currentPage.value = current
      clampPage()
    }
  }
})

const headers = [
  { title: i18n.global.t('out.name'), dataIndex: 'tag', key: 'tag' },
  { title: i18n.global.t('type'), dataIndex: 'type', key: 'type' },
  { title: i18n.global.t('out.addr'), dataIndex: 'server', key: 'server' },
  { title: i18n.global.t('out.port'), dataIndex: 'server_port', key: 'server_port' },
  { title: i18n.global.t('objects.tls'), key: 'tls' },
  { title: i18n.global.t('online'), key: 'online' },
  { title: i18n.global.t('out.delay'), key: 'delay' },
  { title: i18n.global.t('actions.action'), key: 'actions' },
]

const modal = ref({
  visible: false,
  id: 0,
  data: "",
})

const showModal = (id: number) => {
  modal.value.id = id
  modal.value.data = id == 0 ? '' : JSON.stringify(outbounds.value.findLast(o => o.id == id))
  modal.value.visible = true
}

const closeModal = () => {
  modal.value.visible = false
}

const bulkModal = ref({ visible: false })

const showBulkModal = () => {
  bulkModal.value.visible = true
}

const closeBulkModal = () => {
  bulkModal.value.visible = false
}

const stats = ref({
  visible: false,
  resource: "outbound",
  tag: "",
})

const delOutbound = async (tag: string) => {
  const ok = await Data().save("outbounds", "del", tag)
  if (ok) clampPage()
}

const showStats = (tag: string) => {
  stats.value.tag = tag
  stats.value.visible = true
}
const closeStats = () => {
  stats.value.visible = false
}

const sessions = ref({
  visible: false,
  tag: "",
})

const showSessions = (tag: string) => {
  sessions.value.tag = tag
  sessions.value.visible = true
}
const closeSessions = () => {
  sessions.value.visible = false
}
</script>

<style scoped>
.search-count {
  font-size: 12px;
  opacity: 0.6;
}
</style>

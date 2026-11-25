
<template>
  <ClientModal 
    :id="modal.id"
    v-model="modal.visible"
    :visible="modal.visible"
    :groups="groups"
    :inbound-tags="inboundTags"
    @close="closeModal"
  />
  <ClientAddBulk 
    v-model="addBulkModal"
    :visible="addBulkModal"
    :groups="groups"
    :inbound-tags="inboundTags"
    @close="closeAddBulk"
  />
  <ClientEditBulk 
    v-model="editBulkModal"
    :visible="editBulkModal"
    :inbound-tags="inboundTags"
    :clients="clients"
    @close="closeEditBulk"
  />
  <QrCode
    :id="qrcode.id"
    v-model="qrcode.visible"
    :visible="qrcode.visible"
    @close="closeQrCode"
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
    resource="user"
    :tag="sessions.tag"
    @close="closeSessions"
  />
  <UserIps
    :visible="ipUser !== null"
    :user="ipUser ?? ''"
    @close="ipUser = null"
  />
  <a-modal
    v-model:open="resetTrafficModal"
    :title="$t('actions.resetTraffic')"
    :width="400"
  >
    {{ $t('confirm') }}
    <template #footer>
      <a-button
        type="default"
        @click="resetTrafficModal = false"
      >
        {{ $t('no') }}
      </a-button>
      <a-button
        type="primary"
        danger
        :loading="resetTrafficLoading"
        @click="resetTraffic"
      >
        {{ $t('yes') }}
      </a-button>
    </template>
  </a-modal>
  <a-row
    :gutter="[12, 12]"
    justify="center"
    align="middle"
    style="margin-bottom: 12px;"
  >
    <a-col>
      <a-button
        type="primary"
        @click="showModal(0)"
      >
        {{ $t('actions.add') }}
      </a-button>
    </a-col>
    <a-col>
      <a-dropdown
        v-model:open="actionMenu"
        :trigger="['click']"
        placement="bottom"
      >
        <a-button
          type="text"
          shape="circle"
          :aria-label="$t('a11y.tools')"
        >
          <template #icon>
            <AppIcon
              name="tools"
              color="primary"
            />
          </template>
        </a-button>
        <template #overlay>
          <a-menu>
            <a-menu-item
              key="addBulk"
              @click="addBulk"
            >
              <a-space>
                <AppIcon name="account-multiple-plus" />
                <span>{{ $t('actions.addbulk') }}</span>
              </a-space>
            </a-menu-item>
            <a-menu-item
              key="editBulk"
              @click="editBulk"
            >
              <a-space>
                <AppIcon name="account-multiple-check" />
                <span>{{ $t('actions.editbulk') }}</span>
              </a-space>
            </a-menu-item>
            <a-menu-item
              key="resetTraffic"
              @click="confirmResetTraffic"
            >
              <a-space>
                <AppIcon name="restore" />
                <span>{{ $t('actions.resetTraffic') }}</span>
              </a-space>
            </a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
    </a-col>
    <a-col>
      <a-popover
        v-model:open="filterMenu"
        trigger="click"
        placement="bottom"
      >
        <a-button
          type="text"
          shape="circle"
          :aria-label="$t('a11y.filter')"
        >
          <template #icon>
            <AppIcon
              :name="filterSettings.enabled ? 'filter-check-outline' : 'filter-menu-outline'"
              :color="filterSettings.enabled ? 'primary' : undefined"
            />
          </template>
        </a-button>
        <template #content>
          <a-form
            layout="vertical"
            style="min-width: 240px;"
          >
            <a-form-item
              :label="$t('type')"
              style="margin-bottom: 12px;"
            >
              <a-select
                v-model:value="filterSettings.state"
                :options="filterItems"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
            <a-form-item
              :label="$t('client.group')"
              style="margin-bottom: 12px;"
            >
              <a-select
                v-model:value="filterSettings.group"
                :options="[{ title: $t('all'), value: '-' }, ...groups.map(g => ({ title: g.length > 0 ? g : $t('none'), value: g }))]"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
            <a-form-item
              :label="$t('client.name')"
              style="margin-bottom: 12px;"
            >
              <a-input v-model:value="filterSettings.text" />
            </a-form-item>
            <a-flex
              justify="end"
              gap="small"
            >
              <a-button
                type="default"
                danger
                @click="clearFilter"
              >
                {{ $t('actions.del') }}
              </a-button>
              <a-button
                type="primary"
                @click="doFilter"
              >
                {{ $t('actions.update') }}
              </a-button>
            </a-flex>
          </a-form>
        </template>
      </a-popover>
    </a-col>
  </a-row>
  <a-row :gutter="[12, 12]">
    <a-col :span="24">
      <a-table
        :columns="headers"
        :data-source="displayedClients"
        :pagination="displayedClients.length <= 10 ? false : { pageSize: pageSize, showSizeChanger: true, pageSizeOptions: ['10', '25', '50', '100'], onShowSizeChange: (_: number, size: number) => setItemPerPage(size) }"
        :scroll="{ x: 'max-content' }"
        row-key="name"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'enable'">
            <a-switch
              :checked="record.enable"
              :loading="toggling[record.id]"
              :disabled="toggling[record.id]"
              size="small"
              @change="(val: any) => toggleEnable(record, !!val)"
            />
          </template>
          <template v-else-if="column.key === 'inbounds'">
            <a-tooltip
              v-if="record.inbounds?.length != 0"
              placement="left"
            >
              <template #title>
                <div dir="ltr">
                  <div
                    v-for="i in record.inbounds"
                    :key="i"
                  >
                    {{ inbounds.find(inb => inb.id == i)?.tag }}
                  </div>
                </div>
              </template>
              <span>{{ record.inbounds?.length }}</span>
            </a-tooltip>
            <span v-else>{{ record.inbounds?.length }}</span>
          </template>
          <template v-else-if="column.key === 'volume'">
            <a-tooltip :title="'↓' + HumanReadable.sizeFormat(record.down) + ' - ' + HumanReadable.sizeFormat(record.up) + '↑'">
              <a-tag :color="record.volume == 0 ? 'green' : record.volume <= (record.up + record.down) ? 'red' : undefined">
                {{ HumanReadable.sizeFormat(record.up + record.down) + ' / ' + (record.volume == 0 ? $t('unlimited') : HumanReadable.sizeFormat(record.volume)) }}
              </a-tag>
            </a-tooltip>
            <a-progress
              v-if="record.volume > 0"
              :percent="percent(record)"
              :stroke-color="percentColor(record)"
              :show-info="false"
              size="small"
              style="margin-top: 2px;"
            />
          </template>
          <template v-else-if="column.key === 'expiry'">
            <a-tooltip
              v-if="record.expiry > 0"
              :title="new Date(record.expiry * 1000).toLocaleString(locale)"
            >
              <a-tag :color="record.expiry <= Date.now() / 1000 ? 'red' : undefined">
                {{ HumanReadable.remainedDays(record.expiry) }}
              </a-tag>
            </a-tooltip>
            <a-tag
              v-else
              color="green"
            >
              {{ HumanReadable.remainedDays(record.expiry) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'online'">
            <a-tooltip
              v-if="isOnline(record.name).value"
              :title="$t('sessions.title')"
            >
              <a-tag
                color="green"
                style="cursor: pointer;"
                @click="showSessions(record.name)"
              >
                {{ $t('online') }}
              </a-tag>
            </a-tooltip>
            <template v-else>
              -
            </template>
          </template>
          <template v-else-if="column.key === 'connections'">
            <a-tooltip :title="connHint(record)">
              <a-tag
                :color="connColor(record)"
                style="cursor: pointer; font-variant-numeric: tabular-nums;"
                @click="showSessions(record.name)"
              >
                {{ activeConns(record.name) }} / {{ connLimitText(record) }}
              </a-tag>
            </a-tooltip>
          </template>
          <template v-else-if="column.key === 'ips'">
            <a-tooltip :title="ipHint(record)">
              <a-tag
                :color="ipColor(record)"
                style="cursor: pointer; font-variant-numeric: tabular-nums;"
                @click="ipUser = record.name"
              >
                {{ activeIPs(record.name) }} / {{ ipLimitText(record) }}
              </a-tag>
            </a-tooltip>
          </template>
          <template v-else-if="column.key === 'createdAt'">
            <a-tooltip
              v-if="record.createdAt > 0"
              :title="new Date(record.createdAt * 1000).toLocaleString(locale)"
            >
              <span>{{ new Date(record.createdAt * 1000).toLocaleDateString(locale) }}</span>
            </a-tooltip>
            <template v-else>
              -
            </template>
          </template>
          <template v-else-if="column.key === 'onlineAt'">
            <a-tooltip
              v-if="record.onlineAt > 0"
              :title="new Date(record.onlineAt * 1000).toLocaleString(locale)"
            >
              <span>{{ new Date(record.onlineAt * 1000).toLocaleString(locale) }}</span>
            </a-tooltip>
            <template v-else>
              -
            </template>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space :size="4">
              <a-tooltip :title="$t('actions.edit')">
                <a-button
                  type="text"
                  shape="circle"
                  size="small"
                  :aria-label="$t('actions.edit')"
                  @click="showModal(record.id)"
                >
                  <template #icon>
                    <AppIcon name="pencil" />
                  </template>
                </a-button>
              </a-tooltip>
              <a-popconfirm
                v-model:open="delOverlay[clients.findIndex(c => c.id == record.id)]"
                :title="$t('confirm')"
                :ok-text="$t('yes')"
                :cancel-text="$t('no')"
                ok-type="danger"
                @confirm="delClient(record.id)"
              >
                <a-button
                  type="text"
                  shape="circle"
                  size="small"
                  danger
                  :aria-label="$t('actions.del')"
                >
                  <template #icon>
                    <AppIcon name="delete" />
                  </template>
                </a-button>
              </a-popconfirm>
              <a-tooltip :title="$t('a11y.qrCode')">
                <a-button
                  type="text"
                  shape="circle"
                  size="small"
                  :aria-label="$t('a11y.qrCode')"
                  @click="showQrCode(record.id)"
                >
                  <template #icon>
                    <AppIcon name="qrcode" />
                  </template>
                </a-button>
              </a-tooltip>
              <a-tooltip
                v-if="Data().enableTraffic"
                :title="$t('stats.graphTitle')"
              >
                <a-button
                  type="text"
                  shape="circle"
                  size="small"
                  :aria-label="$t('stats.graphTitle')"
                  @click="showStats(record.name)"
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
import ClientModal from '@/layouts/modals/Client.vue'
import ClientAddBulk from '@/layouts/modals/ClientAddBulk.vue'
import ClientEditBulk from '@/layouts/modals/ClientEditBulk.vue'
import QrCode from '@/layouts/modals/QrCode.vue'
import Stats from '@/layouts/modals/Stats.vue'
import Sessions from '@/layouts/modals/Sessions.vue'
import UserIps from '@/layouts/modals/UserIps.vue'
import { Client } from '@/types/clients'
import { Inbound } from '@/types/inbounds'
import { computed, ref } from 'vue'
import { HumanReadable } from '@/plugins/utils'
import { i18n, locale } from '@/locales'
import HttpUtils from '@/plugins/httputil'

// A client as it comes back from the panel: saved, so it carries the id and
// the timestamps that only a client built in the browser is still without.
type ClientRow = Client & {
  id: number
  createdAt: number
  onlineAt: number
}

const clients = computed((): ClientRow[] => {
  return <ClientRow[]> Data().clients
})

// A client is online when it has a live routed connection or recent traffic.
// The connection map covers idle sessions, which the traffic-based list misses.
const activeConns = (cname: string): number => Data().connections?.[cname] ?? 0
const onlineNow = (cname: string): boolean => activeConns(cname) > 0 || (Data().onlines?.user?.includes(cname) ?? false)
const isOnline = (cname: string) => computed(() => onlineNow(cname))

const connLimit = (c: ClientRow): number => (c.maxConns && c.maxConns > 0 ? c.maxConns : 0)
const connLimitText = (c: ClientRow): string => connLimit(c) > 0 ? String(connLimit(c)) : '∞'
const connColor = (c: ClientRow): string | undefined => {
  const limit = connLimit(c)
  const active = activeConns(c.name)
  if (limit > 0 && active >= limit) return 'red'
  if (limit > 0 && active >= Math.ceil(limit * 0.8)) return 'orange'
  return undefined
}
const connHint = (c: ClientRow): string => i18n.global.t('client.connectionsHint', {
  current: activeConns(c.name),
  limit: connLimitText(c)
})

// Distinct client IPs currently used by the client. The backend groups IPv6
// addresses by /64, so this is the number an ipLimit is compared against.
const activeIPs = (cname: string): number => Data().ipCounts?.[cname] ?? 0
const ipLimit = (c: ClientRow): number => (c.ipLimit && c.ipLimit > 0 ? c.ipLimit : 0)
const ipLimitText = (c: ClientRow): string => ipLimit(c) > 0 ? String(ipLimit(c)) : '∞'
const ipColor = (c: ClientRow): string | undefined => {
  const limit = ipLimit(c)
  const active = activeIPs(c.name)
  if (limit > 0 && active >= limit) return 'red'
  if (limit > 0 && active >= Math.ceil(limit * 0.8)) return 'orange'
  return undefined
}
const ipHint = (c: ClientRow): string => i18n.global.t('client.ipHint', {
  current: activeIPs(c.name),
  limit: ipLimitText(c)
})

const inbounds = computed((): Inbound[] => {
  return Data().inbounds?? []
})

// Only the inbound types that declare users are in the union, but any inbound
// can carry one, so the field is read through a shape that only says that.
const inboundTags = computed((): { title: string, value: number }[] => {
  if (!inbounds.value) return []
  return inbounds.value?.filter(i => i.tag != "" && (i as { users?: unknown }).users).map(i => { return { title: i.tag, value: i.id } })
})

const groups = computed((): string[] => {
  if (!clients.value) return []
  if (filterSettings?.value.enabled) return Array.from(new Set(filterSettings.value.filteredClients?.map(c => c.group)))
  return Array.from(new Set(clients.value?.map(c => c.group)))
})

const actionMenu = ref(false)
const filterMenu = ref(false)
const filterSettings = ref({
  enabled: false,
  state: '',
  group: '-',
  text: '',
  filteredClients: <ClientRow[]>[]
})

const filterItems = [
  { title: i18n.global.t('none'), value: '' },
  { title: i18n.global.t('disable'), value: 'disable' },
  { title: i18n.global.t('date.expired'), value: 'expired' },
  { title: i18n.global.t('online'), value: 'online' },
]

const headers = [
  { title: i18n.global.t('client.name'), dataIndex: 'name', key: 'name', sorter: (a: ClientRow, b: ClientRow) => (a.name ?? '').localeCompare(b.name ?? '') },
  { title: i18n.global.t('enable'), dataIndex: 'enable', key: 'enable', sorter: (a: ClientRow, b: ClientRow) => Number(a.enable) - Number(b.enable) },
  { title: i18n.global.t('client.desc'), dataIndex: 'desc', key: 'desc', sorter: (a: ClientRow, b: ClientRow) => (a.desc ?? '').localeCompare(b.desc ?? '') },
  { title: i18n.global.t('client.group'), dataIndex: 'group', key: 'group', sorter: (a: ClientRow, b: ClientRow) => (a.group ?? '').localeCompare(b.group ?? '') },
  { title: i18n.global.t('pages.inbounds'), dataIndex: 'inbounds', key: 'inbounds', sorter: (a: ClientRow, b: ClientRow) => (a.inbounds?.length ?? 0) - (b.inbounds?.length ?? 0) },
  { title: i18n.global.t('actions.action'), key: 'actions' },
  // Sorted by what was used, not by the quota: with unlimited clients every
  // volume is 0 and the column did not sort at all (#1277).
  { title: i18n.global.t('stats.volume'), dataIndex: 'volume', key: 'volume', sorter: (a: ClientRow, b: ClientRow) => (a.up + a.down) - (b.up + b.down) },
  { title: i18n.global.t('date.expiry'), dataIndex: 'expiry', key: 'expiry', sorter: (a: ClientRow, b: ClientRow) => a.expiry - b.expiry },
  { title: i18n.global.t('online'), key: 'online', sorter: (a: ClientRow, b: ClientRow) => Number(isOnline(a.name).value) - Number(isOnline(b.name).value) },
  { title: i18n.global.t('client.connections'), key: 'connections', sorter: (a: ClientRow, b: ClientRow) => activeConns(a.name) - activeConns(b.name) },
  { title: i18n.global.t('client.ips'), key: 'ips', sorter: (a: ClientRow, b: ClientRow) => activeIPs(a.name) - activeIPs(b.name) },
  { title: i18n.global.t('date.created'), dataIndex: 'createdAt', key: 'createdAt', sorter: (a: ClientRow, b: ClientRow) => a.createdAt - b.createdAt },
  { title: i18n.global.t('date.lastOnline'), dataIndex: 'onlineAt', key: 'onlineAt', sorter: (a: ClientRow, b: ClientRow) => a.onlineAt - b.onlineAt },
]

const displayedClients = computed((): ClientRow[] => {
  return filterSettings.value.enabled ? filterSettings.value.filteredClients : clients.value
})

const itemPerPage = ref(localStorage.getItem('items-per-page') || '10')

const pageSize = computed(() => {
  const n = Number(itemPerPage.value)
  return n > 0 ? n : 10
})

const setItemPerPage = (items: number) => {
  itemPerPage.value = items.toString()
  localStorage.setItem('items-per-page', items.toString())
}

const modal = ref({
  visible: false,
  id: 0,
})

const delOverlay = ref(new Array<boolean>(clients.value.length).fill(false))

const showModal = async (id: number) => {
  modal.value.id = id
  modal.value.visible = true
}
const closeModal = () => {
  modal.value.visible = false
}

const toggling = ref<Record<number, boolean>>({})

const toggleEnable = async (item: ClientRow, val: boolean) => {
  if (item.enable === val) return
  toggling.value = { ...toggling.value, [item.id]: true }
  const full = await Data().loadClients(item.id)
  if (full?.id) {
    full.enable = val
    await Data().save("clients", "edit", full)
  }
  toggling.value = { ...toggling.value, [item.id]: false }
}

const delClient = async (id: number) => {
  const index = clients.value.findIndex(c => c.id === id)
  const success = await Data().save("clients", "del", id)
  if (success) delOverlay.value[index] = false
}

const qrcode = ref({
  visible: false,
  id: 0,
})

const showQrCode = (id: number) => {
  qrcode.value.id = id
  qrcode.value.visible = true
}
const closeQrCode = () => {
  qrcode.value.visible = false
}

const stats = ref({
  visible: false,
  resource: "user",
  tag: "",
})

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

// Opening the IP list only needs the client name; the modal loads the rest.
const ipUser = ref<string | null>(null)

const doFilter = () => {
  let filteredClients = clients.value.slice()
  if (filterSettings.value.group != '-') {
    filteredClients = filteredClients.filter(c => c.group == filterSettings.value.group)
  }
  if (filterSettings.value.text.length>0) {
    // includes, not search: String.search compiles its argument as a regular
    // expression, so typing "a.b" matched anything with a character between a
    // and b, and an unbalanced "(" threw out of the filter entirely.
    const txt = filterSettings.value.text.toLowerCase()
    filteredClients = filteredClients.filter(c =>
      (c.name ?? '').toLowerCase().includes(txt) || (c.desc ?? '').toLowerCase().includes(txt))
  }
  switch (filterSettings.value.state) {
    case "disable":
      filteredClients = filteredClients.filter(c => c.enable == false)
      break
    case "expired":
      filteredClients = filteredClients.filter(c => c.expiry > 0 && c.expiry < (Date.now()/1000) )
      break
    case "online":
      filteredClients = filteredClients.filter(c => onlineNow(c.name))
      break
  }
  filterSettings.value.filteredClients = filteredClients
  filterSettings.value.enabled = true
  filterMenu.value = false
}

const clearFilter = () => {
  filterSettings.value = {
    enabled: false,
    state: '',
    group: '-',
    text: '',
    filteredClients: <ClientRow[]>[]
  }
  filterMenu.value = false
}

const addBulkModal = ref(false)

const addBulk = () => {
  addBulkModal.value = true
  actionMenu.value = false
}

const closeAddBulk = () => {
  addBulkModal.value = false
}

const editBulkModal = ref(false)

const editBulk = () => {
  editBulkModal.value = true
  actionMenu.value = false
}

const closeEditBulk = () => {
  editBulkModal.value = false
}

const resetTrafficModal = ref(false)
const resetTrafficLoading = ref(false)

const confirmResetTraffic = () => {
  resetTrafficModal.value = true
  actionMenu.value = false
}

const resetTraffic = async () => {
  resetTrafficLoading.value = true
  const msg = await HttpUtils.post('api/resetTraffic', {})
  resetTrafficLoading.value = false
  if (msg.success) {
    resetTrafficModal.value = false
    await Data().loadData()
  }
}

const percent = (c: Client) => { return c.volume>0 ? Math.round((c.up+c.down) *100 / c.volume) : 0 }
const percentColor = (c: Client) => { return (c.up+c.down) >= c.volume ? '#ff4d4f' : percent(c)>90 ? '#faad14' : '#52c41a' }

</script>
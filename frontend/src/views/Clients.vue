
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
  <ClientImportNodes
    v-model="importNodesModal"
    :visible="importNodesModal"
    :inbound-tags="inboundTags"
    @close="importNodesModal = false"
  />
  <SubscriptionsModal
    :visible="subscriptionsModal"
    @close="subscriptionsModal = false"
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
  <a-modal
    v-model:open="batchDeleteModal"
    :title="$t('actions.delbulk')"
    :width="420"
  >
    <p>{{ $t('client.deleteSelectedConfirm', { count: selectedNames.length }) }}</p>
    <template #footer>
      <a-button
        type="default"
        @click="batchDeleteModal = false"
      >
        {{ $t('no') }}
      </a-button>
      <a-button
        type="primary"
        danger
        :loading="batchDeleteLoading"
        @click="deleteSelected"
      >
        {{ $t('yes') }}
      </a-button>
    </template>
  </a-modal>
  <a-row
    :gutter="[12, 12]"
    justify="start"
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
              key="addNodes"
              @click="importNodesModal = true"
            >
              <a-space>
                <AppIcon name="server-plus" />
                <span>{{ $t('importNodes.entry') }}</span>
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
              key="subscriptions"
              @click="subscriptionsModal = true; actionMenu = false"
            >
              <a-space>
                <AppIcon name="link" />
                <span>{{ $t('subscriptions.title') }}</span>
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
      <a-button
        danger
        :disabled="selectedNames.length === 0"
        @click="confirmBatchDelete"
      >
        {{ $t('actions.delbulk') }}<template v-if="selectedNames.length"> ({{ selectedNames.length }})</template>
      </a-button>
    </a-col>
    <a-col class="client-search-col">
      <a-input
        v-model:value="searchQuery"
        allow-clear
        :placeholder="$t('client.searchPlaceholder')"
        class="client-search"
      >
        <template #prefix>
          <AppIcon name="magnify" />
        </template>
      </a-input>
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
        :pagination="paginationConfig"
        :row-selection="rowSelection"
        :scroll="{ x: 'max-content' }"
        row-key="name"
        size="small"
        @change="onTableChange"
      >
        <template #footer>
          <div class="client-table-footer">
            <span>
              {{ $t('client.totalItems', { total: clients.length }) }}
              <template v-if="searchQuery.trim()">
                · {{ $t('client.filteredItems', { count: displayedClients.length }) }}
              </template>
            </span>
            <span v-if="selectedNames.length" class="client-selected-count">
              {{ $t('client.selectedItems', { count: selectedNames.length }) }}
            </span>
          </div>
        </template>
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
          <template v-else-if="column.key === 'address'">
            <div class="address-cell">
              <template v-if="addressEntry(record.name)?.status === 'split'">
                <a-tooltip :title="$t('client.outboundSplitHint')">
                  <a-tag>{{ $t('client.outboundSplit') }}</a-tag>
                </a-tooltip>
              </template>
              <template v-else-if="addressEntry(record.name)?.status === 'unbound'">
                <a-tooltip :title="$t('client.outboundUnboundHint')">
                  <span class="address-muted">{{ $t('client.outboundUnbound') }}</span>
                </a-tooltip>
              </template>
              <template v-else-if="addressEntry(record.name)?.status === 'direct'">
                <a-tooltip :title="addressEntry(record.name)?.outbound">
                  <a-tag>{{ $t('client.outboundDirect') }}</a-tag>
                </a-tooltip>
              </template>
              <template v-else-if="addressEntry(record.name)?.status === 'multi'">
                <a-tooltip :title="addressEntry(record.name)?.outbound">
                  <a-tag>{{ $t('client.outboundMulti') }}</a-tag>
                </a-tooltip>
              </template>
              <template v-else-if="addressEntry(record.name)?.status === 'ok'">
                <a-tooltip placement="left">
                  <template #title>
                    <div>{{ $t('client.outboundLabel') }}: {{ addressEntry(record.name)?.outbound }}</div>
                    <div>{{ $t('client.outboundEndpoint') }}: {{ addressWithPort(record.name) }}</div>
                  </template>
                  <span class="address-text">{{ addressEntry(record.name)?.address }}</span>
                </a-tooltip>
                <a-tooltip :title="$t('client.outboundCopy')">
                  <a-button
                    type="text"
                    size="small"
                    :aria-label="$t('client.outboundCopy')"
                    @click="copyAddress(record.name)"
                  >
                    <AppIcon name="content-copy" />
                  </a-button>
                </a-tooltip>
              </template>
              <template v-else>
                <a-tooltip :title="addressEntry(record.name)?.outbound">
                  <span class="address-muted">—</span>
                </a-tooltip>
              </template>
            </div>
          </template>
          <template v-else-if="column.key === 'dates'">
            <div class="date-cell">
              <a-tooltip :title="$t('date.created')">
                <a-tag
                  v-if="record.createdAt > 0"
                  class="date-tag"
                >
                  {{ dateTime(record.createdAt) }}
                </a-tag>
                <a-tag
                  v-else
                  class="date-tag"
                >
                  -
                </a-tag>
              </a-tooltip>
              <a-tooltip :title="$t('date.expiry')">
                <a-tag
                  v-if="record.expiry > 0"
                  class="date-tag"
                  :color="record.expiry <= Date.now() / 1000 ? 'red' : undefined"
                >
                  {{ dateTime(record.expiry) }}
                </a-tag>
                <a-tag
                  v-else
                  class="date-tag"
                  color="green"
                >
                  {{ $t('unlimited') }}
                </a-tag>
              </a-tooltip>
            </div>
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
          <template v-else-if="column.key === 'speed'">
            <a-tag
              style="margin-inline-end: 0; font-variant-numeric: tabular-nums; white-space: nowrap;"
            >
              {{ speedPair(record.name) }}
            </a-tag>
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
                :title="`${$t('confirm')} ${record.name}? ${$t('client.deleteHint')}`"
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
              <a-tooltip :title="$t('importNodes.copyPersonalSubscription')">
                <a-button
                  type="text"
                  shape="circle"
                  size="small"
                  :aria-label="$t('importNodes.copyPersonalSubscription')"
                  @click="copyClientSubscription(record.id)"
                >
                  <template #icon>
                    <AppIcon name="link" />
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
import ClientImportNodes from '@/layouts/modals/ClientImportNodes.vue'
import SubscriptionsModal from '@/layouts/modals/Subscriptions.vue'
import ClientEditBulk from '@/layouts/modals/ClientEditBulk.vue'
import QrCode from '@/layouts/modals/QrCode.vue'
import Stats from '@/layouts/modals/Stats.vue'
import Sessions from '@/layouts/modals/Sessions.vue'
import UserIps from '@/layouts/modals/UserIps.vue'
import { Client } from '@/types/clients'
import { Inbound } from '@/types/inbounds'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { HumanReadable } from '@/plugins/utils'
import { i18n, locale } from '@/locales'
import { push } from 'notivue'
import HttpUtils from '@/plugins/httputil'
import api from '@/plugins/api'
import { formatSpeedPair, speedFromBytes } from '@/utils/formatSpeed'
import { outboundAddressSortValue, outboundAddressText, outboundAddressWithPort, type OutboundAddressEntry, type OutboundAddressResponse } from '@/utils/outboundAddress'
import { formatDateTime } from '@/utils/formatDate'

// A client as it comes back from the panel: saved, so it carries the id and
// the timestamps that only a client built in the browser is still without.
type ClientRow = Client & {
  id: number
  createdAt: number
  onlineAt: number
}

// A rate is null until two valid snapshots from the same core instance exist.
// The template renders null as an em dash.
interface SpeedSample {
  up: number | null
  down: number | null
}

interface UserSpeedResponse {
  running: boolean
  instance: string
  sampledAt: number
  users?: Record<string, { up: number, down: number }>
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
const importNodesModal = ref(false)
const subscriptionsModal = ref(false)
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

// Rates are kept per client name. They live only in this page: the two-second
// poll must not reload the whole client list or touch the shared data store.
const speedRates = ref<Record<string, SpeedSample>>({})
const speedValue = (name: string, direction: 'up' | 'down'): number => speedRates.value[name]?.[direction] ?? -1
const speedPair = (name: string): string => formatSpeedPair(speedRates.value[name]?.up ?? null, speedRates.value[name]?.down ?? null)

// The outbound address is read from the current config, not measured. It is
// fetched once per visible page and rendered immediately; there is no polling
// and no network probe.
const addressByUser = ref<Record<string, OutboundAddressEntry>>({})
const addressEntry = (name: string): OutboundAddressEntry | undefined => addressByUser.value[name]
const addressWithPort = (name: string): string => outboundAddressWithPort(addressEntry(name))
const dateTime = (unixSeconds?: number): string => formatDateTime(unixSeconds) || '-'

let addressTimer: ReturnType<typeof setTimeout> | null = null
let addressSeq = 0
const addressSeqByUser: Record<string, number> = {}

const fetchAddresses = async () => {
  const names = pageNames.value
  if (names.length === 0) return
  const seq = ++addressSeq
  for (const name of names) addressSeqByUser[name] = seq
  try {
    const response = await api.get<{ success: boolean, obj: OutboundAddressResponse }>('api/outboundAddresses', {
      params: { users: JSON.stringify(names) },
      timeout: 5000,
    })
    const msg = response.data
    if (!msg?.success || !msg.obj?.users) return
    const next = { ...addressByUser.value }
    for (const name of names) {
      const entry = msg.obj.users[name]
      if (!entry) continue
      if ((addressSeqByUser[name] ?? 0) > seq) continue
      next[name] = entry
    }
    addressByUser.value = next
  } catch {
    // Keep whatever was already shown; a page change will retry.
  }
}

const scheduleAddressFetch = () => {
  if (addressTimer !== null) clearTimeout(addressTimer)
  addressTimer = setTimeout(() => {
    addressTimer = null
    void fetchAddresses()
  }, 0)
}

const copyAddress = async (name: string) => {
  const address = outboundAddressText(addressEntry(name))
  if (!address) return
  try {
    await navigator.clipboard.writeText(address)
    push.success({ title: i18n.global.t('success'), message: i18n.global.t('client.outboundCopy') })
  } catch {
    push.error({ title: i18n.global.t('failed'), message: i18n.global.t('client.outboundCopy') })
  }
}

const headers = [
  { title: i18n.global.t('enable'), dataIndex: 'enable', key: 'enable', width: 64, sorter: (a: ClientRow, b: ClientRow) => Number(a.enable) - Number(b.enable) },
  { title: i18n.global.t('client.name'), dataIndex: 'name', key: 'name', width: 112, sorter: (a: ClientRow, b: ClientRow) => (a.name ?? '').localeCompare(b.name ?? '') },
  { title: i18n.global.t('client.group'), dataIndex: 'group', key: 'group', width: 76, sorter: (a: ClientRow, b: ClientRow) => (a.group ?? '').localeCompare(b.group ?? '') },
  { title: i18n.global.t('pages.inbounds'), dataIndex: 'inbounds', key: 'inbounds', width: 92, sorter: (a: ClientRow, b: ClientRow) => (a.inbounds?.length ?? 0) - (b.inbounds?.length ?? 0) },
  { title: i18n.global.t('client.outboundAddress'), key: 'address', width: 190, sorter: (a: ClientRow, b: ClientRow) => outboundAddressSortValue(addressByUser.value[a.name]).localeCompare(outboundAddressSortValue(addressByUser.value[b.name])) },
  { title: i18n.global.t('client.createdExpiry'), key: 'dates', width: 160 },
  { title: i18n.global.t('online'), key: 'online', width: 66, sorter: (a: ClientRow, b: ClientRow) => Number(isOnline(a.name).value) - Number(isOnline(b.name).value) },
  { title: i18n.global.t('client.connections'), key: 'connections', width: 74, sorter: (a: ClientRow, b: ClientRow) => activeConns(a.name) - activeConns(b.name) },
  { title: i18n.global.t('client.ips'), key: 'ips', width: 72, sorter: (a: ClientRow, b: ClientRow) => activeIPs(a.name) - activeIPs(b.name) },
  // Sorted by what was used, not by the quota: with unlimited clients every
  // volume is 0 and the column did not sort at all (#1277).
  { title: i18n.global.t('stats.volume'), dataIndex: 'volume', key: 'volume', width: 135, sorter: (a: ClientRow, b: ClientRow) => (a.up + a.down) - (b.up + b.down) },
  { title: i18n.global.t('client.speed'), key: 'speed', width: 210, sorter: (a: ClientRow, b: ClientRow) => (speedValue(a.name, 'up') + speedValue(a.name, 'down')) - (speedValue(b.name, 'up') + speedValue(b.name, 'down')) },
  { title: i18n.global.t('actions.action'), key: 'actions', width: 146 },
  { title: i18n.global.t('date.lastOnline'), dataIndex: 'onlineAt', key: 'onlineAt', width: 125, sorter: (a: ClientRow, b: ClientRow) => a.onlineAt - b.onlineAt },
  // The description stays at the far right so the operational columns keep
  // their fixed positions.
  { title: i18n.global.t('client.desc'), dataIndex: 'desc', key: 'desc', width: 120, sorter: (a: ClientRow, b: ClientRow) => (a.desc ?? '').localeCompare(b.desc ?? '') },
]

const searchQuery = ref('')
const displayedClients = computed((): ClientRow[] => {
  let list = filterSettings.value.enabled ? filterSettings.value.filteredClients : clients.value
  const query = searchQuery.value.trim().toLowerCase()
  if (query) {
    list = list.filter(c =>
      (c.name ?? '').toLowerCase().includes(query)
      || (c.group ?? '').toLowerCase().includes(query)
      || (c.desc ?? '').toLowerCase().includes(query)
      || (c.remark ?? '').toLowerCase().includes(query))
  }
  return list
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

// Pagination and the sorted data source are tracked explicitly so the poll
// requests exactly the rows the table is showing. Ant's sorter runs inside the
// table, hence tableData follows `extra.currentDataSource` from @change.
const currentPage = ref(1)
const tableData = ref<ClientRow[]>([])
const clientSignature = computed((): string => displayedClients.value.map(c => c.name).join('\u0000'))
const pageNames = computed((): string[] => {
  const start = (currentPage.value - 1) * pageSize.value
  return tableData.value.slice(start, start + pageSize.value).map(c => c.name)
})
const pageNamesKey = computed((): string => pageNames.value.join('\u0000'))

const paginationConfig = computed(() => {
  const total = displayedClients.value.length
  if (total <= 10) return false
  return {
    current: currentPage.value,
    pageSize: pageSize.value,
    total,
    showSizeChanger: true,
    pageSizeOptions: ['10', '25', '50', '100'],
    showQuickJumper: true,
    showTotal: (count: number, range: [number, number]) => `${range[0]}-${range[1]} / ${count}`,
    onChange: (page: number) => {
      currentPage.value = page
    },
    onShowSizeChange: (_: number, size: number) => {
      setItemPerPage(size)
      currentPage.value = 1
    },
  }
})

const selectedNames = ref<string[]>([])
const rowSelection = computed(() => ({
  selectedRowKeys: selectedNames.value,
  preserveSelectedRowKeys: true,
  onChange: (keys: (string | number)[]) => {
    selectedNames.value = keys.map(String)
  },
}))

watch(searchQuery, () => {
  currentPage.value = 1
})

let speedBaseline: {
  instance: string
  sampledAt: number
  users: Record<string, { up: number, down: number }>
} | null = null
let speedTimer: ReturnType<typeof setInterval> | null = null
let speedRestartTimer: ReturnType<typeof setTimeout> | null = null
let speedInFlight = false
let speedInFlightSeq = 0
let speedSeq = 0

const clearSpeed = () => {
  speedBaseline = null
  speedRates.value = {}
}

const resetSpeedSampling = () => {
  speedSeq++
  speedInFlight = false
  clearSpeed()
}

const stopSpeedPolling = () => {
  if (speedTimer !== null) {
    clearInterval(speedTimer)
    speedTimer = null
  }
}

const fetchSpeed = async () => {
  if (speedInFlight || document.visibilityState !== 'visible') return
  const names = pageNames.value
  if (names.length === 0) {
    clearSpeed()
    return
  }
  const seq = ++speedSeq
  speedInFlight = true
  speedInFlightSeq = seq
  try {
    const response = await api.get<{ success: boolean, obj: UserSpeedResponse }>('api/userRates', {
      params: { users: JSON.stringify(names) },
      timeout: 5000,
    })
    if (seq !== speedSeq) return
    const msg = response.data
    if (!msg?.success || !msg.obj || !msg.obj.running || !msg.obj.instance) {
      clearSpeed()
      return
    }
    const snapshot = msg.obj
    const previous = speedBaseline
    if (!previous || previous.instance !== snapshot.instance) {
      const next: Record<string, SpeedSample> = {}
      for (const name of names) {
        next[name] = { up: null, down: null }
      }
      speedBaseline = { instance: snapshot.instance, sampledAt: snapshot.sampledAt, users: snapshot.users ?? {} }
      speedRates.value = next
      return
    }
    const seconds = (snapshot.sampledAt - previous.sampledAt) / 1000
    const next: Record<string, SpeedSample> = {}
    for (const name of names) {
      const before = previous.users[name]
      const current = snapshot.users?.[name]
      if (!before || !current || seconds <= 0) {
        next[name] = { up: null, down: null }
        continue
      }
      next[name] = {
        up: speedFromBytes(current.up - before.up, seconds),
        down: speedFromBytes(current.down - before.down, seconds),
      }
    }
    speedBaseline = { instance: snapshot.instance, sampledAt: snapshot.sampledAt, users: snapshot.users ?? {} }
    speedRates.value = next
  } catch {
    // A failed poll (core down, network, canceled request) clears the baseline
    // so the next success is a fresh baseline and cannot produce a spike.
    if (seq === speedSeq) clearSpeed()
  } finally {
    if (speedInFlightSeq === seq) speedInFlight = false
  }
}

const startSpeedPolling = () => {
  stopSpeedPolling()
  resetSpeedSampling()
  if (document.visibilityState !== 'visible') return
  void fetchSpeed()
  speedTimer = setInterval(() => { void fetchSpeed() }, 2000)
}

const scheduleSpeedRestart = () => {
  if (speedRestartTimer !== null) clearTimeout(speedRestartTimer)
  speedRestartTimer = setTimeout(() => {
    speedRestartTimer = null
    startSpeedPolling()
  }, 0)
}

const onTableChange = (pagination: unknown, _filters: unknown, _sorter: unknown, extra: unknown) => {
  const page = pagination as { current?: number, pageSize?: number } | undefined
  const nextSize = Number(page?.pageSize)
  if (nextSize > 0 && nextSize !== pageSize.value) setItemPerPage(nextSize)
  const nextPage = Number(page?.current)
  if (nextPage > 0) currentPage.value = nextPage
  const source = (extra as { currentDataSource?: ClientRow[] } | undefined)?.currentDataSource
  if (Array.isArray(source)) {
    tableData.value = source
  }
}

const onVisibilityChange = () => {
  if (document.visibilityState === 'visible') {
    startSpeedPolling()
  } else {
    stopSpeedPolling()
    resetSpeedSampling()
  }
}

watch(clientSignature, (signature, previous) => {
  if (signature !== previous || tableData.value.length === 0) {
    tableData.value = displayedClients.value
    currentPage.value = 1
  }
  scheduleSpeedRestart()
}, { immediate: true })

watch(pageNamesKey, () => scheduleSpeedRestart())

// immediate: opening the page or changing page/filter/sort asks for exactly
// the rows the table now shows. Cached outbounds answer without probing.
watch(pageNamesKey, () => scheduleAddressFetch(), { immediate: true })

onMounted(() => {
  document.addEventListener('visibilitychange', onVisibilityChange)
  scheduleSpeedRestart()
})

onUnmounted(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
  stopSpeedPolling()
  if (speedRestartTimer !== null) {
    clearTimeout(speedRestartTimer)
    speedRestartTimer = null
  }
  resetSpeedSampling()
  if (addressTimer !== null) {
    clearTimeout(addressTimer)
    addressTimer = null
  }
  // Drop late responses after the page is gone.
  addressSeq++
})

const modal = ref({
  visible: false,
  id: 0,
})

const delOverlay = ref(new Array<boolean>(clients.value.length).fill(false))
const batchDeleteModal = ref(false)
const batchDeleteLoading = ref(false)

const confirmBatchDelete = () => {
  if (selectedNames.value.length === 0) return
  batchDeleteModal.value = true
}

const deleteSelected = async () => {
  const selected = new Set(selectedNames.value)
  const ids = clients.value.filter(c => selected.has(c.name)).map(c => c.id)
  if (ids.length === 0) {
    batchDeleteModal.value = false
    return
  }
  batchDeleteLoading.value = true
  const success = await Data().save("clients", "delbulk", ids)
  batchDeleteLoading.value = false
  if (success) {
    selectedNames.value = []
    batchDeleteModal.value = false
  }
}

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
  const name = clients.value[index]?.name
  const success = await Data().save("clients", "del", id)
  if (success) {
    delOverlay.value[index] = false
    selectedNames.value = selectedNames.value.filter(n => n !== name)
  }
}

const qrcode = ref({
  visible: false,
  id: 0,
})

const copyClientSubscription = async (id: number) => {
  const msg = await HttpUtils.post<{ link: string }>('api/enableClientSubscription', { clientId: id })
  if (msg.success && msg.obj?.link) {
    try {
      await navigator.clipboard.writeText(msg.obj.link)
      push.success({ title: i18n.global.t('success'), message: i18n.global.t('copyToClipboard') })
    } catch {
      push.error({ title: i18n.global.t('failed'), message: i18n.global.t('copyToClipboard') })
    }
  }
}

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

<style scoped>
.client-search-col {
  flex: 1 1 220px;
  max-width: 360px;
}
.client-search {
  width: 100%;
}
.client-table-footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: rgba(0, 0, 0, 0.65);
  font-size: 12px;
}
.client-selected-count {
  color: #1677ff;
}
.address-cell {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-wrap: nowrap;
}
.address-text {
  min-width: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
  user-select: all;
}
.address-muted {
  color: rgba(0, 0, 0, 0.45);
}
.date-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}
.date-cell :deep(.ant-tag) {
  margin-inline-end: 0;
  font-size: 11px;
  line-height: 18px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
</style>

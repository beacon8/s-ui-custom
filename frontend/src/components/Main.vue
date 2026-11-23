<template>
  <LogVue
    v-model="logModal.visible"
    :control="logModal"
    :visible="logModal.visible"
  />
  <Backup
    v-model="backupModal.visible"
    :control="backupModal"
    :visible="backupModal.visible"
  />
  <UsageStats v-model:visible="usageStatsModal.visible" />
  <a-modal
    v-model:open="menu"
    :title="$t('main.tiles')"
    :width="800"
    :footer="null"
  >
    <a-flex
      vertical
      gap="small"
    >
      <a-card
        v-for="(items, mi) in menuItems"
        :key="mi"
        size="small"
        :title="items.title"
      >
        <a-row :gutter="[12, 12]">
          <a-col
            v-for="item in items.value"
            :key="item.title ?? item.value"
            :xs="24"
            :md="12"
            :lg="6"
          >
            <a-space>
              <a-switch
                :checked="reloadItems.includes(item.value)"
                @change="(checked: boolean | string | number) => toggleReloadItem(item.value, Boolean(checked))"
              />
              <span>{{ item.title }}</span>
            </a-space>
          </a-col>
        </a-row>
      </a-card>
    </a-flex>
  </a-modal>
  <a-spin :spinning="loading">
    <div
      style="padding: 16px;"
      :style="reloadItems.length === 0 ? { minHeight: '70vh', display: 'flex', flexDirection: 'column', justifyContent: 'center' } : {}"
    >
      <a-flex
        justify="center"
        align="center"
        style="margin-bottom: 16px;"
      >
        <img
          src="@/assets/logo.svg"
          :width="reloadItems.length > 0 ? 100 : 200"
          alt="logo"
        >
      </a-flex>
      <a-flex
        justify="center"
        align="center"
        wrap="wrap"
        gap="small"
        style="margin-bottom: 20px;"
      >
        <a-button
          type="default"
          @click="menu = true"
        >
          <template #icon>
            <AppIcon name="star-plus" />
          </template>
          {{ $t('main.tiles') }}
        </a-button>
        <a-button
          type="default"
          @click="backupModal.visible = true"
        >
          <template #icon>
            <AppIcon name="backup-restore" />
          </template>
          {{ $t('main.backup.title') }}
        </a-button>
        <a-button
          type="default"
          @click="logModal.visible = true"
        >
          <template #icon>
            <AppIcon name="list-box-outline" />
          </template>
          {{ $t('basic.log.title') }}
        </a-button>
        <a-button
          type="default"
          @click="usageStatsModal.visible = true"
        >
          <template #icon>
            <AppIcon name="chart-box-outline" />
          </template>
          {{ $t('main.stats.title') }}
        </a-button>
      </a-flex>
      <a-row :gutter="[16, 16]">
        <a-col
          v-for="i in reloadItems"
          :key="i"
          :xs="24"
          :sm="12"
          :md="6"
        >
          <a-card
            size="small"
            style="height: 210px;"
            :body-style="{ padding: '8px 16px' }"
          >
            <template #title>
              <a-space :size="6">
                <span>{{ menuItems.flatMap(cat => cat.value).find(m => m.value == i)?.title }}</span>
                <a-tooltip
                  v-if="i == 'i-sys'"
                  :title="$t('actions.update')"
                  placement="top"
                >
                  <a-button
                    type="text"
                    size="small"
                    :aria-label="$t('actions.update')"
                    @click="reloadSys()"
                  >
                    <template #icon>
                      <AppIcon
                        name="update"
                        color="primary"
                      />
                    </template>
                  </a-button>
                </a-tooltip>
                <a-tooltip
                  v-if="i == 'h-net'"
                  :title="'↓' + HumanReadable.sizeFormat(tilesData.net?.recv) + ' - ' + HumanReadable.sizeFormat(tilesData.net?.sent) + '↑'"
                  placement="top"
                >
                  <AppIcon
                    name="information"
                    color="primary"
                  />
                </a-tooltip>
              </a-space>
            </template>
            <div style="text-align: center;">
              <Gauge
                v-if="i.charAt(0) == 'g'"
                :tiles-data="tilesData"
                :type="i"
              />
              <History
                v-if="i.charAt(0) == 'h'"
                :tiles-data="tilesData"
                :type="i"
              />
              <template v-if="i == 'i-sys'">
                <a-row
                  :gutter="[8, 6]"
                  align="middle"
                  style="text-align: start;"
                >
                  <a-col :span="6">
                    {{ $t('main.info.host') }}
                  </a-col>
                  <a-col
                    :span="18"
                    style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                  >
                    {{ tilesData.sys?.hostName }}
                  </a-col>
                  <a-col :span="6">
                    {{ $t('main.info.cpu') }}
                  </a-col>
                  <a-col :span="18">
                    <a-tooltip placement="top">
                      <template #title>
                        <span style="direction: ltr;">{{ tilesData.sys?.cpuType }}</span>
                      </template>
                      <a-tag>
                        {{ tilesData.sys?.cpuCount }} {{ $t('main.info.core') }}
                      </a-tag>
                    </a-tooltip>
                  </a-col>
                  <a-col :span="6">
                    IP
                  </a-col>
                  <a-col :span="18">
                    <a-tooltip
                      v-if="tilesData.sys?.ipv4?.length > 0"
                      placement="top"
                    >
                      <template #title>
                        <span style="white-space: pre-line; direction: ltr;">{{ (tilesData.sys?.ipv4 ?? []).join('\n') }}</span>
                      </template>
                      <a-tag color="blue">
                        IPv4
                      </a-tag>
                    </a-tooltip>
                    <a-tooltip
                      v-if="tilesData.sys?.ipv6?.length > 0"
                      placement="top"
                    >
                      <template #title>
                        <span style="white-space: pre-line; direction: ltr;">{{ (tilesData.sys?.ipv6 ?? []).join('\n') }}</span>
                      </template>
                      <a-tag color="blue">
                        IPv6
                      </a-tag>
                    </a-tooltip>
                  </a-col>
                  <a-col :span="6">
                    S-UI
                  </a-col>
                  <a-col :span="18">
                    <a-tooltip
                      :title="tilesData.sys?.coreVersion ? 'Sing-Box v' + tilesData.sys?.coreVersion : ''"
                      placement="top"
                    >
                      <a-tag color="blue">
                        v{{ tilesData.sys?.appVersion }}
                      </a-tag>
                    </a-tooltip>
                    <a-tooltip
                      v-if="Data().update"
                      :title="$t('update.hint', { current: Data().update?.current })"
                      placement="top"
                    >
                      <a-tag
                        color="processing"
                        style="cursor: pointer;"
                        @click="Data().updateVisible = true"
                      >
                        <template #icon>
                          <AppIcon name="update" />
                        </template>
                        v{{ Data().update?.latest }}
                      </a-tag>
                    </a-tooltip>
                  </a-col>
                  <a-col :span="6">
                    {{ $t('main.info.uptime') }}
                  </a-col>
                  <a-col :span="18">
                    <a-tooltip
                      :title="$t('main.info.startupTime') + ': ' + new Date((tilesData.sys?.bootTime || 0) * 1000).toLocaleString(locale)"
                      placement="top"
                    >
                      <span>{{ HumanReadable.formatSecond((Date.now() / 1000) - tilesData.sys?.bootTime) }}</span>
                    </a-tooltip>
                  </a-col>
                </a-row>
              </template>
              <template v-if="i == 'i-sbd'">
                <a-row
                  :gutter="[8, 6]"
                  align="middle"
                  style="text-align: start;"
                >
                  <a-col :span="8">
                    {{ $t('main.info.running') }}
                  </a-col>
                  <a-col :span="16">
                    <a-space :size="4">
                      <a-tag
                        v-if="tilesData.sbd?.running"
                        color="green"
                      >
                        {{ $t('yes') }}
                      </a-tag>
                      <!-- A core stopped on purpose reads the same as a crashed
                           one here, and the operator has to be able to tell. -->
                      <a-tag
                        v-else-if="tilesData.sbd?.maintenance"
                        color="orange"
                      >
                        {{ $t('setting.maintenance') }}
                      </a-tag>
                      <a-tag
                        v-else
                        color="red"
                      >
                        {{ $t('no') }}
                      </a-tag>
                      <a-tooltip
                        v-if="tilesData.sbd?.running && !loading"
                        :title="$t('actions.restartSb')"
                        placement="top"
                      >
                        <a-button
                          type="text"
                          size="small"
                          @click="restartSingbox()"
                        >
                          <template #icon>
                            <AppIcon
                              name="restart"
                              color="warning"
                            />
                          </template>
                        </a-button>
                      </a-tooltip>
                    </a-space>
                  </a-col>
                  <a-col :span="8">
                    {{ $t('main.info.memory') }}
                  </a-col>
                  <a-col :span="16">
                    <a-tag
                      v-if="tilesData.sbd?.stats?.Alloc"
                      color="blue"
                    >
                      {{ HumanReadable.sizeFormat(tilesData.sbd?.stats?.Alloc) }}
                    </a-tag>
                  </a-col>
                  <a-col :span="8">
                    {{ $t('main.info.threads') }}
                  </a-col>
                  <a-col :span="16">
                    <a-tag
                      v-if="tilesData.sbd?.stats?.NumGoroutine"
                      color="blue"
                    >
                      {{ tilesData.sbd?.stats?.NumGoroutine }}
                    </a-tag>
                  </a-col>
                  <a-col :span="8">
                    {{ $t('main.info.uptime') }}
                  </a-col>
                  <a-col :span="16">
                    {{ HumanReadable.formatSecond(tilesData.sbd?.stats?.Uptime) }}
                  </a-col>
                  <a-col :span="8">
                    {{ $t('online') }}
                  </a-col>
                  <a-col :span="16">
                    <template v-if="tilesData.sbd?.running">
                      <a-tooltip
                        v-if="Data().onlines.user"
                        placement="top"
                      >
                        <template #title>
                          <div style="max-height: 200px; overflow-y: auto;">
                            <strong>{{ $t('pages.clients') }}</strong><br>
                            <span
                              v-for="user in Data().onlines.user"
                              :key="user"
                            >{{ user }}<br></span>
                          </div>
                        </template>
                        <a-tag color="blue">
                          {{ Data().onlines.user?.length }}
                        </a-tag>
                      </a-tooltip>
                      <a-tooltip
                        v-if="Data().onlines.inbound"
                        placement="top"
                      >
                        <template #title>
                          <div style="max-height: 200px; overflow-y: auto;">
                            <strong>{{ $t('pages.inbounds') }}</strong><br>
                            <span
                              v-for="tag in Data().onlines.inbound"
                              :key="tag"
                            >{{ tag }}<br></span>
                          </div>
                        </template>
                        <a-tag color="green">
                          {{ Data().onlines.inbound?.length }}
                        </a-tag>
                      </a-tooltip>
                      <a-tooltip
                        v-if="Data().onlines.outbound"
                        placement="top"
                      >
                        <template #title>
                          <div style="max-height: 200px; overflow-y: auto;">
                            <strong>{{ $t('pages.outbounds') }}</strong><br>
                            <span
                              v-for="o in Data().onlines.outbound"
                              :key="o"
                            >{{ o }}<br></span>
                          </div>
                        </template>
                        <a-tag color="processing">
                          {{ Data().onlines.outbound?.length }}
                        </a-tag>
                      </a-tooltip>
                    </template>
                  </a-col>
                </a-row>
              </template>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </div>
  </a-spin>
</template>

<script lang="ts" setup>
import HttpUtils from '@/plugins/httputil'
import { HumanReadable } from '@/plugins/utils'
import Data from '@/store/modules/data'
import Gauge from '@/components/tiles/Gauge.vue'
import History from '@/components/tiles/History.vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { i18n, locale } from '@/locales'
import LogVue from '@/layouts/modals/Logs.vue'
import Backup from '@/layouts/modals/Backup.vue'
import UsageStats from '@/layouts/modals/UsageStats.vue'

const loading = ref(false)
const menu = ref(false)
const menuItems = [
  { title: i18n.global.t('main.gauges'), value: [
    { title: i18n.global.t('main.gauge.cpu'), value: "g-cpu" },
    { title: i18n.global.t('main.gauge.mem'), value: "g-mem" },
    { title: i18n.global.t('main.gauge.dsk'), value: "g-dsk" },
    { title: i18n.global.t('main.gauge.swp'), value: "g-swp" },
    ]
  },
  { title: i18n.global.t('main.charts'), value: [
    { title: i18n.global.t('main.chart.cpu'), value: "h-cpu" },
    { title: i18n.global.t('main.chart.mem'), value: "h-mem" },
    { title: i18n.global.t('main.chart.net'), value: "h-net" },
    { title: i18n.global.t('main.chart.pnet'), value: "hp-net" },
    { title: i18n.global.t('main.chart.dio'), value: "h-dio" },
    ]
  },
  { title: i18n.global.t('main.infos'), value: [
    { title: i18n.global.t('main.info.sys'), value: "i-sys" },
    { title: i18n.global.t('main.info.sbd'), value: "i-sbd" },
    ]
  },
]

// What api/status answers with. The panel asks for only the sections it is
// showing, so a section is absent until its first reload; every reader below
// and in the template guards with ?. before it reaches in.
interface Usage {
  current: number
  total: number
}

interface NetUsage {
  recv: number
  sent: number
  precv: number
  psent: number
}

interface DiskIo {
  read: number
  write: number
}

interface SysInfo {
  hostName: string
  cpuType: string
  cpuCount: number
  ipv4: string[]
  ipv6: string[]
  appVersion: string
  // The linked sing-box module; absent from a binary built without module info
  coreVersion?: string
  bootTime: number
}

interface SingBoxInfo {
  running: boolean
  maintenance: boolean
  stats: {
    Alloc: number
    NumGoroutine: number
    Uptime: number
  }
}

interface TilesData {
  cpu: number
  mem: Usage
  dsk: Usage
  swp: Usage
  net: NetUsage
  dio: DiskIo
  sys: SysInfo
  sbd: SingBoxInfo
}

const tilesData = ref(<TilesData>{})

const reloadItems = computed({
  get() { return Data().reloadItems },
  set(v:string[]) {
    if (Data().reloadItems.length == 0 && v.length>0) startTimer()
    if (Data().reloadItems.length > 0 && v.length == 0) stopTimer()
    Data().reloadItems = v
    if (v.length > 0) localStorage.setItem("reloadItems", v.join(','))
    else localStorage.removeItem("reloadItems")
  }
})

const toggleReloadItem = (val: string, checked: boolean) => {
  const current = reloadItems.value
  if (checked) {
    if (!current.includes(val)) {
      reloadItems.value = [...current, val]
    }
  } else {
    reloadItems.value = current.filter(item => item !== val)
  }
}

const reloadData = async () => {
  const request = [...new Set(reloadItems.value.map(r => r.split('-')[1]))]
  if (tilesData.value?.sys?.appVersion) request.filter(r => r != 'sys')
  const data = await HttpUtils.get<TilesData>('api/status',{ r: request.join(',')})
  if (data.success) {
    tilesData.value = data.obj
  }
}

const reloadSys = async () => {
  const data = await HttpUtils.get<TilesData>('api/status',{ r: 'sys'})
  if (data.success) {
    tilesData.value.sys = data.obj.sys
  }
}

// The next poll is scheduled only once the previous one has answered. A fixed
// interval fired again while a slow api/status was still in flight, and the
// duplicate cancelled it. The generation stops a poll that was in flight
// across stopTimer/startTimer from starting a second chain.
let timerId: ReturnType<typeof setTimeout> | null = null
let pollGen = 0

const startTimer = () => {
  const gen = ++pollGen
  const tick = () => {
    timerId = setTimeout(async () => {
      await reloadData()
      if (gen === pollGen) tick()
    }, 2000)
  }
  tick()
}

const stopTimer = () => {
  pollGen++
  if (timerId) {
    clearTimeout(timerId)
    timerId = null
  }
}

onMounted(async () => {
  loading.value = true
  if (Data().reloadItems.length != 0) {
    await reloadData()
    startTimer()
  }
  loading.value = false
})

onBeforeUnmount(() => {
  stopTimer()
})

const logModal = ref({ visible: false })

const backupModal = ref({ visible: false })

const usageStatsModal = ref({ visible: false })

const restartSingbox = async () => {
  loading.value = true
  await HttpUtils.post('api/restartSb',{})
  loading.value = false
}
</script>

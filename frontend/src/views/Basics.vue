<template>
  <a-row
    :gutter="[12, 12]"
    justify="center"
    style="margin-bottom: 12px;"
  >
    <a-col>
      <a-button
        type="primary"
        :loading="loading"
        :disabled="stateChange"
        @click="saveConfig"
      >
        {{ $t('actions.save') }}
      </a-button>
    </a-col>
  </a-row>
  <HttpClientVue
    v-model="httpClientModal.visible"
    :visible="httpClientModal.visible"
    :index="httpClientModal.index"
    :data="httpClientModal.data"
    :tags="httpClientTags"
    @close="httpClientModal.visible = false"
    @save="saveHttpClient"
  />
  <a-collapse>
    <a-collapse-panel
      key="log"
      :header="$t('basic.log.title')"
    >
      <template #extra>
        <DocLink
          :href="LOG_DOC"
          @click.stop
        />
      </template>
      <a-form layout="vertical">
        <a-row
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-space>
              <a-switch v-model:checked="appConfig.log.disabled" />
              <span>{{ $t('disable') }}</span>
            </a-space>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              :label="$t('basic.log.level')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="appConfig.log.level"
                allow-clear
                style="width: 100%"
                @change="(val: any) => { if (val === undefined) delete appConfig.log.level }"
              >
                <a-select-option
                  v-for="lvl in levels"
                  :key="lvl"
                  :value="lvl"
                >
                  {{ lvl }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              :label="$t('basic.log.output')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="appConfig.log.output" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-space>
              <a-switch v-model:checked="appConfig.log.timestamp" />
              <span>{{ $t('basic.log.timestamp') }}</span>
            </a-space>
          </a-col>
        </a-row>
      </a-form>
    </a-collapse-panel>
    <a-collapse-panel
      key="ntp"
      header="NTP"
    >
      <template #extra>
        <DocLink
          :href="NTP_DOC"
          @click.stop
        />
      </template>
      <a-form layout="vertical">
        <a-row
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-space>
              <a-switch v-model:checked="enableNtp" />
              <span>{{ $t('enable') }}</span>
            </a-space>
          </a-col>
          <a-col
            v-if="appConfig.ntp?.enabled"
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              :label="$t('out.addr')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="appConfig.ntp.server" />
            </a-form-item>
          </a-col>
          <a-col
            v-if="appConfig.ntp?.enabled"
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              :label="$t('out.port')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="appConfig.ntp.server_port"
                style="width: 100%"
                @change="(val: any) => { if (val === null || val === undefined || val === '') delete appConfig.ntp?.server_port }"
              />
            </a-form-item>
          </a-col>
          <a-col
            v-if="appConfig.ntp?.enabled"
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              :label="$t('ruleset.interval')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="ntpInterval"
                :min="0"
                :addon-after="$t('date.m')"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
      <Dial
        v-if="appConfig.ntp?.enabled"
        :dial="appConfig.ntp"
      />
    </a-collapse-panel>
    <a-collapse-panel
      key="httpClient"
      :header="$t('basic.httpClient.title')"
    >
      <!-- Shared transports for everything sing-box downloads: remote
           rule-sets, the API dashboard, certificate providers. Each is
           referenced by its tag. -->
      <a-row
        :gutter="[12, 12]"
        style="margin-bottom: 12px;"
      >
        <a-col :span="24">
          <a-button
            type="primary"
            @click="showHttpClientModal(-1)"
          >
            {{ $t('actions.add') }}
          </a-button>
        </a-col>
      </a-row>
      <a-table
        :columns="httpClientHeaders"
        :data-source="httpClients"
        :pagination="false"
        :locale="{ emptyText: $t('basic.httpClient.none') }"
        row-key="tag"
        size="small"
      >
        <template #bodyCell="{ column, text, index }">
          <template v-if="column.key === 'version'">
            {{ versionName(text) }}
          </template>
          <template v-else-if="column.key === 'detour'">
            {{ text ?? '-' }}
          </template>
          <template v-else-if="column.key === 'engine'">
            {{ text ?? '-' }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space :size="4">
              <a-tooltip :title="$t('actions.edit')">
                <a-button
                  type="text"
                  shape="circle"
                  size="small"
                  :aria-label="$t('actions.edit')"
                  @click="showHttpClientModal(index)"
                >
                  <template #icon>
                    <AppIcon name="file-edit" />
                  </template>
                </a-button>
              </a-tooltip>
              <a-tooltip :title="$t('actions.del')">
                <a-button
                  type="text"
                  shape="circle"
                  size="small"
                  danger
                  :aria-label="$t('actions.del')"
                  @click="delHttpClient(index)"
                >
                  <template #icon>
                    <AppIcon name="file-remove" />
                  </template>
                </a-button>
              </a-tooltip>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-collapse-panel>
    <a-collapse-panel
      key="experimental"
      header="Experimental"
    >
      <template #extra>
        <DocLink
          :href="EXPERIMENTAL_DOC"
          @click.stop
        />
      </template>
      <a-form layout="vertical">
        <a-divider
          orientation="left"
          style="margin: 8px 0 12px;"
        >
          Cache File <DocLink :href="CACHE_FILE_DOC" />
        </a-divider>
        <a-row
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-space>
              <a-switch v-model:checked="enableCacheFile" />
              <span>{{ $t('enable') }}</span>
            </a-space>
          </a-col>
          <a-col
            v-if="appConfig.experimental.cache_file"
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              :label="$t('transport.path')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="appConfig.experimental.cache_file.path" />
            </a-form-item>
          </a-col>
          <a-col
            v-if="appConfig.experimental.cache_file"
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              label="Cache ID"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="appConfig.experimental.cache_file.cache_id" />
            </a-form-item>
          </a-col>
          <a-col
            v-if="appConfig.experimental.cache_file"
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-space>
              <a-switch v-model:checked="appConfig.experimental.cache_file.store_fakeip" />
              <span>{{ $t('basic.exp.storeFakeIp') }}</span>
            </a-space>
          </a-col>
        </a-row>
        <a-divider
          orientation="left"
          style="margin: 16px 0 12px;"
        >
          Clash API <DocLink :href="CLASH_API_DOC" />
        </a-divider>
        <a-row
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-space>
              <a-switch v-model:checked="enableClashApi" />
              <span>{{ $t('enable') }}</span>
            </a-space>
          </a-col>
          <template v-if="appConfig.experimental.clash_api">
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
              :lg="4"
            >
              <a-form-item
                :label="$t('basic.exp.extController')"
                style="margin-bottom: 0;"
              >
                <a-input v-model:value="appConfig.experimental.clash_api.external_controller" />
              </a-form-item>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
              :lg="4"
            >
              <a-form-item
                :label="$t('basic.exp.secret')"
                style="margin-bottom: 0;"
              >
                <a-input v-model:value="appConfig.experimental.clash_api.secret" />
              </a-form-item>
            </a-col>
          </template>
        </a-row>
        <a-row
          v-if="appConfig.experimental.clash_api"
          :gutter="[12, 12]"
          align="middle"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              :label="$t('basic.exp.extUi')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="appConfig.experimental.clash_api.external_ui" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="16"
            :md="8"
          >
            <a-form-item
              :label="$t('basic.exp.extUiDownloadUrl')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="appConfig.experimental.clash_api.external_ui_download_url" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              :label="$t('basic.exp.extUiDownloadDetour')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="appConfig.experimental.clash_api.external_ui_download_detour"
                allow-clear
                style="width: 100%"
                @change="(val: any) => { if (val === undefined && appConfig.experimental.clash_api) delete appConfig.experimental.clash_api.external_ui_download_detour }"
              >
                <a-select-option
                  v-for="tag in outboundTags"
                  :key="tag"
                  :value="tag"
                >
                  {{ tag }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-if="appConfig.experimental.clash_api"
          :gutter="[12, 12]"
          align="middle"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-form-item
              :label="$t('basic.exp.defaultMode')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="appConfig.experimental.clash_api.default_mode" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="16"
            :md="8"
          >
            <a-form-item
              :label="$t('basic.exp.allowOrigin') + ' ' + $t('commaSeparated')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="origin" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-space>
              <a-switch v-model:checked="appConfig.experimental.clash_api.access_control_allow_private_network" />
              <span>{{ $t('basic.exp.allowPrivate') }}</span>
            </a-space>
          </a-col>
        </a-row>
        <a-divider
          orientation="left"
          style="margin: 16px 0 12px;"
        >
          V2Ray API <DocLink :href="V2RAY_API_DOC" />
        </a-divider>
        <a-row
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-space>
              <a-switch v-model:checked="enableV2rayApi" />
              <span>{{ $t('enable') }}</span>
            </a-space>
          </a-col>
          <template v-if="appConfig.experimental.v2ray_api">
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
              :lg="4"
            >
              <a-form-item
                :label="$t('objects.listen')"
                style="margin-bottom: 0;"
              >
                <a-input v-model:value="appConfig.experimental.v2ray_api.listen" />
              </a-form-item>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
              :lg="4"
            >
              <a-space>
                <a-switch v-model:checked="appConfig.experimental.v2ray_api.stats.enabled" />
                <span>{{ $t('stats.enable') }}</span>
              </a-space>
            </a-col>
          </template>
        </a-row>
        <a-row
          v-if="appConfig.experimental.v2ray_api?.stats?.enabled"
          :gutter="[12, 12]"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('pages.inbounds')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="appConfig.experimental.v2ray_api.stats.inbounds"
                mode="multiple"
                style="width: 100%"
              >
                <a-select-option
                  v-for="tag in inboundTags"
                  :key="tag"
                  :value="tag"
                >
                  {{ tag }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('pages.outbounds')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="appConfig.experimental.v2ray_api.stats.outbounds"
                mode="multiple"
                style="width: 100%"
              >
                <a-select-option
                  v-for="tag in outboundTags"
                  :key="tag"
                  :value="tag"
                >
                  {{ tag }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('pages.clients')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="appConfig.experimental.v2ray_api.stats.users"
                mode="multiple"
                style="width: 100%"
              >
                <a-select-option
                  v-for="name in clientNames"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-divider
          orientation="left"
          style="margin: 16px 0 12px;"
        >
          Debug
        </a-divider>
        <a-row
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
            :lg="4"
          >
            <a-space>
              <a-switch v-model:checked="enableDebug" />
              <span>{{ $t('enable') }}</span>
            </a-space>
          </a-col>
          <template v-if="appConfig.experimental.debug">
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
              :lg="4"
            >
              <a-form-item
                :label="$t('objects.listen')"
                style="margin-bottom: 0;"
              >
                <a-input
                  v-model:value="appConfig.experimental.debug.listen"
                  allow-clear
                  placeholder="127.0.0.1:6060"
                  @change="() => { if (!appConfig.experimental.debug?.listen && appConfig.experimental.debug) delete appConfig.experimental.debug.listen }"
                />
              </a-form-item>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
              :lg="4"
            >
              <a-form-item
                :label="$t('types.oomKiller.memoryLimit')"
                style="margin-bottom: 0;"
              >
                <a-input
                  v-model:value="debugMemoryLimit"
                  allow-clear
                  placeholder="200mb"
                />
              </a-form-item>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
              :lg="4"
            >
              <a-form-item
                :label="$t('basic.exp.gcPercent')"
                style="margin-bottom: 0;"
              >
                <a-input-number
                  v-model:value="debugGcPercent"
                  placeholder="100"
                  style="width: 100%"
                />
              </a-form-item>
            </a-col>
          </template>
        </a-row>
      </a-form>
    </a-collapse-panel>
  </a-collapse>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import Dial from '@/components/Dial.vue'
import HttpClientVue from '@/layouts/modals/HttpClient.vue'
import DocLink from '@/components/DocLink.vue'
import { computed, ref, onBeforeMount } from 'vue'
import { Config, Ntp } from '@/types/config'
import { HttpClient } from '@/types/httpClient'
import { FindDiff } from '@/plugins/utils'
import { i18n } from '@/locales'
import { LOG_DOC, NTP_DOC, EXPERIMENTAL_DOC, CACHE_FILE_DOC, CLASH_API_DOC, V2RAY_API_DOC } from '@/plugins/docs'

const oldConfig = ref({})
const loading = ref(false)

const appConfig = computed((): Config => {
  return <Config> Data().config
})

onBeforeMount(async () => {
  loading.value = true
  while (Data().lastLoad == 0) {
    await new Promise(resolve => setTimeout(resolve, 100))
  }
  oldConfig.value = JSON.parse(JSON.stringify(Data().config))
  loading.value = false
})

const stateChange = computed(() => {
  return FindDiff.deepCompare(appConfig.value,oldConfig.value)
})

const saveConfig = async () => {
  loading.value = true
  const success = await Data().save("config", "set", appConfig.value)
  if (success) {
    oldConfig.value = JSON.parse(JSON.stringify(Data().config))
    loading.value = false
  }
}

const inboundTags = computed((): string[] => {
  return [...(Data().inbounds?.map(i => i.tag) ?? []), ...(Data().endpoints?.filter(e => e.listen_port > 0).map(e => e.tag) ?? [])]
})

const clientNames = computed((): string[] => {
  const clients = Data().clients
  return clients?.map(c => c.name)
})

const outboundTags = computed((): string[] => {
  return [...(Data().outbounds?.map(o => o.tag) ?? []), ...(Data().endpoints?.map(e => e.tag) ?? [])]
})

const levels = ["trace", "debug", "info", "warn", "error", "fatal", "panic"]

// Shared HTTP clients live at the top level of the config, beside log and dns.
const httpClients = computed((): HttpClient[] => {
  const config = appConfig.value
  if (!Array.isArray(config.http_clients)) config.http_clients = []
  return config.http_clients
})

const httpClientTags = computed((): string[] => httpClients.value.map(c => c.tag))

const versionNames: Record<string, string> = { 0: 'Auto', 1: 'HTTP/1.1', 2: 'HTTP/2', 3: 'HTTP/3' }
const versionName = (version?: number): string => versionNames[version ?? 0] ?? 'Auto'

const httpClientHeaders = computed(() => [
  { title: i18n.global.t('objects.tag'), dataIndex: 'tag', key: 'tag' },
  { title: i18n.global.t('basic.httpClient.version'), dataIndex: 'version', key: 'version' },
  { title: i18n.global.t('basic.httpClient.engine'), dataIndex: 'engine', key: 'engine' },
  { title: i18n.global.t('listen.detour'), dataIndex: 'detour', key: 'detour' },
  { title: '', key: 'actions', align: 'right' as const },
])

const httpClientModal = ref({ visible: false, index: -1, data: "" })

const showHttpClientModal = (index: number) => {
  httpClientModal.value.index = index
  httpClientModal.value.data = index == -1 ? '' : JSON.stringify(httpClients.value[index])
  httpClientModal.value.visible = true
}

const saveHttpClient = (data: HttpClient) => {
  if (httpClientModal.value.index == -1) httpClients.value.push(data)
  else httpClients.value[httpClientModal.value.index] = data
  httpClientModal.value.visible = false
}

const delHttpClient = (index: number) => { httpClients.value.splice(index, 1) }

const enableNtp = computed({
  get() { return appConfig.value.ntp?.enabled?? false },
  set(v:boolean) { 
    if (v){
      appConfig.value.ntp = <Ntp>{ enabled: true, server: 'time.apple.com', server_port: 123, interval: '30m'}
    } else { appConfig.value.ntp = <Ntp>{}  }
  }
})

const ntpInterval = computed({
  get(): number | null { return appConfig.value.ntp?.interval? parseInt(appConfig.value.ntp?.interval.replace('m','')) : null },
  set(v:number) {
    if (!appConfig.value.ntp) return
    if (v > 0) appConfig.value.ntp.interval = v + 'm'
    else delete appConfig.value.ntp.interval
  }
})

const enableCacheFile = computed({
  get() { return appConfig.value.experimental.cache_file?.enabled?? false },
  set(v:boolean) { 
    if (v){
      appConfig.value.experimental.cache_file = { enabled: true }
    } else { delete appConfig.value.experimental.cache_file  }
  }
})

const enableClashApi = computed({
  get() { return appConfig.value.experimental.clash_api != undefined },
  set(v:boolean) { appConfig.value.experimental.clash_api = v ? { external_controller: '127.0.0.1:9090' } : undefined }
})

const enableV2rayApi = computed({
  get() { return appConfig.value.experimental.v2ray_api != undefined },
  set(v:boolean) { appConfig.value.experimental.v2ray_api = v ? { listen: '127.0.0.1:8080', stats: { enabled: false, inbounds: [], outbounds: [], users: [] }} : undefined }
})

const enableDebug = computed({
  get() { return appConfig.value.experimental.debug != undefined },
  set(v:boolean) { appConfig.value.experimental.debug = v ? {} : undefined }
})

// sing-box rejects an empty memory_limit, so a cleared field drops the key.
const debugMemoryLimit = computed({
  get() { return appConfig.value.experimental.debug?.memory_limit ?? '' },
  set(v:string|null) {
    if (!appConfig.value.experimental.debug) return
    if (v) appConfig.value.experimental.debug.memory_limit = v
    else delete appConfig.value.experimental.debug.memory_limit
  }
})

const debugGcPercent = computed({
  get() { return appConfig.value.experimental.debug?.gc_percent },
  set(v:number|string|null) {
    if (!appConfig.value.experimental.debug) return
    if (typeof v == 'number') appConfig.value.experimental.debug.gc_percent = v
    else delete appConfig.value.experimental.debug.gc_percent
  }
})

const origin = computed({
  get() { return appConfig.value.experimental.clash_api?.access_control_allow_origin &&
    appConfig.value.experimental.clash_api.access_control_allow_origin.length>0 ? appConfig.value.experimental.clash_api.access_control_allow_origin.join(',') : '' },
  set(v:string) {
    // Guarded on clash_api, not on the field itself. The old test required the
    // list to already exist, so the input silently discarded everything typed
    // into it until one was set some other way.
    if (!appConfig.value.experimental.clash_api) return
    appConfig.value.experimental.clash_api.access_control_allow_origin =
      v.length > 0 ? v.split(',').map(o => o.trim()).filter(o => o.length > 0) : undefined
  }
})
</script>
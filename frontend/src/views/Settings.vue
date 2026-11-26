<template>
  <TokenModal
    v-model="tokenModalVisible"
    :visible="tokenModalVisible"
    @close="tokenModalVisible = false"
  />
  <a-card
    :loading="loading"
    :bordered="false"
    class="settings-card"
  >
    <div class="settings-page">
      <div class="settings-toolbar">
        <div
          class="settings-tabs"
          role="tablist"
        >
          <button
            v-for="item in tabs"
            :key="item.key"
            type="button"
            role="tab"
            class="settings-tab"
            :class="{ active: tab === item.key }"
            :aria-selected="tab === item.key"
            @click="tab = item.key"
          >
            {{ $t(item.label) }}
          </button>
        </div>
        <div class="settings-actions">
          <a-button
            type="default"
            @click="tokenModalVisible = true"
          >
            {{ $t('admin.api.title') }}
          </a-button>
          <a-button
            type="primary"
            :loading="loading"
            :disabled="!stateChange"
            @click="save"
          >
            {{ $t('actions.save') }}
          </a-button>
          <a-button
            type="default"
            :loading="loading"
            :disabled="stateChange"
            @click="restartApp"
          >
            {{ $t('actions.restartApp') }}
          </a-button>
          <!-- The core cannot be held down any other way: a plain stop is undone by
               the watchdog within five seconds. -->
          <a-button
            :type="maintenance ? 'primary' : 'default'"
            :danger="!maintenance"
            :loading="loading"
            :disabled="stateChange"
            @click="toggleMaintenance"
          >
            {{ maintenance ? $t('actions.startCore') : $t('actions.stopCore') }}
          </a-button>
        </div>
      </div>

      <a-alert
        v-if="maintenance"
        type="warning"
        show-icon
        class="maintenance-alert"
        :message="$t('setting.maintenanceOnHint')"
      />

      <a-form
        v-if="tab === 't1'"
        layout="vertical"
        class="settings-form"
      >
        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="10"
          >
            <a-form-item
              :label="$t('setting.addr')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.webListen" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="4"
          >
            <a-form-item
              class="short-field"
              :label="$t('setting.port')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="webPort"
                :min="1"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="24"
            :md="24"
            :lg="10"
          >
            <a-form-item
              :label="$t('setting.webPath')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.webPath" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('setting.sslCert')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.webCertFile" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('setting.sslKey')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.webKeyFile" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('setting.domain')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.webDomain" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('setting.webUri')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.webURI" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="8"
          >
            <a-form-item
              class="short-field"
              :label="$t('setting.sessionAge')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="sessionMaxAge"
                :min="0"
                :addon-after="$t('date.m')"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="8"
          >
            <a-form-item
              class="short-field"
              :label="$t('setting.trafficAge')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="trafficAge"
                :min="0"
                :addon-after="$t('date.d')"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="8"
          >
            <a-tooltip :title="$t('setting.statsBucketSecondsHint')">
              <a-form-item
                class="short-field"
                :label="$t('setting.statsBucketSeconds')"
                style="margin-bottom: 0;"
              >
                <a-input-number
                  v-model:value="statsBucketSeconds"
                  :min="1"
                  :addon-after="$t('date.s')"
                  style="width: 100%;"
                />
              </a-form-item>
            </a-tooltip>
          </a-col>
        </a-row>

        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="8"
          >
            <a-form-item
              class="short-field"
              :label="$t('setting.ipv4Prefix')"
              :extra="$t('setting.ipv4PrefixHint')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="ipv4Prefix"
                :min="1"
                :max="32"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="8"
          >
            <a-form-item
              class="short-field"
              :label="$t('setting.ipv6Prefix')"
              :extra="$t('setting.ipv6PrefixHint')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="ipv6Prefix"
                :min="1"
                :max="128"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="8"
          >
            <a-form-item
              :label="$t('setting.ipLimitDryRun')"
              :extra="$t('setting.ipLimitDryRunHint')"
              style="margin-bottom: 0;"
            >
              <a-switch v-model:checked="ipLimitDryRun" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('setting.timeLoc')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.timeLocation" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-tooltip :title="$t('setting.globalResetHint')">
              <a-form-item
                :label="$t('setting.globalReset')"
                style="margin-bottom: 0;"
              >
                <a-input
                  v-model:value="settings.globalReset"
                  placeholder="0 0 1 * *"
                />
              </a-form-item>
            </a-tooltip>
          </a-col>
        </a-row>
      </a-form>

      <a-form
        v-else-if="tab === 't2'"
        layout="vertical"
        class="settings-form"
      >
        <div class="switch-row">
          <a-space>
            <a-switch v-model:checked="subEncode" />
            <span>{{ $t('setting.subEncode') }}</span>
          </a-space>
          <a-space>
            <a-switch v-model:checked="subShowInfo" />
            <span>{{ $t('setting.subInfo') }}</span>
          </a-space>
          <a-space>
            <a-switch v-model:checked="subRemarkPrefix" />
            <span>{{ $t('setting.subRemarkPrefix') }}</span>
          </a-space>
        </div>

        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="10"
          >
            <a-form-item
              :label="$t('setting.addr')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.subListen" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="12"
            :lg="4"
          >
            <a-form-item
              class="short-field"
              :label="$t('setting.port')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="subPort"
                :min="1"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="24"
            :md="24"
            :lg="10"
          >
            <a-form-item
              :label="$t('setting.path')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.subPath" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('setting.sslCert')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.subCertFile" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('setting.sslKey')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.subKeyFile" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('setting.domain')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.subDomain" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('setting.subUri')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="settings.subURI" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row
          :gutter="[16, 12]"
          class="field-row"
        >
          <a-col
            :xs="24"
            :sm="8"
            :md="6"
            :lg="6"
          >
            <a-form-item
              class="short-field"
              :label="$t('setting.update')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="subUpdates"
                :min="0"
                :addon-after="$t('date.h')"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="16"
            :md="18"
            :lg="18"
          >
            <a-tooltip :title="$t('setting.subCustomHeadersHint')">
              <a-form-item
                :label="$t('setting.subCustomHeaders')"
                style="margin-bottom: 0;"
              >
                <a-textarea
                  v-model:value="settings.subCustomHeaders"
                  :placeholder="'announce: Hello!\nannounce-url: https://example.com/'"
                  :auto-size="{ minRows: 3, maxRows: 10 }"
                />
              </a-form-item>
            </a-tooltip>
          </a-col>
        </a-row>
      </a-form>

      <SubJsonExtVue
        v-else-if="tab === 't3'"
        :settings="settings"
      />

      <SubClashExtVue
        v-else-if="tab === 't4'"
        :settings="settings"
      />
    </div>
  </a-card>
</template>

<script lang="ts" setup>
import { i18n } from '@/locales'
import { Ref, computed, inject, onMounted, ref } from 'vue'
import HttpUtils from '@/plugins/httputil'
import { FindDiff } from '@/plugins/utils'
import SubJsonExtVue from '@/components/SubJsonExt.vue'
import SubClashExtVue from '@/components/SubClashExt.vue'
import TokenModal from '@/layouts/modals/Token.vue'
import { push } from 'notivue'
import Data from '@/store/modules/data'

const tabs = [
  { key: 't1', label: 'setting.interface' },
  { key: 't2', label: 'setting.sub' },
  { key: 't3', label: 'setting.jsonSub' },
  { key: 't4', label: 'setting.clashSub' },
]
const tab = ref("t1")
const tokenModalVisible = ref(false)
const loading: Ref = inject('loading') ?? ref(false)
const oldSettings = ref<Record<string, string>>({})

const settings = ref({
  webListen: "",
  webDomain: "",
  webPort: "2095",
  webCertFile: "",
  webKeyFile: "",
  webPath: "/app/",
  webURI: "",
  sessionMaxAge: "0",
  trafficAge: "30",
  statsBucketSeconds: "60",
  ipv4Prefix: "32",
  ipv6Prefix: "64",
  ipLimitDryRun: "false",
  timeLocation: "Asia/Tehran",
  subListen: "",
  subPort: "2096",
  subPath: "/sub/",
  subDomain: "",
  subCertFile: "",
  subKeyFile: "",
  subUpdates: "12",
  subEncode: "true",
  subShowInfo: "false",
  subRemarkPrefix: "true",
  subURI: "",
  subCustomHeaders: "",
  subJsonExt: "",
  subClashExt: "",
  subClashNoDefGrp: "false",
  subClashSprtAll: "false",
  subClashUdp: "false",
  globalReset: "",
})

// The panel settings, exactly as the block above spells them out.
type PanelSettings = typeof settings.value

onMounted(async () => {
  loading.value = true
  await loadData()
  loading.value = false
})

const loadData = async () => {
  loading.value = true
  const msg = await HttpUtils.get<PanelSettings>('api/settings')
  loading.value = false
  if (msg.success) {
    setData(msg.obj)
  }
}

const setData = (data: PanelSettings) => {
  settings.value = data
  oldSettings.value = { ...data }
}

const save = async () => {
  loading.value = true
  const msg = await HttpUtils.post<{ settings: PanelSettings }>('api/save', { object: 'settings', action: 'set', data: JSON.stringify(settings.value) })
  if (msg.success) {
    push.success({
      title: i18n.global.t('success'),
      duration: 5000,
      message: i18n.global.t('actions.set') + " " + i18n.global.t('pages.settings')
    })
    setData(msg.obj.settings)
  }
  loading.value = false
}

const maintenance = computed((): boolean => Data().maintenance)

// Stopping the core is deliberate downtime for every user, so it is confirmed
// rather than done on a single click.
const toggleMaintenance = async () => {
  const turningOn = !maintenance.value
  if (turningOn && !confirm(i18n.global.t('setting.maintenanceConfirm'))) return

  loading.value = true
  const msg = await HttpUtils.post('api/maintenance', { enable: turningOn })
  if (msg.success) {
    Data().maintenance = turningOn
    push.success({
      title: i18n.global.t('success'),
      duration: 5000,
      message: i18n.global.t(turningOn ? 'setting.maintenanceOn' : 'setting.maintenanceOff')
    })
  }
  loading.value = false
}

const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

const restartApp = async () => {
  loading.value = true
  const msg = await HttpUtils.post('api/restartApp', {})
  if (msg.success) {
    let url = settings.value.webURI
    if (url !== "") {
      const isTLS = settings.value.webCertFile !== "" || settings.value.webKeyFile !== ""
      url = buildURL(settings.value.webDomain, settings.value.webPort.toString(), isTLS, settings.value.webPath)
    }
    await sleep(3000)
    window.location.replace(url)
  }
  loading.value = false
}

const buildURL = (host: string, port: string, isTLS: boolean, path: string) => {
  if (!host || host.length == 0) host = window.location.hostname
  if (!port || port.length == 0) port = window.location.port

  const protocol = isTLS ? "https:" : "http:"

  if (port === "" || (isTLS && port === "443") || (!isTLS && port === "80")) {
    port = ""
  } else {
    port = `:${port}`
  }

  return `${protocol}//${host}${port}${path}settings`
}

const subEncode = computed({
  get: () => { return settings.value.subEncode == "true" },
  set: (v: boolean) => { settings.value.subEncode = v ? "true" : "false" }
})

const subShowInfo = computed({
  get: () => { return settings.value.subShowInfo == "true" },
  set: (v: boolean) => { settings.value.subShowInfo = v ? "true" : "false" }
})

const subRemarkPrefix = computed({
  get: () => { return settings.value.subRemarkPrefix == "true" },
  set: (v: boolean) => { settings.value.subRemarkPrefix = v ? "true" : "false" }
})

const webPort = computed({
  get: () => { return settings.value.webPort.length > 0 ? parseInt(settings.value.webPort) : 2095 },
  set: (v: number | null) => { settings.value.webPort = v && v > 0 ? v.toString() : "2095" }
})

const sessionMaxAge = computed({
  get: () => { return settings.value.sessionMaxAge.length > 0 ? parseInt(settings.value.sessionMaxAge) : 0 },
  set: (v: number | null) => { settings.value.sessionMaxAge = v && v > 0 ? v.toString() : "0" }
})

const trafficAge = computed({
  get: () => { return settings.value.trafficAge.length > 0 ? parseInt(settings.value.trafficAge) : 0 },
  set: (v: number | null) => { settings.value.trafficAge = v && v > 0 ? v.toString() : "0" }
})

const statsBucketSeconds = computed({
  get: () => { return settings.value.statsBucketSeconds.length > 0 ? parseInt(settings.value.statsBucketSeconds) : 60 },
  set: (v: number | null) => { settings.value.statsBucketSeconds = v && v > 0 ? v.toString() : "60" }
})

const ipv4Prefix = computed({
  get: () => parseInt(settings.value.ipv4Prefix) || 32,
  set: (v: number | null) => { settings.value.ipv4Prefix = v && v > 0 ? Math.min(v, 32).toString() : "32" }
})

const ipv6Prefix = computed({
  get: () => parseInt(settings.value.ipv6Prefix) || 64,
  set: (v: number | null) => { settings.value.ipv6Prefix = v && v > 0 ? Math.min(v, 128).toString() : "64" }
})

const ipLimitDryRun = computed({
  get: () => settings.value.ipLimitDryRun === "true",
  set: (v: boolean) => { settings.value.ipLimitDryRun = v ? "true" : "false" }
})

const subPort = computed({
  get: () => { return settings.value.subPort.length > 0 ? parseInt(settings.value.subPort) : 2096 },
  set: (v: number | null) => { settings.value.subPort = v && v > 0 ? v.toString() : "2096" }
})

const subUpdates = computed({
  get: () => { return settings.value.subUpdates.length > 0 ? parseInt(settings.value.subUpdates) : 12 },
  set: (v: number | null) => { settings.value.subUpdates = v && v > 0 ? v.toString() : "12" }
})

const stateChange = computed(() => {
  return !FindDiff.deepCompare(settings.value, oldSettings.value)
})
</script>

<style scoped>
.settings-card {
  max-width: 1360px;
  margin: 0 auto;
}

.settings-card :deep(.ant-card-body) {
  padding: 20px 24px;
}

.settings-page {
  min-width: 0;
  overflow-x: clip;
}

.settings-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.16);
  margin-bottom: 12px;
}

.settings-tabs {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 2px;
  min-width: 0;
}

.settings-tab {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  white-space: nowrap;
}

.settings-tab:hover {
  background: rgba(128, 128, 128, 0.1);
}

.settings-tab.active {
  color: #1677ff;
  background: rgba(22, 119, 255, 0.1);
  font-weight: 600;
}

.settings-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.maintenance-alert {
  margin-bottom: 12px;
}

.settings-form {
  min-width: 0;
  overflow-x: clip;
}

.settings-form .field-row + .field-row {
  margin-top: 12px;
}

.settings-form .ant-form-item {
  margin-bottom: 0;
}

.settings-form .ant-form-item-extra {
  overflow-wrap: anywhere;
}

.switch-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px 20px;
  margin-bottom: 12px;
}

.short-field .ant-input-number,
.short-field .ant-input-number-group-wrapper {
  width: 100%;
  max-width: 160px;
}

@media (max-width: 767px) {
  .settings-card :deep(.ant-card-body) {
    padding: 14px;
  }

  .settings-actions {
    width: 100%;
    margin-left: 0;
  }

  .settings-tab {
    padding: 6px 8px;
  }
}
</style>

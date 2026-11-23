<template>
  <a-card
    size="small"
    title="API"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <a-row
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.api.secret')"
            style="margin-bottom: 0;"
          >
            <a-input-password
              v-model:value="data.secret"
              allow-clear
              @change="(e: any) => { if (!e.target?.value) delete data.secret }"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-space>
            <a-switch v-model:checked="data.access_control_allow_private_network" />
            <span>{{ $t('types.api.allowPrivateNetwork') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row :gutter="[12, 12]">
        <a-col :span="24">
          <a-form-item
            :label="$t('types.api.allowOrigin')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.access_control_allow_origin"
              mode="tags"
              allow-clear
              style="width: 100%"
              @change="(v: any) => { if (!v || v.length === 0) delete data.access_control_allow_origin }"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-space>
            <a-switch v-model:checked="dashboardEnabled" />
            <span>{{ $t('types.api.dashboard') }}</span>
          </a-space>
        </a-col>
        <a-col
          v-if="dashboardEnabled"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.api.dashboardPath')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="dashboardPath"
              allow-clear
            />
          </a-form-item>
        </a-col>
      </a-row>
      <!-- Letting sing-box fetch the dashboard itself needs the full object
           form, so it stays behind a toggle: without it the setting keeps the
           short boolean or path shape. -->
      <template v-if="dashboardEnabled">
        <a-row
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-space>
              <a-switch v-model:checked="dashboardDownload" />
              <span>{{ $t('types.api.dashboardDownload') }}</span>
            </a-space>
          </a-col>
        </a-row>
        <a-row
          v-if="dashboardDownload"
          :gutter="[12, 12]"
        >
          <a-col :span="24">
            <a-form-item
              :label="$t('types.api.dashboardUrl')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="dashboardDownloadUrl" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('basic.httpClient.title')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="dashboardHttpClient"
                :not-found-content="$t('basic.httpClient.none')"
                allow-clear
                style="width: 100%"
              >
                <a-select-option
                  v-for="item in httpClients"
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
            :sm="12"
          >
            <a-form-item
              :label="$t('ruleset.interval')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="dashboardInterval"
                :addon-after="$t('date.d')"
                :min="0"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </template>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { httpClientTags, refTag } from '@/plugins/httpClient'
import { HttpClientRef } from '@/types/httpClient'
import { Srv } from '@/types/services'

// The object form of dashboard, the widest of the three shapes it accepts.
interface DashboardOptions {
  enabled?: boolean
  path?: string
  download_url?: string
  http_client?: HttpClientRef
  update_interval?: string
}

// Only the api fields this form reads or writes.
interface ApiData {
  secret?: string
  access_control_allow_origin?: string[]
  access_control_allow_private_network?: boolean
  dashboard?: boolean | string | DashboardOptions
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Srv>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as ApiData)

// The object form when dashboard already holds one, for the fields only it can
// carry.
const dashboardObject = computed((): DashboardOptions | undefined => {
  const dashboard = data.value.dashboard
  return dashboard && typeof dashboard === 'object' ? dashboard : undefined
})

const httpClients = computed((): string[] => httpClientTags())

const dashboardEnabled = computed({
  get: (): boolean => {
    const dashboard = data.value.dashboard
    if (typeof dashboard === 'boolean') return dashboard
    if (typeof dashboard === 'string') return dashboard.length > 0
    return dashboard?.enabled === true
  },
  set: (v: boolean) => {
    if (!v) {
      delete data.value.dashboard
      return
    }
    data.value.dashboard = true
  },
})

const dashboardPath = computed({
  get: (): string => {
    const dashboard = data.value.dashboard
    if (typeof dashboard === 'string') return dashboard
    if (dashboard && typeof dashboard === 'object') return dashboard.path ?? ''
    return ''
  },
  set: (v: string) => {
    const dashboard = data.value.dashboard
    if (dashboard && typeof dashboard === 'object') {
      if (v) dashboard.path = v
      else delete dashboard.path
      return
    }
    data.value.dashboard = v ? v : true
  },
})

// Turning this off drops the download settings and lets the value fall
// back to its short form.
const dashboardDownload = computed({
  get: (): boolean => {
    const dashboard = dashboardObject.value
    if (!dashboard) return false
    return dashboard.download_url != undefined
      || dashboard.http_client != undefined
      || dashboard.update_interval != undefined
  },
  set: (v: boolean) => {
    if (v) {
      asObject().download_url = ''
      return
    }
    const dashboard = dashboardObject.value
    if (!dashboard) return
    const path = dashboard.path
    data.value.dashboard = path ? path : true
  },
})

const dashboardDownloadUrl = computed({
  get: (): string => dashboardObject.value?.download_url ?? '',
  set: (v: string) => {
    const dashboard = asObject()
    if (v) dashboard.download_url = v
    else delete dashboard.download_url
  },
})

const dashboardHttpClient = computed({
  get: (): string | undefined => refTag(dashboardObject.value?.http_client),
  set: (v: string | undefined) => {
    const dashboard = asObject()
    if (v) dashboard.http_client = v
    else delete dashboard.http_client
  },
})

const dashboardInterval = computed({
  get: (): number => {
    const interval = dashboardObject.value?.update_interval
    return interval ? parseInt(String(interval).replace('d', '')) : 0
  },
  set: (v: number) => {
    const dashboard = asObject()
    if (v > 0) dashboard.update_interval = v + 'd'
    else delete dashboard.update_interval
  },
})

// dashboard accepts a bool, a path string or an object. Anything beyond
// enabled and path needs the object, so the short forms are widened on
// demand rather than up front.
function asObject(): DashboardOptions {
  const dashboard = data.value.dashboard
  if (dashboard && typeof dashboard === 'object') return dashboard
  const widened: DashboardOptions = { enabled: true }
  if (typeof dashboard === 'string' && dashboard.length > 0) widened.path = dashboard
  data.value.dashboard = widened
  return widened
}
</script>

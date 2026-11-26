<template>
  <Editor
    v-model="enableEditor"
    :data="model.subClashExt"
    :visible="enableEditor"
    :title="$t('editor') + ' - ' + $t('setting.clashSub')"
    @close="enableEditor = false"
    @save="saveEditor"
  />
  <div class="sub-pane clash-pane">
    <a-form
      layout="vertical"
      class="sub-form"
    >
      <div class="clash-default-row">
        <div class="clash-switches">
          <a-space>
            <a-switch v-model:checked="subClashNoDefGrp" />
            <span>{{ $t('setting.clashNoDefGrp') }}</span>
          </a-space>
          <a-space>
            <a-switch v-model:checked="subClashSprtAll" />
            <span>{{ $t('setting.clashSprtAll') }}</span>
          </a-space>
          <a-space>
            <a-switch v-model:checked="subClashUdp" />
            <span>{{ $t('setting.clashUdp') }}</span>
          </a-space>
        </div>
        <a-flex
          class="sub-actions"
          justify="end"
          gap="small"
        >
          <a-button
            type="default"
            @click="openEditor"
          >
            {{ $t('editor') }}
          </a-button>
          <a-popover
            v-model:open="menu"
            trigger="click"
            placement="left"
          >
            <template #content>
              <a-flex
                vertical
                gap="small"
              >
                <a-space>
                  <a-switch v-model:checked="optionMixed" />
                  <span>{{ $t('setting.mixedPort') }}</span>
                </a-space>
                <a-space>
                  <a-switch v-model:checked="optionTun" />
                  <span>{{ $t('setting.tun') }}</span>
                </a-space>
                <a-space>
                  <a-switch v-model:checked="optionExt" />
                  <span>{{ $t('basic.exp.extController') }}</span>
                </a-space>
                <a-space>
                  <a-switch v-model:checked="optionLog" />
                  <span>{{ $t('basic.log.title') }}</span>
                </a-space>
                <a-space>
                  <a-switch v-model:checked="optionDns" />
                  <span>{{ $t('pages.dns') }}</span>
                </a-space>
                <a-space>
                  <a-switch v-model:checked="optionRules" />
                  <span>{{ $t('pages.rules') }}</span>
                </a-space>
              </a-flex>
            </template>
            <a-button type="default">
              {{ $t('setting.jsonSubOptions') }}
            </a-button>
          </a-popover>
        </a-flex>
      </div>
      <a-row
        v-if="optionMixed"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('setting.mixedPort')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="mixedPort"
              :min="1"
              :max="65535"
              style="width: 100%"
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
            :label="$t('types.ts.allowLanAccess')"
            style="margin-bottom: 0;"
          >
            <div class="switch-control">
              <a-switch v-model:checked="allowLan" />
            </div>
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionExt || optionLog"
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          v-if="optionExt"
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('basic.exp.extController')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="externalController" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionLog"
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('basic.log.title') + ' - ' + $t('basic.log.level')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="logLevel"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in ['debug', 'info', 'warning', 'error']"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionTun || optionDns"
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          v-if="optionTun"
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('setting.tun')"
            style="margin-bottom: 0;"
          >
            <div class="switch-control">
              <a-switch v-model:checked="tun" />
            </div>
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionDns"
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('pages.dns')"
            style="margin-bottom: 0;"
          >
            <div class="switch-control">
              <a-switch v-model:checked="dns" />
            </div>
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionRules"
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          :xs="24"
          :sm="24"
          :md="12"
          :lg="8"
        >
          <a-form-item
            :label="$t('pages.rules')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="rules"
              :options="rulesIP"
              :field-names="{ label: 'title', value: 'value' }"
              mode="multiple"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { push } from 'notivue'
import Editor from './Editor.vue'
import yaml from 'yaml'
import { i18n } from '@/locales'

// Only the settings rows this form touches. The parent owns the whole
// settings object and keeps the rest of it.
interface ClashSettings {
  subClashExt: string
  subClashNoDefGrp: string
  subClashSprtAll: string
  subClashUdp: string
}

// The clash config the editor round-trips, parsed out of the stored YAML.
// Its values are whatever the operator typed, so every read asserts the
// shape it expects.
type MetaJson = Record<string, unknown>

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<ClashSettings>('settings', { required: true })

const enableEditor = ref(false)
const menu = ref(false)

const defaultConfig = {
  "mixed-port": 7890,
  "allow-lan": false,
  "mode": "rule",
  "log-level": "info",
  "external-controller": "127.0.0.1:9090",
  "tun": {
    "enable": true,
    "stack": "system",
    "auto-route": true,
    "auto-detect-interface": true,
    "dns-hijack": ["any:53"],
  },
  "dns": {
    "enable": true,
    "ipv6": false,
    "enhanced-mode": "fake-ip",
    "fake-ip-range": "198.18.0.1/16",
    "default-nameserver": ["8.8.8.8","1.1.1.1"],
    "nameserver": [
      "https://doh.pub/dns-query",
      "https://1.0.0.1/dns-query"
    ],
    "fallback": ["tcp://9.9.9.9:53"],
    "fake-ip-filter": ["*.lan", "localhost", "*.local"]
  },
  "rules": [
    "GEOIP,Private,DIRECT",
    "MATCH,Proxy"
  ]
}

const rulesIP = [
  { title: 'Private-Direct', value: 'GEOIP,Private,DIRECT' },
  { title: 'Private-Block', value: 'GEOIP,Private,REJECT' },
  { title: 'LAN-Direct', value: 'GEOIP,LAN,DIRECT' },
  { title: 'LAN-Block', value: 'GEOIP,LAN,REJECT' },
  { title: 'Ads-Direct', value: 'GEOIP,Ads,DIRECT' },
  { title: 'Ads-Block', value: 'GEOIP,Ads,REJECT' },
  { title: '🇨🇳 China-Direct', value: 'GEOIP,CN,DIRECT' },
  { title: '🇨🇳 China-Block', value: 'GEOIP,CN,REJECT' },
  { title: '🇮🇷 Iran-Direct', value: 'GEOIP,CATEGORY-IR,DIRECT' },
  { title: '🇮🇷 Iran-Block', value: 'GEOIP,CATEGORY-IR,REJECT' },
  { title: '🇻🇳 Vietnam-Direct', value: 'GEOIP,CATEGORY-VN,DIRECT' },
  { title: '🇻🇳 Vietnam-Block', value: 'GEOIP,CATEGORY-VN,REJECT' },
  { title: '🇯🇵 Japan-Direct', value: 'GEOIP,JP,DIRECT' },
  { title: '🇯🇵 Japan-Block', value: 'GEOIP,JP,REJECT' },
]

const metaJson = computed({
  get: (): MetaJson => {
    try {
      return yaml.parse(model.value.subClashExt)??{}
    } catch {
      return {}
    }
  },
  set: (v:MetaJson) => {
    model.value.subClashExt = Object.keys(v).length==0 ? "" : yaml.stringify(v)
  },
})

const optionMixed = computed({
  get: (): boolean => (metaJson.value['mixed-port'] as number)>0,
  set: (v:boolean) => {
    updateMetaJson(v ? defaultConfig['mixed-port'] : null, 'mixed-port')
    updateMetaJson(v ? defaultConfig['allow-lan'] : null, 'allow-lan')
  },
})

const optionTun = computed({
  get: (): boolean => (metaJson.value['tun'] as MetaJson | undefined)?.['enable'] as boolean?? false,
  set: (v:boolean) => { updateMetaJson(v ? defaultConfig['tun'] : null, 'tun') },
})

const optionExt = computed({
  get: (): boolean => ((metaJson.value['external-controller'] as string | undefined)?.length?? 0)>0,
  set: (v:boolean) => { updateMetaJson(v ? defaultConfig['external-controller'] : null, 'external-controller') },
})

const optionLog = computed({
  get: (): boolean => ((metaJson.value['log-level'] as string | undefined)?.length?? 0)>0,
  set: (v:boolean) => { updateMetaJson(v ? defaultConfig['log-level'] : null, 'log-level') },
})

const optionDns = computed({
  get: (): boolean => (metaJson.value['dns'] as MetaJson | undefined)?.['enable'] as boolean?? false,
  set: (v:boolean) => { updateMetaJson(v ? defaultConfig['dns'] : null, 'dns') },
})

const optionRules = computed({
  get: (): boolean => ((metaJson.value['rules'] as string[] | undefined)?.length?? 0)>0,
  set: (v:boolean) => {
    updateMetaJson(v ? defaultConfig['rules'] : null, 'rules')
    updateMetaJson(v ? defaultConfig['mode'] : null, 'mode')
  },
})

const mixedPort = computed({
  get: (): number => metaJson.value['mixed-port'] as number,
  set: (v:number) => { updateMetaJson(v, 'mixed-port') },
})

const allowLan = computed({
  get: (): boolean => metaJson.value['allow-lan'] as boolean,
  set: (v:boolean) => { updateMetaJson(v, 'allow-lan') },
})

const externalController = computed({
  get: (): string => metaJson.value['external-controller'] as string,
  set: (v:string) => { updateMetaJson(v, 'external-controller') },
})

const logLevel = computed({
  get: (): string => metaJson.value['log-level'] as string,
  set: (v:string) => { updateMetaJson(v, 'log-level') },
})

const dns = computed({
  get: (): boolean => (metaJson.value['dns'] as MetaJson | undefined)?.['enable'] as boolean ?? false,
  set: (v:boolean) => { updateMetaJson({ ...(metaJson.value['dns'] as MetaJson | undefined), 'enable': v }, 'dns') },
})

const tun = computed({
  get: (): boolean => (metaJson.value['tun'] as MetaJson | undefined)?.['enable'] as boolean ?? false,
  set: (v:boolean) => { updateMetaJson({ ...(metaJson.value['tun'] as MetaJson | undefined), 'enable': v }, 'tun') },
})

const rules = computed({
  get: (): string[] => {
    const current = metaJson.value.rules as string[]
    return current.length > 0 ? current.filter((r:string) => r != "MATCH,Proxy") : []
  },
  set: (v:string[]) => {
    let newRules = <string[]>[]
    v.forEach((r:string) => { newRules.push(r) })
    updateMetaJson([ ...newRules, "MATCH,Proxy" ], 'rules')
  },
})

const subClashNoDefGrp = computed({
  get: (): boolean => model.value.subClashNoDefGrp == "true",
  set: (v:boolean) => { model.value.subClashNoDefGrp = v ? "true" : "false" },
})

const subClashSprtAll = computed({
  get: (): boolean => model.value.subClashSprtAll == "true",
  set: (v:boolean) => { model.value.subClashSprtAll = v ? "true" : "false" },
})

const subClashUdp = computed({
  get: (): boolean => model.value.subClashUdp == "true",
  set: (v:boolean) => { model.value.subClashUdp = v ? "true" : "false" },
})

function openEditor() {
  enableEditor.value = true
}

function saveEditor(data:string) {
  try {
    const result = yaml.parse(data)
    if (typeof result != 'object' || Array.isArray(result)) throw new Error()
  } catch {
    push.error({
      message: i18n.global.t('failed') + ": " + i18n.global.t('error.invalidData'),
      duration: 5000,
    })
    return
  }
  model.value.subClashExt = data
  enableEditor.value = false
}

function updateMetaJson(data:unknown, key:string) {
  let newMetaJson = metaJson.value
  if (data==null) {
    delete newMetaJson[key]
  } else {
    newMetaJson[key] = data
  }
  metaJson.value = newMetaJson
}
</script>
<style scoped>
.sub-pane {
  min-width: 0;
}

.clash-default-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.clash-switches {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px 20px;
}

.sub-actions {
  width: 100%;
  flex-wrap: wrap;
}

.clash-pane > .ant-row {
  margin-top: 12px;
}

.sub-form {
  min-width: 0;
  overflow-x: clip;
}

.sub-form .ant-form-item {
  margin-bottom: 0;
}

.switch-control {
  display: flex;
  align-items: center;
  min-height: 32px;
}

@media (max-width: 575px) {
  .clash-default-row {
    align-items: flex-start;
  }

  .clash-switches {
    width: 100%;
  }

  .sub-actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>

<style scoped>
.clash-default-row .sub-actions {
  width: auto;
}

@media (max-width: 575px) {
  .clash-default-row .sub-actions {
    width: 100%;
  }
}
</style>

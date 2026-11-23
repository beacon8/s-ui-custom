<template>
  <a-card
    size="small"
    title="Hysteria2"
  >
    <a-row
      :gutter="[12, 12]"
      align="middle"
    >
      <a-col
        v-if="direction == 'in'"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-space>
          <a-switch v-model:checked="data.ignore_client_bandwidth" />
          <span>{{ $t('types.hy.ignoreBw') }}</span>
        </a-space>
      </a-col>
      <a-col
        v-if="!data.ignore_client_bandwidth"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('stats.upload')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="up_mbps"
            :min="0"
            :addon-after="$t('stats.Mbps')"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
      <a-col
        v-if="!data.ignore_client_bandwidth"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('stats.download')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="down_mbps"
            :min="0"
            :addon-after="$t('stats.Mbps')"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>
    <a-row
      v-if="data.obfs != undefined"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('type')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="obfsType"
            style="width: 100%"
          >
            <a-select-option
              v-for="item in obfsTypes"
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
        :md="8"
      >
        <a-form-item
          :label="$t('types.hy.obfs')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.obfs.password" />
        </a-form-item>
      </a-col>
      <template v-if="data.obfs.type == 'gecko'">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.hy.minPacketSize')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="minPacketSize"
              placeholder="512"
              :min="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.hy.maxPacketSize')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="maxPacketSize"
              placeholder="1200"
              :min="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </template>
    </a-row>
    <template v-if="direction == 'in'">
      <a-card
        v-if="data.masquerade != undefined"
        size="small"
        title="Hysteria2 Masquerade"
        style="margin-top: 12px;"
      >
        <a-row :gutter="[12, 12]">
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('type')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="masqueradeType"
                :options="masqTypes"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col
            v-if="masqueradeType == ''"
            :xs="24"
            :sm="16"
          >
            <a-form-item
              label="HTTP3 server on auth fails"
              style="margin-bottom: 0;"
            >
              <a-input
                v-model:value="masqueradeStr"
                placeholder="file:///var/www | http://127.0.0.1:8080"
              />
            </a-form-item>
          </a-col>
          <a-col
            v-if="masqueradeType == 'file'"
            :xs="24"
            :sm="16"
          >
            <a-form-item
              label="File server root directory"
              style="margin-bottom: 0;"
            >
              <a-input
                v-model:value="masquerade.directory"
                placeholder="/var/www"
              />
            </a-form-item>
          </a-col>
          <a-col
            v-if="masqueradeType == 'string'"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              label="HTTP Code"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="masquerade.status_code"
                :min="100"
                :max="599"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-if="masqueradeType == 'proxy'"
          :gutter="[12, 12]"
          align="middle"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              label="Target URL"
              style="margin-bottom: 0;"
            >
              <a-input
                v-model:value="masquerade.url"
                placeholder="http://example.com:8080"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-space>
              <a-switch v-model:checked="masquerade.rewrite_host" />
              <span>Rewrite Host</span>
            </a-space>
          </a-col>
        </a-row>
        <template v-if="masqueradeType == 'string'">
          <a-row
            :gutter="[12, 12]"
            style="margin-top: 12px;"
          >
            <a-col
              :xs="24"
              :sm="16"
            >
              <a-form-item
                label="Content"
                style="margin-bottom: 0;"
              >
                <a-input v-model:value="masquerade.content" />
              </a-form-item>
            </a-col>
          </a-row>
          <Headers :data="masquerade" />
        </template>
      </a-card>
    </template>
    <template v-else>
      <a-row
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.pw')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.password" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <Network :data="model" />
        </a-col>
        <a-col
          v-if="optionMPort"
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('rule.portRange') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="server_ports" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionMPort"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('ruleset.interval')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="hop_interval"
              :min="0"
              :addon-after="$t('date.s')"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
    </template>
    <a-flex
      justify="end"
      style="margin-top: 12px;"
    >
      <a-popover
        v-model:open="menu"
        trigger="click"
        placement="left"
      >
        <template #content>
          <a-space direction="vertical">
            <a-space>
              <a-switch v-model:checked="optionObfs" />
              <span>{{ $t('types.hy.obfs') }}</span>
            </a-space>
            <template v-if="direction == 'in'">
              <a-space>
                <a-switch v-model:checked="optionMasq" />
                <span>Masquerade</span>
              </a-space>
            </template>
            <template v-else>
              <a-space v-if="!optionRealm">
                <a-switch v-model:checked="optionMPort" />
                <span>{{ $t('rule.portRange') }}</span>
              </a-space>
            </template>
            <a-space>
              <a-switch v-model:checked="optionRealm" />
              <span>{{ $t('types.hy.realm') }}</span>
            </a-space>
          </a-space>
        </template>
        <a-button type="default">
          {{ $t('types.hy.hy2Options') }}
        </a-button>
      </a-popover>
    </a-flex>
    <Hysteria2Realm
      v-if="data.realm != undefined"
      :data="data.realm"
    />
    <QuicFields
      :data="model"
      quic
    />
  </a-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import QuicFields from '@/components/QuicFields.vue'
import Network from '@/components/Network.vue'
import Headers from '@/components/Headers.vue'
import Hysteria2Realm from '@/components/Hysteria2Realm.vue'
import { i18n } from '@/locales'
import { Inbound, Hysteria2Obfs } from '@/types/inbounds'
import { Outbound, Hysteria2Realm as Realm } from '@/types/outbounds'

// The masquerade is either a bare URL string or one of the object forms; the
// type field picks which of the object fields apply.
interface Hysteria2Masquerade {
  type: string
  directory?: string
  url?: string
  rewrite_host?: boolean
  status_code?: number
  content?: string
  headers?: Record<string, string | string[]>
}

// One form serves both directions, so it takes the widest shape either side
// can hold: only the inbound side carries the masquerade and the bandwidth
// override, only the outbound side the password and the port hopping.
interface Hysteria2Data {
  ignore_client_bandwidth?: boolean
  up_mbps?: number
  down_mbps?: number
  obfs?: Hysteria2Obfs
  masquerade?: string | Hysteria2Masquerade
  password?: string
  server?: string
  server_port?: number
  server_ports?: string[]
  hop_interval?: string
  realm?: Realm
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound | Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as Hysteria2Data)

const options = defineProps<{
  direction: 'in' | 'out' | 'out_json'
}>()

const menu = ref(false)

const masqTypes = [
  { title: i18n.global.t('rule.simple'), value: '' },
  { title: "File server", value: "file" },
  { title: "Reverse Proxy", value: "proxy" },
  { title: "Fixed response", value: "string" },
]

// The masquerade type select decides which shape the field holds, which the
// template checker cannot follow either, so expose the object form once here.
const masquerade = computed(() => data.value.masquerade as Hysteria2Masquerade)

const masqueradeStr = computed({
  get: (): string => typeof data.value.masquerade === 'string' ? data.value.masquerade : '',
  set: (v: string) => { data.value.masquerade = v },
})

const down_mbps = computed({
  get: (): number => data.value.down_mbps ?? 0,
  set: (v: number) => { data.value.down_mbps = v > 0 ? v : undefined },
})

const up_mbps = computed({
  get: (): number => data.value.up_mbps ?? 0,
  set: (v: number) => { data.value.up_mbps = v > 0 ? v : undefined },
})

const server_ports = computed({
  get: (): string => data.value.server_ports?.join(',') ?? '',
  set: (v: string) => { data.value.server_ports = v.length > 0 ? v.split(',') : undefined },
})

const masqueradeType = computed({
  get: (): string => typeof data.value.masquerade === 'object' ? data.value.masquerade.type ?? '' : '',
  set: (v: string) => {
    if (v == '') {
      data.value.masquerade = ''
    } else {
      data.value.masquerade = { type: v }
    }
  },
})

const hop_interval = computed({
  get: (): number => {
    const v = data.value.hop_interval
    return v ? parseInt(v.replace('s', '')) : 0
  },
  set: (v: number) => { data.value.hop_interval = v > 0 ? v + 's' : undefined },
})

const optionObfs = computed({
  get: (): boolean => data.value.obfs != undefined,
  set: (v: boolean) => { data.value.obfs = v ? { type: "salamander", password: "" } : undefined },
})

const obfsTypes = ['salamander', 'gecko']

// The packet sizes exist only for gecko, so switching away drops them
const obfsType = computed({
  get: (): string => data.value.obfs?.type ?? 'salamander',
  set: (v: string) => {
    if (!data.value.obfs) return
    data.value.obfs.type = <Hysteria2Obfs['type']>v
    if (v != 'gecko') {
      delete data.value.obfs.min_packet_size
      delete data.value.obfs.max_packet_size
    }
  },
})

// Empty leaves the field out, so sing-box applies its own default
const minPacketSize = computed({
  get: (): number | undefined => data.value.obfs?.min_packet_size,
  set: (v: number) => { if (data.value.obfs) data.value.obfs.min_packet_size = v > 0 ? v : undefined },
})

const maxPacketSize = computed({
  get: (): number | undefined => data.value.obfs?.max_packet_size,
  set: (v: number) => { if (data.value.obfs) data.value.obfs.max_packet_size = v > 0 ? v : undefined },
})

const optionMasq = computed({
  get: (): boolean => data.value.masquerade != undefined,
  set: (v: boolean) => { data.value.masquerade = v ? "" : undefined },
})

// An outbound reaching the server through a realm must not name the server
// itself, so turning the realm on clears the address and the port hopping.
const optionRealm = computed({
  get: (): boolean => data.value.realm != undefined,
  set: (v: boolean) => {
    data.value.realm = v ? { server_url: '', realm_id: '', stun_servers: [] } : undefined
    if (v && options.direction == 'out') {
      delete data.value.server
      delete data.value.server_port
      delete data.value.server_ports
      delete data.value.hop_interval
    }
  },
})

const optionMPort = computed({
  get: (): boolean => data.value.server_ports != undefined,
  set: (v: boolean) => { data.value.server_ports = v ? [] : undefined },
})
</script>

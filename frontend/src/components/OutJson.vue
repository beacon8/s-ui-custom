<template>
  <a-card
    size="small"
    :title="$t('pages.basics')"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        v-if="type == inTypes.SOCKS"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('version')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="inData.out_json.version"
            style="width: 100%"
          >
            <a-select-option
              v-for="item in ['4', '4a', '5']"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col
        v-if="needNetwork"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <Network :data="inData.out_json" />
      </a-col>
      <a-col
        v-if="needUot"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <UoT :data="inData.out_json" />
      </a-col>
      <a-col
        v-if="type == inTypes.HTTP"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('transport.path')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="inData.out_json.path" />
        </a-form-item>
      </a-col>
      <a-col
        v-if="type == inTypes.VMess || type == inTypes.VLESS"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.vless.udpEnc')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="packet_encoding"
            style="width: 100%"
          >
            <a-select-option
              v-for="item in ['none', 'packetaddr', 'xudp']"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
      <template v-if="type == inTypes.VMess">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.vmess.security')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="inData.out_json.security"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in vmessSecurities"
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
          style="display: flex; align-items: center;"
        >
          <a-space>
            <a-switch v-model:checked="inData.out_json.global_padding" />
            <span>{{ $t('types.vmess.globalPadding') }}</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
          style="display: flex; align-items: center;"
        >
          <a-space>
            <a-switch v-model:checked="inData.out_json.authenticated_length" />
            <span>{{ $t('types.vmess.authLen') }}</span>
          </a-space>
        </a-col>
      </template>
      <!-- The QUIC stream window; recv_window was hysteria's own name for it. -->
      <a-col
        v-if="type == inTypes.Hysteria"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('quic.streamWindow')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="hysteriaStreamWindow"
            placeholder="8mb"
            allow-clear
          />
        </a-form-item>
      </a-col>
      <template v-if="type == inTypes.TUIC">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            label="UDP Relay Mode"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="inData.out_json.udp_relay_mode"
              allow-clear
              style="width: 100%"
              @change="(v: unknown) => { if (!v) delete inData.out_json.udp_relay_mode }"
            >
              <a-select-option
                v-for="item in ['native', 'quic']"
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
          style="display: flex; align-items: center;"
        >
          <a-space>
            <a-switch v-model:checked="inData.out_json.udp_over_stream" />
            <span>UDP Over Stream</span>
          </a-space>
        </a-col>
      </template>
    </a-row>
    <a-row
      v-if="viaRealm"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.hy.clientToken')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="realmToken" />
        </a-form-item>
      </a-col>
    </a-row>
    <a-row
      v-else-if="[inTypes.Hysteria, inTypes.Hysteria2].includes(type)"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
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
    <Headers
      v-if="type == inTypes.HTTP"
      :data="inData.out_json"
    />
    <AnyTls
      v-if="type == inTypes.AnyTls"
      :data="outJson"
      direction="out_json"
    />
    <Naive
      v-if="type == inTypes.Naive"
      :data="outJson"
      direction="out_json"
    />
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { InTypes, Inbound } from '@/types/inbounds'
import { Outbound } from '@/types/outbounds'
import Network from './Network.vue'
import UoT from './UoT.vue'
import Headers from './Headers.vue'
import AnyTls from './protocols/AnyTls.vue'
import Naive from './protocols/Naive.vue'

// The outbound sing-box config this inbound hands to its clients: only the
// fields the panel lets the operator override are listed here.
interface OutJsonFields {
  version?: string
  path?: string
  network?: string
  udp_over_tcp?: { enabled: boolean, version: number }
  headers?: Record<string, string | string[]>
  security?: string
  global_padding?: boolean
  authenticated_length?: boolean
  packet_encoding?: string
  stream_receive_window?: number | string
  server_ports?: string[]
  hop_interval?: string
  udp_relay_mode?: 'native' | 'quic'
  udp_over_stream?: boolean
  realm?: { token?: string }
}

interface OutJsonHolder {
  out_json: OutJsonFields
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound>('inData', { required: true })

// The parent holds a union and out_json is untyped on it, so narrow once here.
// The computed returns the same object, so edits still reach the parent.
const inData = computed(() => model.value as unknown as OutJsonHolder)

const props = defineProps<{
  type: string
}>()

// AnyTls and Naive take a whole outbound; they edit the very same object.
const outJson = computed(() => inData.value.out_json as unknown as Outbound)

const inTypes = InTypes

const vmessSecurities = [
  "auto",
  "none",
  "zero",
  "aes-128-gcm",
  "aes-128-ctr",
  "chacha20-poly1305",
]

const haveNetwork: string[] = [
  InTypes.SOCKS,
  InTypes.Shadowsocks,
  InTypes.VMess,
  InTypes.Trojan,
  InTypes.Hysteria,
  InTypes.VLESS,
  InTypes.TUIC,
  InTypes.Hysteria2,
]

const havUoT: string[] = [
  InTypes.SOCKS,
  InTypes.Shadowsocks,
]

// Takes a number or a string with a unit ("8mb"); an empty value is
// removed, since sing-box rejects "" for a byte size.
const hysteriaStreamWindow = computed({
  get: (): string => String(inData.value.out_json.stream_receive_window ?? ''),
  set: (v: string) => {
    const trimmed = (v ?? '').trim()
    if (trimmed) inData.value.out_json.stream_receive_window = trimmed
    else delete inData.value.out_json.stream_receive_window
  },
})

// The rest of the client's realm is copied from the inbound when it is saved;
// only a token of the clients' own is set here.
const viaRealm = computed((): boolean => props.type == InTypes.Hysteria2 && (model.value as { realm?: object }).realm != undefined)

const realmToken = computed({
  get: (): string => inData.value.out_json.realm?.token ?? '',
  set: (v: string) => {
    if (v) inData.value.out_json.realm = { ...inData.value.out_json.realm, token: v }
    else delete inData.value.out_json.realm?.token
  },
})

const needNetwork = computed((): boolean => haveNetwork.includes(props.type))

const needUot = computed((): boolean => havUoT.includes(props.type))

const packet_encoding = computed({
  get: () => inData.value.out_json.packet_encoding != undefined ? inData.value.out_json.packet_encoding : 'none',
  set: (v: string) => { inData.value.out_json.packet_encoding = v != "none" ? v : undefined },
})

const server_ports = computed({
  get: (): string => inData.value.out_json.server_ports?.join(',') ?? '',
  set: (v: string) => { inData.value.out_json.server_ports = v.length > 0 ? v.split(',') : undefined },
})

const hop_interval = computed({
  get: () => inData.value.out_json.hop_interval ? parseInt(inData.value.out_json.hop_interval.replace('s', '')) : 0,
  set: (v: number) => { inData.value.out_json.hop_interval = v > 0 ? v + 's' : undefined },
})
</script>

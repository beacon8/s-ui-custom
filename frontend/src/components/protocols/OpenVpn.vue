<template>
  <a-card
    size="small"
    :title="isServer ? 'OpenVPN Server' : 'OpenVPN Client'"
  >
    <!-- Client dials a remote server; server listens locally. -->
    <a-row
      v-if="!isServer"
      :gutter="[12, 12]"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.server')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.server" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.serverPort')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="data.server_port"
            :min="1"
            :max="65535"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>
    <a-row
      v-else
      :gutter="[12, 12]"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('in.addr')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.listen" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('in.port')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="data.listen_port"
            :min="1"
            :max="65535"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>

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
          :label="$t('types.openvpn.mode')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="mode"
            :options="modes"
            :field-names="{ label: 'title', value: 'value' }"
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
          :label="$t('network')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.network"
            :options="networks"
            :field-names="{ label: 'title', value: 'value' }"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
      <!-- static_key has no TLS session to negotiate a cipher, so its single
           fixed one is always set (the mode switch fills in a default). -->
      <a-col
        v-if="data.mode == 'static_key'"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.cipher')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.cipher"
            style="width: 100%"
          >
            <a-select-option
              v-for="item in staticCiphers"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
    </a-row>

    <!-- The server always assigns addresses; a static_key client must set its
         own, since there is no TLS session to push one over. -->
    <a-row
      v-if="isServer || data.mode == 'static_key'"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col :xs="24">
        <a-form-item
          :label="$t('types.openvpn.address')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.address"
            mode="tags"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>

    <!-- static_key mode has no TLS session, so the peer address cannot be
         negotiated and has to be configured on both ends. -->
    <a-row
      v-if="!isServer && data.mode == 'static_key'"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.peerAddress')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="data.peer_address"
            placeholder="10.8.0.1"
          />
        </a-form-item>
      </a-col>
      <a-col
        v-if="optionPeerIpv6"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.peerAddressIpv6')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.peer_address_ipv6" />
        </a-form-item>
      </a-col>
    </a-row>

    <!-- static_key mode uses a shared key instead of TLS -->
    <a-row
      v-if="data.mode == 'static_key'"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.staticKeyPath')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="data.static_key_path"
            allow-clear
            @change="() => { if (!data.static_key_path) delete data.static_key_path }"
          />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.keyDirection')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.key_direction"
            :options="keyDirections"
            :field-names="{ label: 'title', value: 'value' }"
            allow-clear
            style="width: 100%"
            @clear="delete data.key_direction"
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-row
      v-if="!isServer && data.mode == 'tls' && optionCredentials"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.un')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.username" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.pw')"
          style="margin-bottom: 0;"
        >
          <a-input-password v-model:value="data.password" />
        </a-form-item>
      </a-col>
    </a-row>

    <!-- A TLS session negotiates one cipher out of `data_ciphers`, in the
         order listed, so the full width keeps several chips readable. -->
    <template v-if="data.mode != 'static_key' && optionDataCiphers">
      <a-row
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col :xs="24">
          <a-form-item
            :label="$t('types.openvpn.dataCiphers')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.data_ciphers"
              mode="multiple"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in tlsCiphers"
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
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.openvpn.dataCiphersFallback')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.data_ciphers_fallback"
              allow-clear
              style="width: 100%"
              @clear="delete data.data_ciphers_fallback"
            >
              <a-select-option
                v-for="item in tlsCiphers"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
    </template>

    <a-row
      v-if="optionAuth || optionName || optionMtu || (isServer && optionMaxClients)"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        v-if="optionAuth"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.auth')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.auth"
            style="width: 100%"
          >
            <a-select-option
              v-for="item in digests"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col
        v-if="optionName"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.interfaceName')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.name" />
        </a-form-item>
      </a-col>
      <a-col
        v-if="optionMtu"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          label="MTU"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="data.mtu"
            :min="0"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
      <a-col
        v-if="isServer && optionMaxClients"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openvpn.maxClients')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="data.max_clients"
            :min="0"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-row
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-space>
          <a-switch v-model:checked="data.system" />
          <span>{{ $t('types.openvpn.system') }}</span>
        </a-space>
      </a-col>
      <a-col
        v-if="isServer"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-space>
          <a-switch v-model:checked="data.duplicate_cn" />
          <span>{{ $t('types.openvpn.duplicateCn') }}</span>
        </a-space>
      </a-col>
    </a-row>

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
            <a-space v-if="!isServer && data.mode == 'tls'">
              <a-switch v-model:checked="optionCredentials" />
              <span>{{ $t('types.openvpn.credentials') }}</span>
            </a-space>
            <a-space v-if="!isServer && data.mode == 'static_key'">
              <a-switch v-model:checked="optionPeerIpv6" />
              <span>{{ $t('types.openvpn.peerAddressIpv6') }}</span>
            </a-space>
            <a-space v-if="data.mode != 'static_key'">
              <a-switch v-model:checked="optionDataCiphers" />
              <span>{{ $t('types.openvpn.dataCiphers') }}</span>
            </a-space>
            <a-space>
              <a-switch v-model:checked="optionAuth" />
              <span>{{ $t('types.openvpn.auth') }}</span>
            </a-space>
            <a-space>
              <a-switch v-model:checked="optionName" />
              <span>{{ $t('types.openvpn.interfaceName') }}</span>
            </a-space>
            <a-space>
              <a-switch v-model:checked="optionMtu" />
              <span>MTU</span>
            </a-space>
            <a-space v-if="isServer">
              <a-switch v-model:checked="optionMaxClients" />
              <span>{{ $t('types.openvpn.maxClients') }}</span>
            </a-space>
          </a-space>
        </template>
        <a-button type="default">
          {{ $t('types.openvpn.options') }}
        </a-button>
      </a-popover>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Endpoint } from '@/types/endpoints'

// One form serves both openvpn-client and openvpn-server, so it takes the
// widest shape either side can hold: the client dials `server`/`server_port`
// while the server listens on `listen`/`listen_port`, and only the server
// carries `max_clients`/`duplicate_cn`.
interface OpenVpnData {
  type: string
  server?: string
  server_port?: number
  listen?: string
  listen_port?: number
  max_clients?: number
  duplicate_cn?: boolean
  mode?: 'tls' | 'static_key'
  network?: 'udp' | 'udp4' | 'udp6' | 'tcp' | 'tcp4' | 'tcp6'
  name?: string
  address?: string[]
  peer_address?: string
  peer_address_ipv6?: string
  static_key_path?: string
  key_direction?: 'server' | 'client'
  username?: string
  password?: string
  mtu?: number
  cipher?: string
  data_ciphers?: string[]
  data_ciphers_fallback?: string
  auth?: string
  system?: boolean
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Endpoint>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as OpenVpnData)

const menu = ref(false)

const modes = [
  { title: 'TLS', value: 'tls' },
  { title: 'Static Key', value: 'static_key' },
]

const keyDirections = [
  { title: 'Server', value: 'server' },
  { title: 'Client', value: 'client' },
]

// The names sing-openvpn accepts (options_validation.go). A static_key tunnel
// can only use the CBC ones; a TLS session also takes the AEAD ciphers, which
// lead the list as the ones to prefer, and the CFB/OFB stream forms.
const cbcCiphers = [
  'AES-256-CBC', 'AES-192-CBC', 'AES-128-CBC',
  'CAMELLIA-256-CBC', 'CAMELLIA-192-CBC', 'CAMELLIA-128-CBC',
  'ARIA-256-CBC', 'ARIA-192-CBC', 'ARIA-128-CBC',
  'SEED-CBC', 'SM4-CBC', 'CAST5-CBC', 'BF-CBC',
  'DES-EDE3-CBC', 'DES-EDE-CBC', 'DES-CBC',
]
const staticCiphers = [...cbcCiphers, 'NONE']
const tlsCiphers = [
  'AES-256-GCM', 'AES-192-GCM', 'AES-128-GCM', 'CHACHA20-POLY1305',
  ...cbcCiphers,
  'AES-256-CFB', 'AES-192-CFB', 'AES-128-CFB',
  'AES-256-OFB', 'AES-192-OFB', 'AES-128-OFB',
]
const digests = ['SHA256', 'SHA384', 'SHA512', 'SHA224', 'SHA1', 'RIPEMD160', 'MD5', 'NONE']

// Each mode rejects the other's cipher options outright, so switching has
// to clear the ones that no longer apply. static_key additionally requires
// a CBC cipher: GCM relies on the TLS key exchange for IV uniqueness, so
// sing-box will not use it without a TLS session.
const mode = computed({
  get: (): string | undefined => data.value.mode,
  set: (v: 'tls' | 'static_key') => {
    data.value.mode = v
    if (v === 'static_key') {
      delete data.value.data_ciphers
      delete data.value.data_ciphers_fallback
      if (!data.value.cipher || !staticCiphers.includes(data.value.cipher)) {
        data.value.cipher = 'AES-256-CBC'
      }
    } else {
      delete data.value.cipher
    }
  },
})

const isServer = computed((): boolean => data.value.type === 'openvpn-server')

// The server accepts tcp/udp only; the client also takes the v4/v6 forms.
const networks = computed((): { title: string; value: string }[] => {
  const base = [
    { title: 'UDP', value: 'udp' },
    { title: 'TCP', value: 'tcp' },
  ]
  if (isServer.value) return base
  return base.concat([
    { title: 'UDPv4', value: 'udp4' },
    { title: 'UDPv6', value: 'udp6' },
    { title: 'TCPv4', value: 'tcp4' },
    { title: 'TCPv6', value: 'tcp6' },
  ])
})

// Optional fields stay out of the form, and out of the saved config, until
// switched on from the options menu.
const optionCredentials = computed({
  get: (): boolean => data.value.username !== undefined || data.value.password !== undefined,
  set: (v: boolean) => {
    if (v) {
      data.value.username = ''
      data.value.password = ''
    } else {
      delete data.value.username
      delete data.value.password
    }
  },
})

const optionPeerIpv6 = computed({
  get: (): boolean => data.value.peer_address_ipv6 !== undefined,
  set: (v: boolean) => {
    if (v) data.value.peer_address_ipv6 = ''
    else delete data.value.peer_address_ipv6
  },
})

// Starts from OpenVPN's own default list, so turning it on changes nothing
// until the list is edited.
const optionDataCiphers = computed({
  get: (): boolean => data.value.data_ciphers !== undefined || data.value.data_ciphers_fallback !== undefined,
  set: (v: boolean) => {
    if (v) {
      data.value.data_ciphers = ['AES-256-GCM', 'AES-128-GCM', 'CHACHA20-POLY1305']
    } else {
      delete data.value.data_ciphers
      delete data.value.data_ciphers_fallback
    }
  },
})

const optionAuth = computed({
  get: (): boolean => data.value.auth !== undefined,
  set: (v: boolean) => {
    if (v) data.value.auth = 'SHA256'
    else delete data.value.auth
  },
})

const optionName = computed({
  get: (): boolean => data.value.name !== undefined,
  set: (v: boolean) => {
    if (v) data.value.name = ''
    else delete data.value.name
  },
})

const optionMtu = computed({
  get: (): boolean => data.value.mtu !== undefined,
  set: (v: boolean) => {
    if (v) data.value.mtu = 1500
    else delete data.value.mtu
  },
})

const optionMaxClients = computed({
  get: (): boolean => data.value.max_clients !== undefined,
  set: (v: boolean) => {
    if (v) data.value.max_clients = 100
    else delete data.value.max_clients
  },
})
</script>

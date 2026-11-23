<template>
  <a-card
    size="small"
    title="Wireguard"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.wg.privKey')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.private_key">
              <template #suffix>
                <a-button
                  type="text"
                  size="small"
                  @click="newKey()"
                >
                  <template #icon>
                    <AppIcon name="key-star" />
                  </template>
                </a-button>
              </template>
            </a-input>
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('tls.pubKey')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="public_key"
              readonly
            >
              <template #suffix>
                <a-button
                  type="text"
                  size="small"
                  @click="getWgPubKey()"
                >
                  <template #icon>
                    <AppIcon name="refresh" />
                  </template>
                </a-button>
              </template>
            </a-input>
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.wg.localIp') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="address" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row :gutter="[12, 12]">
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
              v-model:value="listenPort"
              :min="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="data.udp_timeout != undefined"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            label="UDP Timeout"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="udp_timeout"
              :min="0"
              :addon-after="$t('date.m')"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="data.workers != undefined || data.mtu != undefined"
        :gutter="[12, 12]"
      >
        <a-col
          v-if="data.workers != undefined"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.wg.worker')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="data.workers"
              :min="1"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="data.mtu != undefined"
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
      </a-row>
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.wg.server')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.ext.server" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('dns.title') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.ext.dns" />
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
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.system" />
            <span>{{ $t('types.wg.sysIf') }}</span>
          </a-space>
        </a-col>
        <a-col
          v-if="data.system"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.wg.ifName')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="ifName" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-flex justify="end">
        <a-popover
          v-model:open="menu"
          trigger="click"
          placement="left"
        >
          <template #content>
            <a-flex
              vertical
              :gap="8"
            >
              <a-space>
                <a-switch v-model:checked="optionUdp" />
                <span>UDP Timeout</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionWorker" />
                <span>{{ $t('types.wg.worker') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionMtu" />
                <span>MTU</span>
              </a-space>
            </a-flex>
          </template>
          <a-button>
            {{ $t('types.wg.options') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
  <a-card
    v-if="data.peers != undefined"
    size="small"
    style="margin-top: 12px;"
  >
    <template #title>
      <a-space>
        <span>{{ $t('types.wg.peers') }}</span>
        <a-button
          type="primary"
          size="small"
          shape="circle"
          @click="addPeer"
        >
          <template #icon>
            <AppIcon name="plus" />
          </template>
        </a-button>
      </a-space>
    </template>
    <a-flex
      vertical
      :gap="12"
    >
      <a-card
        v-for="(p, index) in data.peers"
        :key="index"
        size="small"
      >
        <template #title>
          <a-space>
            <span>{{ $t('types.wg.peer') + ' ' + (Number(index) + 1) }}</span>
            <a-button
              type="text"
              danger
              size="small"
              shape="circle"
              :aria-label="$t('actions.del')"
              @click="delPeer(Number(index))"
            >
              <template #icon>
                <AppIcon name="delete" />
              </template>
            </a-button>
          </a-space>
        </template>
        <Peer
          :data="p"
          :ext="data.ext"
          @refresh-peer-key="refreshPeerKey(index)"
        />
      </a-card>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import Peer from '@/components/WgPeer.vue'
import { Endpoint, WgPeer } from '@/types/endpoints'

// The panel-only extras carried beside the endpoint's sing-box options: the
// public key derived from the private one, the key store the peers read their
// private keys from, and the hints handed to the generated client config.
interface WgExt {
  public_key?: string
  server?: string
  dns?: string
  keys: { public_key: string, private_key: string }[]
}

// The WireGuard endpoint this form edits. An option that is switched off is
// removed rather than written out, so those fields are optional here.
interface WireguardData {
  private_key: string
  listen_port: number
  address?: string[]
  name?: string
  system?: boolean
  mtu?: number
  workers?: number
  udp_timeout?: string
  peers?: WgPeer[]
  ext: WgExt
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Endpoint>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as WireguardData)

const emit = defineEmits<{
  newWgKey: []
  getWgPubKey: [privKey: string]
  addPeer: []
  delPeer: [id: number]
  refreshPeerKey: [id: number]
}>()

const menu = ref(false)

const optionUdp = computed({
  get: (): boolean => data.value.udp_timeout != undefined,
  set: (v: boolean) => { data.value.udp_timeout = v ? "5m" : undefined },
})

const optionWorker = computed({
  get: (): boolean => data.value.workers != undefined,
  set: (v: boolean) => { data.value.workers = v ? 2 : undefined },
})

const optionMtu = computed({
  get: (): boolean => data.value.mtu != undefined,
  set: (v: boolean) => { data.value.mtu = v ? 1408 : undefined },
})

const ifName = computed({
  get: () => data.value.name ?? '',
  set: (v: string) => { data.value.name = v.length > 0 ? v : undefined },
})

const listenPort = computed({
  get: () => data.value.listen_port > 0 ? data.value.listen_port : '',
  set: (v: number | string) => {
    const port = Number(v)
    data.value.listen_port = Number.isInteger(port) && port > 0 ? port : 0
  },
})

const address = computed({
  get: () => data.value.address?.join(','),
  set: (v: string) => { data.value.address = v.length > 0 ? v.split(',') : undefined },
})

const udp_timeout = computed({
  get: () => data.value.udp_timeout ? parseInt(data.value.udp_timeout.replace('m', '')) : 5,
  set: (v: number) => { data.value.udp_timeout = v > 0 ? v + 'm' : '5m' },
})

const public_key = computed({
  get: () => data.value.ext?.public_key ?? '',
  set: (v: string) => { data.value.ext.public_key = v },
})

function addPeer() {
  emit('addPeer')
}

function delPeer(id: number) {
  emit('delPeer', id)
}

function refreshPeerKey(id: number) {
  emit('refreshPeerKey', id)
}

function newKey() {
  emit('newWgKey')
}

function getWgPubKey() {
  const privKey = data.value.private_key
  if (privKey.length == 0) return
  emit('getWgPubKey', privKey)
}
</script>

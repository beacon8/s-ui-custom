<template>
  <a-card
    size="small"
    title="Warp"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <template v-if="data.id > 0">
        <table
          dir="ltr"
          width="100%"
        >
          <tbody>
            <tr>
              <td>Device ID</td>
              <td>{{ data.ext.device_id }}</td>
            </tr>
            <tr>
              <td>Access Token</td>
              <td>{{ data.ext.access_token }}</td>
            </tr>
            <tr>
              <td>{{ $t('types.wg.privKey') }}</td>
              <td>{{ data.private_key }}</td>
            </tr>
            <tr>
              <td>{{ $t('types.wg.localIp') }}</td>
              <td>{{ data.address.join(',') }}</td>
            </tr>
            <tr>
              <td colspan="2">
                <a-form-item
                  label="License Key"
                  style="margin-bottom: 0;"
                >
                  <a-input v-model:value="data.ext.license_key" />
                </a-form-item>
              </td>
            </tr>
          </tbody>
        </table>
        <a-card
          size="small"
          :title="$t('types.wg.peer')"
        >
          <a-flex
            vertical
            :gap="12"
          >
            <a-row :gutter="[12, 12]">
              <a-col
                :xs="24"
                :sm="12"
              >
                <a-form-item
                  :label="$t('out.addr')"
                  style="margin-bottom: 0;"
                >
                  <a-input v-model:value="data.peers[0].address" />
                </a-form-item>
              </a-col>
              <a-col
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item
                  :label="$t('out.port')"
                  style="margin-bottom: 0;"
                >
                  <a-input-number
                    v-model:value="data.peers[0].port"
                    :min="1"
                    style="width: 100%"
                  />
                </a-form-item>
              </a-col>
            </a-row>
            <table
              dir="ltr"
              width="100%"
            >
              <tbody>
                <tr>
                  <td>{{ $t('types.wg.pubKey') }}</td>
                  <td>{{ data.peers[0].public_key }}</td>
                </tr>
                <tr>
                  <td>{{ $t('types.wg.allowedIp') }}</td>
                  <td>{{ data.peers[0].allowed_ips.join(',') }}</td>
                </tr>
                <tr>
                  <td>Reserved</td>
                  <td>[{{ data.peers[0].reserved.join(',') }}]</td>
                </tr>
              </tbody>
            </table>
          </a-flex>
        </a-card>
      </template>
      <a-row
        v-if="data.udp_timeout != undefined || data.workers != undefined || data.mtu != undefined"
        :gutter="[12, 12]"
      >
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
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Endpoint } from '@/types/endpoints'

// The Warp-specific bookkeeping the panel keeps alongside the endpoint.
interface WarpExt {
  device_id?: string
  access_token?: string
  license_key?: string
}

// Only the Warp endpoint fields this form reads or writes. The peer block and
// `ext` are filled in by the registration step, so they are read as present.
interface WarpPeer {
  address: string
  port: number
  public_key: string
  allowed_ips: string[]
  reserved: number[]
}

interface WarpData {
  id: number
  ext: WarpExt
  address: string[]
  private_key: string
  peers: WarpPeer[]
  udp_timeout?: string
  workers?: number
  mtu?: number
  system?: boolean
  name?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Endpoint>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as WarpData)

const menu = ref(false)

const optionUdp = computed({
  get: (): boolean => data.value.udp_timeout != undefined,
  set: (v:boolean) => { data.value.udp_timeout = v ? "5m" : undefined },
})

const optionWorker = computed({
  get: (): boolean => data.value.workers != undefined,
  set: (v:boolean) => { data.value.workers = v ? 2 : undefined },
})

const optionMtu = computed({
  get: (): boolean => data.value.mtu != undefined,
  set: (v:boolean) => { data.value.mtu = v ? 1408 : undefined },
})

const ifName = computed({
  get: (): string => data.value.name ?? '',
  set: (v:string) => { data.value.name = v.length > 0 ? v : undefined },
})

const udp_timeout = computed({
  get: (): number => data.value.udp_timeout ? parseInt(data.value.udp_timeout.replace('m','')) : 5,
  set: (v:number) => { data.value.udp_timeout = v > 0 ? v + 'm' : '5m' },
})
</script>

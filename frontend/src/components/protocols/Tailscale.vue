<template>
  <a-card
    size="small"
    title="Talescale"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.ephemeral" />
            <span>{{ $t('types.ts.ephemeral') }}</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.accept_routes" />
            <span>{{ $t('types.ts.acceptRoutes') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row
        v-if="optionStateDir"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.ts.stateDir')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.state_directory" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionAuth"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.ts.authKey')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.auth_key" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionCtrlUrl"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.ts.controlUrl')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.control_url" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionHostname || optionUdpTimeout"
        :gutter="[12, 12]"
      >
        <a-col
          v-if="optionHostname"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.ts.hostname')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.hostname" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionUdpTimeout"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.ts.udpTimeout')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="udpTimeout"
              :min="1"
              :addon-after="$t('date.s')"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionExitNode"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.ts.exitNode')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.exit_node" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.exit_node_allow_lan_access" />
            <span>{{ $t('types.ts.allowLanAccess') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row
        v-if="optionRelay"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.ts.relayServerPort')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="data.relay_server_port"
              :min="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.ts.relayEndpoints') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="relay_endpoints" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionSysIf"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.system_interface" />
            <span>{{ $t('types.ts.systemInterface') }}</span>
          </a-space>
        </a-col>
        <a-col
          v-if="data.system_interface"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.ts.sysIfName')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.system_interface_name" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="data.system_interface"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.ts.sysIfMtu')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="data.system_interface_mtu"
              :min="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionAdvRoutes"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.ts.advRoutes') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="advertise_routes" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.advertise_exit_node" />
            <span>{{ $t('types.ts.advExitNode') }}</span>
          </a-space>
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
                <a-switch v-model:checked="optionStateDir" />
                <span>{{ $t('types.ts.stateDir') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionAuth" />
                <span>{{ $t('types.ts.authKey') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionCtrlUrl" />
                <span>{{ $t('types.ts.controlUrl') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionHostname" />
                <span>{{ $t('types.ts.hostname') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionExitNode" />
                <span>{{ $t('types.ts.exitNode') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionRelay" />
                <span>{{ $t('types.ts.relayServer') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionSysIf" />
                <span>{{ $t('types.ts.systemInterface') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionAdvRoutes" />
                <span>{{ $t('types.ts.advRoutes') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionUdpTimeout" />
                <span>{{ $t('types.ts.udpTimeout') }}</span>
              </a-space>
            </a-flex>
          </template>
          <a-button>
            {{ $t('types.ts.options') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Endpoint } from '@/types/endpoints'

// Only the Tailscale endpoint fields this form reads or writes.
interface TailscaleData {
  state_directory?: string
  auth_key?: string
  control_url?: string
  ephemeral?: boolean
  hostname?: string
  accept_routes?: boolean
  exit_node?: string
  exit_node_allow_lan_access?: boolean
  advertise_routes?: string[]
  advertise_exit_node?: boolean
  relay_server_port?: number
  relay_server_static_endpoints?: string[]
  system_interface?: boolean
  system_interface_name?: string
  system_interface_mtu?: number
  udp_timeout?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Endpoint>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as TailscaleData)

const menu = ref(false)

const optionStateDir = computed({
  get: (): boolean => data.value.state_directory !== undefined,
  set: (v: boolean) => { data.value.state_directory = v ? "$HOME/.tailscale" : undefined },
})

const optionAuth = computed({
  get: (): boolean => data.value.auth_key !== undefined,
  set: (v: boolean) => { data.value.auth_key = v ? "" : undefined },
})

const optionCtrlUrl = computed({
  get: (): boolean => data.value.control_url !== undefined,
  set: (v: boolean) => { data.value.control_url = v ? "https://controlplane.tailscale.com" : undefined },
})

const optionHostname = computed({
  get: (): boolean => data.value.hostname !== undefined,
  set: (v: boolean) => { data.value.hostname = v ? "localhost" : undefined },
})

const optionExitNode = computed({
  get: (): boolean => data.value.exit_node !== undefined,
  set: (v: boolean) => {
    if (v) {
      data.value.exit_node = ""
    } else {
      delete data.value.exit_node
      delete data.value.exit_node_allow_lan_access
    }
  },
})

const optionAdvRoutes = computed({
  get: (): boolean => data.value.advertise_routes !== undefined,
  set: (v: boolean) => {
    if (v) {
      data.value.advertise_routes = []
    } else {
      delete data.value.advertise_routes
      delete data.value.advertise_exit_node
    }
  },
})

const optionRelay = computed({
  get: (): boolean => data.value.relay_server_port !== undefined || (data.value.relay_server_static_endpoints?.length ?? 0) > 0,
  set: (v: boolean) => {
    if (v) {
      data.value.relay_server_port = 0
      data.value.relay_server_static_endpoints = []
    } else {
      delete data.value.relay_server_port
      delete data.value.relay_server_static_endpoints
    }
  },
})

const optionSysIf = computed({
  get: (): boolean => data.value.system_interface !== undefined,
  set: (v: boolean) => {
    if (v) {
      data.value.system_interface = false
    } else {
      delete data.value.system_interface
      delete data.value.system_interface_name
      delete data.value.system_interface_mtu
    }
  },
})

const optionUdpTimeout = computed({
  get: (): boolean => data.value.udp_timeout !== undefined,
  set: (v: boolean) => { data.value.udp_timeout = v ? '30s' : undefined },
})

const udpTimeout = computed({
  get: (): string => data.value.udp_timeout ? data.value.udp_timeout.replace('s','') : '',
  set: (v: number) => { data.value.udp_timeout = v>1 ? v + 's' : '30s' },
})

const advertise_routes = computed({
  get: (): string => data.value.advertise_routes?.join(',') ?? "",
  set: (v: string) => { data.value.advertise_routes = v.length > 0 ? v.split(',') : [] },
})

const relay_endpoints = computed({
  get: (): string => data.value.relay_server_static_endpoints?.join(',') ?? "",
  set: (v: string) => { data.value.relay_server_static_endpoints = v.length > 0 ? v.split(',') : [] },
})
</script>

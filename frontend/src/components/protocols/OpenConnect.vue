<template>
  <a-card
    size="small"
    title="OpenConnect"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openconnect.server')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="data.server"
            placeholder="vpn.example.com"
          />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openconnect.flavor')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.flavor"
            :options="flavors"
            :field-names="{ label: 'title', value: 'value' }"
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
          :label="$t('types.un')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="data.username"
            allow-clear
            @change="() => { if (!data.username) delete data.username }"
          />
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
          <a-input
            v-model:value="data.password"
            :type="showPassword ? 'text' : 'password'"
            allow-clear
            @change="() => { if (!data.password) delete data.password }"
          >
            <template #suffix>
              <a-button
                type="text"
                size="small"
                @click="genPassword"
              >
                <template #icon>
                  <AppIcon name="refresh" />
                </template>
              </a-button>
            </template>
          </a-input>
        </a-form-item>
      </a-col>
    </a-row>
    <a-row
      v-if="optionCookie"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col :xs="24">
        <a-form-item
          :label="$t('types.openconnect.cookie')"
          style="margin-bottom: 0;"
        >
          <a-input-password v-model:value="data.cookie" />
        </a-form-item>
      </a-col>
    </a-row>
    <a-row
      v-if="optionAuthGroup || optionName || optionMtu || optionUdpTimeout"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        v-if="optionAuthGroup"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openconnect.authGroup')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.auth_group" />
        </a-form-item>
      </a-col>
      <a-col
        v-if="optionName"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openconnect.interfaceName')"
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
        v-if="optionUdpTimeout"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.openconnect.udpTimeout')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="data.udp_timeout"
            placeholder="5m"
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
          <span>{{ $t('types.openconnect.system') }}</span>
        </a-space>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-space>
          <a-switch v-model:checked="data.no_udp" />
          <span>{{ $t('types.openconnect.noUdp') }}</span>
        </a-space>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-space>
          <a-switch v-model:checked="data.ipv6_disabled" />
          <span>{{ $t('types.openconnect.ipv6Disabled') }}</span>
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
            <a-space>
              <a-switch v-model:checked="optionAuthGroup" />
              <span>{{ $t('types.openconnect.authGroup') }}</span>
            </a-space>
            <a-space>
              <a-switch v-model:checked="optionCookie" />
              <span>{{ $t('types.openconnect.cookie') }}</span>
            </a-space>
            <a-space>
              <a-switch v-model:checked="optionName" />
              <span>{{ $t('types.openconnect.interfaceName') }}</span>
            </a-space>
            <a-space>
              <a-switch v-model:checked="optionMtu" />
              <span>MTU</span>
            </a-space>
            <a-space>
              <a-switch v-model:checked="optionUdpTimeout" />
              <span>{{ $t('types.openconnect.udpTimeout') }}</span>
            </a-space>
          </a-space>
        </template>
        <a-button type="default">
          {{ $t('types.openconnect.options') }}
        </a-button>
      </a-popover>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Endpoint } from '@/types/endpoints'
import RandomUtil from '@/plugins/randomUtil'

// Only the OpenConnect endpoint fields this form edits.
interface OpenConnectData {
  server: string
  flavor?: 'anyconnect' | 'gp' | 'fortinet' | 'f5' | 'pulse' | 'nc'
  name?: string
  username?: string
  password?: string
  auth_group?: string
  cookie?: string
  mtu?: number
  udp_timeout?: string
  system?: boolean
  no_udp?: boolean
  ipv6_disabled?: boolean
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Endpoint>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as OpenConnectData)

const flavors = [
  { title: 'Cisco AnyConnect', value: 'anyconnect' },
  { title: 'Palo Alto GlobalProtect', value: 'gp' },
  { title: 'Fortinet', value: 'fortinet' },
  { title: 'F5 BIG-IP', value: 'f5' },
  { title: 'Juniper Pulse', value: 'pulse' },
  { title: 'Junos Network Connect', value: 'nc' },
]

const menu = ref(false)

// A typed-in password stays masked, but a generated one is shown, since it
// still has to be copied to the account on the server.
const showPassword = ref(false)

const genPassword = () => {
  data.value.password = RandomUtil.randomSeq(16)
  showPassword.value = true
}

// Optional fields stay out of the form, and out of the saved config, until
// switched on from the options menu.
const optionAuthGroup = computed({
  get: (): boolean => data.value.auth_group !== undefined,
  set: (v: boolean) => {
    if (v) data.value.auth_group = ''
    else delete data.value.auth_group
  },
})

const optionCookie = computed({
  get: (): boolean => data.value.cookie !== undefined,
  set: (v: boolean) => {
    if (v) data.value.cookie = ''
    else delete data.value.cookie
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

const optionUdpTimeout = computed({
  get: (): boolean => data.value.udp_timeout !== undefined,
  set: (v: boolean) => {
    if (v) data.value.udp_timeout = '5m'
    else delete data.value.udp_timeout
  },
})
</script>

<template>
  <a-card
    size="small"
    title="Tun"
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
            :label="$t('types.tun.addr') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="addrs"
              placeholder="172.18.0.1/30"
            />
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
            :label="$t('types.tun.ifName')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.interface_name"
              placeholder="tun0"
              allow-clear
              @change="() => { if (!data.interface_name) delete data.interface_name }"
            />
          </a-form-item>
        </a-col>
        <a-col
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
              style="width: 100%"
            />
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
            label="UDP timeout"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="udpTimeout"
              :min="1"
              :addon-after="$t('date.m')"
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
            label="Stack"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.stack"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in ['system', 'gvisor', 'mixed']"
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
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="autoRoute" />
            <span>Auto Route</span>
          </a-space>
        </a-col>
        <a-col
          v-if="autoRoute"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.auto_redirect" />
            <span>Auto Redirect</span>
          </a-space>
        </a-col>
        <a-col
          v-if="autoRoute"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.strict_route" />
            <span>Strict Route</span>
          </a-space>
        </a-col>
        <a-col
          v-if="autoRoute && data.auto_redirect"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.exclude_mptcp" />
            <span>{{ $t('types.tun.excludeMptcp') }}</span>
          </a-space>
        </a-col>
        <a-col
          v-if="autoRoute && data.auto_redirect"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.tun.fallbackRuleIndex')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="fallbackRuleIndex"
              :min="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Inbound } from '@/types/inbounds'

// Only the Tun inbound fields this form reads or writes.
interface TunData {
  address?: string[]
  interface_name?: string
  mtu?: number
  udp_timeout?: string
  stack?: string
  auto_route?: boolean
  strict_route?: boolean
  auto_redirect?: boolean
  exclude_mptcp?: boolean
  auto_redirect_iproute2_fallback_rule_index?: number
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as TunData)

const addrs = computed({
  get: (): string | undefined => data.value.address?.join(','),
  set: (v:string) => { data.value.address = v.length > 0 ? v.split(',') : undefined },
})

const udpTimeout = computed({
  get: (): number => data.value.udp_timeout ? parseInt(data.value.udp_timeout.replace('m','')) : 5,
  set: (v:number) => { data.value.udp_timeout = v > 0 ? v + 'm' : '5m' },
})

const autoRoute = computed({
  get: (): boolean => data.value.auto_route ?? false,
  set: (v:boolean) => {
    data.value.auto_route = v
    data.value.auto_redirect = v ? false : undefined
    data.value.strict_route = v ? false : undefined
  },
})

const fallbackRuleIndex = computed({
  get: (): number => data.value.auto_redirect_iproute2_fallback_rule_index ?? 32768,
  set: (v: number) => {
    const val = typeof v === 'number' && !isNaN(v) && v >= 0 ? v : undefined
    data.value.auto_redirect_iproute2_fallback_rule_index = val
  },
})
</script>

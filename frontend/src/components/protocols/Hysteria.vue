<template>
  <a-card
    size="small"
    title="Hysteria"
  >
    <a-row :gutter="[12, 12]">
      <a-col
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
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.hy.obfs')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.obfs" />
        </a-form-item>
      </a-col>
      <a-col
        v-if="direction=='out'"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.hy.auth')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.auth_str" />
        </a-form-item>
      </a-col>
    </a-row>
    <a-row
      v-if="direction=='out'"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <Network :data="model" />
      </a-col>
    </a-row>
    <a-row
      v-if="direction=='out' && optionMPort"
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
          label="Hop interval"
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
            <a-space v-if="direction=='out'">
              <a-switch v-model:checked="optionMPort" />
              <span>{{ $t('rule.portRange') }}</span>
            </a-space>
          </a-space>
        </template>
        <a-button type="default">
          {{ $t('types.hy.hyOptions') }}
        </a-button>
      </a-popover>
    </a-flex>
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
import { Inbound } from '@/types/inbounds'
import { Outbound } from '@/types/outbounds'

// One form serves both directions, so it takes the widest shape either side
// can hold: only the outbound side carries the auth string, the port hopping
// and the network field.
interface HysteriaData {
  up_mbps?: number
  // v-model.number hands back the raw string when the field does not parse.
  down_mbps?: number | string
  down?: string
  obfs?: string
  auth_str?: string
  server_ports?: string[]
  hop_interval?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound | Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as HysteriaData)

defineProps<{
  direction: 'in' | 'out' | 'out_json'
}>()

const menu = ref(false)

const optionMPort = computed({
  get: (): boolean => data.value.server_ports != undefined,
  set: (v: boolean) => { data.value.server_ports = v ? [] : undefined },
})

const server_ports = computed({
  get: (): string => data.value.server_ports?.join(',') ?? '',
  set: (v: string) => { data.value.server_ports = v.length > 0 ? v.split(',').map((s: string) => s.trim()) : undefined },
})

const hop_interval = computed({
  get: (): number => {
    const v = data.value.hop_interval
    return v ? parseInt(v.replace('s', '')) : 0
  },
  set: (v: number) => { data.value.hop_interval = v > 0 ? v + 's' : undefined },
})

const down_mbps = computed({
  get: (): number | string => data.value.down_mbps ? data.value.down_mbps : 0,
  set: (newValue: number | string) => {
    if (typeof newValue != 'string' || newValue.length != 0) {
      data.value.down_mbps = newValue
      data.value.down = "" + newValue + " Mbps"
    } else {
      data.value.down_mbps = 0
      data.value.down = "0 Mbps"
    }
  },
})

const up_mbps = computed({
  get: (): number => data.value.up_mbps ? data.value.up_mbps : 0,
  set: (newValue: number) => { data.value.up_mbps = newValue > 0 ? newValue : 0 },
})
</script>

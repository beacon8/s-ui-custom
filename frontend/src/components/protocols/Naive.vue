<template>
  <a-card
    size="small"
    :title="direction != 'out_json' ? 'Naive' : undefined"
  >
    <!-- Inbound -->
    <template v-if="direction === 'in'">
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <Network :data="data" />
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.naive.quicCongestion')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.quic_congestion_control"
              :options="inbCngs"
              :field-names="{ label: 'title', value: 'value' }"
              allow-clear
              style="width: 100%"
              @clear="delete data.quic_congestion_control"
            />
          </a-form-item>
        </a-col>
      </a-row>
    </template>
    <!-- Outbound -->
    <template v-if="['out', 'out_json'].includes(direction)">
      <a-row
        v-if="direction === 'out'"
        :gutter="[12, 12]"
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
      <a-row
        :gutter="[12, 12]"
        align="middle"
        :style="direction === 'out' ? 'margin-top: 12px;' : undefined"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.naive.insecureConcurrency')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="insecure_concurrency"
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
          <a-space>
            <a-switch v-model:checked="udpOverTcp" />
            <span>{{ $t('types.naive.udpOverTcp') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row
        v-if="direction === 'out'"
        :gutter="[12, 12]"
        align="middle"
        style="margin-top: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.quic" />
            <span>{{ $t('types.naive.quic') }}</span>
          </a-space>
        </a-col>
        <a-col
          v-if="data.quic"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.naive.quicCongestion')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.quic_congestion_control"
              :options="outCngs"
              :field-names="{ label: 'title', value: 'value' }"
              allow-clear
              style="width: 100%"
              @clear="delete data.quic_congestion_control"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <Headers :data="extra_headers" />
    </template>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Network from '@/components/Network.vue'
import Headers from '@/components/Headers.vue'
import { Inbound } from '@/types/inbounds'
import { Outbound } from '@/types/outbounds'

// One form serves both directions, so it takes the widest shape either side
// can hold: the inbound side only picks a congestion control, while the
// client fields, the extra headers and the QUIC toggle belong to the
// outbound side.
interface NaiveData {
  quic_congestion_control?: 'bbr' | 'bbr_standard' | 'bbr2' | 'bbr2_variant' | 'cubic' | 'reno'
  username?: string
  password?: string
  insecure_concurrency?: number
  // sing-box accepts either a bare boolean or the versioned object form.
  udp_over_tcp?: boolean | { enabled?: boolean, version?: number }
  quic?: boolean
  extra_headers?: Record<string, string | string[]>
}

// Headers.vue edits a `headers` field, which this protocol calls
// `extra_headers`, so it is handed a view carrying that one name.
interface HeadersView {
  headers?: Record<string, string | string[]>
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound | Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as NaiveData)

defineProps<{
  direction: 'in' | 'out' | 'out_json'
}>()

const inbCngs = [
  { title: 'BBR', value: 'bbr'},
  { title: 'BBR Standard', value: 'bbr_standard'},
  { title: 'BBRv2', value: 'bbr2'},
  { title: 'BBRv2 variant', value: 'bbr2_variant'},
  { title: 'Cubic', value: 'cubic'},
  { title: 'New Reno', value: 'reno'},
]

const outCngs = [
  { title: 'BBR', value: 'bbr'},
  { title: 'BBR2', value: 'bbr2'},
  { title: 'Cubic', value: 'cubic'},
  { title: 'Reno', value: 'reno'},
]

const udpOverTcp = computed({
  get: (): boolean => {
    const u = data.value.udp_over_tcp
    return u === true || (typeof u === 'object' && u != null && u.enabled === true)
  },
  set: (v: boolean) => {
    if (v) {
      data.value.udp_over_tcp = { enabled: true }
    } else {
      data.value.udp_over_tcp = false
    }
  },
})

const insecure_concurrency = computed({
  get: (): number => data.value.insecure_concurrency ?? 0,
  set: (v: number) => {
    data.value.insecure_concurrency = v > 0 ? v : undefined
  },
})

const extra_headers = computed((): HeadersView => {
  const d = data.value
  return new Proxy({} as HeadersView, {
    get(_, prop) {
      if (prop === 'headers') return d?.extra_headers ?? {}
      return undefined
    },
    set(_, prop, value) {
      if (prop === 'headers') {
        d.extra_headers = value
        return true
      }
      return false
    }
  })
})
</script>

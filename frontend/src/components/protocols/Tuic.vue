<template>
  <a-card
    size="small"
    title="TUIC"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <a-row
        v-if="direction == 'out'"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            label="UUID"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.uuid" />
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
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            label="UDP Relay Mode"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.udp_relay_mode"
              allow-clear
              style="width: 100%"
              @change="(v: any) => { if (v === undefined) delete data.udp_relay_mode }"
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
        >
          <a-space>
            <a-switch v-model:checked="data.udp_over_stream" />
            <span>UDP Over Stream</span>
          </a-space>
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
          <a-form-item
            :label="$t('types.tuic.congControl')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.congestion_control"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in congestion_controls"
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
          <a-space>
            <a-switch v-model:checked="data.zero_rtt_handshake" />
            <span>Zero-RTT Handshake</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row :gutter="[12, 12]">
        <a-col
          v-if="direction == 'in'"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.tuic.authTimeout')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="auth_timeout"
              :min="1"
              :addon-after="$t('date.s')"
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
            :label="$t('types.tuic.hb')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="heartbeat"
              :min="1"
              :addon-after="$t('date.s')"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <QuicFields
        :data="model"
        quic
      />
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import QuicFields from '@/components/QuicFields.vue'
import Network from '@/components/Network.vue'
import { Inbound } from '@/types/inbounds'
import { Outbound } from '@/types/outbounds'

// One form serves both directions, so it takes the widest shape either side
// can hold: only the outbound side carries the client credentials and the UDP
// relay options, only the inbound side carries the auth timeout.
interface TuicData {
  uuid?: string
  password?: string
  udp_relay_mode?: "native" | "quic"
  udp_over_stream?: boolean
  congestion_control?: ""|"cubic"|"new_reno"|"bbr"
  zero_rtt_handshake?: boolean
  auth_timeout?: string
  heartbeat?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound | Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as TuicData)

defineProps<{
  direction: 'in' | 'out'
}>()

const congestion_controls = [
  "cubic","new_reno", "bbr"
]

const auth_timeout = computed({
  get: (): number | '' => data.value.auth_timeout ? parseInt(data.value.auth_timeout.replace('s','')) : '',
  set: (newValue:number) => { data.value.auth_timeout = newValue ? newValue + 's' : '' },
})

const heartbeat = computed({
  get: (): number | '' => data.value.heartbeat ? parseInt(data.value.heartbeat.replace('s','')) : '',
  set: (newValue:number) => { data.value.heartbeat = newValue ? newValue + 's' : '' },
})
</script>

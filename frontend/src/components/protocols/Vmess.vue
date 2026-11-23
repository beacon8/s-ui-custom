<template>
  <a-card
    size="small"
    title="VMESS"
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
            label="Alter ID"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="data.alter_id"
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
          <a-form-item
            :label="$t('types.vmess.security')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.security"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in securities"
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
          <a-space>
            <a-switch v-model:checked="data.global_padding" />
            <span>{{ $t('types.vmess.globalPadding') }}</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.authenticated_length" />
            <span>{{ $t('types.vmess.authLen') }}</span>
          </a-space>
        </a-col>
      </a-row>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Network from '@/components/Network.vue'
import { Outbound } from '@/types/outbounds'

// Only the VMess outbound fields this form edits.
interface VmessData {
  uuid: string
  alter_id: number
  security?: string
  global_padding?: boolean
  authenticated_length?: boolean
  packet_encoding?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as VmessData)

const securities = [
  "auto",
  "none",
  "zero",
  "aes-128-gcm",
  "aes-128-ctr",
  "chacha20-poly1305",
]

const packet_encoding = computed({
  get: (): string => data.value.packet_encoding != undefined ? data.value.packet_encoding : 'none',
  set: (newValue: string) => { data.value.packet_encoding = newValue != "none" ? newValue : undefined },
})
</script>
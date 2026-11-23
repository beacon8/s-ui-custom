<template>
  <a-card
    size="small"
    title="Shadowsocks"
  >
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
          :label="$t('in.ssMethod')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.method"
            style="width: 100%"
            @change="direction == 'in' ? changeMethod(String($event)) : undefined"
          >
            <a-select-option
              v-for="m in ssMethods"
              :key="m"
              :value="m"
            >
              {{ m }}
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
        v-if="direction == 'out'"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <UoT :data="data" />
      </a-col>
      <a-col
        v-if="direction == 'in'"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-space>
          <a-switch v-model:checked="data.managed" />
          <span>{{ $t('in.ssManageable') }}</span>
        </a-space>
      </a-col>
    </a-row>
    <a-row
      v-if="data.method != 'none' || direction == 'out'"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="16"
      >
        <a-form-item
          :label="$t('types.pw')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.password">
            <template
              v-if="direction == 'in'"
              #suffix
            >
              <a-button
                type="text"
                size="small"
                @click="changeMethod(data.method)"
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
      v-if="direction == 'out'"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          label="Plugin"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="data.plugin"
            allow-clear
          />
        </a-form-item>
      </a-col>
      <a-col
        v-if="data.plugin"
        :xs="24"
        :sm="16"
      >
        <a-form-item
          label="Plugin Options"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.plugin_opts" />
        </a-form-item>
      </a-col>
    </a-row>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Network from '@/components/Network.vue'
import UoT from '@/components/UoT.vue'
import RandomUtil from '@/plugins/randomUtil'
import { Inbound } from '@/types/inbounds'
import { Outbound } from '@/types/outbounds'

// One form serves both directions, so it takes the widest shape either side
// can hold: only the inbound side is manageable, and only the outbound side
// carries the SIP003 plugin fields.
interface ShadowsocksData {
  method: string
  password?: string
  managed?: boolean
  plugin?: string
  plugin_opts?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound | Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as ShadowsocksData)

defineProps<{
  direction: 'in' | 'out'
}>()

const ssMethods = [
  "none",
  "aes-128-gcm",
  "aes-192-gcm",
  "aes-256-gcm",
  "chacha20-ietf-poly1305",
  "xchacha20-ietf-poly1305",
  "2022-blake3-aes-128-gcm",
  "2022-blake3-aes-256-gcm",
  "2022-blake3-chacha20-poly1305"
]

function changeMethod(ssMethod :string) {
  if (ssMethod.startsWith('2022')) {
    data.value.password = ssMethod == "2022-blake3-aes-128-gcm" ? RandomUtil.randomShadowsocksPassword(16) : RandomUtil.randomShadowsocksPassword(32)
  } else if (ssMethod == 'none') {
    delete data.value.password
  } else {
    data.value.password = RandomUtil.randomSeq(10)
  }
}
</script>
<template>
  <a-card
    size="small"
    title="SOCKS"
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
          <a-form-item
            :label="$t('types.un')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="username" />
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
            <a-input v-model:value="password" />
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
            :label="$t('version')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.version"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in ['4', '4a', '5']"
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
          <UoT :data="data" />
        </a-col>
      </a-row>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Network from '@/components/Network.vue'
import UoT from '@/components/UoT.vue'
import { Outbound } from '@/types/outbounds'

// Only the SOCKS outbound fields this form edits. Credentials are dropped
// entirely when left blank, so they are optional.
interface SocksData {
  version?: '4' | '4a' | '5'
  username?: string
  password?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as SocksData)

const username = computed({
  get: (): string => {
    const v = data.value.username
    return v != undefined && v.length > 0 ? v : ''
  },
  set: (v: string) => { data.value.username = v.length > 0 ? v : undefined },
})

const password = computed({
  get: (): string => {
    const v = data.value.password
    return v != undefined && v.length > 0 ? v : ''
  },
  set: (v: string) => { data.value.password = v.length > 0 ? v : undefined },
})
</script>
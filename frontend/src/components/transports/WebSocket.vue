<template>
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
          :label="$t('transport.path')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="transport.path" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('transport.host')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="host" />
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
          label="Max Early Data"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="max_early_data"
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
        <a-form-item
          label="Early Data Header Name"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="transport.early_data_header_name" />
        </a-form-item>
      </a-col>
    </a-row>
    <Headers :data="transport" />
  </a-flex>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Transport } from '../../types/transport'
import Headers from '../Headers.vue'

// Only the WebSocket fields this form reads or writes; headers is handed to
// Headers.vue, which edits it.
interface WebSocketData {
  path?: string
  headers?: { Host: string }
  max_early_data?: number
  early_data_header_name?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Transport>('transport', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const transport = computed(() => model.value as unknown as WebSocketData)

const max_early_data = computed({
  get: (): number | '' => transport.value.max_early_data ? transport.value.max_early_data : '',
  set: (newValue: number) => { transport.value.max_early_data = newValue ? newValue : undefined },
})

const host = computed({
  get: (): string => transport.value.headers?.Host ? transport.value.headers.Host : '',
  set: (newValue: string) => {
    transport.value.headers = newValue != "" ? { Host: newValue } : undefined
  },
})

onMounted(() => {
  transport.value.early_data_header_name ??= 'Sec-WebSocket-Protocol'
  transport.value.path ??= '/'
})
</script>

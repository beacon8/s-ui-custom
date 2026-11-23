<template>
  <a-flex
    vertical
    :gap="12"
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
          :label="$t('transport.grpcServiceName')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="transport.service_name" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-space>
          <a-switch v-model:checked="transport.permit_without_stream" />
          <span>{{ $t('transport.grpcPws') }}</span>
        </a-space>
      </a-col>
    </a-row>
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('transport.idleTimeout')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="idle_timeout"
            addon-after="s"
            :min="1"
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
          :label="$t('transport.pingTimeout')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="ping_timeout"
            addon-after="s"
            :min="1"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>
  </a-flex>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Transport } from '../../types/transport'

// Only the gRPC fields this form reads or writes.
interface GrpcData {
  service_name?: string
  permit_without_stream?: boolean
  idle_timeout?: string
  ping_timeout?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Transport>('transport', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const transport = computed(() => model.value as unknown as GrpcData)

const idle_timeout = computed({
  get: (): number | '' => transport.value.idle_timeout ? parseInt(transport.value.idle_timeout.replace('s', '')) : '',
  set: (newValue: number) => { transport.value.idle_timeout = newValue ? newValue + 's' : '' },
})

const ping_timeout = computed({
  get: (): number | '' => transport.value.ping_timeout ? parseInt(transport.value.ping_timeout.replace('s', '')) : '',
  set: (newValue: number) => { transport.value.ping_timeout = newValue ? newValue + 's' : '' },
})
</script>

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
          :label="$t('transport.hosts')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="transport.host" />
        </a-form-item>
      </a-col>
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
    </a-row>
    <Headers :data="transport" />
  </a-flex>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Transport } from '../../types/transport'
import Headers from '../Headers.vue'

// Only the httpupgrade fields this form reads or writes; headers is handed to
// Headers.vue, which edits it.
interface HttpUpgradeData {
  host?: string
  path?: string
  headers?: Record<string, string | string[]>
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Transport>('transport', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const transport = computed(() => model.value as unknown as HttpUpgradeData)
</script>

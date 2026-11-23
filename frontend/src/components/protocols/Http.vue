<template>
  <a-card
    size="small"
    title="HTTP"
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
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('transport.path')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.path" />
        </a-form-item>
      </a-col>
    </a-row>
    <Headers :data="model" />
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Headers from '@/components/Headers.vue'
import { Outbound } from '@/types/outbounds'

// Outbound HTTP proxy: credentials are optional and removed when emptied, and
// the headers are owned by the Headers child.
interface HttpData {
  username?: string
  password?: string
  path?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as HttpData)

const username = computed({
  get: (): string => {
    const v = data.value.username
    return v && v.length > 0 ? v : ''
  },
  set: (v: string) => { data.value.username = v.length > 0 ? v : undefined },
})

const password = computed({
  get: (): string => {
    const v = data.value.password
    return v && v.length > 0 ? v : ''
  },
  set: (v: string) => { data.value.password = v.length > 0 ? v : undefined },
})
</script>

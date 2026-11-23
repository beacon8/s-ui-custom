<template>
  <a-card
    size="small"
    title="Direct"
  >
    <a-row :gutter="[12, 12]">
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
          :label="$t('types.direct.overrideAddr')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.override_address" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.direct.overridePort')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="override_port"
            :min="0"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Network from '@/components/Network.vue'
import { Inbound } from '@/types/inbounds'

// Direct is inbound only; the network field is owned by the Network child.
interface DirectData {
  override_address?: string
  override_port?: number
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as DirectData)

// v-model.number hands back a number when the field parses and the raw string
// when it does not, so both have to be accepted here.
const override_port = computed({
  get: (): number | string => data.value.override_port ? data.value.override_port : '',
  set: (newValue: number | string) => {
    const text = String(newValue)
    data.value.override_port = text.length == 0 || Number(text) == 0 ? undefined : parseInt(text)
  },
})
</script>

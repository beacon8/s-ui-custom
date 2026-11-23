<template>
  <a-card
    size="small"
    title="Selector"
  >
    <a-row
      :gutter="[12, 12]"
      align="middle"
    >
      <a-col
        :xs="24"
        :sm="12"
      >
        <a-form-item
          :label="$t('pages.outbounds')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.outbounds"
            mode="tags"
            style="width: 100%"
            @change="updateDefault"
          >
            <a-select-option
              v-for="tag in tags"
              :key="tag"
              :value="tag"
            >
              {{ tag }}
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
          :label="$t('types.lb.defaultOut')"
          style="margin-bottom: 0;"
        >
          <a-auto-complete
            v-model:value="data.default"
            :options="(data.outbounds ?? []).map((o) => ({ value: o }))"
            allow-clear
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
      >
        <a-space>
          <a-switch v-model:checked="data.interrupt_exist_connections" />
          <span>{{ $t('types.lb.interruptConn') }}</span>
        </a-space>
      </a-col>
    </a-row>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Outbound } from '@/types/outbounds'

// Selector is outbound only: it groups other outbounds and names one of them
// as the default.
interface SelectorData {
  outbounds?: string[]
  default?: string
  interrupt_exist_connections?: boolean
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as SelectorData)

defineProps<{
  tags: string[]
}>()

function updateDefault() {
  const def = data.value.default
  if (def === undefined || !data.value.outbounds?.includes(def)) {
    delete data.value.default
  }
}
</script>

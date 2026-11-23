<template>
  <a-card
    size="small"
    :title="direction != 'out_json' ? 'Bridge' : undefined"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.bridge.interface')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="data.interface"
            allow-clear
            @change="() => { if (!data.interface) delete data.interface }"
          />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.bridge.bridgeName')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="data.bridge_name"
            allow-clear
            @change="() => { if (!data.bridge_name) delete data.bridge_name }"
          />
        </a-form-item>
      </a-col>
    </a-row>
    <a-row
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.bridge.tableIndex')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="tableIndex"
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
          :label="$t('types.bridge.ruleIndex')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="ruleIndex"
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
import { Outbound } from '@/types/outbounds'

// Only the outbound side carries a bridge, so this is the outbound shape.
interface BridgeData {
  interface?: string
  bridge_name?: string
  iproute2_table_index?: number
  iproute2_rule_index?: number
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as BridgeData)

defineProps<{
  direction: 'in' | 'out' | 'out_json'
}>()

const tableIndex = computed({
  get: (): number => data.value.iproute2_table_index ?? 0,
  set: (v: number) => { data.value.iproute2_table_index = v > 0 ? v : undefined },
})

const ruleIndex = computed({
  get: (): number => data.value.iproute2_rule_index ?? 0,
  set: (v: number) => { data.value.iproute2_rule_index = v > 0 ? v : undefined },
})
</script>

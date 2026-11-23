<template>
  <a-card
    size="small"
    title="ShadowTls"
  >
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
            v-model:value="version"
            style="width: 100%"
          >
            <a-select-option
              v-for="v in [1, 2, 3]"
              :key="v"
              :value="v"
            >
              {{ v }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col
        v-if="data.version != undefined && data.version > 1"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.pw')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.password" />
        </a-form-item>
      </a-col>
    </a-row>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Outbound } from '@/types/outbounds'

// Outbound ShadowTLS: v1 carries no password, v2 and v3 do.
interface ShadowTlsData {
  version?: number
  password?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as ShadowTlsData)

const version = computed({
  get: (): number => data.value.version ?? 3,
  set: (v: number) => {
    data.value.version = v
    if (v == 1) {
      delete data.value.password
    } else if (data.value.password === undefined) {
      data.value.password = ""
    }
  },
})
</script>

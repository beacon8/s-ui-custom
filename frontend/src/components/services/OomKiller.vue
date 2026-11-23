<template>
  <a-card
    size="small"
    title="OOM Killer"
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
            :label="$t('types.oomKiller.memoryLimit')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.memory_limit"
              allow-clear
              placeholder="1gb"
              @change="(e: any) => { if (!e.target?.value) delete data.memory_limit }"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.oomKiller.safetyMargin')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.safety_margin"
              allow-clear
              placeholder="128mb"
              @change="(e: any) => { if (!e.target?.value) delete data.safety_margin }"
            />
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
            :label="$t('types.oomKiller.minInterval')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.min_interval"
              allow-clear
              placeholder="10s"
              @change="(e: any) => { if (!e.target?.value) delete data.min_interval }"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.oomKiller.maxInterval')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.max_interval"
              allow-clear
              placeholder="1m"
              @change="(e: any) => { if (!e.target?.value) delete data.max_interval }"
            />
          </a-form-item>
        </a-col>
      </a-row>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Srv } from '@/types/services'

// Only the oom-killer fields this form edits.
interface OomKillerData {
  memory_limit?: string
  safety_margin?: string
  min_interval?: string
  max_interval?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Srv>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as OomKillerData)
</script>

<template>
  <a-card
    size="small"
    title="URL Test"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <a-row :gutter="[12, 12]">
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
      </a-row>
      <a-row
        v-if="optionUrl"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.lb.testUrl')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.url" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionInterval || optionTolerance || optionIdle"
        :gutter="[12, 12]"
      >
        <a-col
          v-if="optionInterval"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.lb.interval')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="interval"
              :min="3"
              :addon-after="$t('date.s')"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionTolerance"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.lb.tolerance')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="tolerance"
              :min="0"
              :addon-after="$t('date.ms')"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionIdle"
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
              :min="0"
              :addon-after="$t('date.m')"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row :gutter="[12, 12]">
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
      <a-flex justify="end">
        <a-popover
          v-model:open="menu"
          trigger="click"
          placement="left"
        >
          <template #content>
            <a-flex
              vertical
              :gap="8"
            >
              <a-space>
                <a-switch v-model:checked="optionUrl" />
                <span>{{ $t('types.lb.testUrl') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionInterval" />
                <span>{{ $t('types.lb.interval') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionTolerance" />
                <span>{{ $t('types.lb.tolerance') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionIdle" />
                <span>{{ $t('transport.idleTimeout') }}</span>
              </a-space>
            </a-flex>
          </template>
          <a-button>
            {{ $t('types.lb.urlTestOptions') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Outbound } from '@/types/outbounds'

// Only the URLTest outbound fields this form reads or writes.
interface UrlTestData {
  outbounds?: string[]
  url?: string
  interval?: string
  tolerance?: number
  idle_timeout?: string
  interrupt_exist_connections?: boolean
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as UrlTestData)

defineProps<{
  tags: string[]
}>()

const menu = ref(false)

const optionUrl = computed({
  get: (): boolean => data.value.url != undefined,
  set: (v:boolean) => { data.value.url = v ? 'https://www.gstatic.com/generate_204' : undefined },
})

const optionInterval = computed({
  get: (): boolean => data.value.interval != undefined,
  set: (v:boolean) => { data.value.interval = v ? '3s' : undefined },
})

const optionTolerance = computed({
  get: (): boolean => data.value.tolerance != undefined,
  set: (v:boolean) => { data.value.tolerance = v ? 50 : undefined },
})

const optionIdle = computed({
  get: (): boolean => data.value.idle_timeout != undefined,
  set: (v:boolean) => { data.value.idle_timeout = v ? '30m' : undefined },
})

const interval = computed({
  get: (): number => data.value.interval ? parseInt(data.value.interval.replace('s','')) : 3,
  set: (v:number) => { data.value.interval = v > 0 ? v + 's' : '3s' },
})

const tolerance = computed({
  get: (): number => data.value.tolerance ? parseInt(String(data.value.tolerance)) : 0,
  set: (v:number) => { data.value.tolerance = v > 0 ? v : 0 },
})

const idle_timeout = computed({
  get: (): number => data.value.idle_timeout ? parseInt(data.value.idle_timeout.replace('m','')) : 30,
  set: (v:number) => { data.value.idle_timeout = v > 0 ? v + 'm' : '0m' },
})
</script>

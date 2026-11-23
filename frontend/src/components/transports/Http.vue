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
          <a-input v-model:value="hosts" />
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
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('transport.httpMethod')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="transport.method"
            allow-clear
            style="width: 100%"
            @change="(v: any) => { if (!v) delete transport.method }"
          >
            <a-select-option
              v-for="item in methodList"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-select-option>
          </a-select>
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
    <Headers :data="transport" />
  </a-flex>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Transport } from '../../types/transport'
import Headers from '../Headers.vue'

// Only the HTTP fields this form reads or writes; headers is handed to
// Headers.vue, which edits it.
interface HttpData {
  host?: string[]
  path?: string
  method?: string
  headers?: Record<string, string | string[]>
  idle_timeout?: string
  ping_timeout?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Transport>('transport', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const transport = computed(() => model.value as unknown as HttpData)

const methodList = ['POST', 'GET', 'PUT', 'PATCH', 'DELETE']

const hosts = computed({
  get: (): string => transport.value.host ? transport.value.host.join(',') : '',
  set: (newValue: string) => { transport.value.host = newValue.length > 0 ? newValue.split(',') : [] },
})

const idle_timeout = computed({
  get: (): number | '' => transport.value.idle_timeout ? parseInt(transport.value.idle_timeout.replace('s', '')) : '',
  set: (newValue: number) => { transport.value.idle_timeout = newValue ? newValue + 's' : '' },
})

const ping_timeout = computed({
  get: (): number | '' => transport.value.ping_timeout ? parseInt(transport.value.ping_timeout.replace('s', '')) : '',
  set: (newValue: number) => { transport.value.ping_timeout = newValue ? newValue + 's' : '' },
})
</script>

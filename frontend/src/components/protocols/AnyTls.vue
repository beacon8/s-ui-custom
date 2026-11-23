<template>
  <a-card
    size="small"
    :title="direction != 'out_json' ? 'AnyTls' : undefined"
  >
    <a-row
      v-if="direction == 'in'"
      :gutter="[12, 12]"
    >
      <a-col
        :xs="24"
        :sm="16"
      >
        <a-form-item
          label="Padding scheme"
          style="margin-bottom: 0;"
        >
          <a-textarea
            v-model:value="padding_scheme"
            auto-size
          />
        </a-form-item>
      </a-col>
    </a-row>
    <a-row
      v-else
      :gutter="[12, 12]"
    >
      <a-col
        v-if="direction == 'out'"
        :xs="24"
        :sm="16"
      >
        <a-form-item
          :label="$t('types.pw')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.password" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.anytls.idleInterval')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="idleInterval"
            :min="0"
            :addon-after="$t('date.s')"
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
          :label="$t('types.anytls.idleTimeout')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="idleTimeout"
            :min="0"
            :addon-after="$t('date.s')"
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
          :label="$t('types.anytls.minIdle')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="minIdle"
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
import { Inbound } from '@/types/inbounds'
import { Outbound } from '@/types/outbounds'

// One form serves both directions: the inbound side carries the padding
// scheme, the outbound side the password and the idle session tuning.
interface AnyTlsData {
  padding_scheme?: string[]
  password?: string
  idle_session_check_interval?: string
  idle_session_timeout?: string
  min_idle_session?: number
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound | Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as AnyTlsData)

defineProps<{
  direction: 'in' | 'out' | 'out_json'
}>()

const padding_scheme = computed({
  get: (): string => {
    const scheme = data.value.padding_scheme
    return scheme && scheme.length > 0 ? scheme.join("\n") : ''
  },
  set: (v: string) => { data.value.padding_scheme = v.length > 0 ? v.split("\n") : undefined },
})

const idleInterval = computed({
  get: (): number => {
    const v = data.value.idle_session_check_interval
    return v && v.length > 0 ? parseInt(v.replace('s', '')) : 30
  },
  set: (v: number) => { data.value.idle_session_check_interval = v && v >= 0 ? `${v}s` : undefined },
})

const idleTimeout = computed({
  get: (): number => {
    const v = data.value.idle_session_timeout
    return v && v.length > 0 ? parseInt(v.replace('s', '')) : 30
  },
  set: (v: number) => { data.value.idle_session_timeout = v && v >= 0 ? `${v}s` : undefined },
})

const minIdle = computed({
  get: (): number => data.value.min_idle_session != undefined ? data.value.min_idle_session : 0,
  set: (v: number) => { data.value.min_idle_session = v > 0 ? v : undefined },
})
</script>

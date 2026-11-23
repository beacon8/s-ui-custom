<template>
  <a-card
    size="small"
    :title="$t('objects.tls')"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('template')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="inbound.tls_id"
            :options="tlsItems"
            :field-names="{ label: 'title', value: 'value' }"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { i18n } from '@/locales'
import { Inbound } from '@/types/inbounds'
import { Srv } from '@/types/services'

// Inbounds and services both reference a TLS template by id, and that is the
// only field this form touches.
interface TlsTarget {
  tls_id: number
}

// A saved TLS template, as listed in the select.
interface TlsConfig {
  id: number
  name: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound | Srv>('inbound', { required: true })

// The parent holds a union the template checker cannot follow, so narrow once
// here. The computed returns the same object, so edits still reach the parent.
const inbound = computed(() => model.value as unknown as TlsTarget)

const props = defineProps<{
  tlsConfigs?: TlsConfig[]
}>()

const tlsItems = computed((): { title: string, value: number }[] => {
  return [ { title: i18n.global.t('none'), value: 0 }, ...(props.tlsConfigs?.map((t) => { return { title: t.name, value: t.id } }) ?? [])]
})
</script>

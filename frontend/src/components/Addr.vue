<template>
  <a-row :gutter="[12, 12]">
    <a-col
      :xs="24"
      :sm="12"
      :md="8"
    >
      <a-form-item
        :label="$t('out.addr')"
        style="margin-bottom: 0;"
      >
        <a-input
          v-model:value="addr.server"
          required
        />
      </a-form-item>
    </a-col>
    <a-col
      :xs="24"
      :sm="12"
      :md="8"
    >
      <a-form-item
        :label="$t('out.port')"
        style="margin-bottom: 0;"
      >
        <a-input-number
          v-model:value="addr.server_port"
          style="width: 100%;"
          required
        />
      </a-form-item>
    </a-col>
    <a-col
      v-if="optionRemark"
      :xs="24"
      :sm="12"
      :md="8"
    >
      <a-form-item
        :label="$t('in.remark')"
        style="margin-bottom: 0;"
      >
        <a-input v-model:value="addr.remark" />
      </a-form-item>
    </a-col>
  </a-row>
  <OutTLS
    v-if="optionTLS"
    :outbound="addr"
  />
  <a-flex
    justify="end"
    style="margin-top: 12px;"
  >
    <a-popover
      v-model:open="menu"
      trigger="click"
      placement="left"
    >
      <template #content>
        <a-flex
          vertical
          gap="small"
        >
          <a-space>
            <a-switch v-model:checked="optionRemark" />
            <span>{{ $t('in.remark') }}</span>
          </a-space>
          <a-space v-if="hasTls">
            <a-switch v-model:checked="optionTLS" />
            <span>{{ $t('objects.tls') }}</span>
          </a-space>
        </a-flex>
      </template>
      <a-button type="default">
        {{ $t('in.mdOption') }}
      </a-button>
    </a-popover>
  </a-flex>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import OutTLS from '@/components/tls/OutTLS.vue'
import { Addr } from '@/types/inbounds'
import { oTls } from '@/types/tls'

// The address row this form edits. The TLS switch stores the outbound TLS
// block itself, which OutTLS then fills in.
interface AddrData {
  server: string
  server_port: number
  remark?: string
  tls?: oTls
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Addr>('addr', { required: true })

// Addr declares tls as a flag while the form keeps the TLS block there, so
// narrow once here. The computed returns the same object, so edits still reach
// the parent.
const addr = computed(() => model.value as unknown as AddrData)

defineProps<{
  hasTls: boolean
}>()

const menu = ref(false)

const optionTLS = computed({
  get: (): boolean => addr.value.tls != undefined,
  set: (v: boolean) => { addr.value.tls = v ? { enabled: true } : undefined },
})

const optionRemark = computed({
  get: (): boolean => addr.value.remark != undefined,
  set: (v: boolean) => { addr.value.remark = v ? '' : undefined },
})
</script>

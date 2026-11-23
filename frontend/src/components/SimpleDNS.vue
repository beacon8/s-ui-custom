<template>
  <a-row :gutter="[8, 8]">
    <a-col :span="24">
      <a-typography-text type="secondary">
        {{ label }}
      </a-typography-text>
    </a-col>
    <a-col :span="data.type == 'local' ? 24 : 8">
      <a-form-item
        :label="$t('type')"
        style="margin-bottom: 0;"
      >
        <a-select
          v-model:value="data.type"
          size="small"
          style="width: 100%"
          @change="updateType($event as string)"
        >
          <a-select-option
            v-for="item in ['udp', 'tcp', 'local', 'tls', 'quic', 'h3']"
            :key="item"
            :value="item"
          >
            {{ item }}
          </a-select-option>
        </a-select>
      </a-form-item>
    </a-col>
    <a-col
      v-if="data.type != 'local'"
      :span="10"
    >
      <a-form-item
        :label="$t('in.addr')"
        style="margin-bottom: 0;"
      >
        <a-input
          v-model:value="data.server"
          size="small"
        />
      </a-form-item>
    </a-col>
    <a-col
      v-if="data.type != 'local'"
      :span="6"
    >
      <a-form-item
        :label="$t('in.port')"
        style="margin-bottom: 0;"
      >
        <a-input-number
          v-model:value="data.server_port"
          size="small"
          :min="1"
          style="width: 100%"
        />
      </a-form-item>
    </a-col>
  </a-row>
</template>

<script setup lang="ts">
// The one DNS server entry this row edits.
interface DnsServer {
  type?: string
  server?: string
  server_port?: number
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const data = defineModel<DnsServer>('data', { required: true })

defineProps<{
  label: string
}>()

function updateType(t: string) {
  if (t == 'local') {
    delete data.value.server
    delete data.value.server_port
  }
}
</script>
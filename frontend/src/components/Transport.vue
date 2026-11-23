<template>
  <a-card
    size="small"
    :title="$t('objects.transport')"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
        style="display: flex; align-items: center;"
      >
        <a-space>
          <a-switch v-model:checked="tpEnable" />
          <span>{{ $t('transport.enable') }}</span>
        </a-space>
      </a-col>
      <a-col
        v-if="tpEnable"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('type')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="transportType"
            :options="Object.keys(trspTypes).map((key, index) => ({ title: key, value: Object.values(trspTypes)[index] }))"
            :field-names="{ label: 'title', value: 'value' }"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>
    <Http
      v-if="Transport.type == trspTypes.HTTP"
      :transport="Transport"
    />
    <WebSocket
      v-if="Transport.type == trspTypes.WebSocket"
      :transport="Transport"
    />
    <GRPC
      v-if="Transport.type == trspTypes.gRPC"
      :transport="Transport"
    />
    <HttpUpgrade
      v-if="Transport.type == trspTypes.HTTPUpgrade"
      :transport="Transport"
    />
  </a-card>
</template>

<script lang="ts">
import { PropType } from 'vue'
import { TrspTypes, Transport } from '@/types/transport'
import Http from './transports/Http.vue'
import WebSocket from './transports/WebSocket.vue'
import GRPC from './transports/gRPC.vue'
import HttpUpgrade from './transports/HttpUpgrade.vue'

// The inbound or outbound that owns the transport block. The panel writes an
// empty object when the transport is switched off, so the key is always there.
interface TransportHolder {
  transport: Partial<Transport>
}

export default {
  components: { Http, WebSocket, GRPC, HttpUpgrade },
  props: {
    // Inbounds and outbounds both bind here, so the prop asks only for an
    // object and the narrowing happens below.
    data: { type: Object as PropType<object>, required: true }
  },
  data() {
    return {
      trspTypes: TrspTypes
    }
  },
  computed: {
    // Narrow once, since no caller hands over a type the template checker could
    // follow. The computed returns the same object, so edits reach the parent.
    holder(): TransportHolder { return <TransportHolder>this.$props.data },
    Transport() {
      return <Transport>this.holder.transport
    },
    tpEnable: {
      get() { return Object.hasOwn(this.holder.transport, 'type') },
      set(newValue: boolean) { this.holder.transport = newValue ? { type: 'http' } : {} }
    },
    transportType: {
      get() { return this.Transport.type },
      set(newValue: string) { this.holder.transport = { type: newValue } }
    }
  }
}
</script>
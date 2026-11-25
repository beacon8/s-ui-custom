<template>
  <a-card
    size="small"
    :title="$t('in.proxyProtocol')"
    style="margin-top: 12px;"
  >
    <a-flex
      align="center"
      gap="small"
    >
      <a-switch v-model:checked="enabled" />
      <span>{{ $t('in.proxyProtocolEnable') }}</span>
    </a-flex>
    <div
      v-if="enabled"
      style="margin-top: 12px;"
    >
      <a-form-item
        :label="$t('in.proxyProtocolTrusted')"
        :extra="$t('in.proxyProtocolTrustedHint')"
      >
        <a-select
          v-model:value="trustedCidrs"
          mode="tags"
          :placeholder="'203.0.113.0/24'"
          :token-separators="[',', ' ']"
          style="width: 100%;"
        />
      </a-form-item>
      <a-form-item
        :label="$t('in.proxyProtocolTimeout')"
        :extra="$t('in.proxyProtocolTimeoutHint')"
        style="margin-bottom: 0;"
      >
        <a-input-number
          v-model:value="headerTimeout"
          :min="1"
          :max="60"
          addon-after="s"
          style="width: 100%;"
        />
      </a-form-item>
      <a-alert
        v-if="trustedCidrs.length === 0"
        type="warning"
        show-icon
        :message="$t('in.proxyProtocolEmpty')"
        style="margin-top: 12px;"
      />
    </div>
  </a-card>
</template>

<script lang="ts">
/* eslint-disable vue/no-mutating-props -- the inbound form is one object edited in place by every section, like Listen/Transport */
import { defineComponent } from 'vue'
import type { PropType } from 'vue'
import type { Inbound, ProxyProtocol } from '@/types/inbounds'

const defaults = (source?: ProxyProtocol): ProxyProtocol => ({
  enabled: source?.enabled ?? false,
  trustedCidrs: source?.trustedCidrs ?? [],
  headerTimeout: source?.headerTimeout && source.headerTimeout > 0 ? source.headerTimeout : 5,
})

export default defineComponent({
  name: 'ProxyProtocolOptions',
  props: {
    data: { type: Object as PropType<Inbound>, required: true },
  },
  computed: {
    enabled: {
      get(): boolean {
        return this.data.proxyProtocol?.enabled ?? false
      },
      set(value: boolean) {
        this.data.proxyProtocol = { ...defaults(this.data.proxyProtocol), enabled: value }
      },
    },
    trustedCidrs: {
      get(): string[] {
        return this.data.proxyProtocol?.trustedCidrs ?? []
      },
      set(value: string[]) {
        this.data.proxyProtocol = { ...defaults(this.data.proxyProtocol), trustedCidrs: value }
      },
    },
    headerTimeout: {
      get(): number {
        return this.data.proxyProtocol?.headerTimeout && this.data.proxyProtocol.headerTimeout > 0 ? this.data.proxyProtocol.headerTimeout : 5
      },
      set(value: number | null) {
        const timeout = value && value > 0 ? Math.floor(value) : 5
        this.data.proxyProtocol = { ...defaults(this.data.proxyProtocol), headerTimeout: timeout }
      },
    },
  },
})
</script>

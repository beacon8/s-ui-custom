<template>
  <a-card
    size="small"
    :title="$t('objects.multiplex')"
  >
    <a-row
      :gutter="[12, 12]"
      align="middle"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-space>
          <a-switch v-model:checked="muxEnable" />
          <span>{{ $t('mux.enable') }}</span>
        </a-space>
      </a-col>
      <template v-if="muxEnable">
        <template v-if="direction=='out'">
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('protocol')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="mux.protocol"
                allow-clear
                style="width: 100%;"
                @clear="delete mux?.protocol"
              >
                <a-select-option
                  v-for="p in ['smux', 'yamux', 'h2mux']"
                  :key="p"
                  :value="p"
                >
                  {{ p }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('mux.maxConn')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="max_connections"
                :min="0"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('mux.minStr')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="min_streams"
                :min="0"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('mux.maxStr')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="max_streams"
                :min="min_streams"
                style="width: 100%;"
              />
            </a-form-item>
          </a-col>
        </template>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="padding" />
            <span>{{ $t('mux.padding') }}</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="burtalEnable" />
            <span>{{ $t('mux.enableBrutal') }}</span>
          </a-space>
        </a-col>
      </template>
    </a-row>
    <a-row
      v-if="mux?.brutal?.enabled"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('stats.upload')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="up_mbps"
            :addon-after="$t('stats.Mbps')"
            style="width: 100%;"
          />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('stats.download')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="down_mbps"
            :min="0"
            :addon-after="$t('stats.Mbps')"
            style="width: 100%;"
          />
        </a-form-item>
      </a-col>
    </a-row>
  </a-card>
</template>

<script lang="ts">
import { PropType } from 'vue'
import { oMultiplex } from '@/types/multiplex'

// The inbound or outbound that owns the multiplex block.
interface MultiplexHolder {
  multiplex?: oMultiplex
}

export default {
  props: {
    // Inbounds, outbounds and the out_json of an inbound all bind here, so the
    // prop asks only for an object and the narrowing happens below.
    data: { type: Object as PropType<object>, required: true },
    direction: { type: String as PropType<'in' | 'out'>, required: true }
  },
  data() {
    return {}
  },
  computed: {
    // Narrow once, since no caller hands over a type the template checker could
    // follow. The computed returns the same object, so edits reach the parent.
    holder(): MultiplexHolder { return <MultiplexHolder>this.$props.data },
    mux(): oMultiplex {
      return <oMultiplex> this.holder.multiplex ?? null
    },
    muxEnable: {
      get(): boolean { return this.mux ? this.mux.enabled : false },
      set(newValue:boolean) { this.holder.multiplex = newValue ? { enabled: newValue } : undefined }
    },
    max_connections: {
      get(): number { return this.mux?.max_connections ? this.mux.max_connections : 0 },
      set(newValue:number) { this.mux.max_connections = newValue > 0 ? newValue : undefined }
    },
    min_streams: {
      get(): number { return this.mux?.min_streams ? this.mux.min_streams : 0 },
      set(newValue:number) { this.mux.min_streams = newValue > 0 ? newValue : undefined }
    },
    max_streams: {
      get(): number { return this.mux?.max_streams ? this.mux.max_streams : 0 },
      set(newValue:number) { this.mux.max_streams = newValue > 0 ? newValue : undefined }
    },
    padding: {
      get(): boolean { return this.mux?.padding ? this.mux.padding : false },
      set(newValue:boolean) { this.mux.padding = newValue ? true : undefined }
    },
    burtalEnable: {
      get(): boolean { return this.mux?.brutal ? this.mux.brutal.enabled : false },
      set(newValue:boolean) { this.mux.brutal = newValue ? { enabled: newValue, up_mbps: 100, down_mbps: 100 } : undefined }
    },
    down_mbps: {
      get() { return this.mux?.brutal && this.mux.brutal.down_mbps ? this.mux.brutal.down_mbps : 0 },
      // An emptied number field arrives as '' or null, which reads back as 0.
      set(newValue:number|string|null) {
        if (this.mux.brutal){
          this.mux.brutal.down_mbps = newValue != null && String(newValue).length != 0 ? Number(newValue) : 0
        }
      }
    },
    up_mbps: {
      get() { return this.mux?.brutal && this.mux.brutal.up_mbps ? this.mux.brutal.up_mbps : 0 },
      // An emptied number field arrives as '' or null, which reads back as 0.
      set(newValue:number|string|null) {
        if (this.mux.brutal){
          this.mux.brutal.up_mbps = newValue != null && String(newValue).length != 0 ? Number(newValue) : 0
        }
      }
    },
  }
}
</script>

<template>
  <a-card
    size="small"
    :title="$t('objects.listen')"
  >
    <a-flex
      vertical
      gap="small"
    >
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('in.addr')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.listen"
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
            :label="$t('in.port')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="data.listen_port"
              :min="1"
              :max="65535"
              style="width: 100%;"
              required
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionDetour"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('listen.detourText')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.detour"
              style="width: 100%;"
            >
              <a-select-option
                v-for="tag in inTags"
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
        v-if="optionTCP"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.tcp_fast_open" />
            <span>TCP Fast Open</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.tcp_multi_path" />
            <span>TCP Multi Path</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row
        v-if="optionUDP"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.udp_fragment" />
            <span>UDP Fragment</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            label="UDP NAT expiration"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="udpTimeout"
              :min="1"
              :addon-after="$t('date.m')"
              style="width: 100%;"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionTcpKeepAlive"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.disable_tcp_keep_alive" />
            <span>{{ $t('listen.disableTcpKeepAlive') }}</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('listen.tcpKeepAlive')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.tcp_keep_alive" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('listen.tcpKeepAliveInterval')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.tcp_keep_alive_interval" />
          </a-form-item>
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
              gap="small"
            >
              <a-space>
                <a-switch v-model:checked="optionDetour" />
                <span>{{ $t('listen.detour') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionTCP" />
                <span>{{ $t('listen.tcpOptions') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionUDP" />
                <span>{{ $t('listen.udpOptions') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionTcpKeepAlive" />
                <span>{{ $t('listen.tcpKeepAlive') }}</span>
              </a-space>
            </a-flex>
          </template>
          <a-button type="default">
            {{ $t('listen.options') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Inbound } from '@/types/inbounds'
import { Srv } from '@/types/services'

// The listen options this form edits. sing-box flattens them into whatever
// carries them, and an option that is switched off leaves no trace.
interface ListenData {
  listen: string
  listen_port: number
  detour?: string
  tcp_fast_open?: boolean
  tcp_multi_path?: boolean
  udp_fragment?: boolean
  udp_timeout?: string
  disable_tcp_keep_alive?: boolean
  tcp_keep_alive?: string
  tcp_keep_alive_interval?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound | Srv>('data', { required: true })

// The parent holds a union and picks this component with a v-if, which the
// template checker cannot follow, so narrow once here. The computed returns the
// same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as ListenData)

const { inTags } = defineProps<{
  inTags: string[]
}>()

const menu = ref(false)

const udpTimeout = computed({
  get: (): number => data.value.udp_timeout ? parseInt(data.value.udp_timeout.replace('m', '')) : 5,
  set: (newValue: number) => { data.value.udp_timeout = newValue > 0 ? newValue + 'm' : '5m' },
})

const optionTCP = computed({
  get: (): boolean => {
    return data.value.tcp_fast_open != undefined &&
           data.value.tcp_multi_path != undefined
  },
  set: (v: boolean) => {
    data.value.tcp_fast_open = v ? false : undefined
    data.value.tcp_multi_path = v ? false : undefined
  },
})

const optionUDP = computed({
  get: (): boolean => {
    return data.value.udp_fragment != undefined &&
           data.value.udp_timeout != undefined
  },
  set: (v: boolean) => {
    data.value.udp_fragment = v ? false : undefined
    data.value.udp_timeout = v ? '5m' : undefined
  },
})

const optionDetour = computed({
  get: (): boolean => data.value.detour != undefined,
  set: (v: boolean) => { data.value.detour = v ? inTags[0] ?? '' : undefined },
})

const optionTcpKeepAlive = computed({
  get: (): boolean => {
    return data.value.disable_tcp_keep_alive != undefined ||
           data.value.tcp_keep_alive != undefined ||
           data.value.tcp_keep_alive_interval != undefined
  },
  set: (v: boolean) => {
    if (v) {
      data.value.tcp_keep_alive = '5m'
      data.value.tcp_keep_alive_interval = '75s'
    } else {
      delete data.value.disable_tcp_keep_alive
      delete data.value.tcp_keep_alive
      delete data.value.tcp_keep_alive_interval
    }
  },
})
</script>

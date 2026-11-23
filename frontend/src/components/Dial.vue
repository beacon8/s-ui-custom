<template>
  <a-card
    size="small"
    :title="$t('objects.dial')"
  >
    <a-flex
      vertical
      gap="small"
    >
      <a-row
        v-if="optionDetour || optionBind"
        :gutter="[12, 12]"
      >
        <a-col
          v-if="optionDetour"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dial.detourText')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="dial.detour"
              style="width: 100%;"
            >
              <a-select-option
                v-for="tag in outTags"
                :key="tag"
                :value="tag"
              >
                {{ tag }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionBind"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dial.bindIf')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dial.bind_interface" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionIPV4 || optionIPV6 || optionBindNoPort"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          v-if="optionIPV4"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dial.bindIp4')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dial.inet4_bind_address" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionIPV6"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dial.bindIp6')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dial.inet6_bind_address" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionBindNoPort"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="dial.bind_address_no_port" />
            <span>{{ $t('dial.bindNoPort') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row
        v-if="optionRM || optionRA"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          v-if="optionRM"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            label="Linux Routing Mark"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="routingMark"
              :min="0"
              style="width: 100%;"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionRA"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="dial.reuse_addr" />
            <span>{{ $t('dial.reuseAddr') }}</span>
          </a-space>
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
            <a-switch v-model:checked="dial.tcp_fast_open" />
            <span>TCP Fast Open</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="dial.tcp_multi_path" />
            <span>TCP Multi Path</span>
          </a-space>
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
            <a-switch v-model:checked="dial.disable_tcp_keep_alive" />
            <span>{{ $t('dial.disableTcpKeepAlive') }}</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dial.tcpKeepAlive')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dial.tcp_keep_alive" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dial.tcpKeepAliveInterval')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dial.tcp_keep_alive_interval" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionUDP || optionCT"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          v-if="optionUDP"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="dial.udp_fragment" />
            <span>UDP Fragment</span>
          </a-space>
        </a-col>
        <a-col
          v-if="optionCT"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dial.connTimeout')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="connectTimeout"
              :min="1"
              :addon-after="$t('date.s')"
              style="width: 100%;"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionDR"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dial.domainResolver')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="dial.domain_resolver"
              style="width: 100%;"
            >
              <a-select-option
                v-for="tag in dnsTags"
                :key="tag"
                :value="tag"
              >
                {{ tag }}
              </a-select-option>
            </a-select>
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
              <a-space v-if="mode != 'client'">
                <a-switch v-model:checked="optionDetour" />
                <span>{{ $t('listen.detour') }}</span>
              </a-space>
              <a-space v-if="mode != 'client'">
                <a-switch v-model:checked="optionBind" />
                <span>{{ $t('dial.bindIf') }}</span>
              </a-space>
              <a-space v-if="mode != 'client'">
                <a-switch v-model:checked="optionIPV4" />
                <span>{{ $t('dial.bindIp4') }}</span>
              </a-space>
              <a-space v-if="mode != 'client'">
                <a-switch v-model:checked="optionIPV6" />
                <span>{{ $t('dial.bindIp6') }}</span>
              </a-space>
              <a-space v-if="mode != 'client'">
                <a-switch v-model:checked="optionBindNoPort" />
                <span>{{ $t('dial.bindNoPort') }}</span>
              </a-space>
              <a-space v-if="mode != 'client'">
                <a-switch v-model:checked="optionRM" />
                <span>Routing Mark</span>
              </a-space>
              <a-space v-if="mode != 'client'">
                <a-switch v-model:checked="optionRA" />
                <span>{{ $t('dial.reuseAddr') }}</span>
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
                <a-switch v-model:checked="optionCT" />
                <span>{{ $t('dial.connTimeout') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionTcpKeepAlive" />
                <span>{{ $t('dial.tcpKeepAlive') }}</span>
              </a-space>
              <a-space v-if="mode != 'client'">
                <a-switch v-model:checked="optionDR" />
                <span>{{ $t('dial.domainResolver') }}</span>
              </a-space>
            </a-flex>
          </template>
          <a-button type="default">
            {{ $t('dial.options') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Data from '@/store/modules/data'

// The dial options a sing-box object can carry. Only the fields this form
// reads or writes are listed; the object the parent owns holds plenty more.
interface DialOptions {
  detour?: string
  bind_interface?: string
  inet4_bind_address?: string
  inet6_bind_address?: string
  bind_address_no_port?: boolean
  routing_mark?: number
  reuse_addr?: boolean
  connect_timeout?: string
  tcp_fast_open?: boolean
  tcp_multi_path?: boolean
  disable_tcp_keep_alive?: boolean
  tcp_keep_alive?: string
  tcp_keep_alive_interval?: string
  udp_fragment?: boolean
  domain_resolver?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop. Inbounds, outbounds, endpoints, DNS servers,
// NTP and the HTTP clients each bind their own shape here, so the model asks
// only for an object and the narrowing happens below.
const model = defineModel<object>('dial', { required: true })

// Narrow once, since no caller hands over a type the template checker could
// follow. The computed returns the same object, so edits still reach the parent.
const dial = computed(() => model.value as DialOptions)

defineProps<{
  mode?: string
}>()

const menu = ref(false)

const outTags = computed((): string[] => [...(Data().outbounds?.map((o) => o.tag) ?? []), ...(Data().endpoints?.map((e) => e.tag) ?? [])])

const connectTimeout = computed({
  get: (): number => dial.value.connect_timeout ? parseInt(dial.value.connect_timeout.replace('s','')) : 5,
  set: (newValue:number) => { dial.value.connect_timeout = newValue > 0 ? newValue + 's' : '5s' },
})

const routingMark = computed({
  get: (): number => dial.value.routing_mark?? 0,
  set: (newValue:number) => { dial.value.routing_mark = newValue > 0 ? newValue : 0 },
})

const optionDetour = computed({
  get: (): boolean => dial.value.detour != undefined,
  set: (v:boolean) => {
    if (v) dial.value.detour = outTags.value[0]?? ''
    else delete dial.value.detour
  },
})

const optionBind = computed({
  get: (): boolean => dial.value.bind_interface != undefined,
  set: (v:boolean) => {
    if (v) dial.value.bind_interface = ''
    else delete dial.value.bind_interface
  },
})

const optionIPV4 = computed({
  get: (): boolean => dial.value.inet4_bind_address != undefined,
  set: (v:boolean) => {
    if (v) dial.value.inet4_bind_address = ''
    else delete dial.value.inet4_bind_address
  },
})

const optionIPV6 = computed({
  get: (): boolean => dial.value.inet6_bind_address != undefined,
  set: (v:boolean) => {
    if (v) dial.value.inet6_bind_address = ''
    else delete dial.value.inet6_bind_address
  },
})

const optionBindNoPort = computed({
  get: (): boolean => dial.value.bind_address_no_port != undefined,
  set: (v:boolean) => {
    if (v) dial.value.bind_address_no_port = true
    else delete dial.value.bind_address_no_port
  },
})

const optionTcpKeepAlive = computed({
  get: (): boolean => {
    return dial.value.disable_tcp_keep_alive != undefined ||
           dial.value.tcp_keep_alive != undefined ||
           dial.value.tcp_keep_alive_interval != undefined
  },
  set: (v:boolean) => {
    if (v) {
      dial.value.tcp_keep_alive = '5m'
      dial.value.tcp_keep_alive_interval = '75s'
    } else {
      delete dial.value.disable_tcp_keep_alive
      delete dial.value.tcp_keep_alive
      delete dial.value.tcp_keep_alive_interval
    }
  },
})

const optionRM = computed({
  get: (): boolean => dial.value.routing_mark != undefined,
  set: (v:boolean) => {
    if (v) dial.value.routing_mark = 0
    else delete dial.value.routing_mark
  },
})

const optionRA = computed({
  get: (): boolean => dial.value.reuse_addr != undefined,
  set: (v:boolean) => {
    if (v) dial.value.reuse_addr = true
    else delete dial.value.reuse_addr
  },
})

const optionTCP = computed({
  get: (): boolean => {
    return dial.value.tcp_fast_open != undefined &&
           dial.value.tcp_multi_path != undefined
  },
  set: (v:boolean) => {
    if (v) {
      dial.value.tcp_fast_open = false
      dial.value.tcp_multi_path = false
    } else {
      delete dial.value.tcp_fast_open
      delete dial.value.tcp_multi_path
    }
  },
})

const optionUDP = computed({
  get: (): boolean => dial.value.udp_fragment != undefined,
  set: (v:boolean) => {
    if (v) dial.value.udp_fragment = true
    else delete dial.value.udp_fragment
  },
})

const optionCT = computed({
  get: (): boolean => dial.value.connect_timeout != undefined,
  set: (v:boolean) => {
    if (v) dial.value.connect_timeout = '5s'
    else delete dial.value.connect_timeout
  },
})

const optionDR = computed({
  get: (): boolean => dial.value.domain_resolver != undefined,
  set: (v:boolean) => { dial.value.domain_resolver = v ? dnsTags.value[0]?? '' : undefined },
})

const dnsTags = computed((): string[] => Data().config.dns?.servers?.map((d) => d.tag) ?? [])
</script>

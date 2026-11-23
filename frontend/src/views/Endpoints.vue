<template>
  <EndpointVue 
    :id="modal.id"
    v-model="modal.visible"
    :visible="modal.visible"
    :data="modal.data"
    :tags="endpointTags"
    @close="closeModal"
  />
  <Stats
    v-model="stats.visible"
    :visible="stats.visible"
    :resource="stats.resource"
    :tag="stats.tag"
    @close="closeStats"
  />
  <Sessions
    v-model="sessions.visible"
    :visible="sessions.visible"
    resource="endpoint"
    :tag="sessions.tag"
    @close="closeSessions"
  />
  <QrCode
    v-model="qrcode.visible"
    :visible="qrcode.visible"
    :data="qrcode.data"
    @close="closeQrCode"
  />
  <a-row
    :gutter="[12, 12]"
    justify="center"
    align="middle"
    style="margin-bottom: 12px;"
  >
    <a-col>
      <a-button
        type="primary"
        @click="showModal(0)"
      >
        {{ $t('actions.add') }}
      </a-button>
    </a-col>
    <a-col>
      <a-button
        type="default"
        :loading="testingAll"
        :disabled="testingAll || endpoints.length === 0"
        @click="checkAllEndpoints"
      >
        <a-space :size="4">
          <span>{{ $t('actions.testAll') || 'Test all' }}</span>
          <AppIcon name="speedometer" />
        </a-space>
      </a-button>
    </a-col>
    <UrlTestTarget />
  </a-row>
  <a-row :gutter="[12, 12]">
    <a-col
      v-for="(item, index) in <any[]>endpoints"
      :key="item.tag"
      :xs="24"
      :sm="8"
      :md="6"
      :lg="4"
    >
      <a-card
        size="small"
        :title="item.tag"
      >
        <div style="font-size: 12px; opacity: 0.65; margin-bottom: 8px;">
          {{ item.type }}
        </div>
        <a-row :gutter="[8, 6]">
          <a-col :span="12">
            {{ $t('in.addr') }}
          </a-col>
          <a-col :span="12">
            {{ item.address?.length > 0 ? item.address[0] : '-' }}
          </a-col>
          <a-col :span="12">
            {{ $t('in.port') }}
          </a-col>
          <a-col :span="12">
            {{ item.listen_port > 0 ? item.listen_port : '-' }}
          </a-col>
          <a-col :span="12">
            {{ $t('types.wg.peers') }}
          </a-col>
          <a-col :span="12">
            {{ item.peers?.length ?? '-' }}
          </a-col>
          <a-col :span="12">
            {{ $t('online') }}
          </a-col>
          <a-col :span="12">
            <a-tooltip
              v-if="onlines.includes(item.tag)"
              :title="$t('sessions.title')"
            >
              <a-tag
                color="green"
                style="cursor: pointer;"
                @click="showSessions(item.tag)"
              >
                {{ $t('online') }}
              </a-tag>
            </a-tooltip>
            <template v-else>
              -
            </template>
          </a-col>
          <a-col :span="12">
            {{ $t('out.delay') }}
          </a-col>
          <a-col :span="12">
            <a-space :size="4">
              <a-spin
                v-if="checkResults[item.tag]?.loading"
                size="small"
              />
              <a-tooltip
                v-else
                :title="$t('actions.test')"
              >
                <a-button
                  type="text"
                  shape="circle"
                  size="small"
                  :aria-label="$t('actions.test')"
                  @click="checkEndpoint(item.tag)"
                >
                  <template #icon>
                    <AppIcon name="speedometer" />
                  </template>
                </a-button>
              </a-tooltip>
              <template v-if="checkResults[item.tag]?.loading == false && checkResults[item.tag]">
                <a-tag
                  v-if="checkResults[item.tag].success"
                  color="green"
                >
                  {{ checkResults[item.tag].data?.Delay + $t('date.ms') }}
                </a-tag>
                <a-tooltip
                  v-else
                  :title="checkResults[item.tag].errorMessage || $t('failed')"
                >
                  <AppIcon
                    name="close-circle"
                    color="error"
                    role="img"
                    :aria-label="checkResults[item.tag].errorMessage || $t('failed')"
                  />
                </a-tooltip>
              </template>
            </a-space>
          </a-col>
        </a-row>
        <a-divider style="margin: 10px 0 6px;" />
        <a-space :size="4">
          <a-tooltip :title="$t('actions.edit')">
            <a-button
              type="text"
              shape="circle"
              :aria-label="$t('actions.edit')"
              @click="showModal(item.id)"
            >
              <template #icon>
                <AppIcon name="file-edit" />
              </template>
            </a-button>
          </a-tooltip>
          <a-popconfirm
            v-model:open="delOverlay[index]"
            :title="$t('confirm')"
            :ok-text="$t('yes')"
            :cancel-text="$t('no')"
            ok-type="danger"
            @confirm="delEndpoint(item.tag)"
          >
            <a-tooltip :title="$t('actions.del')">
              <a-button
                type="text"
                shape="circle"
                danger
                :aria-label="$t('actions.del')"
              >
                <template #icon>
                  <AppIcon name="file-remove" />
                </template>
              </a-button>
            </a-tooltip>
          </a-popconfirm>
          <a-tooltip
            v-if="item.type == 'wireguard' && item.peers?.length > 0"
            :title="$t('a11y.qrCode')"
          >
            <a-button
              type="text"
              shape="circle"
              :aria-label="$t('a11y.qrCode')"
              @click="showQrCode(item.id)"
            >
              <template #icon>
                <AppIcon name="qrcode" />
              </template>
            </a-button>
          </a-tooltip>
          <a-tooltip
            v-if="Data().enableTraffic"
            :title="$t('stats.graphTitle')"
          >
            <a-button
              type="text"
              shape="circle"
              :aria-label="$t('stats.graphTitle')"
              @click="showStats(item.tag)"
            >
              <template #icon>
                <AppIcon name="chart-line" />
              </template>
            </a-button>
          </a-tooltip>
        </a-space>
      </a-card>
    </a-col>
  </a-row>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import HttpUtils from '@/plugins/httputil'
import EndpointVue from '@/layouts/modals/Endpoint.vue'
import Stats from '@/layouts/modals/Stats.vue'
import Sessions from '@/layouts/modals/Sessions.vue'
import QrCode from '@/layouts/modals/WgQrCode.vue'
import { Endpoint } from '@/types/endpoints'
import { computed, ref } from 'vue'
import UrlTestTarget from '@/components/UrlTestTarget.vue'
import { useUrlTestTarget } from '@/composables/useUrlTestTarget'

// What api/checkOutbound answers with.
interface CheckReply {
  OK: boolean
  Delay?: number
  Error?: string
}

interface CheckResult {
  loading?: boolean
  success: boolean
  data?: CheckReply | null
  errorMessage?: string
}

const checkResults = ref<Record<string, CheckResult>>({})
const { link: testLink } = useUrlTestTarget()

// Endpoints (e.g. generated WARP nodes) register as outbounds in the core, so
// the same api/checkOutbound latency test used on the Outbounds page works here.
const checkEndpoint = async (tag: string) => {
  checkResults.value = { ...checkResults.value, [tag]: { loading: true, success: false } }
  const msg = await HttpUtils.get<CheckReply>('api/checkOutbound', { tag, link: testLink.value })
  const success = msg.success && msg.obj?.OK
  const errorMessage = success ? undefined : (msg.obj?.Error ?? msg.msg ?? '')
  checkResults.value = {
    ...checkResults.value,
    [tag]: { loading: false, success, data: msg.obj ?? null, errorMessage }
  }
}

const testingAll = ref(false)

const checkAllEndpoints = async () => {
  const list = endpoints.value
  if (list.length === 0) return
  testingAll.value = true
  try {
    await Promise.all(list.map((e) => checkEndpoint(e.tag)))
  } finally {
    testingAll.value = false
  }
}

const endpoints = computed((): Endpoint[] => {
  return <Endpoint[]> Data().endpoints
})

const endpointTags = computed((): string[] => {
  return endpoints.value?.map((o:Endpoint) => o.tag)
})


const onlines = computed(() => {
  return [...Data().onlines.inbound?? [], ...Data().onlines.outbound??[] ]
})

const modal = ref({
  visible: false,
  id: 0,
  data: "",
})

let delOverlay = ref(new Array<boolean>)

const showModal = (id: number) => {
  modal.value.id = id
  modal.value.data = id == 0 ? '' : JSON.stringify(endpoints.value.findLast(o => o.id == id))
  modal.value.visible = true
}

const closeModal = () => {
  modal.value.visible = false
}

const stats = ref({
  visible: false,
  resource: "endpoint",
  tag: "",
})

const delEndpoint = async (tag: string) => {
  const index = endpoints.value.findIndex(i => i.tag == tag)
  const success = await Data().save("endpoints", "del", tag)
  if (success) delOverlay.value[index] = false
}

const showStats = (tag: string) => {
  stats.value.tag = tag
  stats.value.visible = true
}
const closeStats = () => {
  stats.value.visible = false
}

const sessions = ref({
  visible: false,
  tag: "",
})

const showSessions = (tag: string) => {
  sessions.value.tag = tag
  sessions.value.visible = true
}
const closeSessions = () => {
  sessions.value.visible = false
}

const qrcode = ref({
  visible: false,
  data: <Endpoint>{},
})

const showQrCode = (id: number) => {
  qrcode.value.data = <Endpoint>endpoints.value.findLast(o => o.id == id)
  qrcode.value.visible = true
}
const closeQrCode = () => {
  qrcode.value.visible = false
}
</script>
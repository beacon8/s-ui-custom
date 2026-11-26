<template>
  <InboundVue 
    :id="modal.id"
    v-model="modal.visible"
    :visible="modal.visible"
    :in-tags="inTags"
    :tls-configs="tlsConfigs"
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
    resource="inbound"
    :tag="sessions.tag"
    @close="closeSessions"
  />
  <a-row
    :gutter="[12, 12]"
    justify="start"
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
  </a-row>
  <a-table
    :data-source="inbounds"
    :columns="columns"
    :pagination="false"
    row-key="id"
    size="middle"
    :row-class-name="rowClassName"
    :scroll="{ x: 'max-content' }"
  >
    <template #bodyCell="{ column, record }">
      <template v-if="column.key === 'enable'">
        <a-switch
          :checked="record.enable !== false"
          :loading="toggling[record.id]"
          :disabled="toggling[record.id]"
          :aria-label="$t('enable')"
          size="small"
          @change="(val: any) => toggleEnable(record, !!val)"
        />
      </template>
      <template v-else-if="column.key === 'tag'">
        <a-space :size="6">
          <span class="inbound-tag">{{ record.tag }}</span>
          <a-tag
            v-if="record.enable === false"
            color="default"
          >
            {{ $t('disable') }}
          </a-tag>
        </a-space>
      </template>
      <template v-else-if="column.key === 'tls'">
        <a-tag :color="record.tls_id > 0 ? 'green' : 'default'">
          {{ record.tls_id > 0 ? $t('enable') : $t('disable') }}
        </a-tag>
      </template>
      <template v-else-if="column.key === 'proxy'">
        <a-tooltip
          v-if="record.proxyProtocol?.enabled"
          :title="proxyProtocolHint(record)"
        >
          <a-tag color="blue">
            {{ $t('enable') }}
          </a-tag>
        </a-tooltip>
        <a-tag
          v-else
          color="default"
        >
          {{ $t('disable') }}
        </a-tag>
      </template>
      <template v-else-if="column.key === 'users'">
        <a-tooltip
          v-if="record.users && record.users.length > 0"
          placement="bottom"
        >
          <template #title>
            <div dir="ltr">
              <div
                v-for="u in record.users"
                :key="u"
              >
                {{ u }}
              </div>
            </div>
          </template>
          <span>{{ record.users.length }}</span>
        </a-tooltip>
        <span v-else>{{ record.users?.length ?? '-' }}</span>
      </template>
      <template v-else-if="column.key === 'online'">
        <a-tooltip
          v-if="onlines.includes(record.tag)"
          :title="$t('sessions.title')"
        >
          <a-tag
            color="green"
            style="cursor: pointer;"
            @click="showSessions(record.tag)"
          >
            {{ $t('online') }}
          </a-tag>
        </a-tooltip>
        <span v-else>-</span>
      </template>
      <template v-else-if="column.key === 'actions'">
        <a-space :size="2">
          <a-tooltip :title="$t('actions.edit')">
            <a-button
              type="text"
              shape="circle"
              :aria-label="$t('actions.edit')"
              @click="showModal(record.id)"
            >
              <template #icon>
                <AppIcon name="file-edit" />
              </template>
            </a-button>
          </a-tooltip>
          <a-popconfirm
            :title="$t('confirm')"
            :ok-text="$t('yes')"
            :cancel-text="$t('no')"
            ok-type="danger"
            @confirm="delInbound(record.id)"
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
          <a-tooltip :title="$t('actions.clone')">
            <a-button
              type="text"
              shape="circle"
              :aria-label="$t('actions.clone')"
              :loading="cloneLoading"
              @click="clone(record.id)"
            >
              <template #icon>
                <AppIcon name="content-duplicate" />
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
              @click="showStats(record.tag)"
            >
              <template #icon>
                <AppIcon name="chart-line" />
              </template>
            </a-button>
          </a-tooltip>
        </a-space>
      </template>
    </template>
  </a-table>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import InboundVue from '@/layouts/modals/Inbound.vue'
import Stats from '@/layouts/modals/Stats.vue'
import Sessions from '@/layouts/modals/Sessions.vue'
import { computed, ref } from 'vue'
import { i18n } from '@/locales'
import { createInbound, Inbound } from '@/types/inbounds'
import { tls } from '@/types/tls'
import RandomUtil from '@/plugins/randomUtil'

const columns = computed(() => [
  { title: i18n.global.t('enable'), key: 'enable', width: 70 },
  { title: i18n.global.t('objects.tag'), dataIndex: 'tag', key: 'tag', width: 180 },
  { title: i18n.global.t('type'), dataIndex: 'type', key: 'type', width: 130 },
  { title: i18n.global.t('in.addr'), dataIndex: 'listen', key: 'listen', width: 150 },
  { title: i18n.global.t('in.port'), dataIndex: 'listen_port', key: 'port', width: 90 },
  { title: i18n.global.t('objects.tls'), key: 'tls', width: 90 },
  { title: i18n.global.t('in.proxyProtocol'), key: 'proxy', width: 130 },
  { title: i18n.global.t('pages.clients'), key: 'users', width: 90 },
  { title: i18n.global.t('online'), key: 'online', width: 90 },
  { title: i18n.global.t('actions.operations'), key: 'actions', width: 170 },
])

const inbounds = computed((): Inbound[] => {
  return <Inbound[]> Data().inbounds
})

const tlsConfigs = computed((): tls[] => {
  return Data().tlsConfigs
})

const inTags = computed((): string[] => {
  return [...(inbounds.value?.map(i => i.tag) ?? []), ...(Data().endpoints?.filter(e => e.listen_port > 0).map(e => e.tag) ?? [])]
})

const onlines = computed(() => {
  return Data().onlines.inbound?? []
})

const modal = ref({
  visible: false,
  id: 0,
})

const showModal = (id: number) => {
  modal.value.id = id
  modal.value.visible = true
}
const closeModal = () => {
  modal.value.visible = false
}

const rowClassName = (record: Inbound) => record.enable === false ? 'inbound-disabled' : ''

const proxyProtocolHint = (item: Inbound) => {
  const trusted = item.proxyProtocol?.trustedCidrs ?? []
  return trusted.length > 0 ? trusted.join(', ') : i18n.global.t('in.proxyProtocolEmpty')
}

const delInbound = async (id: number) => {
  const tag = inbounds.value.find(i => i.id == id)?.tag
  if (!tag) return

  await Data().save("inbounds", "del", tag)
}

const toggling = ref<Record<number, boolean>>({})

const toggleEnable = async (item: Inbound, val: boolean) => {
  if ((item.enable !== false) === val) return
  toggling.value = { ...toggling.value, [item.id]: true }
  const full = (await Data().loadInbounds([item.id]))[0]
  if (full?.id) {
    full.enable = val
    await Data().save("inbounds", "edit", full)
  }
  toggling.value = { ...toggling.value, [item.id]: false }
}

let cloneLoading = ref(false)

const clone = async (id: number) => {
  cloneLoading.value = true
  const inboundArray = await Data().loadInbounds([id])
  const inbound = inboundArray[0]
  let newTag = inbound.type + "-" + RandomUtil.randomSeq(3)
  const newInbound = createInbound(inbound.type, { ...inbound,
    id: 0,
    tag: newTag,
    listen_port: RandomUtil.randomIntRange(10000, 60000),
  })
  await Data().save("inbounds", "new", newInbound)
  cloneLoading.value = false
}

const stats = ref({
  visible: false,
  resource: "inbound",
  tag: "",
})

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
</script>
<style scoped>
.inbound-tag {
  font-weight: 500;
}

:deep(.inbound-disabled) {
  opacity: 0.55;
}
</style>

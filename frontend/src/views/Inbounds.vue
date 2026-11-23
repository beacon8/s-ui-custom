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
    justify="center"
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
  <a-row :gutter="[12, 12]">
    <a-col
      v-for="(item, index) in <any[]>inbounds"
      :key="item.tag"
      :xs="24"
      :sm="8"
      :md="6"
      :lg="4"
    >
      <a-card
        size="small"
        :title="item.tag"
        :class="{ 'inbound-disabled': item.enable === false }"
      >
        <template #extra>
          <a-switch
            :checked="item.enable !== false"
            :loading="toggling[item.id]"
            :disabled="toggling[item.id]"
            :aria-label="$t('enable')"
            size="small"
            @change="(val: any) => toggleEnable(item, !!val)"
          />
        </template>
        <div style="font-size: 12px; opacity: 0.65; margin-bottom: 8px;">
          {{ item.type }}
        </div>
        <a-row :gutter="[8, 6]">
          <a-col :span="12">
            {{ $t('in.addr') }}
          </a-col>
          <a-col :span="12">
            {{ item.listen }}
          </a-col>
          <a-col :span="12">
            {{ $t('in.port') }}
          </a-col>
          <a-col :span="12">
            {{ item.listen_port }}
          </a-col>
          <a-col :span="12">
            {{ $t('objects.tls') }}
          </a-col>
          <a-col :span="12">
            {{ item.tls_id > 0 ? $t('enable') : $t('disable') }}
          </a-col>
          <a-col :span="12">
            {{ $t('pages.clients') }}
          </a-col>
          <a-col :span="12">
            <template v-if="item.users">
              <a-tooltip
                v-if="item.users.length > 0"
                placement="bottom"
              >
                <template #title>
                  <div dir="ltr">
                    <div
                      v-for="u in item.users"
                      :key="u"
                    >
                      {{ u }}
                    </div>
                  </div>
                </template>
                <span>{{ item.users.length }}</span>
              </a-tooltip>
              <span v-else>{{ item.users.length }}</span>
            </template>
            <template v-else>
              -
            </template>
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
            @confirm="delInbound(item.id)"
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
              @click="clone(item.id)"
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
import InboundVue from '@/layouts/modals/Inbound.vue'
import Stats from '@/layouts/modals/Stats.vue'
import Sessions from '@/layouts/modals/Sessions.vue'
import { computed, ref } from 'vue'
import { createInbound, Inbound } from '@/types/inbounds'
import { tls } from '@/types/tls'
import RandomUtil from '@/plugins/randomUtil'

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

let delOverlay = ref(new Array<boolean>)

const showModal = (id: number) => {
  modal.value.id = id
  modal.value.visible = true
}
const closeModal = () => {
  modal.value.visible = false
}

const delInbound = async (id: number) => {
  const index = inbounds.value.findIndex(i => i.id == id)
  const tag = inbounds.value[index].tag

  const success = await Data().save("inbounds", "del", tag)
  if (success) delOverlay.value[index] = false
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
.inbound-disabled {
  opacity: 0.6;
}
</style>

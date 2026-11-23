<template>
  <ServiceVue
    :id="modal.id"
    v-model="modal.visible"
    :visible="modal.visible"
    :data="modal.data"
    :in-tags="inTags"
    :ts-tags="tsTags"
    :ss-tags="ssTags"
    :tls-configs="tlsConfigs"
    @close="closeModal"
  />
  <a-row
    :gutter="[12, 12]"
    justify="center"
    style="margin-bottom: 12px;"
  >
    <a-col flex="none">
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
      v-for="item in <any[]>services"
      :key="item.tag"
      :xs="24"
      :sm="8"
      :md="6"
      :lg="4"
    >
      <a-card
        size="small"
        hoverable
        :title="item.tag"
      >
        <div style="margin-bottom: 8px; opacity: 0.65; font-size: 12px;">
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
        </a-row>
        <a-divider style="margin: 8px 0;" />
        <a-space size="small">
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
            :title="$t('confirm')"
            :ok-text="$t('yes')"
            :cancel-text="$t('no')"
            ok-type="danger"
            @confirm="delSrv(item.id)"
          >
            <a-tooltip :title="$t('actions.del')">
              <a-button
                type="text"
                danger
                shape="circle"
                :aria-label="$t('actions.del')"
              >
                <template #icon>
                  <AppIcon name="file-remove" />
                </template>
              </a-button>
            </a-tooltip>
          </a-popconfirm>
        </a-space>
      </a-card>
    </a-col>
  </a-row>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import { Srv } from '@/types/services'
import { tls } from '@/types/tls'
import { computed, ref } from 'vue'
import ServiceVue from '@/layouts/modals/Service.vue'

const services = computed((): Srv[] => {
  return <Srv[]> Data().services
})

const tsTags = computed((): string[] => {
  return Data().endpoints?.filter(o => o.type == "tailscale")?.map(o => o.tag)
})

// Only the inbound types that declare users are in the union, but any inbound
// can carry one, so the field is read through a shape that only says that.
const ssTags = computed((): string[] => {
  return Data().inbounds?.filter(o => o.type == "shadowsocks" && !(o as { users?: unknown }).users)?.map(o => o.tag)
})

const inTags = computed((): string[] => {
  return [...(Data().inbounds?.map(o => o.tag).filter(t => t != null) ?? []), ...(Data().endpoints?.filter(e => e.listen_port > 0).map(e => e.tag) ?? [])]
})

const tlsConfigs = computed((): tls[] => {
  return Data().tlsConfigs
})

const modal = ref({
  visible: false,
  id: 0,
  data: "",
})

const showModal = (id: number) => {
  modal.value.id = id
  modal.value.data = id == 0 ? '' : JSON.stringify(services.value.findLast(o => o.id == id))
  modal.value.visible = true
}

const closeModal = () => {
  modal.value.visible = false
}

const delSrv = async (id: number) => {
  const index = services.value.findIndex(i => i.id == id)
  const tag = services.value[index].tag

  await Data().save("services", "del", tag)
}
</script>

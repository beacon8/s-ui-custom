<template>
  <TlsVue
    :id="modal.id"
    v-model="modal.visible"
    :visible="modal.visible"
    :data="modal.data"
    :providers="providerTags"
    @close="closeModal"
    @save="saveModal"
  />
  <CertProviderVue
    v-model="providerModal.visible"
    :visible="providerModal.visible"
    :index="providerModal.index"
    :data="providerModal.data"
    :tags="providerTags"
    @close="closeProviderModal"
    @save="saveProviderModal"
  />

  <div class="tls-page">
    <section class="tls-section">
      <div class="tls-section-head">
        <h2 class="tls-section-title">
          {{ $t('tls.configs') }}
        </h2>
        <a-button
          type="primary"
          class="tls-section-action"
          @click="showModal(0)"
        >
          {{ $t('tls.addConfig') }}
        </a-button>
      </div>

      <a-row
        v-if="tlsConfigs.length > 0"
        :gutter="[16, 16]"
      >
        <a-col
          v-for="item in <any[]>tlsConfigs"
          :key="item.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="6"
          :xl="4"
        >
          <a-card
            size="small"
            hoverable
            :title="item.name"
          >
            <div style="margin-bottom: 8px; opacity: 0.65; font-size: 12px;">
              {{ item.server?.server_name?.length > 0 ? item.server.server_name : "-" }}
            </div>
            <a-row :gutter="[8, 6]">
              <a-col :span="12">
                {{ $t('pages.inbounds') }}
              </a-col>
              <a-col :span="12">
                <template v-if="tlsInbounds(item.id).length > 0">
                  <a-tooltip placement="bottom">
                    <template #title>
                      <div dir="ltr">
                        <span
                          v-for="i in tlsInbounds(item.id)"
                          :key="i"
                        >{{ i }}<br></span>
                      </div>
                    </template>
                    <span>{{ tlsInbounds(item.id).length }}</span>
                  </a-tooltip>
                </template>
                <template v-else>
                  -
                </template>
              </a-col>
              <a-col :span="12">
                {{ $t('tls.provider.title') }}
              </a-col>
              <a-col :span="12">
                {{ item.server?.certificate_provider ?? '-' }}
              </a-col>
              <a-col :span="12">
                ECH
              </a-col>
              <a-col :span="12">
                {{ $t(item.server?.ech == undefined ? 'no' : 'yes') }}
              </a-col>
              <a-col :span="12">
                Reality
              </a-col>
              <a-col :span="12">
                {{ $t(item.server?.reality == undefined ? 'no' : 'yes') }}
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
                v-if="tlsInbounds(item.id).length == 0"
                :title="$t('confirm')"
                :ok-text="$t('yes')"
                :cancel-text="$t('no')"
                ok-type="danger"
                @confirm="delTls(item.id)"
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
              <a-tooltip :title="$t('actions.clone')">
                <a-button
                  type="text"
                  shape="circle"
                  :aria-label="$t('actions.clone')"
                  @click="clone(item)"
                >
                  <template #icon>
                    <AppIcon name="content-duplicate" />
                  </template>
                </a-button>
              </a-tooltip>
            </a-space>
          </a-card>
        </a-col>
      </a-row>
      <p
        v-else
        class="tls-empty"
      >
        {{ $t('tls.emptyConfigs') }}
      </p>
    </section>

    <a-divider class="tls-divider" />

    <section class="tls-section">
      <div class="tls-section-head">
        <h2 class="tls-section-title">
          {{ $t('tls.provider.title') }}
        </h2>
        <a-button
          type="default"
          class="tls-section-action"
          :loading="providerLoading"
          @click="showProviderModal(-1)"
        >
          {{ $t('tls.addProvider') }}
        </a-button>
      </div>

      <a-row
        v-if="providers.length > 0"
        :gutter="[16, 16]"
      >
        <a-col
          v-for="(item, index) in <any[]>providers"
          :key="item.tag"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="6"
          :xl="4"
        >
          <a-card
            size="small"
            hoverable
            :title="item.tag"
          >
            <div style="margin-bottom: 8px; opacity: 0.65; font-size: 12px;">
              {{ providerTypeName(item.type) }}
            </div>
            <a-row :gutter="[8, 6]">
              <a-col :span="12">
                {{ $t('rule.domain') }}
              </a-col>
              <a-col :span="12">
                <template v-if="item.domain?.length > 0">
                  <a-tooltip placement="bottom">
                    <template #title>
                      <div dir="ltr">
                        <span
                          v-for="d in item.domain"
                          :key="d"
                        >{{ d }}<br></span>
                      </div>
                    </template>
                    <span>{{ item.domain.length }}</span>
                  </a-tooltip>
                </template>
                <template v-else>
                  -
                </template>
              </a-col>
              <a-col :span="12">
                {{ $t('objects.tls') }}
              </a-col>
              <a-col :span="12">
                <template v-if="providerUsers(item.tag).length > 0">
                  <a-tooltip placement="bottom">
                    <template #title>
                      <div dir="ltr">
                        <span
                          v-for="t in providerUsers(item.tag)"
                          :key="t"
                        >{{ t }}<br></span>
                      </div>
                    </template>
                    <span>{{ providerUsers(item.tag).length }}</span>
                  </a-tooltip>
                </template>
                <template v-else>
                  -
                </template>
              </a-col>
            </a-row>
            <a-divider style="margin: 8px 0;" />
            <a-space size="small">
              <a-tooltip :title="$t('actions.edit')">
                <a-button
                  type="text"
                  shape="circle"
                  :aria-label="$t('actions.edit')"
                  @click="showProviderModal(index)"
                >
                  <template #icon>
                    <AppIcon name="file-edit" />
                  </template>
                </a-button>
              </a-tooltip>
              <!-- Deleting a provider a TLS config still references would leave a
                   dangling tag, which stops the core from starting. -->
              <a-popconfirm
                v-if="providerUsers(item.tag).length == 0"
                :title="$t('confirm')"
                :ok-text="$t('yes')"
                :cancel-text="$t('no')"
                ok-type="danger"
                @confirm="delProvider(index)"
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
              <a-tooltip :title="$t('actions.clone')">
                <a-button
                  type="text"
                  shape="circle"
                  :aria-label="$t('actions.clone')"
                  @click="cloneProvider(index)"
                >
                  <template #icon>
                    <AppIcon name="content-duplicate" />
                  </template>
                </a-button>
              </a-tooltip>
            </a-space>
          </a-card>
        </a-col>
      </a-row>
      <p
        v-else
        class="tls-empty"
      >
        {{ $t('tls.emptyProviders') }}
      </p>
    </section>
  </div>
</template>

<script lang="ts" setup>
import TlsVue from '@/layouts/modals/Tls.vue'
import CertProviderVue from '@/layouts/modals/CertProvider.vue'
import Data from '@/store/modules/data'
import { computed, ref } from 'vue'
import { Inbound } from '@/types/inbounds'
import { tls, certProvider } from '@/types/tls'
import { Config } from '@/types/config'

// Certificate providers are not part of the generated Config type yet, so the
// config is read through a view of it that adds them.
type ConfigWithProviders = Config & { certificate_providers?: certProvider[] }

const tlsConfigs = computed((): tls[] => {
  return Data().tlsConfigs
})

const inbounds = computed((): Inbound[] => {
  return Data().inbounds
})

const tlsInbounds = (id: number): string[] => {
  return inbounds.value.filter(i => i.tls_id == id).map(i => i.tag)
}

const modal = ref({
  visible: false,
  id: 0,
  data: "",
})

const showModal = (id: number) => {
  modal.value.id = id
  modal.value.data = id == 0 ? '{}' : JSON.stringify(tlsConfigs.value.findLast(t => t.id == id))
  modal.value.visible = true
}
const clone = (obj: tls) => {
  const data = <tls>JSON.parse(JSON.stringify(obj))
  data.id = 0
  while (tlsConfigs.value.findIndex(t => t.name == data.name) != -1) {
    data.name += "-copy"
  }
  saveModal(data)
}
const closeModal = () => {
  modal.value.visible = false
}
const saveModal = async (data: tls) => {
  const success = await Data().save("tls", data.id > 0 ? "edit" : "new", data)
  if (success) modal.value.visible = false
}

const delTls = async (id: number) => {
  await Data().save("tls", "del", id)
}

// Certificate providers live in the base config, so each change is written
// back through the config object as a whole, the same way the rules page does.
const appConfig = computed((): ConfigWithProviders => Data().config)

const providers = computed((): certProvider[] => {
  const config = appConfig.value
  if (!Array.isArray(config.certificate_providers)) config.certificate_providers = []
  return config.certificate_providers
})

const providerTags = computed((): string[] => providers.value.map(p => p.tag))

const providerUsers = (tag: string): string[] =>
  tlsConfigs.value.filter(t => t.server?.certificate_provider == tag).map(t => t.name)

const providerTypeNames: Record<string, string> = {
  'acme': 'ACME',
  'tailscale': 'Tailscale',
  'cloudflare-origin-ca': 'Cloudflare Origin CA',
}
const providerTypeName = (type: string): string => providerTypeNames[type] ?? type

const providerLoading = ref(false)
const providerModal = ref({ visible: false, index: -1, data: "" })

const showProviderModal = (index: number) => {
  providerModal.value.index = index
  providerModal.value.data = index == -1 ? '' : JSON.stringify(providers.value[index])
  providerModal.value.visible = true
}
const closeProviderModal = () => { providerModal.value.visible = false }

const saveProviderModal = async (data: certProvider) => {
  const index = providerModal.value.index
  if (index == -1) providers.value.push(data)
  else providers.value[index] = data
  const success = await saveProviders()
  if (success) providerModal.value.visible = false
}

const cloneProvider = async (index: number) => {
  const copy = <certProvider>JSON.parse(JSON.stringify(providers.value[index]))
  while (providerTags.value.includes(copy.tag)) copy.tag += "-copy"
  providers.value.push(copy)
  await saveProviders()
}

const delProvider = async (index: number) => {
  providers.value.splice(index, 1)
  await saveProviders()
}

const saveProviders = async (): Promise<boolean> => {
  providerLoading.value = true
  const success = await Data().save("config", "set", appConfig.value)
  providerLoading.value = false
  return success
}
</script>

<style scoped>
.tls-page {
  display: flex;
  flex-direction: column;
  padding-bottom: 24px;
}

/* 两个分区共用同一套宽度、内边距和头部对齐 */
.tls-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tls-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 32px;
}

.tls-section-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  line-height: 32px;
}

.tls-section-action {
  margin-inline-start: auto;
}

.tls-divider {
  margin: 20px 0;
}

.tls-empty {
  margin: 0;
  padding: 2px 0 4px;
  font-size: 13px;
  line-height: 22px;
  opacity: 0.55;
}
</style>

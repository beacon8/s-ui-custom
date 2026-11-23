<template>
  <DnsVue
    v-model="dnsModal.visible"
    :visible="dnsModal.visible"
    :index="dnsModal.index"
    :data="dnsModal.data"
    :ts-tags="tsTags"
    :rslvd-tags="rslvdTags"
    @close="closeDnsModal"
    @save="saveDnsModal"
  />
  <DnsRuleVue
    v-model="dnsRuleModal.visible"
    :visible="dnsRuleModal.visible"
    :index="dnsRuleModal.index"
    :data="dnsRuleModal.data"
    :clients="clients"
    :in-tags="inboundTags"
    :server-tags="dnsServerTags"
    :rule-sets="ruleSets"
    @close="closeDnsRuleModal"
    @save="saveDnsRuleModal"
  />
  <a-row
    :gutter="[12, 12]"
    justify="center"
    style="margin-bottom: 12px;"
  >
    <a-col>
      <a-space>
        <a-button
          type="primary"
          @click="showDnsModal(-1)"
        >
          {{ $t('dns.add') }}
        </a-button>
        <a-button
          type="primary"
          @click="showDnsRuleModal(-1)"
        >
          {{ $t('dns.rule.add') }}
        </a-button>
        <a-button
          type="primary"
          :loading="loading"
          :disabled="stateChange"
          @click="saveConfig"
        >
          {{ $t('actions.save') }}
        </a-button>
      </a-space>
    </a-col>
  </a-row>
  <a-divider
    orientation="left"
    style="margin: 12px 0;"
  >
    {{ $t('pages.basics') }}
  </a-divider>
  <a-form layout="vertical">
    <a-row
      :gutter="[12, 12]"
      align="middle"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="6"
        :lg="4"
      >
        <a-form-item
          :label="$t('dns.final')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="finalDns"
            style="width: 100%"
          >
            <a-select-option value="">
              {{ $t('dns.firstServer') }}
            </a-select-option>
            <a-select-option
              v-for="tag in dnsServerTags"
              :key="tag"
              :value="tag"
            >
              {{ tag }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="6"
        :lg="4"
      >
        <a-form-item
          :label="$t('dns.domainStrategy')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="dns.strategy"
            allow-clear
            style="width: 100%"
            @change="(val: any) => { if (val === undefined) delete dns.strategy }"
          >
            <a-select-option
              v-for="s in ['prefer_ipv4', 'prefer_ipv6', 'ipv4_only', 'ipv6_only']"
              :key="s"
              :value="s"
            >
              {{ s }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="6"
        :lg="4"
      >
        <a-form-item
          :label="$t('dns.rule.action.clientSubnet')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="dns.client_subnet"
            allow-clear
            @change="() => { if (!dns.client_subnet) delete dns.client_subnet }"
          />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="6"
        :lg="4"
      >
        <a-form-item
          :label="$t('dns.cacheCapacity')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="dns.cache_capacity"
            :min="1024"
            style="width: 100%"
            @change="(val: any) => { if (val === null || val === undefined || val === '') delete dns.cache_capacity }"
          />
        </a-form-item>
      </a-col>
      <a-col>
        <a-checkbox v-model:checked="dns.disable_cache">
          {{ $t('dns.disableCache') }}
        </a-checkbox>
      </a-col>
      <a-col>
        <a-checkbox v-model:checked="dns.disable_expire">
          {{ $t('dns.disableExpire') }}
        </a-checkbox>
      </a-col>
      <a-col>
        <a-checkbox v-model:checked="dns.reverse_mapping">
          {{ $t('dns.reverseMapping') }}
        </a-checkbox>
      </a-col>
    </a-row>
  </a-form>
  <a-divider
    orientation="left"
    style="margin: 16px 0 12px;"
  >
    {{ $t('dns.title') }}
  </a-divider>
  <a-row :gutter="[12, 12]">
    <a-col
      v-for="(item, index) in <any[]>dns.servers"
      :key="item.id ?? item.tag ?? index"
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
            {{ $t('dns.server') }}
          </a-col>
          <a-col :span="12">
            {{ item.server ?? '-' }}
          </a-col>
          <a-col :span="12">
            {{ $t('in.port') }}
          </a-col>
          <a-col :span="12">
            {{ item.server_port ?? '-' }}
          </a-col>
          <a-col :span="12">
            {{ $t('objects.tls') }}
          </a-col>
          <a-col :span="12">
            {{ Object.hasOwn(item, 'tls') ? $t(item.tls?.enabled ? 'enable' : 'disable') : '-' }}
          </a-col>
        </a-row>
        <a-divider style="margin: 10px 0 6px;" />
        <a-space :size="4">
          <a-tooltip :title="$t('actions.edit')">
            <a-button
              type="text"
              shape="circle"
              :aria-label="$t('actions.edit')"
              @click="showDnsModal(index)"
            >
              <template #icon>
                <AppIcon name="file-edit" />
              </template>
            </a-button>
          </a-tooltip>
          <a-popconfirm
            v-model:open="delDnsOverlay[index]"
            :title="$t('confirm')"
            :ok-text="$t('yes')"
            :cancel-text="$t('no')"
            ok-type="danger"
            @confirm="delDns(index)"
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
        </a-space>
      </a-card>
    </a-col>
  </a-row>
  <a-divider
    orientation="left"
    style="margin: 16px 0 12px;"
  >
    {{ $t('dns.rule.title') }}
  </a-divider>
  <a-row :gutter="[12, 12]">
    <a-col
      v-for="(item, index) in <any[]>dnsRules"
      :key="item.id ?? index"
      :xs="24"
      :sm="8"
      :md="6"
      :lg="4"
      :draggable="true"
      @dragstart="onDragStart(index)"
      @dragover.prevent
      @drop="onDrop(index)"
    >
      <a-card
        size="small"
        :title="String(index + 1)"
      >
        <div style="font-size: 12px; opacity: 0.65; margin-bottom: 8px;">
          {{ item.type != undefined ? $t('rule.logical') + ' (' + item.mode + ')' : $t('rule.simple') }}
        </div>
        <a-row :gutter="[8, 6]">
          <a-col :span="12">
            {{ $t('admin.action') }}
          </a-col>
          <a-col :span="12">
            {{ item.action }}
          </a-col>
          <a-col :span="12">
            {{ $t('dns.server') }}
          </a-col>
          <a-col :span="12">
            {{ item.server ?? '-' }}
          </a-col>
          <a-col :span="12">
            {{ $t('pages.rules') }}
          </a-col>
          <a-col :span="12">
            {{ item.rules ? item.rules.length : Object.keys(item).filter(r => !actionDnsRuleKeys.includes(r)).length }}
          </a-col>
          <a-col :span="12">
            {{ $t('rule.invert') }}
          </a-col>
          <a-col :span="12">
            {{ $t((item.invert ?? false) ? 'yes' : 'no') }}
          </a-col>
        </a-row>
        <a-divider style="margin: 10px 0 6px;" />
        <a-flex
          justify="space-between"
          align="center"
        >
          <a-space :size="4">
            <a-tooltip :title="$t('actions.edit')">
              <a-button
                type="text"
                shape="circle"
                :aria-label="$t('actions.edit')"
                @click="showDnsRuleModal(index)"
              >
                <template #icon>
                  <AppIcon name="file-edit" />
                </template>
              </a-button>
            </a-tooltip>
            <a-popconfirm
              v-model:open="delDnsRuleOverlay[index]"
              :title="$t('confirm')"
              :ok-text="$t('yes')"
              :cancel-text="$t('no')"
              ok-type="danger"
              @confirm="delDnsRule(index)"
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
          </a-space>
          <a-space :size="4">
            <a-tooltip :title="$t('actions.moveUp')">
              <a-button
                type="text"
                shape="circle"
                :aria-label="$t('actions.moveUp')"
                :disabled="index == 0"
                @click="moveRule(index, -1)"
              >
                <template #icon>
                  <AppIcon name="arrow-up" />
                </template>
              </a-button>
            </a-tooltip>
            <a-tooltip :title="$t('actions.moveDown')">
              <a-button
                type="text"
                shape="circle"
                :aria-label="$t('actions.moveDown')"
                :disabled="index == dnsRules.length - 1"
                @click="moveRule(index, 1)"
              >
                <template #icon>
                  <AppIcon name="arrow-down" />
                </template>
              </a-button>
            </a-tooltip>
          </a-space>
        </a-flex>
      </a-card>
    </a-col>
  </a-row>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import { computed, ref, onBeforeMount } from 'vue'
import DnsVue from '@/layouts/modals/Dns.vue'
import DnsRuleVue from '@/layouts/modals/DnsRule.vue'
import { Config } from '@/types/config'
import { actionDnsRuleKeys, Dns, dnsRule, DnsServer } from '@/types/dns'
import { FindDiff } from '@/plugins/utils'

const oldConfig = ref(<Partial<Config>>{})
const loading = ref(false)

const appConfig = computed((): Config => {
  return <Config> Data().config
})

onBeforeMount( async () => {
  // fix old configs
  if (!appConfig.value.dns) appConfig.value.dns = { servers: [], rules: [] }
  if (!appConfig.value.dns.servers) appConfig.value.dns.servers = []
  if (!appConfig.value.dns.rules) appConfig.value.dns.rules = []

  loading.value = true
  while (Data().lastLoad == 0) {
    await new Promise(resolve => setTimeout(resolve, 100))
  }
  oldConfig.value = JSON.parse(JSON.stringify(Data().config))
  loading.value = false
})

const tsTags = computed((): string[] => {
  return Data().endpoints?.filter(e => e.type == "tailscale").map(e => e.tag)
})

const rslvdTags = computed((): string[] => {
  return Data().services?.filter(e => e.type == "resolved").map(e => e.tag)
})

const clients = computed((): string[] => {
  return Data().clients.map(c => c.name)
})

const stateChange = computed(() => {
  return FindDiff.deepCompare(appConfig.value.dns,oldConfig.value.dns)
})

const saveConfig = async () => {
  loading.value = true
  const success = await Data().save("config", "set", appConfig.value)
  if (success) {
    oldConfig.value = JSON.parse(JSON.stringify(Data().config))
  }
  loading.value = false
}

const inboundTags = computed((): string[] => {
  return [...(Data().inbounds?.map(o => o.tag) ?? []), ...(Data().endpoints?.filter(e => e.listen_port > 0).map(e => e.tag) ?? [])]
})

const dns = computed((): Dns => {
  return appConfig.value.dns
})

const dnsServerTags = computed((): string[] => {
  return dns.value?.servers?.filter(s => s.tag && s.tag != "")?.map(s => s.tag) ?? []
})

const finalDns = computed({
  get() { return dns.value?.final?? '' },
  set(v:string) { dns.value.final = v.length>0 ? v : undefined }
})


const dnsRules = computed((): dnsRule[] => {
  return <dnsRule[]>dns.value.rules
})

const ruleSets = computed((): string[] => {
  return appConfig.value?.route?.rule_set?.map(r => r.tag) ?? []
})

let delDnsOverlay = ref(new Array<boolean>)
let delDnsRuleOverlay = ref(new Array<boolean>)

const dnsModal = ref({
  visible: false,
  index: -1,
  data: "",
})

const showDnsModal = (index: number) => {
  dnsModal.value.index = index
  dnsModal.value.data = index == -1 ? '' : JSON.stringify(dns.value.servers[index])
  dnsModal.value.visible = true
}

const closeDnsModal = () => {
  dnsModal.value.visible = false
}

const saveDnsModal = (data:DnsServer) => {
  // New or Edit
  if (dnsModal.value.index == -1) {
    dns.value.servers.push(data)
  } else {
    dns.value.servers[dnsModal.value.index] = data
  }
  dnsModal.value.visible = false
}

const delDns = (index: number) => {
  dns.value.servers.splice(index,1)
  delDnsOverlay.value[index] = false
}

const dnsRuleModal = ref({
  visible: false,
  index: -1,
  data: "",
})

const showDnsRuleModal = (index: number) => {
  dnsRuleModal.value.index = index
  dnsRuleModal.value.data = index == -1 ? '' : JSON.stringify(dnsRules.value[index])
  dnsRuleModal.value.visible = true
}

const closeDnsRuleModal = () => {
  dnsRuleModal.value.visible = false
}

const saveDnsRuleModal = (data:dnsRule) => {
  // New or Edit
  if (dnsRuleModal.value.index == -1) {
    dnsRules.value.push(data)
  } else {
    dnsRules.value[dnsRuleModal.value.index] = data
  }
  dnsRuleModal.value.visible = false
}

const delDnsRule = (index: number) => {
  dnsRules.value.splice(index,1)
  delDnsRuleOverlay.value[index] = false
}

// Dragging is mouse-only: HTML5 drag and drop never starts from a touch in
// most mobile browsers, so the buttons are the way to reorder there (#1283).
const moveRule = (index: number, delta: number) => {
  const target = index + delta
  if (target < 0 || target >= dnsRules.value.length) return
  const [item] = dnsRules.value.splice(index, 1)
  dnsRules.value.splice(target, 0, item)
}

const draggedItemIndex = ref<number | null>(null)

const onDragStart = (index: number) => {
  draggedItemIndex.value = index
}

const onDrop = (index: number) => {
  if (draggedItemIndex.value !== null) {
    // Swap the dragged item with the dropped one
    const draggedItem = dnsRules.value[draggedItemIndex.value]
    dnsRules.value.splice(draggedItemIndex.value, 1)
    dnsRules.value.splice(index, 0, draggedItem)
    draggedItemIndex.value = null
  }
}
</script>
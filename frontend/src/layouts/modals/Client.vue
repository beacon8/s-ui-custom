<template>
  <a-modal
    :open="visible"
    :width="800"
    :body-style="{ padding: 0 }"
    :title="$t('actions.' + title) + ' ' + $t('objects.client')"
    @cancel="closeModal"
  >
    <a-skeleton
      v-if="loading"
      class="client-loading"
      active
      :paragraph="{ rows: 6 }"
    />
    <div :hidden="loading">
      <a-tabs
        v-model:activeKey="tab"
        class="client-tabs"
        :destroy-inactive-tab-pane="false"
      >
        <a-tab-pane
          key="t1"
          :tab="$t('client.basics')"
        >
          <div class="client-pane">
            <a-form
              layout="vertical"
              class="client-form"
            >
              <!-- 基础信息：名称、组、启用同排，描述与备注并排 -->
              <div class="form-group">
                <a-row :gutter="[16, 16]">
                  <a-col
                    :xs="24"
                    :sm="10"
                    :md="11"
                  >
                    <a-form-item :label="$t('client.name')">
                      <a-input v-model:value="client.name" />
                    </a-form-item>
                  </a-col>
                  <a-col
                    :xs="24"
                    :sm="8"
                    :md="8"
                  >
                    <a-form-item :label="$t('client.group')">
                      <a-auto-complete
                        v-model:value="client.group"
                        :options="groups.map(g => ({ value: g }))"
                        style="width: 100%"
                        allow-clear
                      />
                    </a-form-item>
                  </a-col>
                  <a-col
                    :xs="24"
                    :sm="6"
                    :md="5"
                  >
                    <a-form-item :label="$t('enable')">
                      <div class="switch-field">
                        <a-switch v-model:checked="client.enable" />
                      </div>
                    </a-form-item>
                  </a-col>
                </a-row>
                <a-row :gutter="[16, 16]">
                  <a-col
                    :xs="24"
                    :sm="12"
                  >
                    <a-form-item :label="$t('client.desc')">
                      <a-input v-model:value="client.desc" />
                    </a-form-item>
                  </a-col>
                  <a-col
                    :xs="24"
                    :sm="12"
                  >
                    <a-form-item :label="$t('client.remark')">
                      <a-input v-model:value="client.remark" />
                    </a-form-item>
                  </a-col>
                </a-row>
              </div>

              <!-- 使用限制：三项等宽，自动重置贴近流量，延迟启动贴近到期 -->
              <div class="form-group">
                <div class="form-group-title">
                  {{ $t('client.usageLimit') }}
                </div>
                <a-row :gutter="[16, 16]">
                  <a-col
                    :xs="24"
                    :sm="8"
                  >
                    <a-form-item :label="$t('stats.volume')">
                      <a-input-number
                        v-model:value="Volume"
                        :min="0"
                        addon-after="GiB"
                        style="width: 100%"
                      />
                    </a-form-item>
                    <div class="inline-toggle">
                      <a-switch v-model:checked="autoReset" />
                      <span>{{ $t('client.autoReset') }}</span>
                    </div>
                    <a-form-item
                      v-if="autoReset"
                      :label="$t('client.resetDays')"
                      class="inline-setting"
                    >
                      <a-input-number
                        v-model:value="resetDays"
                        :min="1"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col
                    :xs="24"
                    :sm="8"
                  >
                    <DatePick
                      v-if="!(client.delayStart && !client.autoReset)"
                      :label="$t('client.expiry')"
                      :expiry="expDate"
                      :placeholder="$t('unlimited')"
                      @submit="setDate"
                    />
                    <a-form-item
                      v-else
                      :label="$t('client.expiry')"
                    >
                      <div class="muted-control">
                        {{ $t('client.delayStartHint') }}
                      </div>
                    </a-form-item>
                    <div class="inline-toggle">
                      <a-switch
                        v-model:checked="delayStart"
                        :disabled="client.up + client.down > 0"
                      />
                      <span>{{ $t('client.delayStart') }}</span>
                    </div>
                    <a-form-item
                      v-if="delayStart && !autoReset"
                      :label="$t('client.resetDays')"
                      class="inline-setting"
                    >
                      <a-input-number
                        v-model:value="resetDays"
                        :min="1"
                        style="width: 100%"
                      />
                    </a-form-item>
                    <div
                      v-if="delayStart && autoReset"
                      class="inline-hint"
                    >
                      {{ $t('client.delayStartHint') }}
                    </div>
                  </a-col>
                  <a-col
                    :xs="24"
                    :sm="8"
                  >
                    <a-form-item :label="$t('client.maxConns')">
                      <a-input-number
                        v-model:value="maxConns"
                        :min="0"
                        :placeholder="$t('unlimited')"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <a-row :gutter="[16, 16]">
                  <a-col
                    :xs="24"
                    :sm="12"
                  >
                    <a-form-item :label="$t('client.upRate')">
                      <a-input-number
                        v-model:value="upMbps"
                        :min="0"
                        :max="17179"
                        :step="1"
                        placeholder="0"
                        addon-after="Mbps"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col
                    :xs="24"
                    :sm="12"
                  >
                    <a-form-item :label="$t('client.downRate')">
                      <a-input-number
                        v-model:value="downMbps"
                        :min="0"
                        :max="17179"
                        :step="1"
                        placeholder="0"
                        addon-after="Mbps"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <div class="inline-hint">
                  {{ $t('client.rateHint') }}
                </div>

                <!-- 已有客户端的只读用量摘要 -->
                <div
                  v-if="id > 0"
                  class="usage-summary"
                >
                  <a-row :gutter="[16, 16]">
                    <a-col
                      :xs="24"
                      :sm="8"
                    >
                      <div class="summary-row">
                        <span class="summary-label">
                          {{ $t('stats.usage') }}: {{ total }}<sup
                            v-if="percent > 0"
                            dir="ltr"
                          >({{ percent }}%)</sup>
                        </span>
                        <a-tooltip :title="$t('reset')">
                          <a-button
                            :aria-label="$t('reset')"
                            type="text"
                            size="small"
                            shape="circle"
                            @click="resetUsage"
                          >
                            <template #icon>
                              <AppIcon name="restore" />
                            </template>
                          </a-button>
                        </a-tooltip>
                      </div>
                      <a-progress
                        v-if="client.volume > 0"
                        :percent="percent"
                        :stroke-color="percentColor === 'error' ? '#ff4d4f' : percentColor === 'warning' ? '#faad14' : '#52c41a'"
                        :show-info="false"
                        size="small"
                      />
                    </a-col>
                    <a-col
                      :xs="24"
                      :sm="8"
                    >
                      <div class="summary-label">
                        {{ $t('stats.upload') }} / {{ $t('stats.download') }}
                      </div>
                      <div class="summary-value">
                        <AppIcon
                          name="upload"
                          color="#faad14"
                        />
                        <span style="color: #faad14;">{{ up }}</span>
                        /
                        <AppIcon
                          name="download"
                          color="#52c41a"
                        />
                        <span style="color: #52c41a;">{{ down }}</span>
                      </div>
                    </a-col>
                    <a-col
                      v-if="autoReset"
                      :xs="24"
                      :sm="8"
                    >
                      <div class="summary-label">
                        {{ $t('client.nextReset') }}
                      </div>
                      <div class="summary-value">
                        {{ nextResetFormatted }}
                      </div>
                      <div class="summary-label summary-label-tight">
                        {{ $t('main.stats.totalUsage') }}
                      </div>
                      <div class="summary-value">
                        <AppIcon
                          name="upload"
                          color="#faad14"
                        />
                        <span style="color: #faad14;">{{ totalUp }}</span>
                        /
                        <AppIcon
                          name="download"
                          color="#52c41a"
                        />
                        <span style="color: #52c41a;">{{ totalDown }}</span>
                      </div>
                    </a-col>
                  </a-row>
                </div>
              </div>

              <!-- 入站标签独占一行；右侧是全选，不是单个勾选 -->
              <div class="form-group">
                <a-form-item :label="$t('client.inboundTags')">
                  <a-flex
                    gap="small"
                    align="center"
                  >
                    <a-select
                      v-model:value="clientInbounds"
                      :options="inboundTags"
                      :field-names="{ label: 'title', value: 'value' }"
                      mode="multiple"
                      allow-clear
                      style="width: 100%"
                    />
                    <a-tooltip :title="$t('client.selectAllHint')">
                      <a-button
                        :aria-label="$t('client.selectAllHint')"
                        type="text"
                        style="flex: none;"
                        @click="setAllInbounds"
                      >
                        <template #icon>
                          <AppIcon name="set-all" />
                        </template>
                        {{ $t('client.selectAll') }}
                      </a-button>
                    </a-tooltip>
                  </a-flex>
                </a-form-item>
              </div>
            </a-form>
          </div>
        </a-tab-pane>

        <a-tab-pane
          key="t2"
          :tab="$t('client.config')"
        >
          <div class="client-pane">
            <div class="config-toolbar">
              <a-button
                type="default"
                @click="shuffle()"
              >
                <template #icon>
                  <AppIcon name="refresh" />
                </template>
                {{ $t('client.resetAllCredentials') }}
              </a-button>
            </div>

            <div
              class="config-head"
              aria-hidden="true"
            >
              <span>{{ $t('protocol') }}</span>
              <span>{{ $t('client.credentials') }}</span>
              <span class="config-head-action">{{ $t('actions.action') }}</span>
            </div>

            <div class="config-list">
              <div
                v-for="key in Object.keys(clientConfig)"
                :key="key"
                class="config-row"
              >
                <div class="config-protocol">
                  {{ key }}
                </div>
                <div class="config-credentials">
                  <label
                    v-if="clientConfig[key].password != undefined"
                    class="credential-field"
                  >
                    <span class="credential-label">{{ $t('client.password') }}</span>
                    <a-input v-model:value="clientConfig[key].password" />
                  </label>
                  <label
                    v-if="clientConfig[key].uuid != undefined"
                    class="credential-field"
                  >
                    <span class="credential-label">UUID</span>
                    <a-input v-model:value="clientConfig[key].uuid" />
                  </label>
                  <label
                    v-if="key == 'vless'"
                    class="credential-field"
                  >
                    <span class="credential-label">Flow</span>
                    <a-input v-model:value="clientConfig[key].flow" />
                  </label>
                  <label
                    v-if="key == 'hysteria'"
                    class="credential-field"
                  >
                    <span class="credential-label">{{ $t('client.password') }}</span>
                    <a-input v-model:value="clientConfig[key].auth_str" />
                  </label>
                  <label
                    v-if="key == 'snell'"
                    class="credential-field"
                  >
                    <span class="credential-label">{{ $t('client.userKey') }}</span>
                    <a-input v-model:value="clientConfig[key].userkey" />
                  </label>
                </div>
                <div class="config-actions">
                  <a-tooltip :title="$t('client.resetCredential')">
                    <a-button
                      :aria-label="$t('client.resetCredential')"
                      type="text"
                      size="small"
                      shape="circle"
                      @click="shuffle(key)"
                    >
                      <template #icon>
                        <AppIcon name="refresh" />
                      </template>
                    </a-button>
                  </a-tooltip>
                </div>
              </div>
            </div>
          </div>
        </a-tab-pane>

        <a-tab-pane
          key="t3"
          :tab="$t('client.links')"
        >
          <div class="client-pane">
            <a-form layout="vertical">
              <a-row
                v-for="(lnk, index) in links"
                :key="index"
                :gutter="[16, 16]"
                style="margin-bottom: 8px;"
              >
                <a-col flex="none">
                  {{ index + 1 }}
                </a-col>
                <a-col
                  :flex="1"
                  style="direction: ltr; overflow-y: hidden; word-break: break-all;"
                >
                  {{ lnk.uri }}
                </a-col>
              </a-row>

              <a-row
                :gutter="[16, 16]"
                style="margin-bottom: 16px;"
              >
                <a-col :span="24">
                  <a-button
                    type="primary"
                    @click="extLinks.push({ type: 'external', uri: '' })"
                  >
                    {{ $t('actions.add') }} {{ $t('client.external') }}
                  </a-button>
                </a-col>
              </a-row>

              <a-row
                v-for="(lnk, index) in extLinks"
                :key="index"
                :gutter="[16, 16]"
                style="margin-bottom: 16px;"
              >
                <a-col :span="24">
                  <a-form-item
                    :label="$t('client.external') + ' ' + (index + 1)"
                    style="margin-bottom: 0;"
                  >
                    <a-flex
                      gap="small"
                      align="center"
                    >
                      <a-input
                        v-model:value="lnk.uri"
                        dir="ltr"
                        placeholder="<protocol>://<data>"
                      />
                      <a-button
                        type="text"
                        danger
                        style="flex: none;"
                        @click="extLinks.splice(index, 1)"
                      >
                        <template #icon>
                          <AppIcon name="delete" />
                        </template>
                      </a-button>
                    </a-flex>
                  </a-form-item>
                </a-col>
              </a-row>

              <a-row
                :gutter="[16, 16]"
                style="margin-bottom: 16px;"
              >
                <a-col :span="24">
                  <a-button
                    type="primary"
                    @click="subLinks.push({ type: 'sub', uri: '' })"
                  >
                    {{ $t('actions.add') }} {{ $t('client.sub') }}
                  </a-button>
                </a-col>
              </a-row>

              <a-row
                v-for="(lnk, index) in subLinks"
                :key="index"
                :gutter="[16, 16]"
                style="margin-bottom: 16px;"
              >
                <a-col :span="24">
                  <a-form-item
                    :label="$t('client.sub') + ' ' + (index + 1)"
                    style="margin-bottom: 0;"
                  >
                    <a-flex
                      gap="small"
                      align="center"
                    >
                      <a-input
                        v-model:value="lnk.uri"
                        dir="ltr"
                        placeholder="http[s]://<domain>[:]<port>/<path>"
                      />
                      <a-button
                        type="text"
                        danger
                        style="flex: none;"
                        @click="subLinks.splice(index, 1)"
                      >
                        <template #icon>
                          <AppIcon name="delete" />
                        </template>
                      </a-button>
                    </a-flex>
                  </a-form-item>
                </a-col>
              </a-row>
            </a-form>
          </div>
        </a-tab-pane>
      </a-tabs>
    </div>
    <template #footer>
      <a-button
        type="default"
        @click="closeModal"
      >
        {{ $t('actions.close') }}
      </a-button>
      <a-button
        type="primary"
        :loading="loading"
        @click="saveChanges"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import { createClient, randomConfigs, updateConfigs, Link, shuffleConfigs } from '@/types/clients'
import DatePick from '@/components/DateTime.vue'
import { HumanReadable } from '@/plugins/utils'
import Data from '@/store/modules/data'
import { locale } from '@/locales'
import type { PropType } from 'vue'

// The per-protocol credentials of a client. The type is internal to
// @/types/clients, so it is taken from what builds one.
type ClientConfig = ReturnType<typeof randomConfigs>

// An inbound the client can be attached to, as the parent lists them.
interface InboundTag {
  title: string
  value: number
}

export default {
  components: { DatePick },
  props: {
    visible: { type: Boolean, required: true },
    id: { type: Number, required: true },
    inboundTags: { type: Array as PropType<InboundTag[]>, required: true },
    groups: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['close'],
  data() {
    return {
      client: createClient(),
      title: "add",
      loading: false,
      tab: "t1",
      clientConfig: <ClientConfig>{},
      links: <Link[]>[],
      extLinks: <Link[]>[],
      subLinks: <Link[]>[],
    }
  },
  computed: {
    clientInbounds: {
      get() { return this.client.inbounds.length>0 ? [...this.client.inbounds].sort() : [] },
      set(v:number[]) { this.client.inbounds = v.length == 0 ?  [] : v.sort() }
    },
    expDate: {
      get() { return this.client.expiry},
      set(v:number) { this.client.expiry = v }
    },
    Volume: {
      get() { return this.client.volume == 0 ? 0 : (this.client.volume / (1024 ** 3)) },
      set(v:number) { this.client.volume = v > 0 ? v*(1024 ** 3) : 0 }
    },
    maxConns: {
      get() { return this.client.maxConns ?? 0 },
      set(v:number) { this.client.maxConns = v > 0 ? Math.floor(v) : 0 }
    },
    upMbps: {
      get() { return this.client.upMbps ?? 0 },
      set(v: number | null) { this.client.upMbps = v && v > 0 ? Math.floor(v) : 0 }
    },
    downMbps: {
      get() { return this.client.downMbps ?? 0 },
      set(v: number | null) { this.client.downMbps = v && v > 0 ? Math.floor(v) : 0 }
    },
    delayStart: {
      get() { return this.client.delayStart?? false },
      set(v:boolean) {
        this.client.delayStart = v
        this.client.resetDays = v ? 1 : 0
        if (v && !this.autoReset) this.client.expiry = 0
      }
    },
    autoReset: {
      get() { return this.client.autoReset?? false },
      set(v:boolean) {
        this.client.autoReset = v
        this.client.resetDays = v ? 1 : 0
        if (!v) this.client.nextReset = 0
      }
    },
    resetDays: {
      get() { return this.client.resetDays?? 1 },
      set(v:number|null) {
        if (!v) v = 1
        if (this.client.nextReset && this.client.nextReset > 0) {
          this.client.nextReset += (v-(this.client.resetDays?? 0))*24*60*60
        }
        this.client.resetDays = v
      }
    },
    up() :string { return HumanReadable.sizeFormat(this.client.up) },
    down() :string { return HumanReadable.sizeFormat(this.client.down) },
    total() :string { return HumanReadable.sizeFormat(this.client.down + this.client.up) },
    totalUp() :string { return HumanReadable.sizeFormat((this.client.totalUp ?? 0) + this.client.up) },
    totalDown() :string { return HumanReadable.sizeFormat((this.client.totalDown ?? 0) + this.client.down) },
    nextResetFormatted() :string {
      const ts = this.client.nextReset?? 0
      if (ts == 0) return '-'
      const date = new Date(ts*1000)
      return date.toLocaleString(locale)
    },
    percent() :number { return this.client.volume>0 ? Math.round((this.client.up + this.client.down) *100 / this.client.volume) : 0 },
    percentColor() :string { return (this.client.up+this.client.down) >= this.client.volume ? 'error' : this.percent>90 ? 'warning' : 'success' },
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.updateData(this.$props.id)
      }
    },
  },
  methods: {
    async updateData(id: number) {
      if (id > 0) {
        this.loading = true
        const newData = await Data().loadClients(id)
        this.client = createClient(newData)
        this.title = "edit"
        this.clientConfig = <ClientConfig>this.client.config
        this.loading = false
      }
      else {
        this.client = createClient()
        this.title = "add"
        this.clientConfig = randomConfigs('client')
      }
      this.links = this.client.links?.filter(l => l.type == 'local')?? []
      this.extLinks = this.client.links?.filter(l => l.type == 'external')?? []
      this.subLinks = this.client.links?.filter(l => l.type == 'sub')?? []
      this.tab = "t1"
      this.loading = false
    },
    closeModal() {
      this.updateData(0) // reset
      this.$emit('close')
    },
    async saveChanges() {
      if (!this.$props.visible) return
      // check duplicate name
      const isDuplicateName = Data().checkClientName(this.$props.id, this.client.name)
      if (isDuplicateName) return

      // check if delayStart is true and autoReset is false, set expiry to 0
      if (this.client.delayStart && !this.client.autoReset) this.client.expiry = 0

      // save data
      this.loading = true
      this.client.config = updateConfigs(this.clientConfig, this.client.name)
      this.client.links = [
                        ...this.extLinks.filter(l => l.uri != ''),
                        ...this.subLinks.filter(l => l.uri != '')]
      const success = await Data().save("clients", this.$props.id == 0 ? "new" : "edit", this.client)
      if (success) this.closeModal()
      this.loading = false
    },
    setDate(newDate:number){
      this.client.expiry = newDate
    },
    setAllInbounds(){
      this.client.inbounds = this.inboundTags.map(i => i.value).sort()
    },
    shuffle(k?:string) {
      shuffleConfigs(this.clientConfig, k)
    },
    resetUsage(){
      this.client.totalUp = (this.client.totalUp ?? 0) + this.client.up
      this.client.totalDown = (this.client.totalDown ?? 0) + this.client.down
      this.client.up = 0
      this.client.down = 0
    }
  },
}

</script>

<style scoped>
.client-loading {
  padding: 24px;
}

/* 弹窗正文清零内边距，由标签栏与滚动正文各自控制，保证标题、标签、正文同一左边线 */
.client-tabs {
  display: flex;
  flex-direction: column;
  height: 70vh;
  max-height: 760px;
}

.client-tabs :deep(.ant-tabs-nav) {
  flex: none;
  margin: 0;
  padding: 8px 0 0;
}

.client-tabs :deep(.ant-tabs-content-holder) {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.client-tabs :deep(.ant-tabs-content),
.client-tabs :deep(.ant-tabs-tabpane) {
  height: 100%;
}

/* 只有这一层滚动；右侧留白让控件与滚动条之间保持空隙 */
.client-pane {
  height: 100%;
  overflow-x: hidden;
  overflow-y: auto;
  padding: 16px 24px 24px 0;
}

.client-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.client-form :deep(.ant-form-item) {
  margin-bottom: 0;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group-title {
  font-size: 13px;
  font-weight: 600;
  line-height: 22px;
  opacity: 0.65;
}

.switch-field {
  display: flex;
  align-items: center;
  height: 32px;
}

.inline-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 22px;
  margin-top: 8px;
}

.inline-setting {
  margin-top: 8px;
}

.inline-hint {
  margin-top: 4px;
  font-size: 12px;
  line-height: 18px;
  opacity: 0.55;
}

.muted-control {
  display: flex;
  align-items: center;
  height: 32px;
  padding: 0 11px;
  font-size: 13px;
  border: 1px dashed rgba(128, 128, 128, 0.35);
  border-radius: 6px;
  opacity: 0.7;
}

.usage-summary {
  padding-top: 16px;
  border-top: 1px solid rgba(128, 128, 128, 0.15);
}

.summary-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.summary-label {
  font-size: 12px;
  line-height: 20px;
  opacity: 0.6;
}

.summary-label-tight {
  margin-top: 4px;
}

.summary-value {
  font-size: 13px;
  line-height: 22px;
}

/* 配置页：协议 | 凭据 | 操作 */
.config-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
}

.config-head,
.config-row {
  display: grid;
  grid-template-columns: 112px minmax(0, 1fr) 40px;
  grid-template-areas: "protocol credentials actions";
  gap: 16px;
  align-items: center;
}

.config-head {
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.18);
  font-size: 12px;
  line-height: 20px;
  opacity: 0.6;
}

.config-head > :nth-child(1) {
  grid-area: protocol;
}

.config-head > :nth-child(2) {
  grid-area: credentials;
}

.config-head > :nth-child(3) {
  grid-area: actions;
}

.config-head-action {
  text-align: end;
}

.config-list {
  display: flex;
  flex-direction: column;
}

.config-row {
  padding: 10px 0;
  border-bottom: 1px solid rgba(128, 128, 128, 0.12);
}

.config-row:last-child {
  border-bottom: 0;
}

.config-protocol {
  grid-area: protocol;
  overflow: hidden;
  font-size: 13px;
  font-weight: 500;
  line-height: 32px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.config-credentials {
  display: flex;
  grid-area: credentials;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 16px;
  min-width: 0;
}

.credential-field {
  display: flex;
  flex: 1 1 220px;
  align-items: center;
  gap: 8px;
  min-width: 0;
  margin: 0;
  cursor: text;
}

.credential-label {
  flex: 0 0 52px;
  font-size: 12px;
  line-height: 22px;
  text-align: end;
  white-space: nowrap;
  opacity: 0.6;
}

.credential-field :deep(.ant-input) {
  flex: 1;
  min-width: 0;
}

.config-actions {
  display: flex;
  grid-area: actions;
  justify-content: flex-end;
}

@media (max-width: 575px) {
  .client-pane {
    padding-right: 16px;
  }

  .config-head {
    display: none;
  }

  .config-row {
    grid-template-columns: minmax(0, 1fr) auto;
    grid-template-areas:
      "protocol actions"
      "credentials credentials";
    row-gap: 8px;
    padding: 12px 0;
  }

  .config-protocol {
    line-height: 24px;
  }

  .credential-field {
    flex: 1 1 100%;
  }
}
</style>

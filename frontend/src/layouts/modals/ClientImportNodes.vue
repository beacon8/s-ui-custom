<template>
  <a-modal
    :open="visible"
    :width="940"
    :title="$t('importNodes.title')"
    :body-style="{ padding: '16px 24px 0' }"
    @cancel="closeModal"
  >
    <div class="import-body">
      <a-alert
        type="info"
        show-icon
        :message="$t('importNodes.intro')"
        style="margin-bottom: 16px;"
      />

      <div class="section-title">
        {{ $t('importNodes.batch') }}
      </div>
      <a-form layout="vertical">
        <a-row :gutter="[16, 12]">
          <a-col :xs="24" :sm="8" :md="5">
            <a-form-item :label="$t('importNodes.prefix')">
              <a-input
                v-model:value="form.prefix"
                :placeholder="'US'"
              />
            </a-form-item>
          </a-col>
          <a-col :xs="12" :sm="8" :md="4">
            <a-form-item :label="$t('importNodes.startId')">
              <a-input-number
                v-model:value="form.startId"
                :min="0"
                :max="999999999"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :xs="12" :sm="8" :md="4">
            <a-form-item
              :label="$t('importNodes.idWidth')"
              :extra="$t('importNodes.idWidthHint')"
            >
              <a-input-number
                v-model:value="form.idWidth"
                :min="3"
                :max="12"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :sm="12" :md="5">
            <a-form-item :label="$t('importNodes.defaultProtocol')">
              <a-select v-model:value="form.defaultProtocol">
                <a-select-option value="socks5">SOCKS5</a-select-option>
                <a-select-option value="http">HTTP</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :xs="24" :sm="12" :md="6">
            <a-form-item :label="$t('importNodes.group')">
              <a-input
                :value="groupPreview"
                readonly
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="[16, 12]">
          <a-col :span="24">
            <a-form-item
              :label="$t('client.inboundTags')"
              :extra="$t('importNodes.inboundHint')"
            >
              <a-select
                v-model:value="form.inboundIds"
                :options="inboundTags"
                :field-names="{ label: 'title', value: 'value' }"
                mode="multiple"
                allow-clear
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <div class="section-title">
          {{ $t('importNodes.sharedSettings') }}
        </div>
        <a-row :gutter="[16, 12]">
          <a-col :xs="12" :sm="6" :md="4">
            <a-form-item :label="$t('enable')">
              <div class="switch-field">
                <a-switch v-model:checked="form.enable" />
              </div>
            </a-form-item>
          </a-col>
          <a-col :xs="12" :sm="6" :md="5">
            <a-form-item :label="$t('stats.volume')">
              <a-input-number
                v-model:value="form.volumeGiB"
                :min="0"
                addon-after="GiB"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :sm="12" :md="6">
            <DatePick
              :label="$t('client.expiry')"
              :expiry="form.expiry"
              :placeholder="$t('unlimited')"
              @submit="form.expiry = $event"
            />
          </a-col>
          <a-col :xs="12" :sm="6" :md="4">
            <a-form-item :label="$t('client.maxConns')">
              <a-input-number
                v-model:value="form.maxConns"
                :min="0"
                :placeholder="$t('unlimited')"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :xs="12" :sm="6" :md="5">
            <a-form-item
              :label="$t('client.ipLimit')"
              :extra="$t('client.ipLimitHint')"
            >
              <a-input-number
                v-model:value="form.ipLimit"
                :min="0"
                :placeholder="$t('unlimited')"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="[16, 12]">
          <a-col :xs="12" :sm="6" :md="4">
            <a-form-item :label="$t('client.upRate')">
              <a-input-number
                v-model:value="form.upMbps"
                :min="0"
                :max="17179"
                addon-after="Mbps"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :xs="12" :sm="6" :md="4">
            <a-form-item :label="$t('client.downRate')">
              <a-input-number
                v-model:value="form.downMbps"
                :min="0"
                :max="17179"
                addon-after="Mbps"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :sm="12" :md="8">
            <a-form-item :label="$t('client.desc')">
              <a-input v-model:value="form.desc" />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :sm="12" :md="8">
            <a-form-item :label="$t('client.remark')">
              <a-input v-model:value="form.remark" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="[16, 12]">
          <a-col :xs="24" :sm="8" :md="6">
            <div class="inline-toggle">
              <a-switch v-model:checked="form.delayStart" />
              <span>{{ $t('client.delayStart') }}</span>
            </div>
          </a-col>
          <a-col :xs="24" :sm="8" :md="6">
            <div class="inline-toggle">
              <a-switch v-model:checked="form.autoReset" />
              <span>{{ $t('client.autoReset') }}</span>
            </div>
          </a-col>
          <a-col
            v-if="form.delayStart || form.autoReset"
            :xs="24"
            :sm="8"
            :md="6"
          >
            <a-form-item :label="$t('client.resetDays')">
              <a-input-number
                v-model:value="form.resetDays"
                :min="1"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <div class="section-title">
          {{ $t('importNodes.nodes') }}
        </div>
        <div class="nodes-field">
          <textarea
            v-model="form.nodesText"
            class="ant-input nodes-input"
            :rows="7"
            :placeholder="$t('importNodes.nodesPlaceholder')"
            spellcheck="false"
          ></textarea>
          <div class="nodes-hint">
            {{ $t('importNodes.nodesHint') }}
          </div>
        </div>
      </a-form>
    </div>

    <div
      v-if="preview || result || issues.length > 0 || errorMessage"
      class="result-panel"
    >
      <template v-if="preview && !result">
        <div class="section-title">
          {{ $t('importNodes.previewTitle') }}
        </div>
        <a-descriptions
          bordered
          size="small"
          :column="4"
          style="margin-bottom: 12px;"
        >
          <a-descriptions-item :label="$t('importNodes.group')">
            {{ preview.group || '—' }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('count')">
            {{ preview.count }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('importNodes.endId')">
            {{ preview.count > 0 ? preview.endId : '—' }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('pages.inbounds')">
            {{ (preview.inboundTags ?? []).join(', ') || '—' }}
          </a-descriptions-item>
        </a-descriptions>
        <a-table
          :data-source="preview.items"
          :columns="previewColumns"
          :pagination="false"
          :row-key="(r: ImportItem) => String(r.index)"
          size="small"
          :scroll="{ x: 'max-content' }"
        />
      </template>

      <template v-if="result">
        <div class="section-title">
          {{ $t('importNodes.resultTitle') }}
          <a-tag
            :color="result.batchStatus === 'applied' ? 'green' : result.batchStatus === 'apply_failed' ? 'red' : 'blue'"
          >
            {{ result.batchStatus ?? result.core.status }}
          </a-tag>
        </div>
        <a-descriptions
          bordered
          size="small"
          :column="4"
          style="margin-bottom: 12px;"
        >
          <a-descriptions-item :label="$t('importNodes.batchId')">
            <code>{{ result.batchId }}</code>
          </a-descriptions-item>
          <a-descriptions-item :label="$t('importNodes.group')">
            {{ result.group }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('count')">
            {{ result.count }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('importNodes.coreStatus')">
            {{ $t('importNodes.core.' + result.core.status) }}
            <span v-if="result.core.error"> — {{ result.core.error }}</span>
            <span v-else-if="result.core.reason"> — {{ result.core.reason }}</span>
          </a-descriptions-item>
          <a-descriptions-item :label="$t('importNodes.groupSubscriptionId')">
            <code v-if="result.groupSubscription?.subscriptionId">{{ result.groupSubscription.subscriptionId }}</code>
            <span v-else>—</span>
          </a-descriptions-item>
        </a-descriptions>
        <a-space
          v-if="result.groupSubscription?.link"
          style="margin-bottom: 12px;"
        >
          <a-button
            size="small"
            @click="copyText(result.groupSubscription!.link!)"
          >
            {{ $t('importNodes.copyGroupSubscription') }}
          </a-button>
          <span class="subscription-id">{{ $t('importNodes.groupSubscriptionHint') }}</span>
        </a-space>
        <a-table
          :data-source="result.items"
          :columns="resultColumns"
          :pagination="false"
          :row-key="(r: ImportItem) => String(r.clientId)"
          size="small"
          :scroll="{ x: 'max-content' }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'subscription'">
              <a-space :size="4">
                <code>{{ record.subscriptionId || '—' }}</code>
                <a-button
                  v-if="record.subscriptionLink"
                  type="link"
                  size="small"
                  @click="copyText(record.subscriptionLink)"
                >
                  {{ $t('importNodes.copyPersonalSubscription') }}
                </a-button>
              </a-space>
            </template>
          </template>
        </a-table>
      </template>

      <template v-if="issues.length > 0">
        <div class="section-title error-title">
          {{ $t('importNodes.errors') }}
        </div>
        <a-table
          :data-source="issues"
          :columns="issueColumns"
          :pagination="false"
          :row-key="(r: ImportIssue, i: number) => String(r.code + '-' + (r.index ?? i))"
          size="small"
          :scroll="{ x: 'max-content' }"
        />
      </template>

      <a-alert
        v-if="errorMessage"
        type="error"
        show-icon
        :message="errorMessage"
        style="margin-top: 12px;"
      >
        <template
          v-if="failedBatchId"
          #action
        >
          <a-button
            size="small"
            :loading="reapplying"
            @click="reapply"
          >
            {{ $t('importNodes.reapply') }}
          </a-button>
        </template>
      </a-alert>
    </div>

    <template #footer>
      <a-button @click="closeModal">
        {{ $t('actions.close') }}
      </a-button>
      <a-button
        :loading="previewing"
        :disabled="!canSubmit"
        @click="runPreview"
      >
        {{ $t('importNodes.preview') }}
      </a-button>
      <a-button
        type="primary"
        :loading="submitting"
        :disabled="!canSubmit"
        @click="submit"
      >
        {{ $t('actions.submit') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts" setup>
import DatePick from '@/components/DateTime.vue'
import Data from '@/store/modules/data'
import api from '@/plugins/api'
import { i18n } from '@/locales'
import { push } from 'notivue'
import { computed, reactive, ref, watch, type PropType } from 'vue'
import { classifyNodeLines, importGroupName } from '@/utils/importNodes'

interface InboundTag {
  title: string
  value: number
}

interface ImportIssue {
  code: string
  index?: number
  line?: number
  field?: string
  conflict?: string
  message: string
}

interface ImportItem {
  index: number
  line?: number
  clientId: number
  clientName: string
  outboundTag: string
  protocol: string
  address: string
  subscriptionId?: string
  subscriptionLink?: string
  subscriptionRevoked?: boolean
}

interface GroupSubscription {
  subscriptionId?: string
  name?: string
  link?: string
  revoked?: boolean
}

interface CoreStatus {
  status: string
  reason?: string
  error?: string
  at?: number
}

interface ImportResult {
  batchId: string
  batchStatus?: string
  prefix: string
  group: string
  count: number
  startId: number
  endId: number
  idWidth: number
  inboundIds: number[]
  inboundTags: string[]
  items: ImportItem[]
  groupSubscription?: GroupSubscription
  errors?: ImportIssue[]
  core: CoreStatus
  dryRun: boolean
  valid: boolean
  replayed?: boolean
  retry?: string
}

interface ImportFailure {
  code: string
  message: string
  errors?: ImportIssue[]
  batchId?: string
  core?: CoreStatus
  retry?: string
}

interface Envelope<T> {
  success: boolean
  msg: string
  obj: T
}

const props = defineProps({
  visible: { type: Boolean, required: true },
  inboundTags: { type: Array as PropType<InboundTag[]>, required: true },
})
const emit = defineEmits(['close'])

const emptyForm = () => ({
  prefix: 'US',
  startId: 1,
  idWidth: 3,
  defaultProtocol: 'socks5',
  inboundIds: <number[]>[],
  enable: true,
  desc: '',
  remark: '',
  volumeGiB: 0,
  expiry: 0,
  delayStart: false,
  autoReset: false,
  resetDays: 1,
  maxConns: 0,
  ipLimit: 0,
  upMbps: 0,
  downMbps: 0,
  nodesText: '',
})

const form = reactive(emptyForm())
const preview = ref<ImportResult | null>(null)
const result = ref<ImportResult | null>(null)
const errorMessage = ref('')
const failedBatchId = ref('')
const previewing = ref(false)
const submitting = ref(false)
const reapplying = ref(false)
const idempotencyKey = ref(newKey())

function newKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `imp-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function nodeLines() {
  return classifyNodeLines(form.nodesText)
}

const canSubmit = computed(() => nodeLines().length > 0 && form.inboundIds.length > 0)
const groupPreview = computed(() => importGroupName(form.prefix, form.startId, nodeLines().length) || '—')
const issues = computed((): ImportIssue[] => {
  const list = result.value?.errors ?? preview.value?.errors ?? []
  return list.length > 0 ? list : (failedBatch.value?.errors ?? [])
})
const failedBatch = ref<ImportFailure | null>(null)

const previewColumns = [
  { title: 'Line', dataIndex: 'line', key: 'line', width: 64 },
  { title: i18n.global.t('client.name'), dataIndex: 'clientName', key: 'clientName' },
  { title: i18n.global.t('outbounds'), dataIndex: 'outboundTag', key: 'outboundTag' },
  { title: i18n.global.t('protocol'), dataIndex: 'protocol', key: 'protocol' },
  { title: i18n.global.t('importNodes.address'), dataIndex: 'address', key: 'address' },
]
const resultColumns = [
  { title: 'ID', dataIndex: 'clientId', key: 'clientId', width: 70 },
  { title: i18n.global.t('client.name'), dataIndex: 'clientName', key: 'clientName' },
  { title: i18n.global.t('importNodes.personalSubscription'), key: 'subscription' },
  { title: i18n.global.t('outbounds'), dataIndex: 'outboundTag', key: 'outboundTag' },
  { title: i18n.global.t('protocol'), dataIndex: 'protocol', key: 'protocol' },
  { title: i18n.global.t('importNodes.address'), dataIndex: 'address', key: 'address' },
]
const issueColumns = [
  { title: 'Line', dataIndex: 'line', key: 'line', width: 64 },
  { title: i18n.global.t('importNodes.errorCode'), dataIndex: 'code', key: 'code' },
  { title: i18n.global.t('importNodes.errorMessage'), dataIndex: 'message', key: 'message' },
  { title: i18n.global.t('importNodes.conflict'), dataIndex: 'conflict', key: 'conflict' },
]

watch(form, () => {
  if (submitting.value) return
  preview.value = null
  result.value = null
  errorMessage.value = ''
  failedBatch.value = null
  failedBatchId.value = ''
  idempotencyKey.value = newKey()
}, { deep: true })

watch(() => props.visible, (visible) => {
  if (visible) reset()
})

function reset() {
  Object.assign(form, emptyForm())
  preview.value = null
  result.value = null
  errorMessage.value = ''
  failedBatch.value = null
  failedBatchId.value = ''
  idempotencyKey.value = newKey()
}

function closeModal() {
  emit('close')
}

function payload(dryRun: boolean) {
  return {
    prefix: form.prefix.trim(),
    startId: form.startId,
    idWidth: form.idWidth,
    inboundIds: form.inboundIds,
    defaultProtocol: form.defaultProtocol,
    clientDefaults: {
      enable: form.enable,
      desc: form.desc,
      remark: form.remark,
      volumeGiB: form.volumeGiB,
      expiry: form.expiry,
      delayStart: form.delayStart,
      autoReset: form.autoReset,
      resetDays: form.resetDays,
      maxConns: form.maxConns,
      ipLimit: form.ipLimit,
      upMbps: form.upMbps,
      downMbps: form.downMbps,
    },
    nodes: nodeLines(),
    dryRun,
  }
}

async function callImport(dryRun: boolean, key: string): Promise<{ ok: boolean, status: number, msg: Envelope<ImportResult | ImportFailure> }> {
  try {
    const response = await api.post('api/importClientNodes', payload(dryRun), {
      headers: {
        'Content-Type': 'application/json',
        ...(key ? { 'Idempotency-Key': key } : {}),
      },
    })
    return { ok: true, status: response.status, msg: response.data as Envelope<ImportResult> }
  } catch (error: unknown) {
    const err = error as { response?: { status?: number, data?: Envelope<ImportFailure> } }
    if (err.response?.data) {
      return { ok: false, status: err.response.status ?? 0, msg: err.response.data }
    }
    return {
      ok: false,
      status: 0,
      msg: { success: false, msg: String(error), obj: { code: 'NETWORK', message: String(error) } },
    }
  }
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    push.success({ title: i18n.global.t('success'), message: i18n.global.t('copyToClipboard') })
  } catch {
    push.error({ title: i18n.global.t('failed'), message: i18n.global.t('copyToClipboard') })
  }
}

async function refreshData() {
  // The incremental poll compares whole seconds; an import can land in the
  // same second as the last poll, so force one full load instead.
  const store = Data()
  store.lastLoad = 0
  await store.loadData()
}

async function runPreview() {
  if (!canSubmit.value) return
  previewing.value = true
  errorMessage.value = ''
  failedBatch.value = null
  failedBatchId.value = ''
  result.value = null
  const response = await callImport(true, '')
  previewing.value = false
  if (response.ok && response.msg.success) {
    preview.value = response.msg.obj as ImportResult
  } else {
    failedBatch.value = (response.msg.obj as ImportFailure) ?? null
    errorMessage.value = response.msg.msg || (response.msg.obj as ImportFailure)?.message || i18n.global.t('failed')
  }
}

async function submit() {
  if (!canSubmit.value) return
  submitting.value = true
  errorMessage.value = ''
  failedBatch.value = null
  failedBatchId.value = ''
  const response = await callImport(false, idempotencyKey.value)
  submitting.value = false
  if (response.ok && response.msg.success) {
    result.value = response.msg.obj as ImportResult
    preview.value = null
    idempotencyKey.value = newKey()
    push.success({
      title: i18n.global.t('success'),
      message: i18n.global.t('importNodes.created', { count: result.value.count, group: result.value.group }),
      duration: 5000,
    })
    await refreshData()
    return
  }
  const failure = (response.msg.obj as ImportFailure) ?? null
  failedBatch.value = failure
  errorMessage.value = response.msg.msg || failure?.message || i18n.global.t('failed')
  if (failure?.batchId) {
    failedBatchId.value = failure.batchId
  }
}

async function reapply() {
  const batchId = failedBatchId.value || result.value?.batchId || ''
  if (batchId === '') return
  reapplying.value = true
  try {
    const response = await api.post('api/reapplyImportBatch', { batchId }, {
      headers: { 'Content-Type': 'application/json' },
    })
    const envelope = response.data as Envelope<ImportResult>
    if (envelope.success) {
      result.value = envelope.obj
      failedBatch.value = null
      failedBatchId.value = ''
      errorMessage.value = ''
      push.success({
        title: i18n.global.t('success'),
        message: i18n.global.t('importNodes.reapplied'),
        duration: 5000,
      })
      await refreshData()
    }
  } catch (error: unknown) {
    const err = error as { response?: { data?: Envelope<ImportFailure> } }
    const failure = err.response?.data?.obj
    errorMessage.value = err.response?.data?.msg || failure?.message || String(error)
  } finally {
    reapplying.value = false
  }
}
</script>

<style scoped>
.import-body {
  max-height: 58vh;
  overflow-y: auto;
  padding-right: 8px;
}

.section-title {
  margin: 4px 0 12px;
  font-size: 13px;
  font-weight: 600;
  line-height: 22px;
  opacity: 0.7;
}

.result-panel {
  max-height: 40vh;
  overflow-y: auto;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid rgba(128, 128, 128, 0.18);
}

.error-title {
  color: #ff4d4f;
}

.nodes-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.nodes-hint {
  font-size: 12px;
  line-height: 18px;
  opacity: 0.55;
}

.nodes-input {
  resize: vertical;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  line-height: 1.5;
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
  min-height: 32px;
  margin-top: 4px;
}

:deep(.ant-form-item) {
  margin-bottom: 0;
}
</style>

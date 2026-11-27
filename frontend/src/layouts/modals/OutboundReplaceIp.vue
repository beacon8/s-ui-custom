<template>
  <a-modal
    :open="visible"
    :width="980"
    :title="$t('outboundReplace.title')"
    @cancel="closeModal"
  >
    <a-alert
      v-if="countMismatch"
      type="warning"
      show-icon
      :message="$t('outboundReplace.countMismatch', { selected: targets.length, input: parsed.length })"
      style="margin-bottom: 12px;"
    />
    <a-alert
      v-else-if="hasErrors"
      type="error"
      show-icon
      :message="$t('outboundReplace.errors')"
      style="margin-bottom: 12px;"
    />
    <a-alert
      v-if="submitError"
      type="error"
      show-icon
      :message="submitError"
      style="margin-bottom: 12px;"
    />
    <a-alert
      v-if="applyWarning"
      type="warning"
      show-icon
      :message="$t('outboundReplace.savedNotApplied', { error: applyWarning })"
      style="margin-bottom: 12px;"
    />
    <a-form layout="vertical">
      <a-form-item
        :label="$t('outboundReplace.inputLabel')"
        style="margin-bottom: 12px;"
      >
        <a-textarea
          v-model:value="input"
          :rows="6"
          :placeholder="$t('outboundReplace.placeholder')"
          style="font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;"
        />
      </a-form-item>
    </a-form>
    <a-table
      :columns="columns"
      :data-source="rows"
      :pagination="false"
      :scroll="{ x: 'max-content' }"
      row-key="tag"
      size="small"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'index'">
          {{ record.index }}
        </template>
        <template v-else-if="column.key === 'tag'">
          {{ record.tag }}
        </template>
        <template v-else-if="column.key === 'type'">
          {{ record.type }}
        </template>
        <template v-else-if="column.key === 'original'">
          {{ record.original }}
        </template>
        <template v-else-if="column.key === 'newAddress'">
          <span>{{ record.newAddress }}</span>
          <div
            v-if="record.error"
            class="replace-error"
          >
            {{ $t('outboundReplace.' + record.error) }}
          </div>
        </template>
        <template v-else-if="column.key === 'username'">
          {{ record.username }}
        </template>
        <template v-else-if="column.key === 'password'">
          {{ record.password }}
        </template>
      </template>
    </a-table>
    <template #footer>
      <a-button
        type="default"
        :disabled="submitting"
        @click="closeModal"
      >
        {{ $t('actions.cancel') }}
      </a-button>
      <a-button
        type="primary"
        danger
        :loading="submitting"
        :disabled="!canSubmit"
        @click="submit"
      >
        {{ $t('outboundReplace.confirm', { count: targets.length }) }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import { computed, defineComponent, ref, watch, type PropType } from 'vue'
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'
import Data from '@/store/modules/data'
import { Outbound } from '@/types/outbounds'
import { parseOutboundReplaceInput, replaceAddress } from '@/utils/outboundReplace'
import { push } from 'notivue'

interface ReplaceReply {
  saved: number
  applied: boolean
  warning?: string
  outbounds?: Outbound[]
}

interface ReplyMsg {
  success: boolean
  msg: string
  obj?: ReplaceReply
}

export default defineComponent({
  props: {
    visible: { type: Boolean, required: true },
    targets: { type: Array as PropType<Outbound[]>, required: true },
  },
  emits: ['close', 'done'],
  setup(props, { emit }) {
    const input = ref('')
    const submitting = ref(false)
    const submitError = ref('')
    const applyWarning = ref('')

    const parsed = computed(() => parseOutboundReplaceInput(input.value))
    const countMismatch = computed(() => parsed.value.length !== props.targets.length)
    const hasErrors = computed(() => parsed.value.some(line => line.error !== undefined))
    const canSubmit = computed(() => props.targets.length > 0
      && !countMismatch.value
      && !hasErrors.value
      && !submitting.value)

    const rows = computed(() => props.targets.map((target, index) => {
      const line = parsed.value[index]
      return {
        tag: target.tag,
        index: index + 1,
        type: target.type,
        original: replaceAddress(target.server ?? '', target.server_port),
        newAddress: line ? replaceAddress(line.server, line.port) : '—',
        username: line?.username ?? '',
        password: line?.password ?? '',
        error: line?.error,
      }
    }))

    const columns = computed(() => [
      { title: i18n.global.t('outboundReplace.index'), key: 'index', width: 64 },
      { title: i18n.global.t('outboundReplace.outbound'), key: 'tag', width: 160 },
      { title: i18n.global.t('outboundReplace.protocol'), key: 'type', width: 90 },
      { title: i18n.global.t('outboundReplace.original'), key: 'original', width: 180 },
      { title: i18n.global.t('outboundReplace.newAddress'), key: 'newAddress', width: 200 },
      { title: i18n.global.t('outboundReplace.username'), key: 'username', width: 150 },
      { title: i18n.global.t('outboundReplace.password'), key: 'password', width: 170 },
    ])

    watch(() => props.visible, (visible) => {
      if (visible) {
        input.value = ''
        submitError.value = ''
        applyWarning.value = ''
        submitting.value = false
      }
    })

    const closeModal = () => {
      if (submitting.value) return
      emit('close')
    }

    const submit = async () => {
      if (!canSubmit.value) return
      submitting.value = true
      submitError.value = ''
      applyWarning.value = ''
      const items = props.targets.map((target, index) => {
        const line = parsed.value[index]
        return {
          tag: target.tag,
          server: line.server,
          server_port: line.port,
          username: line.username,
          password: line.password,
        }
      })
      try {
        const msg = await HttpUtils.post<ReplaceReply>('api/replaceOutbounds', { items }, { headers: { 'Content-Type': 'application/json' } }) as ReplyMsg
        if (!msg.success) {
          submitError.value = msg.msg || i18n.global.t('failed')
          return
        }
        if (msg.obj?.warning) {
          applyWarning.value = msg.obj.warning
          push.error({ title: i18n.global.t('outboundReplace.savedNotAppliedTitle'), message: msg.obj.warning })
        } else {
          push.success({ title: i18n.global.t('success'), message: i18n.global.t('outboundReplace.success', { count: msg.obj?.saved ?? props.targets.length }) })
        }
        if (msg.obj?.outbounds) {
          Data().outbounds = msg.obj.outbounds
        }
        emit('done')
        emit('close')
      } catch (error) {
        submitError.value = error instanceof Error ? error.message : String(error)
      } finally {
        submitting.value = false
      }
    }

    return {
      input,
      submitting,
      submitError,
      applyWarning,
      parsed,
      countMismatch,
      hasErrors,
      canSubmit,
      rows,
      columns,
      closeModal,
      submit,
    }
  },
})
</script>

<style scoped>
.replace-error {
  color: #ff4d4f;
  font-size: 12px;
}
</style>

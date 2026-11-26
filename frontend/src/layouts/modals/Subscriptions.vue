<template>
  <a-modal
    :open="visible"
    :width="860"
    :title="$t('subscriptions.title')"
    @cancel="emit('close')"
  >
    <template v-if="!detail">
      <a-table
        :data-source="groups"
        :columns="groupColumns"
        :pagination="false"
        row-key="id"
        size="small"
        :scroll="{ x: 'max-content' }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="record.revoked ? 'red' : 'green'">
              {{ record.revoked ? $t('subscriptions.revoked') : $t('subscriptions.active') }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-button
              type="link"
              size="small"
              @click="openGroup(record.id)"
            >
              {{ $t('subscriptions.open') }}
            </a-button>
          </template>
        </template>
      </a-table>
    </template>

    <template v-else>
      <a-space style="margin-bottom: 12px;">
        <a-button
          size="small"
          @click="detail = null"
        >
          {{ $t('subscriptions.back') }}
        </a-button>
        <a-tag>{{ detail.name }}</a-tag>
        <code>{{ detail.subscriptionId }}</code>
        <a-button
          size="small"
          @click="renameGroup"
        >
          {{ $t('subscriptions.rename') }}
        </a-button>
        <a-button
          size="small"
          danger
          @click="revoke('group', detail.id, !detail.revoked)"
        >
          {{ detail.revoked ? $t('subscriptions.unrevoke') : $t('subscriptions.revoke') }}
        </a-button>
      </a-space>
      <a-space
        v-if="detail.link"
        style="margin-bottom: 12px;"
      >
        <a-button
          size="small"
          @click="copy(detail.link)"
        >
          {{ $t('subscriptions.copyGroup') }}
        </a-button>
        <a-button
          size="small"
          @click="rotate('group', detail.id)"
        >
          {{ $t('subscriptions.rotate') }}
        </a-button>
      </a-space>
      <a-table
        :data-source="detail.members"
        :columns="memberColumns"
        :pagination="false"
        row-key="id"
        size="small"
        :scroll="{ x: 'max-content' }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="record.revoked ? 'red' : record.enabled ? 'green' : 'orange'">
              {{ record.revoked ? $t('subscriptions.revoked') : record.enabled ? $t('subscriptions.active') : $t('subscriptions.disabled') }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space :size="2">
              <a-button
                v-if="record.link"
                type="link"
                size="small"
                @click="copy(record.link)"
              >
                {{ $t('subscriptions.copy') }}
              </a-button>
              <a-button
                type="link"
                size="small"
                @click="rotate('member', record.id)"
              >
                {{ $t('subscriptions.rotate') }}
              </a-button>
              <a-button
                type="link"
                size="small"
                @click="revoke('member', record.id, !record.revoked)"
              >
                {{ record.revoked ? $t('subscriptions.unrevoke') : $t('subscriptions.revoke') }}
              </a-button>
              <a-popconfirm
                :title="$t('confirm')"
                @confirm="removeMember(record.id)"
              >
                <a-button
                  type="link"
                  danger
                  size="small"
                >
                  {{ $t('actions.del') }}
                </a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </template>
  </a-modal>
</template>

<script lang="ts" setup>
import HttpUtils from '@/plugins/httputil'
import { i18n } from '@/locales'
import { push } from 'notivue'
import { computed, ref, watch } from 'vue'

interface GroupRow {
  id: number
  subscriptionId: string
  name: string
  batchId?: string
  revoked: boolean
  memberCount: number
  link?: string
  members?: MemberRow[]
}
interface MemberRow {
  id: number
  subscriptionId: string
  clientId: number
  clientName: string
  enabled: boolean
  revoked: boolean
  link?: string
}

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits(['close'])
const groups = ref<GroupRow[]>([])
const detail = ref<GroupRow | null>(null)

const groupColumns = [
  { title: i18n.global.t('subscriptions.name'), dataIndex: 'name', key: 'name' },
  { title: i18n.global.t('subscriptions.id'), dataIndex: 'subscriptionId', key: 'subscriptionId' },
  { title: i18n.global.t('subscriptions.members'), dataIndex: 'memberCount', key: 'memberCount' },
  { title: i18n.global.t('subscriptions.status'), key: 'status' },
  { title: i18n.global.t('actions.action'), key: 'actions' },
]
const memberColumns = computed(() => [
  { title: i18n.global.t('client.name'), dataIndex: 'clientName', key: 'clientName' },
  { title: i18n.global.t('subscriptions.id'), dataIndex: 'subscriptionId', key: 'subscriptionId' },
  { title: i18n.global.t('subscriptions.status'), key: 'status' },
  { title: i18n.global.t('actions.action'), key: 'actions' },
])

watch(() => props.visible, (v) => { if (v) loadGroups() })

async function loadGroups() {
  detail.value = null
  const msg = await HttpUtils.get<GroupRow[]>('api/subscriptionGroups')
  if (msg.success) groups.value = msg.obj ?? []
}

async function openGroup(id: number) {
  const msg = await HttpUtils.get<GroupRow>('api/subscriptionGroup', { groupId: id })
  if (msg.success) detail.value = msg.obj
}

async function copy(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    push.success({ title: i18n.global.t('success'), message: i18n.global.t('copyToClipboard') })
  } catch {
    push.error({ title: i18n.global.t('failed'), message: i18n.global.t('copyToClipboard') })
  }
}

async function rotate(kind: 'group' | 'member', id: number) {
  const msg = await HttpUtils.post<{ link: string }>('api/rotateSubscription', { kind, id })
  if (msg.success) {
    push.success({ title: i18n.global.t('success'), message: i18n.global.t('subscriptions.rotated') })
    if (msg.obj?.link) await copy(msg.obj.link)
    if (detail.value) await openGroup(detail.value.id)
  }
}

async function revoke(kind: 'group' | 'member', id: number, revoke: boolean) {
  const msg = await HttpUtils.post('api/revokeSubscription', { kind, id, revoke })
  if (msg.success && detail.value) await openGroup(detail.value.id)
}

async function removeMember(memberId: number) {
  const msg = await HttpUtils.post('api/removeSubscriptionMember', { memberId })
  if (msg.success && detail.value) await openGroup(detail.value.id)
}

async function renameGroup() {
  const name = detail.value?.name ?? ''
  const next = window.prompt(i18n.global.t('subscriptions.renamePrompt'), name)
  if (!next || next.trim() === '' || !detail.value) return
  const msg = await HttpUtils.post<GroupRow>('api/renameSubscriptionGroup', { groupId: detail.value.id, name: next.trim() })
  if (msg.success) {
    detail.value = msg.obj
    await loadGroups()
    if (detail.value) await openGroup(detail.value.id)
  }
}
</script>

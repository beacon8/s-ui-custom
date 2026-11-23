<template>
  <a-modal
    :open="visible"
    :width="800"
    :title="$t('admin.api.title')"
    @cancel="closeModal"
  >
    <a-spin :spinning="loading">
      <a-alert
        v-if="newToken.token.length > 0"
        type="success"
        show-icon
        style="margin-bottom: 12px;"
      >
        <template #message>
          {{ $t('admin.api.msg') }}
        </template>
        <template #description>
          <a-input
            :value="newToken.token"
            readonly
            style="margin-top: 8px;"
          >
            <template #suffix>
              <AppIcon
                name="content-copy"
                style="cursor: pointer;"
                @click="copyToClipboard(newToken.token)"
              />
            </template>
          </a-input>
        </template>
      </a-alert>

      <a-table
        :data-source="tokens"
        :columns="tokenColumns"
        :pagination="false"
        size="small"
        row-key="id"
        style="margin-bottom: 12px;"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'expiry'">
            {{ dateFormatted(record.expiry) }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-popconfirm
              :title="$t('confirm')"
              :ok-text="$t('yes')"
              :cancel-text="$t('no')"
              @confirm="deleteToken(record.id)"
            >
              <a-button
                type="text"
                danger
                size="small"
                :aria-label="$t('actions.del')"
              >
                <template #icon>
                  <AppIcon name="delete" />
                </template>
              </a-button>
            </a-popconfirm>
          </template>
        </template>
      </a-table>

      <a-button
        type="primary"
        @click="showAddToken()"
      >
        {{ $t('actions.add') }}
      </a-button>
    </a-spin>

    <a-modal
      v-model:open="showNewToken"
      :width="340"
      :title="$t('admin.api.token')"
    >
      <a-form layout="vertical">
        <a-form-item :label="$t('client.desc')">
          <a-input v-model:value="newToken.desc" />
        </a-form-item>
        <a-form-item
          :label="$t('date.expiry')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="newToken.expiry"
            :min="0"
            :addon-after="$t('date.d')"
            style="width: 100%;"
          />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button
          type="default"
          @click="showNewToken = false"
        >
          {{ $t('actions.close') }}
        </a-button>
        <a-button
          type="primary"
          :loading="loading"
          @click="addToken"
        >
          {{ $t('actions.add') }}
        </a-button>
      </template>
    </a-modal>

    <template #footer>
      <a-button
        type="default"
        @click="closeModal"
      >
        {{ $t('actions.close') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'
import Clipboard from 'clipboard'
import { push } from 'notivue';
import type { PropType } from 'vue'

// One row of api/tokens, mirroring the backend's Tokens model.
interface ApiToken {
  id: number
  desc: string
  token: string
  expiry: number
  userId?: number
}

// The owning admin, as the parent knows it. Nothing is read from it here.
interface AdminUser {
  id?: number
  username?: string
}

export default {
  props: {
    visible: { type: Boolean, required: true },
    user: { type: Object as PropType<AdminUser>, default: () => ({}) },
  },
  emits: ['close'],
  data() {
    return {
      loading: false,
      tokens: <ApiToken[]>[],
      showNewToken: false,
      newToken: {
        desc: '',
        token: '',
        expiry: 0,
      },
      delOverlay: new Array<boolean>(0),
    }
  },
  computed: {
    tokenColumns() {
      return [
        { title: '#', dataIndex: 'id', key: 'id', width: 60 },
        { title: this.$t('admin.api.token'), dataIndex: 'token', key: 'token' },
        { title: this.$t('client.desc'), dataIndex: 'desc', key: 'desc' },
        { title: this.$t('date.expiry'), dataIndex: 'expiry', key: 'expiry' },
        { title: this.$t('actions.del'), key: 'actions', width: 80 },
      ]
    },
    locale() {
      const l = i18n.global.locale.value
      switch (l) {
        case "zhHans":
          return "zh-cn"
        case "zhHant":
          return "zh-tw"
        default:
          return l
      }
    },
  },
  watch: {
    visible(v) {
      if (v) {
        this.resetNewToken()
        this.loadData()
      }
    },
  },
  methods: {
    async loadData() {
      this.loading = true
      const data = await HttpUtils.get<ApiToken[]>('api/tokens')
      if (data.success) {
        this.tokens = data.obj ?? []
        this.delOverlay = new Array<boolean>(this.tokens.length).fill(false)
      }
      this.loading = false
    },
    resetNewToken() {
      this.newToken={
          desc: '',
          token: '',
          expiry: 30,
        }
    },
    showAddToken() {
      this.resetNewToken()
      this.showNewToken = true
    },
    async addToken() {
      this.loading = true
      this.newToken.expiry = this.newToken.expiry>0 ? this.newToken.expiry : 0
      const response = await HttpUtils.post<string>('api/addToken', { desc: this.newToken.desc, expiry: this.newToken.expiry })
      if (response.success) {
        this.newToken.token = response.obj
        this.loadData()
        this.showNewToken = false
      }
      this.loading = false
    },
    async deleteToken(id: number) {
      this.loading = true
      const response = await HttpUtils.post('api/deleteToken', { id: id })
      if (response.success) {
        this.loadData()
      }
      this.loading = false
    },
    dateFormatted(expiry: number) {
      if (expiry == 0) return i18n.global.t('unlimited')
      const date = new Date(expiry*1000)
      return date.toLocaleString(this.locale, {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
      })
    },
    copyToClipboard(txt:string) {
      const hiddenButton = document.createElement('button')
      hiddenButton.className = 'clipboard-btn'
      document.body.appendChild(hiddenButton)

      const clipboard = new Clipboard('.clipboard-btn', {
        text: () => txt,
        container: document.getElementById('qrcode-modal')?? undefined
      });

      clipboard.on('success', () => {
        clipboard.destroy()
        push.success({
          message: i18n.global.t('success') + ": " + i18n.global.t('copyToClipboard'),
          duration: 5000,
        })
      })

      clipboard.on('error', () => {
        clipboard.destroy()
        push.error({
          message: i18n.global.t('failed') + ": " + i18n.global.t('copyToClipboard'),
          duration: 5000,
        })
      })

      // Perform click on hidden button to trigger copy
      hiddenButton.click()
      document.body.removeChild(hiddenButton)
    },
    closeModal() {
      this.$emit('close')
    },
  },
}
</script>

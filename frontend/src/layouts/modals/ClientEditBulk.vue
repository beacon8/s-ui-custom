<template>
  <a-modal
    :open="visible"
    :width="800"
    :title="$t('actions.editbulk')"
    @cancel="closeModal"
  >
    <a-card
      size="small"
      :title="$t('actions.action')"
      style="margin-bottom: 16px;"
    >
      <a-form layout="vertical">
        <a-row
          :gutter="[12, 12]"
          style="margin-bottom: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('actions.action')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="actionMode"
                :options="actionModes"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
                @change="onActionChange"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-if="actionMode === 'change_limits'"
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('bulk.addDays')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="editData.addDays"
                :addon-after="$t('date.d')"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('bulk.addVolume')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="editData.addVolume"
                :addon-after="$t('stats.GB')"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-space>
              <a-switch v-model:checked="editData.enable" />
              <span>{{ $t('enable') }}</span>
            </a-space>
          </a-col>
        </a-row>
        <a-row
          v-if="actionMode === 'add_inbounds' || actionMode === 'remove_inbounds'"
          :gutter="[12, 12]"
        >
          <a-col
            :xs="24"
            :sm="16"
          >
            <a-form-item
              :label="$t('client.inboundTags')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="editData.inboundTags"
                :options="inboundTags"
                :field-names="{ label: 'title', value: 'value' }"
                mode="multiple"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-card>

    <!-- Section two: Clients (like init users in Inbound modal) -->
    <Users
      :clients="clients"
      :data="selectedClients"
    />

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
        :disabled="selectedClients.model != 'all' && selectedClients.values.length == 0"
        @click="saveChanges"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import Users from '@/components/Users.vue'
import { i18n } from '@/locales'
import Data from '@/store/modules/data';
import { Client } from '@/types/clients';
import type { PropType } from 'vue'

// An inbound the clients can be attached to, as the parent lists them.
interface InboundTag {
  title: string
  value: number
}

export default {
  components: { Users },
  props: {
    visible: { type: Boolean, required: true },
    clients: { type: Array as PropType<Client[]>, required: true },
    inboundTags: { type: Array as PropType<InboundTag[]>, required: true },
  },
  emits: ['close'],
  data() {
    return {
      loading: false,
      actionMode: 'change_limits',
      actionModes: [
        { title: i18n.global.t('bulk.changeLimits'), value: 'change_limits' },
        { title: i18n.global.t('bulk.addInbounds'), value: 'add_inbounds' },
        { title: i18n.global.t('bulk.removeInbounds'), value: 'remove_inbounds' },
        { title: i18n.global.t('actions.delbulk'), value: 'delete_bulk' },
      ],
      editData: {
        enable: true,
        addDays: 0,
        addVolume: 0,
        inboundTags: [] as number[],
      },
      selectedClients: {
        model: 'none',
        // Group names when picking by group, client ids when picking clients.
        values: [] as (string | number)[],
      },
    }
  },
  watch: {
    visible(newVal) {
      if (newVal) {
        this.actionMode = 'change_limits'
        this.editData = { enable: true, addDays: 0, addVolume: 0, inboundTags: [] }
        this.selectedClients = { model: 'none', values: [] }
      }
    },
  },
  methods: {
    onActionChange() {
      this.editData.inboundTags = []
    },
    closeModal() {
      this.$emit('close')
    },
    getTargetClients(): Client[] {
      const clients = this.clients ?? []
      switch (this.selectedClients.model) {
        case 'all':
          return clients
        case 'group':
          return clients
            .filter(c => this.selectedClients.values.includes(c.group))
        case 'client':
          return clients.filter(c => c.id != undefined && this.selectedClients.values.includes(c.id))
        default:
          return []
      }
    },
    async saveChanges() {
      this.loading = true
      // Deep copies. getTargetClients returns the store's own objects, so the
      // edits below were applied to the live list the moment they were typed:
      // a cancelled dialog, or a save that failed, left the table showing
      // values that were never written.
      const targetClients: Client[] = JSON.parse(JSON.stringify(this.getTargetClients()))
      switch (this.actionMode) {
        case 'change_limits':
          targetClients.forEach((c: Client) => {
            if (this.editData.addVolume != 0 && c.volume > 0)
              c.volume += this.editData.addVolume*(1024 ** 3)
            if (this.editData.addDays != 0 && c.expiry > 0)
              c.expiry += this.editData.addDays*(24 * 60 * 60)
            if (this.editData.enable)
              c.enable = (c.volume == 0 || c.up + c.down < c.volume) && (c.expiry == 0 || c.expiry > Date.now()/1000)
          })
          break
        case 'add_inbounds':
          targetClients.forEach((c: Client) => {
            this.editData.inboundTags.forEach((t: number) => {
              if (!c.inbounds.includes(t)) {
                c.inbounds.push(t)
              }
            })
            c.inbounds = c.inbounds.sort()
          })
          break
        case 'remove_inbounds':
          targetClients.forEach((c: Client) => {
            c.inbounds = c.inbounds.filter((i: number) => !this.editData.inboundTags.includes(i))
          })
          break
        case 'delete_bulk': {
          const success = await Data().save("clients", "delbulk", targetClients.map((c: Client) => c.id))
          if (success) this.closeModal()
          this.loading = false
          return
        }
      }
      const success = await Data().save("clients", 'editbulk', targetClients)
      if (success) this.closeModal()
      this.loading = false
    },
  },
}
</script>

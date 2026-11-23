<template>
  <a-modal
    :open="visible"
    :width="800"
    :title="$t('actions.addbulk')"
    @cancel="closeModal"
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
            :label="$t('count')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="count"
              :min="1"
              :max="100"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        :gutter="[12, 12]"
        style="margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('client.name')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="namePattern"
              mode="tags"
              style="width: 100%"
            >
              <a-select-option
                v-for="p in patterns"
                :key="p.value"
                :value="p.value"
              >
                {{ p.title }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        :gutter="[12, 12]"
        style="margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('client.desc')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="descPattern"
              mode="tags"
              style="width: 100%"
            >
              <a-select-option
                v-for="p in patterns"
                :key="p.value"
                :value="p.value"
              >
                {{ p.title }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
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
            :label="$t('client.group')"
            style="margin-bottom: 0;"
          >
            <a-auto-complete
              v-model:value="bulkData.group"
              :options="groups.map(g => ({ value: g }))"
              style="width: 100%"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('stats.volume')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="bulkData.Volume"
              :min="0"
              addon-after="GiB"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="!(bulkData.delayStart && !bulkData.autoReset)"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <DatePick
            :expiry="bulkData.expiry"
            :placeholder="$t('unlimited')"
            @submit="setDate"
          />
        </a-col>
      </a-row>
      <a-row
        :gutter="[12, 12]"
        align="middle"
        style="margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="bulkData.delayStart" />
            <span>{{ $t('client.delayStart') }}</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="bulkData.autoReset" />
            <span>{{ $t('client.autoReset') }}</span>
          </a-space>
        </a-col>
        <a-col
          v-if="bulkData.autoReset || bulkData.delayStart"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('client.resetDays')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="bulkData.resetDays"
              :min="1"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row :gutter="[12, 12]">
        <a-col :span="24">
          <a-form-item
            :label="$t('client.inboundTags')"
            style="margin-bottom: 0;"
          >
            <a-flex
              gap="small"
              align="center"
            >
              <a-select
                v-model:value="bulkData.clientInbounds"
                :options="inboundTags"
                :field-names="{ label: 'title', value: 'value' }"
                mode="multiple"
                style="width: 100%"
              />
              <a-tooltip :title="$t('all')">
                <a-button
                  :aria-label="$t('all')"
                  type="text"
                  @click="setAllInbounds"
                >
                  <template #icon>
                    <AppIcon name="set-all" />
                  </template>
                </a-button>
              </a-tooltip>
            </a-flex>
          </a-form-item>
        </a-col>
      </a-row>
    </a-form>
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
import DatePick from '@/components/DateTime.vue'
import { push } from 'notivue'
import RandomUtil from '@/plugins/randomUtil'
import { Client, createClient, randomConfigs } from '@/types/clients'
import { i18n } from '@/locales'
import Data from '@/store/modules/data'
import type { PropType } from 'vue'

// An inbound the clients can be attached to, as the parent lists them.
interface InboundTag {
  title: string
  value: number
}

// A name or description is built from literal text and generator tokens picked
// out of `patterns`.
interface PatternToken {
  title: string
  value: string
}
type PatternPart = string | PatternToken

export default {
  components: { DatePick },
  props: {
    visible: { type: Boolean, required: true },
    inboundTags: { type: Array as PropType<InboundTag[]>, required: true },
    groups: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['close'],
  data() {
    return {
      count: 1,
      clients: <Client[]>[],
      bulkData: {
        name: <PatternPart[]>[],
        desc: <PatternPart[]>[],
        group: '',
        clientInbounds: <number[]>[],
        expiry: 0,
        Volume: 0,
        delayStart: false,
        autoReset: false,
        resetDays: 0,
      },
      patterns: [
        { title: i18n.global.t("bulk.random"), value: "random" },
        { title: i18n.global.t("bulk.order"), value: "order" },
      ],
      loading: false,
    }
  },
  computed: {
    namePattern: {
      get(): string[] {
        return this.bulkData.name.map(p => typeof p === 'object' ? p.value : p)
      },
      set(vals: string[]) {
        this.bulkData.name = vals.map(v => this.patterns.find(p => p.value === v) ?? v)
      },
    },
    descPattern: {
      get(): string[] {
        return this.bulkData.desc.map(p => typeof p === 'object' ? p.value : p)
      },
      set(vals: string[]) {
        this.bulkData.desc = vals.map(v => this.patterns.find(p => p.value === v) ?? v)
      },
    },
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.resetData()
      }
    },
  },
  methods: {
    resetData() {
      this.count = 1
      this.clients = []
      this.bulkData = {
        name: [this.patterns[1], "-", this.patterns[0]],
        desc: [],
        group: '',
        clientInbounds: [],
        expiry: 0,
        Volume: 0,
        delayStart: false,
        autoReset: false,
        resetDays: 0,
      }
    },
    closeModal() {
      this.$emit('close')
    },
    async saveChanges() {
      if (!this.$props.visible) return
      if (this.bulkData.name.findIndex(n => typeof(n) == 'object') == -1) {
        push.error(i18n.global.t('error.dplData'))
        return
      }
      this.clients = []
      this.loading = true
      for(let i=0;i<this.count;i++){
        const name = this.genByPattern(this.bulkData.name, i)
        this.clients.push(createClient({
          enable: true,
          name: name,
          config: randomConfigs(name),
          inbounds: this.bulkData.clientInbounds.length > 0 ? this.bulkData.clientInbounds.sort() : [],
          links: [],
          volume: this.bulkData.Volume*(1024 ** 3),
          expiry: (this.bulkData.delayStart && !this.bulkData.autoReset) ? 0 : this.bulkData.expiry,
          up: 0,
          down: 0,
          desc: this.genByPattern(this.bulkData.desc, i),
          group: this.bulkData.group,
          delayStart: this.bulkData.delayStart,
          autoReset: this.bulkData.autoReset,
          resetDays: this.bulkData.resetDays,
        }))
      }
      // Check duplicate names
      const isDuplicateName = Data().checkBulkClientNames(this.clients.map(c => c.name))
      if (isDuplicateName) return
      const success = await Data().save("clients", "addbulk", this.clients)
      if (success) this.closeModal()
      this.loading = false
    },
    genByPattern(pattern: PatternPart[], order :number){
      if (pattern.length == 0) return RandomUtil.randomSeq(8)
      let result = ''
      pattern.forEach(p => {
        switch(typeof p){
          case 'object':
            switch(p.value){
              case "random":
                result += RandomUtil.randomSeq(8)
                break
              case "order":
                result += order+1
            }
            break
          default:
            result += p
        }
      })
      return result
    },
    setDate(v:number){
      this.bulkData.expiry = v
    },
    setAllInbounds(){
      this.bulkData.clientInbounds = this.inboundTags.map(i => i.value).sort()
    }
  },
}

</script>

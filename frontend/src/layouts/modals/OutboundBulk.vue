<template>
  <a-modal
    :open="visible"
    :width="800"
    :title="$t('actions.addbulk') + ' ' + $t('objects.outbound')"
    @cancel="closeModal"
  >
    <div style="max-height: 70vh; overflow-y: auto;">
      <a-form
        v-if="outbounds.length == 0"
        layout="vertical"
      >
        <a-row :gutter="[12, 12]">
          <a-col :span="24">
            <a-form-item
              :label="$t('client.sub')"
              style="margin-bottom: 0;"
            >
              <a-input
                v-model:value="link"
                dir="ltr"
                placeholder="http[s]://<domain>[:]<port>/<path>"
              />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-checkbox v-model:checked="addUrlTest">
              {{ $t('out.addUrlTest') }}
            </a-checkbox>
          </a-col>
          <a-col
            :span="24"
            style="text-align: center;"
          >
            <a-button
              type="primary"
              :loading="loading"
              @click="linkCheck"
            >
              {{ $t('submit') }}
            </a-button>
          </a-col>
        </a-row>
      </a-form>
      <a-table
        v-if="outbounds.length > 0"
        :data-source="outbounds"
        :loading="loading"
        :pagination="false"
        size="small"
        :row-key="(_record: Outbound, index?: number) => index ?? 0"
        :columns="[
          { title: '', key: 'check', width: 48 },
          { title: $t('type'), key: 'type', dataIndex: 'type' },
          { title: $t('objects.tag'), key: 'tag', dataIndex: 'tag' },
          { title: $t('out.addr'), key: 'server', dataIndex: 'server' },
          { title: $t('objects.tls'), key: 'tls', dataIndex: 'tls' },
        ]"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'check'">
            <AppIcon
              v-if="outChecks[index] == 1"
              color="success"
              name="check"
            />
            <AppIcon
              v-else-if="outChecks[index] == 2"
              color="error"
              name="close"
            />
            <a-spin
              v-else-if="outChecks[index] == 3"
              size="small"
            />
            <AppIcon
              v-else
              name="help"
            />
          </template>
          <template v-else-if="column.key === 'type'">
            {{ record.type }}
          </template>
          <template v-else-if="column.key === 'tag'">
            {{ record.tag }}
          </template>
          <template v-else-if="column.key === 'tls'">
            {{ Object.hasOwn(record, 'tls') ? $t(record.tls?.enabled ? 'enable' : 'disable') : '-' }}
          </template>
          <template v-else-if="column.key === 'server'">
            {{ record.server }}{{ record.server_port ? ':' + record.server_port : '' }}
          </template>
        </template>
      </a-table>
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
        :disabled="outbounds.length == 0"
        @click="saveChanges"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import HttpUtils from '@/plugins/httputil'
import RandomUtil from '@/plugins/randomUtil';
import Data from '@/store/modules/data'
import { createOutbound, Outbound } from '@/types/outbounds'
import type { PropType } from 'vue'

export default {
  props: {
    visible: { type: Boolean, required: true },
    outboundTags: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['close'],
  data() {
    return {
      loading: false,
      link: "",
      outbounds: <Outbound[]>[],
      outChecks: <number[]>[],
      addUrlTest: false,
    }
  },
  computed: {
    newOutboundTags(): string[] {
      return this.outbounds.map((o:Outbound) => o.tag)
    }
  },
  watch: {
    visible(v) {
      if (v) {
        this.resetData()
      }
    },
  },
  methods: {
    resetData() {
      this.outbounds = []
      this.outChecks = []
      this.link = ""
      this.addUrlTest = false
      this.loading = false
    },
    closeModal() {
      this.resetData()
      this.$emit('close')
    },
    async linkCheck() {
      this.loading = true
      this.outbounds = []
      const msg = await HttpUtils.post<Outbound[]>('api/subConvert', { link: this.link })
      if (msg.success) {
        if (msg.obj?.length>0) {
          msg.obj.forEach((o:Outbound, index:number) => {
            if (this.newOutboundTags.includes(o.tag)) o.tag = o.tag + "-" + (index+1)
            this.outbounds.push(createOutbound(o.type, o))
            this.outChecks.push(0)
          })
          if (this.addUrlTest) {
            const urlTestTsg = "urltest-" + RandomUtil.randomSeq(3)
            this.outbounds.push(createOutbound("urltest", {
              tag: urlTestTsg,
              outbounds: this.outbounds.map((o:Outbound) => o.tag),
              interrupt_exist_connections: false,
              interval: "30s"
            }))
          }
        }
      }
      this.loading = false
    },
    async saveChanges() {
      if (!this.$props.visible) return
      // check duplicate tag
      this.outbounds.forEach((o:Outbound, index:number) => {
        const isDuplicatedTag = Data().checkTag("outbound",0, o.tag)
        this.outChecks[index] = isDuplicatedTag ? 2 : 0
      })

      // save data
      this.loading = true
      this.outbounds.forEach(async (o:Outbound, index:number) => {
        if (this.outChecks[index] == 2) return
        this.outChecks[index] = 3
        const success = await Data().save("outbounds",  "new", o)
        if (success) this.outChecks[index] = 1
        else this.outChecks[index] = 2
      })
      this.loading = false
    }
  },
}
</script>

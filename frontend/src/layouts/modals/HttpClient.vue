<template>
  <a-modal
    :open="visible"
    :width="800"
    :title="$t('actions.' + title) + ' ' + $t('basic.httpClient.title')"
    @cancel="closeModal"
  >
    <a-form layout="vertical">
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="8"
        >
          <!-- The tag is how everything else points at this client. -->
          <a-form-item
            :label="$t('objects.tag')"
            :validate-status="tagError.length > 0 ? 'error' : undefined"
            :help="tagError || undefined"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="client.tag" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="8"
        >
          <a-form-item
            :label="$t('basic.httpClient.version')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="client.version"
              :options="versions"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
              @change="changeVersion($event as any)"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="8"
        >
          <a-form-item
            :label="$t('basic.httpClient.engine')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="client.engine"
              :options="engines"
              :field-names="{ label: 'title', value: 'value' }"
              allow-clear
              style="width: 100%"
              @clear="delete client.engine"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="client.version != undefined && client.version > 0"
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-space>
            <a-switch v-model:checked="client.disable_version_fallback" />
            <span>{{ $t('basic.httpClient.noFallback') }}</span>
          </a-space>
        </a-col>
      </a-row>

      <QuicFields
        v-if="fieldGroup != 'none'"
        :data="client"
        :quic="fieldGroup == 'quic'"
      />
      <Headers :data="client" />
      <Dial :dial="client" />
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
        :disabled="tagError.length > 0"
        @click="saveChanges"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import Dial from '@/components/Dial.vue'
import Headers from '@/components/Headers.vue'
import QuicFields from '@/components/QuicFields.vue'
import RandomUtil from '@/plugins/randomUtil'
import { i18n } from '@/locales'
import type { PropType } from 'vue'
import { HttpClient, HttpVersion, QuicFields as QuicFieldNames, createHttpClient, httpFieldGroup } from '@/types/httpClient'

// The tuning fields belong to one transport or the other, so switching version
// drops the ones that no longer apply.
const versionKeys: (keyof QuicFieldNames)[] = [
  'idle_timeout',
  'keep_alive_period',
  'stream_receive_window',
  'connection_receive_window',
  'max_concurrent_streams',
  'initial_packet_size',
  'disable_path_mtu_discovery',
]

export default {
  components: { Dial, Headers, QuicFields },
  props: {
    visible: { type: Boolean, required: true },
    data: { type: String, required: true },
    index: { type: Number, required: true },
    tags: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['close', 'save'],
  data() {
    return {
      title: "add",
      loading: false,
      client: <HttpClient>createHttpClient(''),
      engines: [
        { title: 'Go', value: 'go' },
        { title: 'Apple', value: 'apple' },
      ],
      versions: [
        { title: 'Auto', value: 0 },
        { title: 'HTTP/1.1', value: 1 },
        { title: 'HTTP/2', value: 2 },
        { title: 'HTTP/3', value: 3 },
      ],
    }
  },
  computed: {
    fieldGroup(): string {
      return httpFieldGroup(this.client.version)
    },
    tagError(): string {
      const tag = this.client.tag?.trim() ?? ''
      if (tag.length == 0) return i18n.global.t('error.invalidData') + ": " + i18n.global.t('objects.tag')
      const others = <string[]>(this.$props.tags ?? [])
      const taken = this.$props.index == -1
        ? others
        : others.filter((_, i) => i != this.$props.index)
      if (taken.includes(tag)) return i18n.global.t('error.dplData') + ": " + i18n.global.t('objects.tag')
      return ''
    },
  },
  watch: {
    visible(newValue) {
      if (newValue) this.updateData()
    },
  },
  methods: {
    updateData() {
      if (this.$props.index != -1) {
        this.title = "edit"
        this.client = <HttpClient>JSON.parse(this.$props.data)
      } else {
        this.title = "add"
        this.client = createHttpClient("hc-" + RandomUtil.randomSeq(3))
      }
    },
    changeVersion(version: HttpVersion) {
      const group = httpFieldGroup(version)
      if (group == 'none') {
        versionKeys.forEach(key => delete this.client[key])
        return
      }
      if (group == 'http2') {
        delete this.client.initial_packet_size
        delete this.client.disable_path_mtu_discovery
      }
    },
    closeModal() {
      this.$emit('close')
    },
    saveChanges() {
      this.loading = true
      this.$emit('save', this.client)
      this.loading = false
    },
  },
}
</script>

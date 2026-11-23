<template>
  <a-modal
    :open="visible"
    :width="800"
    @cancel="closeModal"
  >
    <template #title>
      <a-flex
        justify="space-between"
        align="center"
        style="padding-right: 24px;"
      >
        <span>{{ $t('actions.' + title) + " " + $t('objects.ruleset') }}</span>
        <DocLink section="ruleset" />
      </a-flex>
    </template>
    <a-form layout="vertical">
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('type')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="rule_set.type"
              :options="[{ title: $t('ruleset.local'), value: 'local' }, { title: $t('ruleset.remote'), value: 'remote' }]"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
              @change="updateType($event as any)"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('objects.tag')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="rule_set.tag" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('ruleset.format')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="rule_set.format"
              style="width: 100%"
            >
              <a-select-option
                v-for="fmt in ['source', 'binary']"
                :key="fmt"
                :value="fmt"
              >
                {{ fmt }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="rule_set.type == 'local'"
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col :span="24">
          <a-form-item
            :label="$t('transport.path')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="rule_set.path" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-else
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col :span="24">
          <a-form-item
            label="URL"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="rule_set.url" />
          </a-form-item>
        </a-col>
        <!-- Download over a shared client, or over a single outbound. The
             two are alternatives, so choosing one clears the other. -->
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('basic.httpClient.title')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="http_client"
              :not-found-content="$t('basic.httpClient.none')"
              allow-clear
              style="width: 100%"
            >
              <a-select-option
                v-for="hc in httpClients"
                :key="hc"
                :value="hc"
              >
                {{ hc }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('objects.outbound')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="download_detour"
              allow-clear
              style="width: 100%"
            >
              <a-select-option
                v-for="tag in outTags"
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
          :md="8"
        >
          <a-form-item
            :label="$t('ruleset.interval')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="update_intervals"
              :addon-after="$t('date.d')"
              :min="0"
              style="width: 100%"
            />
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
import RandomUtil from '@/plugins/randomUtil'
import { ruleset } from '@/types/rules'
import DocLink from '@/components/DocLink.vue'
import { downloadHttpClient, httpClientTags, refDetour, refTag } from '@/plugins/httpClient'
import type { PropType } from 'vue'
export default {
  components: { DocLink },
  props: {
    visible: { type: Boolean, required: true },
    data: { type: String, required: true },
    index: { type: Number, required: true },
    outTags: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['close', 'save'],
  data() {
    return {
      title: "add",
      loading: false,
      rule_set: <ruleset>{},
    }
  },
  computed: {
    httpClients(): string[] {
      return httpClientTags()
    },
    // Naming a shared client replaces any inline options, since http_client
    // holds one or the other.
    http_client: {
      get() { return refTag(this.rule_set.http_client) },
      set(v: string | undefined) {
        if (v) this.rule_set.http_client = v
        else delete this.rule_set.http_client
      }
    },
    // The inline form, carrying just the outbound to download over. An empty
    // choice drops http_client entirely rather than leaving an empty transport.
    download_detour: {
      get() { return refDetour(this.rule_set.http_client) },
      set(v: string | undefined) {
        const httpClient = v ? downloadHttpClient(v) : undefined
        if (httpClient) this.rule_set.http_client = httpClient
        else delete this.rule_set.http_client
      }
    },
    update_intervals: {
      get() { return this.rule_set.update_interval != undefined ? parseInt(this.rule_set.update_interval.replace('d','')) : 0 },
      set(v:number) { this.rule_set.update_interval = v>0 ?  v + 'd' : undefined }
    },
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.updateData()
      }
    },
  },
  methods: {
    updateData() {
      if (this.$props.index != -1) {
        this.title = "edit"
        this.rule_set = <ruleset>JSON.parse(this.$props.data)
      }
      else {
        this.title = "add"
        this.rule_set = <ruleset>{type: 'local', tag: "rs-" + RandomUtil.randomSeq(3), format: 'binary'}
      }
    },
    updateType(t:string) {
      if (t == 'local') {
        delete this.rule_set.url
        delete this.rule_set.http_client
        delete this.rule_set.update_interval
      } else {
        delete this.rule_set.path
      }
    },
    closeModal() {
      this.$emit('close')
    },
    saveChanges() {
      this.loading = true
      this.$emit('save', this.rule_set)
      this.loading = false
    }
  },
}
</script>
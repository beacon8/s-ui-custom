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
        <span>{{ $t('actions.' + title) + " " + $t('tls.provider.title') }}</span>
        <DocLink section="tls" />
      </a-flex>
    </template>
    <a-form layout="vertical">
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('type')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="provider.type"
              :options="types"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
              @change="changeType($event as CertProviderType)"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
        >
          <!-- The tag is how a TLS config points at this provider, so it has
               to be set and unique. -->
          <a-form-item
            :label="$t('objects.tag')"
            :validate-status="tagError.length > 0 ? 'error' : undefined"
            :help="tagError || undefined"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="provider.tag" />
          </a-form-item>
        </a-col>
      </a-row>

      <AcmeVue
        v-if="provider.type == 'acme'"
        :data="provider"
      />
      <OriginCaVue
        v-else-if="provider.type == 'cloudflare-origin-ca'"
        :data="provider"
      />
      <a-row
        v-else-if="provider.type == 'tailscale'"
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col :span="24">
          <a-form-item
            :label="$t('tls.provider.endpoint')"
            :extra="$t('tls.provider.endpointHint')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="provider.endpoint"
              allow-clear
              @change="onEndpointChange"
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
        :disabled="tagError.length > 0"
        @click="saveChanges"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import AcmeVue from '@/components/tls/Acme.vue'
import OriginCaVue from '@/components/tls/OriginCa.vue'
import DocLink from '@/components/DocLink.vue'
import Data from '@/store/modules/data'
import RandomUtil from '@/plugins/randomUtil'
import { i18n } from '@/locales'
import { certProvider, CertProviderType, createCertProvider } from '@/types/tls'
import type { PropType } from 'vue'

export default {
  components: { AcmeVue, OriginCaVue, DocLink },
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
      provider: <certProvider>createCertProvider('acme'),
    }
  },
  computed: {
    // ACME is hidden on Windows, where it does not work (#1189), but an
    // existing one stays editable rather than silently changing type.
    types(): { title: string; value: CertProviderType }[] {
      const all: { title: string; value: CertProviderType }[] = [
        { title: 'ACME', value: 'acme' },
        { title: 'Tailscale', value: 'tailscale' },
        { title: 'Cloudflare Origin CA', value: 'cloudflare-origin-ca' },
      ]
      if (Data().os === 'windows' && this.provider.type !== 'acme') {
        return all.filter(t => t.value !== 'acme')
      }
      return all
    },
    tagError(): string {
      const tag = this.provider.tag?.trim() ?? ''
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
        this.provider = <certProvider>JSON.parse(this.$props.data)
      } else {
        this.title = "add"
        this.provider = createCertProvider('acme', "cert-" + RandomUtil.randomSeq(3))
      }
    },
    // Each type has its own option set, so the fields of the one being left
    // behind are dropped rather than carried over as unknown keys.
    changeType(type: CertProviderType) {
      this.provider = createCertProvider(type, this.provider.tag)
    },
    onEndpointChange() {
      if (this.provider.type == 'tailscale' && !this.provider.endpoint) {
        delete this.provider.endpoint
      }
    },
    closeModal() {
      this.$emit('close')
    },
    saveChanges() {
      this.loading = true
      this.$emit('save', this.provider)
      this.loading = false
    },
  },
}
</script>

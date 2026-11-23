<template>
  <div class="ech-section">
    <a-flex
      justify="space-between"
      align="center"
      :gap="12"
    >
      <span class="ech-title">ECH</span>
      <a-space>
        <a-switch v-model:checked="enabled" />
        <span>{{ $t('enable') }}</span>
      </a-space>
    </a-flex>

    <template v-if="enabled">
      <a-flex
        justify="space-between"
        align="center"
        :gap="12"
        wrap
      >
        <a-radio-group
          v-model:value="useEchPath"
          option-type="button"
          button-style="solid"
        >
          <a-radio-button
            :value="0"
            @click="delete ech.key"
          >
            {{ $t('tls.usePath') }}
          </a-radio-button>
          <a-radio-button
            :value="1"
            @click="delete ech.key_path"
          >
            {{ $t('tls.useText') }}
          </a-radio-button>
        </a-radio-group>
        <a-button
          :loading="loading"
          @click="genDialog = true"
        >
          <template #icon>
            <AppIcon name="key-star" />
          </template>
          {{ $t('actions.generate') }}
        </a-button>
      </a-flex>

      <a-form-item
        v-if="useEchPath == 0"
        :label="$t('tls.keyPath')"
      >
        <a-input v-model:value="ech.key_path" />
      </a-form-item>
      <a-form-item
        v-else
        :label="$t('tls.key')"
      >
        <a-textarea
          v-model:value="echKeyText"
          :rows="4"
        />
      </a-form-item>

      <a-form-item
        v-if="oTls.ech"
        :label="$t('tls.queryServerName')"
        :extra="hasConfig ? $t('tls.echQueryDisabled') : undefined"
      >
        <a-input
          v-model:value="oTls.ech.query_server_name"
          :disabled="hasConfig"
          placeholder="ech.example.com"
        />
      </a-form-item>

      <a-form-item :label="$t('tls.cert')">
        <a-textarea
          v-model:value="echConfigText"
          :rows="4"
        />
      </a-form-item>
    </template>

    <a-modal
      v-model:open="genDialog"
      :title="$t('tls.echPublicName')"
      :width="520"
    >
      <p style="margin-bottom: 12px;">
        {{ $t('tls.echPublicNameAbout') }}
      </p>
      <p style="margin-bottom: 16px; opacity: 0.65;">
        {{ $t('tls.echQueryAbout') }}
      </p>
      <a-form-item
        :label="$t('tls.echPublicName')"
        style="margin-bottom: 0;"
      >
        <a-input
          v-model:value="publicName"
          placeholder="public.example.com"
          dir="ltr"
          autofocus
          @keyup.enter="genECH"
        />
      </a-form-item>
      <template #footer>
        <a-button
          type="default"
          @click="genDialog = false"
        >
          {{ $t('actions.close') }}
        </a-button>
        <a-button
          type="primary"
          :loading="loading"
          :disabled="publicName.trim() == ''"
          @click="genECH"
        >
          {{ $t('actions.generate') }}
        </a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'
import { iTls as ITlsOptions, oTls as OTlsOptions } from '@/types/tls'
import { push } from 'notivue'

// Only the ECH fields this form reads or writes on the server half of a TLS
// config.
interface EchInTls {
  server_name?: string
  ech?: {
    enabled: boolean
    key?: string[]
    key_path?: string
  }
}

// The matching client half: it carries the published config rather than the
// key.
interface EchOutTls {
  ech?: {
    enabled?: boolean
    config?: string[]
    query_server_name?: string
  }
}

// The parent owns both objects and this component edits them in place, so they
// are models rather than plain props.
const iTlsModel = defineModel<ITlsOptions>('iTls', { required: true })
const oTlsModel = defineModel<OTlsOptions>('oTls', { required: true })

// Narrow once to the fields this form touches. The computeds return the same
// objects, so edits still reach the parent.
const iTls = computed(() => iTlsModel.value as unknown as EchInTls)
const oTls = computed(() => oTlsModel.value as unknown as EchOutTls)

const useEchPath = ref(iTls.value?.ech?.key ? 1 : 0)
const loading = ref(false)
// The cleartext outer SNI baked into the generated config. It is only an
// input to the generator, asked for when generating; sing-box keeps no field
// for it.
const genDialog = ref(false)
const publicName = ref('')

const ech = computed(() => iTls.value.ech as NonNullable<EchInTls['ech']>)
// The client half is written whenever the server half is, so it is present
// wherever this form is shown.
const oEch = computed(() => oTls.value.ech as NonNullable<EchOutTls['ech']>)

const enabled = computed({
  get: (): boolean => ech.value?.enabled ?? false,
  set: (v: boolean) => {
    iTls.value.ech = v ? { enabled: true } : undefined
    oTls.value.ech = v ? {} : undefined
  },
})

const echKeyText = computed({
  get: (): string => ech.value?.key ? ech.value.key.join('\n') : '',
  set: (newValue: string) => { ech.value.key = newValue.split('\n') },
})

// A client with a config uses it and never looks one up, so the query name
// only matters while the config is empty.
const hasConfig = computed((): boolean => (oTls.value.ech?.config ?? []).some(l => l.trim() != ''))

const echConfigText = computed({
  get: (): string => oTls.value.ech?.config ? oTls.value.ech.config.join('\n') : '',
  set: (newValue: string) => { oEch.value.config = newValue.split('\n') },
})

async function genECH() {
  // Was the real server_name, or the literal '' without one, so the outer SNI
  // gave away the name ECH is meant to hide
  const name = publicName.value.trim()
  if (name == '') return
  loading.value = true
  // The endpoint answers with the generated PEM blocks, one line per entry.
  const msg = await HttpUtils.get<string[]>('api/keypairs', {
    k: "ech",
    o: name
  })
  loading.value = false
  if (msg.success && iTls.value.ech && oTls.value.ech) {
    genDialog.value = false
    iTls.value.ech.key_path = undefined
    useEchPath.value = 1
    if (msg.obj.length > 0) {
      let config = <string[]>[]
      let key = <string[]>[]
      let isConfig = false
      let isKey = false

      msg.obj.forEach((line: string) => {
        if (line === "-----BEGIN ECH CONFIGS-----") {
          isConfig = true
          isKey = false
          config.push(line)
        } else if (line === "-----END ECH CONFIGS-----") {
          isConfig = false
          config.push(line)
        } else if (line === "-----BEGIN ECH KEYS-----") {
          isKey = true
          isConfig = false
          key.push(line)
        } else if (line === "-----END ECH KEYS-----") {
          isKey = false
          key.push(line)
        } else if (isConfig) {
          config.push(line)
        } else if (isKey) {
          key.push(line)
        }
      })
      iTls.value.ech.key = key ?? undefined
      oTls.value.ech.config = config ?? undefined
      // Has no effect next to the config just written
      delete oTls.value.ech.query_server_name

    } else {
      push.error({
        message: i18n.global.t('error') + ": " + msg.obj
      })
    }
  }
}
</script>

<style scoped>
.ech-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ech-title {
  font-size: 13px;
  font-weight: 600;
  line-height: 22px;
  opacity: 0.75;
}

.ech-section :deep(.ant-form-item) {
  margin-bottom: 0;
}
</style>

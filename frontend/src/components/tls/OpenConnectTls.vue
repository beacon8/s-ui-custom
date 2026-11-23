<template>
  <a-card
    size="small"
    :title="$t('objects.tls')"
  >
    <!-- OpenConnect defines its own TLS options too, and names the trust anchor
         after its role rather than reusing the bare `certificate` sing-box's own
         TLS config uses. Left unset it trusts the system store, so nothing here
         is required. -->
    <a-flex
      vertical
      :gap="12"
    >
      <a-row :gutter="[12, 12]">
        <a-col :span="24">
          <a-radio-group
            v-model:value="usePath"
            option-type="button"
            button-style="solid"
          >
            <a-radio-button
              :value="0"
              @click="clearText"
            >
              {{ $t('tls.usePath') }}
            </a-radio-button>
            <a-radio-button
              :value="1"
              @click="clearPaths"
            >
              {{ $t('tls.useText') }}
            </a-radio-button>
          </a-radio-group>
        </a-col>
      </a-row>

      <a-row
        v-if="usePath == 0"
        :gutter="[12, 12]"
      >
        <a-col :span="24">
          <a-form-item
            :label="$t('types.openconnect.tls.caPath')"
            :extra="$t('types.openconnect.tls.caHint')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="tls.certificate_authority_path" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-else
        :gutter="[12, 12]"
      >
        <a-col :span="24">
          <a-form-item
            :label="$t('types.openconnect.tls.ca')"
            :extra="$t('types.openconnect.tls.caHint')"
            style="margin-bottom: 0;"
          >
            <a-textarea
              v-model:value="caText"
              :rows="4"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <!-- Mutual TLS: some deployments issue a client certificate instead of, or
           alongside, a password. -->
      <template v-if="optionMutual">
        <a-row
          v-if="usePath == 0"
          :gutter="[12, 12]"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('tls.clientCertPath')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="tls.client_certificate_path" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('tls.clientKeyPath')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="tls.client_key_path" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-else
          :gutter="[12, 12]"
        >
          <a-col :span="24">
            <a-form-item
              :label="$t('tls.clientCert')"
              style="margin-bottom: 0;"
            >
              <a-textarea
                v-model:value="clientCertText"
                :rows="4"
              />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-form-item
              :label="$t('tls.clientKey')"
              style="margin-bottom: 0;"
            >
              <a-textarea
                v-model:value="clientKeyText"
                :rows="4"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          :gutter="[12, 12]"
          justify="space-between"
          align="bottom"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('types.openconnect.tls.clientKeyPassword')"
              style="margin-bottom: 0;"
            >
              <a-input-password v-model:value="tls.client_key_password" />
            </a-form-item>
          </a-col>
          <!-- Self-signed, so the server has to be given this certificate as
               the CA it trusts for clients. -->
          <a-col flex="none">
            <a-tooltip :title="$t('actions.generate')">
              <a-button
                :aria-label="$t('actions.generate')"
                :loading="loading"
                @click="genClientCert"
              >
                <template #icon>
                  <AppIcon name="key-star" />
                </template>
              </a-button>
            </a-tooltip>
          </a-col>
        </a-row>
      </template>

      <a-row
        v-if="optionFingerprint"
        :gutter="[12, 12]"
      >
        <a-col :span="24">
          <a-form-item
            :label="$t('types.openvpn.tls.peerFingerprint')"
            :extra="$t('types.openconnect.tls.peerFingerprintHint')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="peerFingerprint"
              mode="tags"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <a-row
        v-if="optionServerName"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.openconnect.tls.serverName')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="tls.server_name" />
          </a-form-item>
        </a-col>
      </a-row>

      <a-row
        v-if="optionVerify"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-space>
            <a-switch v-model:checked="tls.insecure" />
            <span>{{ $t('tls.insecure') }}</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-space>
            <a-switch v-model:checked="tls.system_trust_disabled" />
            <span>{{ $t('types.openconnect.tls.systemTrustDisabled') }}</span>
          </a-space>
        </a-col>
      </a-row>

      <a-flex justify="end">
        <a-popover
          v-model:open="menu"
          trigger="click"
          placement="left"
        >
          <template #content>
            <a-flex
              vertical
              :gap="8"
            >
              <a-space>
                <a-switch v-model:checked="optionMutual" />
                <span>{{ $t('tls.mutual') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionFingerprint" />
                <span>{{ $t('types.openvpn.tls.peerFingerprint') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionServerName" />
                <span>{{ $t('types.openconnect.tls.serverName') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionVerify" />
                <span>{{ $t('tls.insecure') }}</span>
              </a-space>
            </a-flex>
          </template>
          <a-button>
            {{ $t('tls.options') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
</template>

<script lang="ts">
import { PropType } from 'vue'
import HttpUtils from '@/plugins/httputil'
import { push } from 'notivue'
import { i18n } from '@/locales'
import { OpenConnectTls } from '@/types/endpoints'

// Material that sing-box takes either as a file path or as the PEM itself.
const pathFields = ['certificate_authority_path', 'client_certificate_path', 'client_key_path'] as const
const textFields = ['certificate_authority', 'client_certificate', 'client_key'] as const

type TextField = typeof textFields[number]

// Everything a mutual-authentication setup carries, so the option reads as on
// when the config already holds any of it.
const mutualFields = ['client_certificate', 'client_certificate_path', 'client_key', 'client_key_path', 'client_key_password'] as const

// The endpoint that owns the TLS block. OpenConnect defines its own TLS
// options, so they sit on the endpoint rather than in a panel TLS config.
interface TlsHolder {
  tag?: string
  username?: string
  tls?: OpenConnectTls
}

// The keypair endpoint returns the private key and the certificate as one run
// of lines, so each block is taken from its own BEGIN/END markers.
function pemBlock(lines: string[], type: string): string[] {
  const start = lines.indexOf('-----BEGIN ' + type + '-----')
  const end = lines.indexOf('-----END ' + type + '-----')
  if (start == -1 || end < start) return []
  return lines.slice(start, end + 1)
}

export default {
  props: {
    // The modal holds the endpoint union and picks this form with a v-if on
    // the type, which the template checker cannot follow, so it is narrowed by
    // the computed below.
    data: { type: Object as PropType<object>, required: true }
  },
  data() {
    return {
      menu: false,
      loading: false,
      usePath: 0,
    }
  },
  computed: {
    // Narrow once. The computed returns the same object, so edits reach the
    // parent.
    holder(): TlsHolder { return <TlsHolder>this.$props.data },
    tls(): OpenConnectTls {
      return <OpenConnectTls>this.holder.tls
    },
    caText: {
      get(): string { return this.joined('certificate_authority') },
      set(v: string) { this.split('certificate_authority', v) }
    },
    clientCertText: {
      get(): string { return this.joined('client_certificate') },
      set(v: string) { this.split('client_certificate', v) }
    },
    clientKeyText: {
      get(): string { return this.joined('client_key') },
      set(v: string) { this.split('client_key', v) }
    },
    peerFingerprint: {
      get(): string[] {
        // sing-box takes a single string or a list here, so a config written
        // by hand can carry either.
        const value: string | string[] | undefined = this.tls.peer_fingerprint
        if (value == undefined) return []
        return Array.isArray(value) ? value : [value]
      },
      set(v: string[]) { this.tls.peer_fingerprint = v ?? [] }
    },
    // An option is on when the config already carries it, so opening an
    // existing endpoint shows what it actually has.
    optionMutual: {
      get(): boolean {
        return mutualFields
          .some(f => this.tls[f] != undefined)
      },
      set(v: boolean) {
        if (v) {
          // Switching an option on has to leave something behind, or the
          // section it reveals would be hidden again by its own getter. The
          // empty placeholder is dropped when the endpoint is saved.
          if (this.usePath == 0) this.tls.client_certificate_path = ''
          else this.tls.client_certificate = []
          return
        }
        for (const field of mutualFields) {
          delete this.tls[field]
        }
      }
    },
    optionFingerprint: {
      get(): boolean { return this.tls.peer_fingerprint != undefined },
      set(v: boolean) {
        if (v) this.tls.peer_fingerprint = []
        else delete this.tls.peer_fingerprint
      }
    },
    optionServerName: {
      get(): boolean { return this.tls.server_name != undefined },
      set(v: boolean) {
        if (v) this.tls.server_name = ''
        else delete this.tls.server_name
      }
    },
    optionVerify: {
      get(): boolean { return this.tls.insecure != undefined || this.tls.system_trust_disabled != undefined },
      set(v: boolean) {
        if (v) {
          this.tls.insecure = false
        } else {
          delete this.tls.insecure
          delete this.tls.system_trust_disabled
        }
      }
    },
  },
  created() {
    // Every field below writes into data.tls, so it has to exist first. An
    // empty one is dropped again when the endpoint is saved.
    if (!this.holder.tls) this.holder.tls = {}
    const tls = <OpenConnectTls>this.holder.tls
    this.usePath = textFields.some(f => tls[f] != undefined) ? 1 : 0
  },
  methods: {
    joined(field: TextField): string {
      return this.tls[field]?.join('\n') ?? ''
    },
    split(field: TextField, value: string) {
      if (value) this.tls[field] = value.split('\n')
      else delete this.tls[field]
    },
    clearText() {
      for (const field of textFields) delete this.tls[field]
    },
    clearPaths() {
      for (const field of pathFields) delete this.tls[field]
    },
    // Named after the username where there is one, since servers that map a
    // certificate to an account usually read it from the common name.
    async genClientCert() {
      this.loading = true
      const msg = await HttpUtils.get<string[]>('api/keypairs', { k: 'tls-client', o: this.holder.username || this.holder.tag || '' })
      this.loading = false
      if (!msg.success || msg.obj.length == 0) {
        push.error({ message: i18n.global.t('error') + ': ' + msg.obj })
        return
      }
      const key = pemBlock(msg.obj, 'PRIVATE KEY')
      const certificate = pemBlock(msg.obj, 'CERTIFICATE')
      if (key.length == 0 || certificate.length == 0) {
        push.error({ message: i18n.global.t('error') + ': ' + msg.obj })
        return
      }
      // The generated pair only exists here, so the form has to hold it as
      // text; there is no file on disk to point at. The key it comes with is
      // not encrypted, so a leftover key password would no longer match.
      this.clearPaths()
      this.usePath = 1
      this.tls.client_certificate = certificate
      this.tls.client_key = key
      delete this.tls.client_key_password
    },
  },
}
</script>

<template>
  <a-card
    size="small"
    :title="$t('objects.tls')"
  >
    <!-- OpenVPN does not use sing-box's TLS options; it defines its own set
         with its own names, and the same name means the opposite thing on the
         two sides: on a server `certificate` is its own and `client_certificate`
         is the CA that signs clients, while on a client `certificate` is the CA
         and `client_certificate` is its own. -->
    <a-flex
      vertical
      :gap="12"
    >
      <a-flex
        justify="space-between"
        align="center"
      >
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
        <!-- Only the server has a certificate of its own to issue. A client's
             trust anchor belongs to whoever runs the server. -->
        <a-tooltip
          v-if="isServer"
          :title="$t('actions.generate')"
        >
          <a-button
            :aria-label="$t('actions.generate')"
            :loading="loading"
            @click="genSelfSigned"
          >
            <template #icon>
              <AppIcon name="key-star" />
            </template>
          </a-button>
        </a-tooltip>
      </a-flex>

      <!-- A server has to present a certificate and key; a client has to be able
           to check the one it is shown, which is the CA below or, failing that,
           the peer fingerprint under the options menu. -->
      <a-row
        v-if="usePath == 0"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="isServer ? 12 : 24"
        >
          <a-form-item
            :label="isServer ? $t('tls.certPath') : $t('types.openvpn.tls.caPath')"
            :extra="isServer ? undefined : $t('types.openvpn.tls.caHint')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="tls.certificate_path" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="isServer"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('tls.keyPath')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="tls.key_path" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-else
        :gutter="[12, 12]"
      >
        <a-col :span="24">
          <a-form-item
            :label="isServer ? $t('tls.cert') : $t('types.openvpn.tls.ca')"
            style="margin-bottom: 0;"
          >
            <a-textarea
              v-model:value="certText"
              :rows="4"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="isServer"
          :span="24"
        >
          <a-form-item
            :label="$t('tls.key')"
            style="margin-bottom: 0;"
          >
            <a-textarea
              v-model:value="keyText"
              :rows="4"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <!-- A server asks every client for a certificate unless it is told not to,
           and then needs the CA that signs them; leaving this unset is what made
           an otherwise complete server refuse to start. -->
      <a-row
        v-if="isServer"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.openvpn.tls.verifyClient')"
            :extra="$t('types.openvpn.tls.verifyClientHint')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="verifyClientCertificate"
              :options="verifyModes"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="wantsClientCerts && usePath == 0"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.openvpn.tls.clientCaPath')"
            :extra="$t('types.openvpn.tls.clientCaHint')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="tls.client_certificate_path" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="wantsClientCerts && usePath != 0"
          :span="24"
        >
          <a-form-item
            :label="$t('types.openvpn.tls.clientCa')"
            :extra="$t('types.openvpn.tls.clientCaHint')"
            style="margin-bottom: 0;"
          >
            <a-textarea
              v-model:value="clientCertText"
              :rows="4"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <!-- A client presents a certificate of its own when the server asks. -->
      <template v-if="!isServer && optionMutual">
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
      </template>

      <!-- tls-auth / tls-crypt wrap the control channel in a pre-shared key.
           Most real deployments use one, and both ends have to agree on it. -->
      <template v-if="optionControlWrap && tls.control_wrap">
        <a-row :gutter="[12, 12]">
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('types.openvpn.tls.controlWrap')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="tls.control_wrap.type"
                :options="controlWrapTypes"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col
            v-if="usePath == 0"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('types.openvpn.tls.controlWrapKeyPath')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="tls.control_wrap.key_path" />
            </a-form-item>
          </a-col>
          <a-col
            v-if="tls.control_wrap.type == 'tls_auth'"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('types.openvpn.keyDirection')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="tls.control_wrap.direction"
                :options="keyDirections"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-if="usePath != 0"
          :gutter="[12, 12]"
        >
          <a-col :span="24">
            <a-form-item
              :label="$t('types.openvpn.tls.controlWrapKey')"
              :extra="$t('types.openvpn.tls.controlWrapKeyHint')"
              style="margin-bottom: 0;"
            >
              <a-textarea
                v-model:value="controlWrapKeyText"
                :rows="4"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-flex justify="end">
          <a-tooltip :title="$t('actions.generate')">
            <a-button
              :aria-label="$t('actions.generate')"
              :loading="loading"
              @click="genStaticKey"
            >
              <template #icon>
                <AppIcon name="key-star" />
              </template>
            </a-button>
          </a-tooltip>
        </a-flex>
      </template>

      <a-row
        v-if="optionFingerprint"
        :gutter="[12, 12]"
      >
        <a-col :span="24">
          <a-form-item
            :label="$t('types.openvpn.tls.peerFingerprint')"
            :extra="$t('types.openvpn.tls.peerFingerprintHint')"
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
        v-if="(optionServerName && !isServer) || optionCrl"
        :gutter="[12, 12]"
      >
        <a-col
          v-if="optionServerName && !isServer"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.openvpn.tls.serverName')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="tls.server_name" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionCrl"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.openvpn.tls.crlPath')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="tls.crl_path" />
          </a-form-item>
        </a-col>
      </a-row>

      <a-row
        v-if="optionVersions"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('tls.minVer')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="tls.version_min"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in versions"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('tls.maxVer')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="tls.version_max"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in versions"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>

      <a-row
        v-if="optionProfile"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.openvpn.tls.remoteCertificateTls')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="tls.remote_certificate_tls"
              :options="remoteCertificateTls"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.openvpn.tls.certificateProfile')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="tls.certificate_profile"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in certificateProfiles"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
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
              <a-space v-if="!isServer">
                <a-switch v-model:checked="optionMutual" />
                <span>{{ $t('tls.mutual') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionControlWrap" />
                <span>{{ $t('types.openvpn.tls.controlWrap') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionFingerprint" />
                <span>{{ $t('types.openvpn.tls.peerFingerprint') }}</span>
              </a-space>
              <a-space v-if="!isServer">
                <a-switch v-model:checked="optionServerName" />
                <span>{{ $t('types.openvpn.tls.serverName') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionCrl" />
                <span>{{ $t('types.openvpn.tls.crlPath') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionVersions" />
                <span>{{ $t('tls.minVer') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionProfile" />
                <span>{{ $t('types.openvpn.tls.certificateProfile') }}</span>
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
import { OpenVpnTls } from '@/types/endpoints'

// Material that sing-box takes either as a file path or as the PEM itself.
// Switching between the two clears the other, so only one ever reaches the
// config.
const pathFields = ['certificate_path', 'key_path', 'client_certificate_path', 'client_key_path'] as const
const textFields = ['certificate', 'key', 'client_certificate', 'client_key'] as const

type TextField = typeof textFields[number]

// Everything a mutual-authentication setup carries, so the option reads as on
// when the config already holds any of it.
const mutualFields = ['client_certificate', 'client_certificate_path', 'client_key', 'client_key_path'] as const

// The endpoint that owns the TLS block. OpenVPN defines its own TLS options,
// so they sit on the endpoint rather than in a panel TLS config.
interface TlsHolder {
  type: string
  tag: string
  tls?: OpenVpnTls
}

const staticKeyBegin = '-----BEGIN OpenVPN Static key V1-----'

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
      versions: ['1.0', '1.1', '1.2', '1.3'],
      verifyModes: [
        { title: 'Require', value: 'require' },
        { title: 'Optional', value: 'optional' },
        { title: 'None', value: 'none' },
      ],
      remoteCertificateTls: [
        { title: 'Server', value: 'server' },
        { title: 'Client', value: 'client' },
        { title: 'None', value: 'none' },
      ],
      certificateProfiles: ['legacy', 'preferred', 'insecure', 'suiteb'],
      controlWrapTypes: [
        { title: 'tls-auth', value: 'tls_auth' },
        { title: 'tls-crypt', value: 'tls_crypt' },
        { title: 'tls-crypt-v2', value: 'tls_crypt_v2' },
      ],
      keyDirections: [
        { title: 'Server', value: 'server' },
        { title: 'Client', value: 'client' },
      ],
    }
  },
  computed: {
    // Narrow once. The computed returns the same object, so edits reach the
    // parent.
    holder(): TlsHolder { return <TlsHolder>this.$props.data },
    tls(): OpenVpnTls {
      return <OpenVpnTls>this.holder.tls
    },
    isServer(): boolean {
      return this.holder.type === 'openvpn-server'
    },
    certText: {
      get(): string { return this.joined('certificate') },
      set(v: string) { this.split('certificate', v) }
    },
    keyText: {
      get(): string { return this.joined('key') },
      set(v: string) { this.split('key', v) }
    },
    clientCertText: {
      get(): string { return this.joined('client_certificate') },
      set(v: string) { this.split('client_certificate', v) }
    },
    clientKeyText: {
      get(): string { return this.joined('client_key') },
      set(v: string) { this.split('client_key', v) }
    },
    controlWrapKeyText: {
      get(): string { return this.tls.control_wrap?.key?.join('\n') ?? '' },
      set(v: string) {
        if (!this.tls.control_wrap) return
        if (v) this.tls.control_wrap.key = v.split('\n')
        else delete this.tls.control_wrap.key
      }
    },
    // sing-box takes a single string or a list; the panel always writes a list
    // so the field round-trips through one shape.
    peerFingerprint: {
      get(): string[] {
        const value: string | string[] | undefined = this.tls.peer_fingerprint
        if (value == undefined) return []
        return Array.isArray(value) ? value : [value]
      },
      set(v: string[]) { this.tls.peer_fingerprint = v ?? [] }
    },
    // An option is on when the config already carries it, so opening an
    // existing endpoint shows what it actually has.
    wantsClientCerts(): boolean {
      return this.tls.verify_client_certificate != 'none'
    },
    verifyClientCertificate: {
      get(): string { return this.tls.verify_client_certificate ?? 'none' },
      set(v: string) {
        this.tls.verify_client_certificate = <OpenVpnTls['verify_client_certificate']>v
        if (v == 'none') {
          delete this.tls.client_certificate
          delete this.tls.client_certificate_path
        }
      }
    },
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
        } else {
          for (const field of mutualFields) {
            delete this.tls[field]
          }
        }
      }
    },
    optionControlWrap: {
      get(): boolean { return this.tls.control_wrap != undefined },
      set(v: boolean) {
        if (v) this.tls.control_wrap = { type: 'tls_crypt' }
        else delete this.tls.control_wrap
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
    optionCrl: {
      get(): boolean { return this.tls.crl_path != undefined },
      set(v: boolean) {
        if (v) this.tls.crl_path = ''
        else delete this.tls.crl_path
      }
    },
    optionVersions: {
      get(): boolean { return this.tls.version_min != undefined || this.tls.version_max != undefined },
      set(v: boolean) {
        if (v) {
          this.tls.version_min = '1.2'
        } else {
          delete this.tls.version_min
          delete this.tls.version_max
        }
      }
    },
    optionProfile: {
      get(): boolean { return this.tls.remote_certificate_tls != undefined || this.tls.certificate_profile != undefined },
      set(v: boolean) {
        if (v) {
          // Each side checks that the other's certificate was issued for the
          // opposite role.
          this.tls.remote_certificate_tls = this.isServer ? 'client' : 'server'
        } else {
          delete this.tls.remote_certificate_tls
          delete this.tls.certificate_profile
        }
      }
    },
  },
  created() {
    // Every field below writes into data.tls, so it has to exist first. An
    // empty one is dropped again when the endpoint is saved.
    if (!this.holder.tls) this.holder.tls = {}
    const tls = <OpenVpnTls>this.holder.tls
    this.usePath = textFields.some(f => tls[f] != undefined) || tls.control_wrap?.key != undefined ? 1 : 0
    // sing-box reads an unset verify_client_certificate as "require", so a
    // server left alone would demand client certificates and then fail for
    // want of a CA to check them against. The panel makes the choice explicit.
    if (this.isServer && tls.verify_client_certificate == undefined) {
      tls.verify_client_certificate = 'none'
    }
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
      if (this.tls.control_wrap) delete this.tls.control_wrap.key
    },
    clearPaths() {
      for (const field of pathFields) delete this.tls[field]
      if (this.tls.control_wrap) delete this.tls.control_wrap.key_path
    },
    // A self-signed certificate is its own authority, so the same PEM is what a
    // client puts in its Certificate Authority field. The endpoint tag becomes
    // the common name, which is what a client's Expected Server Name matches.
    async genSelfSigned() {
      this.loading = true
      const msg = await HttpUtils.get<string[]>('api/keypairs', { k: 'tls', o: this.holder.tag ?? '' })
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
      // text; there is no file on disk to point at.
      this.clearPaths()
      this.usePath = 1
      this.tls.key = key
      this.tls.certificate = certificate
    },
    // tls-auth and tls-crypt take an OpenVPN static key, not a PEM. Both ends
    // of the tunnel have to carry the same one, so the other side gets a copy
    // of what this produces.
    async genStaticKey() {
      this.loading = true
      const msg = await HttpUtils.get<string[]>('api/keypairs', { k: 'openvpn' })
      this.loading = false
      if (!msg.success || msg.obj.length == 0 || msg.obj.indexOf(staticKeyBegin) == -1) {
        push.error({ message: i18n.global.t('error') + ': ' + msg.obj })
        return
      }
      // The generated key only exists here, so the form has to hold it as
      // text; there is no file on disk to point at.
      this.clearPaths()
      this.usePath = 1
      // Only reachable while the control-wrap section is showing, which is
      // what puts control_wrap there.
      this.tls.control_wrap!.key = msg.obj
    },
  },
}
</script>

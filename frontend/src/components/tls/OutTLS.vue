<template>
  <a-card
    size="small"
    :title="$t('objects.tls')"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <a-row
        v-if="tlsOptional"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="tlsEnable" />
            <span>{{ $t('tls.enable') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <template v-if="tls.enabled">
        <a-row :gutter="[12, 12]">
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-space>
              <a-switch v-model:checked="disable_sni" />
              <span>{{ $t('tls.disableSni') }}</span>
            </a-space>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-space>
              <a-switch v-model:checked="insecure" />
              <span>{{ $t('tls.insecure') }}</span>
            </a-space>
          </a-col>
        </a-row>
        <a-row
          v-if="insecure"
          :gutter="[12, 12]"
        >
          <a-col
            :xs="24"
            :sm="16"
          >
            <a-form-item
              label="Base64 Pin SHA256 Fingerprint"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="pin_sha256">
                <template #addonAfter>
                  <a-button
                    type="text"
                    size="small"
                    :aria-label="$t('actions.update')"
                    :loading="loading"
                    @click="pingCert()"
                  >
                    <template #icon>
                      <AppIcon name="refresh" />
                    </template>
                  </a-button>
                </template>
              </a-input>
            </a-form-item>
          </a-col>
        </a-row>
        <template v-if="optionCert">
          <a-row :gutter="[12, 12]">
            <a-col flex="none">
              <a-radio-group
                v-model:value="usePath"
                option-type="button"
                button-style="solid"
              >
                <a-radio-button
                  :value="0"
                  @click="tls.certificate=undefined; tls.certificate_path=''"
                >
                  {{ $t('tls.usePath') }}
                </a-radio-button>
                <a-radio-button
                  :value="1"
                  @click="tls.certificate_path=undefined; tls.certificate=''"
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
            <a-col
              :xs="24"
              :sm="12"
            >
              <a-form-item
                :label="$t('tls.certPath')"
                style="margin-bottom: 0;"
              >
                <a-input v-model:value="tls.certificate_path" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-row
            v-else
            :gutter="[12, 12]"
          >
            <a-col
              :xs="24"
              :sm="12"
            >
              <a-form-item
                :label="$t('tls.cert')"
                style="margin-bottom: 0;"
              >
                <a-textarea
                  v-model:value="tls.certificate"
                  :rows="4"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </template>
        <a-row
          v-if="tls.server_name != undefined || tls.alpn"
          :gutter="[12, 12]"
        >
          <a-col
            v-if="tls.server_name != undefined"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              label="SNI"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="tls.server_name" />
            </a-form-item>
          </a-col>
          <a-col
            v-if="tls.alpn"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              label="ALPN"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="tls.alpn"
                mode="multiple"
                :options="alpn"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-if="tls.min_version || tls.max_version"
          :gutter="[12, 12]"
        >
          <a-col
            v-if="tls.min_version"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('tls.minVer')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="tls.min_version"
                style="width: 100%"
              >
                <a-select-option
                  v-for="item in tlsVersions"
                  :key="item"
                  :value="item"
                >
                  {{ item }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col
            v-if="tls.max_version"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('tls.maxVer')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="tls.max_version"
                style="width: 100%"
              >
                <a-select-option
                  v-for="item in tlsVersions"
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
          v-if="tls.cipher_suites != undefined"
          :gutter="[12, 12]"
        >
          <a-col
            :xs="24"
            :md="16"
          >
            <a-form-item
              :label="$t('tls.cs')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="tls.cipher_suites"
                mode="multiple"
                :options="cipher_suites"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-if="tls.utls != undefined"
          :gutter="[12, 12]"
        >
          <a-col
            :xs="24"
            :md="12"
          >
            <a-form-item
              label="Fingerprint"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="tls.utls.fingerprint"
                :options="fingerprints"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-if="tls.reality != undefined"
          :gutter="[12, 12]"
        >
          <a-col
            :xs="24"
            :md="12"
          >
            <a-form-item
              :label="$t('tls.pubKey')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="tls.reality.public_key" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :md="8"
          >
            <a-form-item
              label="Short ID"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="tls.reality.short_id" />
            </a-form-item>
          </a-col>
        </a-row>
        <template v-if="tls.ech != undefined">
          <a-row :gutter="[12, 12]">
            <a-col :span="24">
              <strong>ECH</strong>
            </a-col>
          </a-row>
          <a-row :gutter="[12, 12]">
            <a-col flex="none">
              <a-radio-group
                v-model:value="useEchPath"
                option-type="button"
                button-style="solid"
              >
                <a-radio-button
                  :value="0"
                  @click="delete tls.ech?.config"
                >
                  {{ $t('tls.usePath') }}
                </a-radio-button>
                <a-radio-button
                  :value="1"
                  @click="delete tls.ech?.config_path"
                >
                  {{ $t('tls.useText') }}
                </a-radio-button>
              </a-radio-group>
            </a-col>
          </a-row>
          <a-row
            v-if="useEchPath == 0"
            :gutter="[12, 12]"
          >
            <a-col
              :xs="24"
              :sm="12"
            >
              <a-form-item
                :label="$t('tls.certPath')"
                style="margin-bottom: 0;"
              >
                <a-input v-model:value="tls.ech.config_path" />
              </a-form-item>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
            >
              <a-form-item
                :label="$t('tls.queryServerName')"
                style="margin-bottom: 0;"
              >
                <a-input
                  v-model:value="tls.ech.query_server_name"
                  placeholder="ech.example.com"
                />
              </a-form-item>
            </a-col>
          </a-row>
          <a-row
            v-else
            :gutter="[12, 12]"
          >
            <a-col
              :xs="24"
              :sm="12"
            >
              <a-form-item
                :label="$t('tls.cert')"
                style="margin-bottom: 0;"
              >
                <a-textarea
                  v-model:value="echConfigText"
                  :rows="4"
                />
              </a-form-item>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
            >
              <a-form-item
                :label="$t('tls.queryServerName')"
                style="margin-bottom: 0;"
              >
                <a-input
                  v-model:value="tls.ech.query_server_name"
                  placeholder="ech.example.com"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </template>
        <a-row
          v-if="tls.handshake_timeout != undefined"
          :gutter="[12, 12]"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('tls.handshakeTimeout')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="handshakeTimeout"
                :min="1"
                :addon-after="$t('date.s')"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <!-- Refused by reality, so it is offered only for plain TLS. The forged
             hostname has to differ from the SNI, which must be set. -->
        <a-row
          v-if="tls.reality == undefined && tls.spoof != undefined"
          :gutter="[12, 12]"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('tls.spoof')"
              style="margin-bottom: 0;"
            >
              <a-input
                v-model:value="spoof"
                placeholder="allowed.example.com"
              />
            </a-form-item>
          </a-col>
          <a-col
            v-if="tls.spoof"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('tls.spoofMethod')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="tls.spoof_method"
                :options="spoofMethods"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-if="tls.fragment != undefined"
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-space>
              <a-switch v-model:checked="tls.fragment" />
              <span>{{ $t('tls.fragment') }}</span>
            </a-space>
          </a-col>
          <a-col
            v-if="tls.fragment"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-space>
              <a-switch v-model:checked="tls.record_fragment" />
              <span>{{ $t('tls.recordFragment') }}</span>
            </a-space>
          </a-col>
          <a-col
            v-if="tls.fragment"
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('tls.fragmentDelay')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="fragmentFallbackDelay"
                :min="0"
                :addon-after="$t('date.ms')"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </template>
      <a-flex
        v-if="tls.enabled"
        justify="end"
      >
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
                <a-switch v-model:checked="optionCert" />
                <span>{{ $t('tls.cert') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionSNI" />
                <span>SNI</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionALPN" />
                <span>ALPN</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionMinV" />
                <span>{{ $t('tls.minVer') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionMaxV" />
                <span>{{ $t('tls.maxVer') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionCS" />
                <span>{{ $t('tls.cs') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionFP" />
                <span>UTLS</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionReality" />
                <span>Reality</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionEch" />
                <span>ECH</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionFragment" />
                <span>{{ $t('tls.fragment') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionHandshake" />
                <span>{{ $t('tls.handshakeTimeout') }}</span>
              </a-space>
              <a-space v-if="tls.reality == undefined">
                <a-switch v-model:checked="optionSpoof" />
                <span>{{ $t('tls.spoof') }}</span>
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
import HttpUtils from '@/plugins/httputil';
import { oTls, defaultOutTls, spoofMethods } from '@/types/tls'

// The object that owns the client TLS block: an outbound, one of an inbound's
// extra addresses, or a DERP mesh node. Only what this form reads is listed.
interface TlsHolder {
  type?: string
  server?: string
  server_port?: number
  tls: oTls
}

// What api/getCertPing answers with.
interface CertPing {
  leafHash: string
}

export default {
  props: {
    // Outbounds, address rows and mesh nodes all bind here, so the prop asks
    // only for an object and the narrowing happens below.
    outbound: { type: Object as PropType<object>, required: true }
  },
  data() {
    const outbound = <TlsHolder>this.$props.outbound
    return {
      menu: false,
      usePath: outbound?.tls?.certificate? 1:0,
      useEchPath: outbound?.tls.ech?.config? 1:0,
      defaults: defaultOutTls,
      alpn: [
        { title: "H3", value: 'h3' },
        { title: "H2", value: 'h2' },
        { title: "Http/1.1", value: 'http/1.1' },
      ],
      tlsVersions: [ '1.0', '1.1', '1.2', '1.3' ],
      spoofMethods,
      cipher_suites: [
        { title: "RSA-AES128-CBC-SHA", value: "TLS_RSA_WITH_AES_128_CBC_SHA" },
        { title: "RSA-AES256-CBC-SHA", value: "TLS_RSA_WITH_AES_256_CBC_SHA" },
        { title: "RSA-AES128-GCM-SHA256", value: "TLS_RSA_WITH_AES_128_GCM_SHA256" },
        { title: "RSA-AES256-GCM-SHA384", value: "TLS_RSA_WITH_AES_256_GCM_SHA384" },
        { title: "AES128-GCM-SHA256", value: "TLS_AES_128_GCM_SHA256" },
        { title: "AES256-GCM-SHA384", value: "TLS_AES_256_GCM_SHA384" },
        { title: "CHACHA20-POLY1305-SHA256", value: "TLS_CHACHA20_POLY1305_SHA256" },
        { title: "ECDHE-ECDSA-AES128-CBC-SHA", value: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA" },
        { title: "ECDHE-ECDSA-AES256-CBC-SHA", value: "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA" },
        { title: "ECDHE-RSA-AES128-CBC-SHA", value: "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA" },
        { title: "ECDHE-RSA-AES256-CBC-SHA", value: "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA" },
        { title: "ECDHE-ECDSA-AES128-GCM-SHA256", value: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256" },
        { title: "ECDHE-ECDSA-AES256-GCM-SHA384", value: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384" },
        { title: "ECDHE-RSA-AES128-GCM-SHA256", value: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256" },
        { title: "ECDHE-RSA-AES256-GCM-SHA384", value: "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384" },
        { title: "ECDHE-ECDSA-CHACHA20-POLY1305-SHA256", value: "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256" },
        { title: "ECDHE-RSA-CHACHA20-POLY1305-SHA256", value: "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256" }
      ],
      fingerprints: [
        { title: "Chrome", value: "chrome" },
        { title: "Firefox", value: "firefox" },
        { title: "Microsoft Edge", value: "edge" },
        { title: "Apple Safari", value: "safari" },
        { title: "360", value: "360" },
        { title: "QQ", value: "qq" },
        { title: "Apple IOS", value: "ios" },
        { title: "Android", value: "android" },
        { title: "Random", value: "random" },
        { title: "Randomized", value: "randomized" },
      ],
      loading: false
    }
  },
  computed: {
    // Narrow once, since no caller hands over a type the template checker could
    // follow. The computed returns the same object, so edits reach the parent.
    holder(): TlsHolder { return <TlsHolder>this.$props.outbound },
    tls(): oTls {
      return <oTls> this.holder.tls
    },
    tlsEnable: {
      get() { return Object.hasOwn(this.tls, 'enabled') ? this.tls.enabled : false },
      set(newValue: boolean) { this.holder.tls = newValue ? { enabled: true } : { enabled: false } }
    },
    disable_sni: {
      get() { return this.tls.disable_sni ?? false },
      set(newValue: boolean) { this.holder.tls.disable_sni = newValue ? true : undefined }
    },
    insecure: {
      get() { return this.tls.insecure ?? false },
      set(newValue: boolean) { this.holder.tls.insecure = newValue ? true : undefined }
    },
    tlsOptional(): boolean {
      // Address rows and mesh nodes carry no type, and so match nothing here.
      return !['hysteria','hysteria2','tuic','shadowtls', 'anytls', 'naive'].includes(<string>this.holder.type)
    },
    echConfigText: {
      get(): string { return this.tls.ech?.config ? this.tls.ech.config.join('\n') : '' },
      set(newValue:string) { if (this.tls.ech) this.tls.ech.config = newValue.split('\n') }
    },
    optionCert: {
      get(): boolean { return this.tls.certificate != undefined || this.tls.certificate_path != undefined },
      set(v:boolean) {
        this.usePath = 0
        if (v) {
          this.holder.tls.certificate_path = ""
        } else {
          delete this.holder.tls.certificate_path
          delete this.holder.tls.certificate
        }
      }
    },
    optionSNI: {
      get(): boolean { return this.tls.server_name != undefined },
      set(v:boolean) { this.holder.tls.server_name = v ? '' : undefined }
    },
    optionALPN: {
      get(): boolean { return this.tls.alpn != undefined },
      set(v:boolean) { this.holder.tls.alpn = v ? defaultOutTls.alpn : undefined }
    },
    optionMinV: {
      get(): boolean { return this.tls.min_version != undefined },
      set(v:boolean) { this.holder.tls.min_version = v ? defaultOutTls.min_version : undefined }
    },
    optionMaxV: {
      get(): boolean { return this.tls.max_version != undefined },
      set(v:boolean) { this.holder.tls.max_version = v ? defaultOutTls.max_version : undefined }
    },
    optionCS: {
      get(): boolean { return this.tls.cipher_suites != undefined },
      set(v:boolean) { this.holder.tls.cipher_suites = v ? defaultOutTls.cipher_suites : undefined }
    },
    optionFP: {
      get(): boolean { return this.tls.utls != undefined },
      set(v:boolean) { this.holder.tls.utls = v ? defaultOutTls.utls : undefined }
    },
    optionReality: {
      get(): boolean { return this.tls.reality != undefined },
      set(v:boolean) {
        this.holder.tls.reality = v ? defaultOutTls.reality : undefined
        // reality rejects a spoofed ClientHello, so the two cannot coexist.
        if (v) this.optionSpoof = false
      }
    },
    handshakeTimeout: {
      get(): number { return parseInt(this.tls.handshake_timeout?.replace('s', '') ?? '15') || 15 },
      set(v:number) { this.holder.tls.handshake_timeout = v > 0 ? `${v}s` : '15s' }
    },
    optionHandshake: {
      get(): boolean { return this.tls.handshake_timeout != undefined },
      set(v:boolean) {
        if (v) {
          this.holder.tls.handshake_timeout = '15s'
        } else {
          delete this.holder.tls.handshake_timeout
        }
      }
    },
    // sing-box refuses spoof_method on its own, so the method only exists
    // while a forged hostname is given.
    spoof: {
      get(): string { return this.tls.spoof ?? '' },
      set(v:string) {
        this.holder.tls.spoof = v
        if (v.length > 0) {
          this.holder.tls.spoof_method = this.tls.spoof_method ?? 'wrong-sequence'
        } else {
          delete this.holder.tls.spoof_method
        }
      }
    },
    optionSpoof: {
      get(): boolean { return this.tls.spoof != undefined },
      set(v:boolean) {
        if (v) {
          this.holder.tls.spoof = ''
        } else {
          delete this.holder.tls.spoof
          delete this.holder.tls.spoof_method
        }
      }
    },
    optionEch: {
      get(): boolean { return this.tls.ech != undefined },
      set(v:boolean) { this.holder.tls.ech = v ? defaultOutTls.ech : undefined }
    },
    optionFragment: {
      get(): boolean { return this.tls.fragment != undefined },
      set(v:boolean) { 
        if (v) {
          this.holder.tls.fragment = false
        } else {
          delete this.holder.tls.fragment
          delete this.holder.tls.fragment_fallback_delay
          delete this.holder.tls.record_fragment
        }
      }
    },
    fragmentFallbackDelay: {
      get(): number { return parseInt(this.tls.fragment_fallback_delay?.replace('ms','')?? '500')?? 500 },
      set(v:number) { this.holder.tls.fragment_fallback_delay = v>0 ? `${v}ms` : undefined }
    },
    pin_sha256: {
      get(): string { return this.tls.certificate_public_key_sha256?.join(',') ?? '' },
      set(v: string) { this.holder.tls.certificate_public_key_sha256 = v.length > 0 ? v.split(',') : undefined }
    }
  },
  methods: {
    async pingCert() {
      this.loading = true
      const msg = await HttpUtils.post<CertPing>('api/getCertPing', {
        domain: this.holder.server,
        port: this.holder.server_port,
      })
      this.loading = false
      if (msg.success) {
        this.pin_sha256 = msg.obj.leafHash
      }
    }
  }
}
</script>
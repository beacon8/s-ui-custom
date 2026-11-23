<template>
  <a-card
    size="small"
    title="ACME"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <a-row :gutter="[12, 12]">
        <a-col :span="24">
          <a-form-item
            :label="$t('rule.domain') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="domains" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionDir || optionDefault || optionEmail"
        :gutter="[12, 12]"
      >
        <a-col
          v-if="optionDir"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('tls.acme.dataDir')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="acme.data_directory" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionDefault"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('tls.acme.defaultDomain')"
            style="margin-bottom: 0;"
          >
            <a-auto-complete
              v-model:value="acme.default_server_name"
              :options="(acme.domain ?? []).map(d => ({ value: d }))"
              style="width: 100%"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionEmail"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('email')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="acme.email" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionChallenge"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="acme.disable_http_challenge" />
            <span>{{ $t('tls.acme.httpChallenge') }}</span>
          </a-space>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="acme.disable_tls_alpn_challenge" />
            <span>{{ $t('tls.acme.tlsChallenge') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row
        v-if="optionPorts"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('tls.acme.altHport')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="acme.alternative_http_port"
              :min="1"
              :max="65532"
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
            :label="$t('tls.acme.altTport')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="acme.alternative_tls_port"
              :min="1"
              :max="65532"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionProvider"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('tls.acme.caProvider')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="caProvider"
              :options="providerList"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="caProvider == ''"
          :xs="24"
          :md="16"
        >
          <a-form-item
            :label="$t('tls.acme.customCa')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="acme.provider" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="acme.external_account != undefined"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            label="Key ID"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="acme.external_account.key_id" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            label="MAC Key"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="acme.external_account.mac_key" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionHttpClient"
        :gutter="[12, 12]"
      >
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
              v-model:value="acme.http_client"
              :not-found-content="$t('basic.httpClient.none')"
              allow-clear
              style="width: 100%"
              @change="(v: any) => { if (!v) acme.http_client = undefined }"
            >
              <a-select-option
                v-for="item in httpClients"
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
        v-if="acme.dns01_challenge != undefined"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('tls.acme.dns01Provider')"
            style="margin-bottom: 0;"
          >
            <a-select
              :value="acme.dns01_challenge.provider"
              style="width: 100%"
              @change="(v: any) => { acme.dns01_challenge = { provider: String(v) } }"
            >
              <a-select-option
                v-for="d in dnsProviders"
                :key="d.provider"
                :value="d.provider"
              >
                {{ d.provider }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-for="item in dnsProviders.filter(d => d.provider == acme.dns01_challenge?.provider)[0]?.params"
          :key="item"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('tls.acme.dns01Params.' + item)"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="acme.dns01_challenge[item]" />
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
              <a-space>
                <a-switch v-model:checked="optionDir" />
                <span>{{ $t('tls.acme.dataDir') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionDefault" />
                <span>{{ $t('tls.acme.defaultDomain') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionEmail" />
                <span>{{ $t('email') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionChallenge" />
                <span>{{ $t('tls.acme.disableChallenges') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionPorts" />
                <span>{{ $t('tls.acme.altPorts') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionProvider" />
                <span>{{ $t('tls.acme.caProvider') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionExt" />
                <span>{{ $t('tls.acme.extAcc') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionDns01" />
                <span>{{ $t('tls.acme.dns01') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionHttpClient" />
                <span>{{ $t('basic.httpClient.title') }}</span>
              </a-space>
            </a-flex>
          </template>
          <a-button>
            {{ $t('tls.acme.options') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
</template>

<script lang="ts">
import { PropType } from 'vue'
import { acme, certProvider } from '@/types/tls'
import { httpClientTags } from '@/plugins/httpClient'

// Edits one ACME entry of the config's certificate_providers list. The tag and
// the type live in the modal around it, since every provider type carries them.
export default {
  props: {
    // The modal holds the provider union and picks this form with a v-if on
    // the type, which the template checker cannot follow, so it is narrowed by
    // the computed below.
    data: { type: Object as PropType<certProvider>, required: true }
  },
  data() {
    return {
      menu: false,
      providerList: [
        { title: "Let's Encrypt", value: "letsencrypt" },
        { title: "ZeroSSL", value: "zerossl" },
        { title: "Custom", value: "" }
      ],
      dnsProviders: [
        { provider: "cloudflare", params: [ "api_token", "zone_token" ] },
        { provider: "alidns", params: [ "access_key_id", "access_key_secret", "region_id", "security_token" ] },
        { provider: "acmedns", params: [ "username", "password", "subdomain", "server_url" ] }
      ]
    }
  },
  computed: {
    acme(): acme {
      return <acme>this.$props.data
    },
    httpClients(): string[] {
      return httpClientTags()
    },
    // Preselect a declared client so the option never leaves an empty
    // reference behind; with none declared the select simply has nothing.
    optionHttpClient: {
      get(): boolean { return this.acme.http_client != undefined },
      set(v: boolean) { this.acme.http_client = v ? this.httpClients[0] ?? '' : undefined }
    },
    domains: {
      get() { return this.acme.domain ? this.acme.domain.join(',') : "" },
      set(v: string) {
        if(!v.endsWith(',')) {
          this.acme.domain = v.length > 0 ? v.split(',') : []
        }
      }
    },
    caProvider: {
      get() { return this.acme.provider && ['letsencrypt','zerossl'].includes(this.acme.provider) ? this.acme.provider : '' },
      set(v: string) { this.acme.provider = ['letsencrypt','zerossl'].includes(v) ? v : 'https://' }
    },
    optionDir: {
      get(): boolean { return this.acme.data_directory != undefined },
      set(v:boolean) { this.acme.data_directory = v ? '' : undefined }
    },
    optionDefault: {
      get(): boolean { return this.acme.default_server_name != undefined },
      set(v:boolean) { this.acme.default_server_name = v ? this.acme.domain?.[0] ?? '' : undefined }
    },
    optionEmail: {
      get(): boolean { return this.acme.email != undefined },
      set(v:boolean) { this.acme.email = v ? '' : undefined }
    },
    optionChallenge: {
      get(): boolean { return this.acme.disable_http_challenge != undefined || this.acme.disable_tls_alpn_challenge != undefined },
      set(v:boolean) {
        if (v) {
          this.acme.disable_http_challenge = false
          this.acme.disable_tls_alpn_challenge = false
        } else {
          delete this.acme.disable_http_challenge
          delete this.acme.disable_tls_alpn_challenge
        }
      }
    },
    optionPorts: {
      get(): boolean { return this.acme.alternative_http_port != undefined || this.acme.alternative_tls_port != undefined },
      set(v:boolean) {
        if (v) {
          this.acme.alternative_http_port = 80
          this.acme.alternative_tls_port = 443
        } else {
          delete this.acme.alternative_http_port
          delete this.acme.alternative_tls_port
        }
      }
    },
    optionProvider: {
      get(): boolean { return this.acme.provider != undefined },
      set(v:boolean) { this.acme.provider = v ? 'letsencrypt' : undefined }
    },
    optionExt: {
      get(): boolean { return this.acme.external_account != undefined },
      set(v:boolean) { this.acme.external_account = v ? { key_id: '', mac_key: '' } : undefined }
    },
    optionDns01: {
      get(): boolean { return this.acme.dns01_challenge != undefined },
      set(v:boolean) {
        if (v) {
          this.acme.dns01_challenge = { provider: 'cloudflare' }
        } else {
          delete this.acme.dns01_challenge
        }
      }
    },
  }
}
</script>

<template>
  <a-modal
    :open="visible"
    :width="800"
    class="tls-modal"
    @cancel="closeModal"
  >
    <template #title>
      <a-flex
        justify="space-between"
        align="center"
        style="padding-inline-end: 24px;"
      >
        <span>{{ $t('actions.' + title) + " " + $t('objects.tls') }}</span>
        <DocLink section="tls" />
      </a-flex>
    </template>

    <div class="tls-modal-body">
      <a-form
        ref="formRef"
        layout="vertical"
        :model="tls"
        class="tls-form"
        @submit.prevent
      >
        <!-- 名称与协议类型同排，两类标签都在控件上方；协议相关字段紧随其下 -->
        <div class="tls-section">
          <a-row :gutter="[16, 16]">
            <a-col
              :xs="24"
              :sm="15"
              :md="16"
            >
              <a-form-item
                :label="$t('client.name')"
                name="name"
                :rules="nameRules"
              >
                <a-input v-model:value="tls.name" />
              </a-form-item>
            </a-col>
            <a-col
              :xs="24"
              :sm="9"
              :md="8"
            >
              <a-form-item :label="$t('tls.protocolType')">
                <a-radio-group
                  v-model:value="tlsType"
                  option-type="button"
                  button-style="solid"
                  @change="changeTlsType"
                >
                  <a-radio-button :value="0">
                    TLS
                  </a-radio-button>
                  <a-radio-button :value="1">
                    Reality
                  </a-radio-button>
                </a-radio-group>
              </a-form-item>
            </a-col>
          </a-row>

          <!-- TLS 的 SNI/版本/套件；Reality 下只保留 SNI -->
          <a-row
            v-if="showProtocolRow"
            :gutter="[16, 16]"
          >
            <a-col
              v-if="inTls.server_name != undefined"
              :xs="24"
              :sm="12"
              :md="8"
            >
              <a-form-item label="SNI">
                <a-input v-model:value="inTls.server_name" />
              </a-form-item>
            </a-col>
            <template v-if="tlsType == 0">
              <a-col
                v-if="inTls.min_version"
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item :label="$t('tls.minVer')">
                  <a-select
                    v-model:value="inTls.min_version"
                    style="width: 100%;"
                  >
                    <a-select-option
                      v-for="v in tlsVersions"
                      :key="v"
                      :value="v"
                    >
                      {{ v }}
                    </a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col
                v-if="inTls.max_version"
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item :label="$t('tls.maxVer')">
                  <a-select
                    v-model:value="inTls.max_version"
                    style="width: 100%;"
                  >
                    <a-select-option
                      v-for="v in tlsVersions"
                      :key="v"
                      :value="v"
                    >
                      {{ v }}
                    </a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col
                v-if="inTls.alpn"
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item label="ALPN">
                  <a-select
                    v-model:value="inTls.alpn"
                    mode="multiple"
                    :options="alpn"
                    :field-names="{ label: 'title', value: 'value' }"
                    style="width: 100%;"
                  />
                </a-form-item>
              </a-col>
              <a-col
                v-if="inTls.cipher_suites != undefined"
                :xs="24"
                :md="16"
              >
                <a-form-item :label="$t('tls.cs')">
                  <a-select
                    v-model:value="inTls.cipher_suites"
                    mode="multiple"
                    :options="cipher_suites"
                    :field-names="{ label: 'title', value: 'value' }"
                    style="width: 100%;"
                  />
                </a-form-item>
              </a-col>
            </template>
          </a-row>

          <!-- Reality：握手目标、密钥对与 Short ID 都紧跟协议切换器 -->
          <template v-if="outTls.reality && inTls.reality">
            <a-row
              :gutter="[16, 16]"
              align="bottom"
            >
              <a-col
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item :label="$t('types.shdwTls.hs')">
                  <a-input v-model:value="inTls.reality.handshake.server" />
                </a-form-item>
              </a-col>
              <a-col
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item :label="$t('out.port')">
                  <a-input-number
                    v-model:value="server_port"
                    :min="0"
                    style="width: 100%;"
                  />
                </a-form-item>
              </a-col>
              <a-col flex="none">
                <a-button
                  type="default"
                  :loading="loading"
                  @click="genRealityKey"
                >
                  <template #icon>
                    <AppIcon name="key-star" />
                  </template>
                  {{ $t('actions.generate') }}
                </a-button>
              </a-col>
            </a-row>

            <a-form-item :label="$t('tls.privKey')">
              <a-input v-model:value="inTls.reality.private_key" />
            </a-form-item>
            <a-form-item :label="$t('tls.pubKey')">
              <a-input v-model:value="outTls.reality.public_key" />
            </a-form-item>

            <a-row :gutter="[16, 16]">
              <a-col
                :xs="24"
                :sm="12"
                :md="optionTime ? 16 : 24"
              >
                <a-form-item label="Short IDs">
                  <a-input v-model:value="short_id">
                    <template #suffix>
                      <AppIcon
                        name="refresh"
                        style="cursor: pointer;"
                        @click="randomSID"
                      />
                    </template>
                  </a-input>
                </a-form-item>
              </a-col>
              <a-col
                v-if="optionTime"
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item label="Max Time Difference">
                  <a-input-number
                    v-model:value="max_time"
                    :min="1"
                    :addon-after="$t('date.m')"
                    style="width: 100%;"
                  />
                </a-form-item>
              </a-col>
            </a-row>
          </template>
        </div>

        <!-- 证书：提供者与输入方式同排，路径并排、PEM 内容上下排列 -->
        <template v-if="tlsType == 0">
          <a-divider class="tls-divider" />
          <div class="tls-section">
            <a-row :gutter="[16, 16]">
              <a-col
                :xs="24"
                :sm="12"
                :md="10"
              >
                <a-form-item :label="$t('tls.provider.title')">
                  <a-select
                    v-model:value="inTls.certificate_provider"
                    :not-found-content="$t('tls.provider.none')"
                    allow-clear
                    style="width: 100%;"
                    @clear="inTls.certificate_provider = undefined"
                  >
                    <a-select-option
                      v-for="tag in providerTags"
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
                :md="14"
              >
                <a-form-item :label="$t('tls.inputMode')">
                  <a-flex
                    align="center"
                    justify="space-between"
                    :gap="12"
                    wrap
                  >
                    <a-radio-group
                      v-model:value="usePath"
                      option-type="button"
                      button-style="solid"
                      :disabled="inTls.certificate_provider != undefined"
                    >
                      <a-radio-button
                        :value="0"
                        @click="inTls.key = undefined; inTls.certificate = undefined"
                      >
                        {{ $t('tls.usePath') }}
                      </a-radio-button>
                      <a-radio-button
                        :value="1"
                        @click="inTls.key_path = undefined; inTls.certificate_path = undefined"
                      >
                        {{ $t('tls.useText') }}
                      </a-radio-button>
                    </a-radio-group>
                    <!-- 钥匙图标生成自签名证书并填入下方 PEM 输入框，所以只在文件内容模式出现 -->
                    <a-button
                      v-if="inTls.certificate_provider == undefined && usePath == 1"
                      type="default"
                      :loading="loading"
                      @click="genSelfSigned"
                    >
                      <template #icon>
                        <AppIcon name="key-star" />
                      </template>
                      {{ $t('tls.generateSelfSigned') }}
                    </a-button>
                  </a-flex>
                  <div
                    v-if="inTls.certificate_provider != undefined"
                    class="tls-hint"
                  >
                    {{ $t('tls.providerHint') }}
                  </div>
                </a-form-item>
              </a-col>
            </a-row>

            <a-row
              v-if="inTls.certificate_provider == undefined && usePath == 0"
              :gutter="[16, 16]"
            >
              <a-col
                :xs="24"
                :sm="12"
              >
                <a-form-item :label="$t('tls.certPath')">
                  <a-input v-model:value="inTls.certificate_path" />
                </a-form-item>
              </a-col>
              <a-col
                :xs="24"
                :sm="12"
              >
                <a-form-item :label="$t('tls.keyPath')">
                  <a-input v-model:value="inTls.key_path" />
                </a-form-item>
              </a-col>
            </a-row>
            <template v-else-if="inTls.certificate_provider == undefined">
              <a-form-item :label="$t('tls.cert')">
                <a-textarea
                  v-model:value="certText"
                  :rows="4"
                />
              </a-form-item>
              <a-form-item :label="$t('tls.key')">
                <a-textarea
                  v-model:value="keyText"
                  :rows="4"
                />
              </a-form-item>
            </template>
          </div>
        </template>

        <!-- 安全开关与 TLS 选项同行；选项开关启用的字段直接出现在本行下方 -->
        <a-divider class="tls-divider" />
        <div class="tls-section">
          <div class="tls-security-row">
            <a-space v-if="tlsType == 0">
              <a-switch v-model:checked="disableSni" />
              <span>{{ $t('tls.disableSni') }}</span>
            </a-space>
            <a-space v-if="tlsType == 0">
              <a-switch v-model:checked="insecure" />
              <span>{{ $t('tls.insecure') }}</span>
            </a-space>
            <a-popover
              v-model:open="menu"
              trigger="click"
              placement="top"
            >
              <template #content>
                <a-space direction="vertical">
                  <template v-if="tlsType == 0">
                    <a-space>
                      <a-switch
                        v-model:checked="optionSNI"
                        size="small"
                      />
                      <span>SNI</span>
                    </a-space>
                    <a-space>
                      <a-switch
                        v-model:checked="optionALPN"
                        size="small"
                      />
                      <span>ALPN</span>
                    </a-space>
                    <a-space>
                      <a-switch
                        v-model:checked="optionMinV"
                        size="small"
                      />
                      <span>{{ $t('tls.minVer') }}</span>
                    </a-space>
                    <a-space>
                      <a-switch
                        v-model:checked="optionMaxV"
                        size="small"
                      />
                      <span>{{ $t('tls.maxVer') }}</span>
                    </a-space>
                    <a-space>
                      <a-switch
                        v-model:checked="optionCS"
                        size="small"
                      />
                      <span>{{ $t('tls.cs') }}</span>
                    </a-space>
                    <a-space>
                      <a-switch
                        v-model:checked="optionFP"
                        size="small"
                      />
                      <span>UTLS</span>
                    </a-space>
                    <a-space>
                      <a-switch
                        v-model:checked="optionClientAuth"
                        size="small"
                      />
                      <span>{{ $t('tls.mutual') }}</span>
                    </a-space>
                    <a-space>
                      <a-switch
                        v-model:checked="optionStore"
                        size="small"
                      />
                      <span>{{ $t('tls.store') }}</span>
                    </a-space>
                    <a-space>
                      <a-switch
                        v-model:checked="optionKtls"
                        size="small"
                      />
                      <span>{{ $t('tls.ktls') }}</span>
                    </a-space>
                    <a-space>
                      <a-switch
                        v-model:checked="optionSpoof"
                        size="small"
                      />
                      <span>{{ $t('tls.spoof') }}</span>
                    </a-space>
                  </template>
                  <template v-else>
                    <a-space>
                      <a-switch
                        v-model:checked="optionTime"
                        size="small"
                      />
                      <span>Max Time Difference</span>
                    </a-space>
                  </template>
                  <a-space>
                    <a-switch
                      v-model:checked="optionHandshake"
                      size="small"
                    />
                    <span>{{ $t('tls.handshakeTimeout') }}</span>
                  </a-space>
                </a-space>
              </template>
              <a-button type="default">
                {{ $t('tls.options') }}
              </a-button>
            </a-popover>
          </div>

          <template v-if="hasAdvanced">
            <a-divider class="tls-subdivider" />
            <a-row
              v-if="optionHandshake"
              :gutter="[16, 16]"
            >
              <a-col
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item :label="$t('tls.handshakeTimeout')">
                  <a-input-number
                    v-model:value="handshakeTimeout"
                    :min="1"
                    :addon-after="$t('date.s')"
                    style="width: 100%;"
                  />
                </a-form-item>
              </a-col>
            </a-row>

            <a-row
              v-if="optionStore || optionKtls"
              :gutter="[16, 16]"
              align="bottom"
            >
              <a-col
                v-if="optionStore"
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item :label="$t('tls.store')">
                  <a-select
                    v-model:value="inTls.store"
                    :options="storeItems"
                    :field-names="{ label: 'title', value: 'value' }"
                    style="width: 100%;"
                  />
                </a-form-item>
              </a-col>
              <template v-if="optionKtls">
                <a-col
                  :xs="24"
                  :sm="12"
                  :md="8"
                >
                  <a-space>
                    <a-switch v-model:checked="inTls.kernel_tx" />
                    <span>{{ $t('tls.kernelTx') }}</span>
                  </a-space>
                </a-col>
                <a-col
                  :xs="24"
                  :sm="12"
                  :md="8"
                >
                  <a-space>
                    <a-switch v-model:checked="inTls.kernel_rx" />
                    <span>{{ $t('tls.kernelRx') }}</span>
                  </a-space>
                </a-col>
              </template>
            </a-row>

            <!-- 双向 TLS：服务器校验客户端证书，客户端出示证书 -->
            <template v-if="optionClientAuth">
              <a-divider class="tls-subdivider" />
              <div class="tls-subtitle">
                {{ $t('tls.mutual') }}
              </div>
              <a-row :gutter="[16, 16]">
                <a-col
                  :xs="24"
                  :sm="12"
                  :md="8"
                >
                  <a-form-item :label="$t('tls.clientAuth')">
                    <a-select
                      v-model:value="inTls.client_authentication"
                      :options="clientAuthTypes"
                      :field-names="{ label: 'title', value: 'value' }"
                      allow-clear
                      style="width: 100%;"
                      @clear="inTls.client_authentication = undefined"
                    />
                  </a-form-item>
                </a-col>
                <a-col
                  :xs="24"
                  :sm="12"
                  :md="8"
                >
                  <a-form-item :label="$t('tls.inputMode')">
                    <a-radio-group
                      v-model:value="useClientPath"
                      option-type="button"
                      button-style="solid"
                    >
                      <a-radio-button
                        :value="0"
                        @click="switchClientTls(true)"
                      >
                        {{ $t('tls.usePath') }}
                      </a-radio-button>
                      <a-radio-button
                        :value="1"
                        @click="switchClientTls(false)"
                      >
                        {{ $t('tls.useText') }}
                      </a-radio-button>
                    </a-radio-group>
                  </a-form-item>
                </a-col>
              </a-row>

              <a-form-item
                v-if="useClientPath == 0"
                :label="$t('tls.clientCaPath')"
              >
                <a-select
                  v-model:value="inTls.client_certificate_path"
                  mode="tags"
                  style="width: 100%;"
                />
              </a-form-item>
              <a-form-item
                v-else
                :label="$t('tls.clientCa')"
              >
                <a-textarea
                  v-model:value="clientCaText"
                  :rows="3"
                />
              </a-form-item>

              <a-row
                v-if="useClientPath == 0"
                :gutter="[16, 16]"
              >
                <a-col
                  :xs="24"
                  :sm="12"
                >
                  <a-form-item :label="$t('tls.clientCertPath')">
                    <a-input v-model:value="outTls.client_certificate_path" />
                  </a-form-item>
                </a-col>
                <a-col
                  :xs="24"
                  :sm="12"
                >
                  <a-form-item :label="$t('tls.clientKeyPath')">
                    <a-input v-model:value="outTls.client_key_path" />
                  </a-form-item>
                </a-col>
              </a-row>
              <template v-else>
                <a-form-item :label="$t('tls.clientCert')">
                  <a-textarea
                    v-model:value="clientCertText"
                    :rows="3"
                  />
                </a-form-item>
                <a-form-item :label="$t('tls.clientKey')">
                  <a-textarea
                    v-model:value="clientKeyText"
                    :rows="3"
                  />
                </a-form-item>
              </template>

              <a-form-item :label="$t('tls.clientPin')">
                <a-select
                  v-model:value="inTls.client_certificate_public_key_sha256"
                  mode="tags"
                  allow-clear
                  style="width: 100%;"
                  @clear="inTls.client_certificate_public_key_sha256 = undefined"
                />
              </a-form-item>
            </template>

            <a-row
              v-if="outTls.utls != undefined"
              :gutter="[16, 16]"
            >
              <a-col
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item label="Fingerprint">
                  <a-select
                    v-model:value="outTls.utls.fingerprint"
                    :options="fingerprints"
                    :field-names="{ label: 'title', value: 'value' }"
                    style="width: 100%;"
                  />
                </a-form-item>
              </a-col>
            </a-row>

            <a-row
              v-if="tlsType == 0 && outTls.spoof != undefined"
              :gutter="[16, 16]"
            >
              <a-col
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item :label="$t('tls.spoof')">
                  <a-input
                    v-model:value="spoof"
                    placeholder="allowed.example.com"
                  />
                </a-form-item>
              </a-col>
              <a-col
                v-if="outTls.spoof"
                :xs="24"
                :sm="12"
                :md="8"
              >
                <a-form-item :label="$t('tls.spoofMethod')">
                  <a-select
                    v-model:value="outTls.spoof_method"
                    :options="spoofMethods"
                    :field-names="{ label: 'title', value: 'value' }"
                    style="width: 100%;"
                  />
                </a-form-item>
              </a-col>
            </a-row>
          </template>
        </div>

        <!-- ECH：标题与启用开关合并一行，启用后展开字段 -->
        <a-divider class="tls-divider" />
        <EchVue
          :i-tls="inTls"
          :o-tls="outTls"
        />
      </a-form>
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
        @click="saveChanges"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import { tls, iTls, defaultInTls, oTls, defaultOutTls, spoofMethods } from '@/types/tls'
import DocLink from '@/components/DocLink.vue'
import EchVue from '@/components/tls/Ech.vue'
import HttpUtils from '@/plugins/httputil'
import { push } from 'notivue'
import { i18n } from '@/locales'
import RandomUtil from '@/plugins/randomUtil'
import type { PropType } from 'vue'
export default {
  components: { DocLink, EchVue },
  props: {
    visible: { type: Boolean, required: true },
    data: { type: String, required: true },
    id: { type: Number, required: true },
    providers: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['close', 'save'],
  data() {
    return {
      tls: <tls>{ id: 0, name: '', server: <iTls>{ enabled: true }, client: <oTls>{} },
      title: "add",
      loading: false,
      menu: false,
      tlsType: 0,
      usePath: 0,
      alpn: [
        { title: "H3", value: 'h3' },
        { title: "H2", value: 'h2' },
        { title: "Http/1.1", value: 'http/1.1' },
      ],
      tlsVersions: [ '1.0', '1.1', '1.2', '1.3' ],
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
      storeItems: [
        { title: "Mozilla", value: "mozilla" },
        { title: "Chrome", value: "chrome" },
      ],
      useClientPath: 0,
      clientAuthTypes: [
        { title: "No", value: "no" },
        { title: "Request", value: "request" },
        { title: "Require Any", value: "require-any" },
        { title: "Verify If Given", value: "verify-if-given" },
        { title: "Require And Verify", value: "require-and-verify" },
      ],
      spoofMethods,
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
      ]
    }
  },
  computed: {
    providerTags(): string[] {
      return <string[]>(this.$props.providers ?? [])
    },
    inTls(): iTls {
      return this.tls.server
    },
    outTls(): oTls {
      return this.tls.client
    },
    // The protocol block sits directly under the name/type row and collapses
    // entirely when every optional TLS field is off.
    showProtocolRow(): boolean {
      return this.inTls.server_name != undefined
        || (this.tlsType == 0 && (
          Boolean(this.inTls.min_version)
          || Boolean(this.inTls.max_version)
          || Boolean(this.inTls.alpn)
          || this.inTls.cipher_suites != undefined
        ))
    },
    // Fields revealed by the "TLS 选项" popover. They are rendered under the
    // switch row, so the options button stays next to the switches it drives.
    hasAdvanced(): boolean {
      return this.optionHandshake
        || this.optionStore
        || this.optionKtls
        || this.optionClientAuth
        || this.optionFP
        || (this.tlsType == 0 && this.optionSpoof)
    },
    nameRules() {
      return [{
        required: true,
        whitespace: true,
        message: i18n.global.t('tls.nameRequired'),
        trigger: 'blur',
      }]
    },
    clientCaText: {
      get(): string { return this.inTls.client_certificate ? this.inTls.client_certificate.join('\n') : '' },
      set(v: string) { this.inTls.client_certificate = v ? v.split('\n') : undefined }
    },
    clientCertText: {
      get(): string { return this.outTls.client_certificate ? this.outTls.client_certificate.join('\n') : '' },
      set(v: string) { this.outTls.client_certificate = v ? v.split('\n') : undefined }
    },
    clientKeyText: {
      get(): string { return this.outTls.client_key ? this.outTls.client_key.join('\n') : '' },
      set(v: string) { this.outTls.client_key = v ? v.split('\n') : undefined }
    },
    certText: {
      get(): string { return this.inTls.certificate ? this.inTls.certificate.join('\n') : '' },
      set(v:string) { this.inTls.certificate = v.split('\n') }
    },
    keyText: {
      get(): string { return this.inTls.key ? this.inTls.key.join('\n') : '' },
      set(v:string) { this.inTls.key = v.split('\n') }
    },
    disableSni: {
      get() { return this.outTls.disable_sni ?? false },
      set(v: boolean) { this.tls.client.disable_sni = v ? true : undefined }
    },
    insecure: {
      get() { return this.outTls.insecure ?? false },
      set(v: boolean) { this.tls.client.insecure = v ? true : undefined }
    },
    server_port: {
      get() { return this.inTls.reality?.handshake?.server_port ? this.inTls.reality.handshake.server_port : 443 },
      // The field is a plain v-model on a number input, so it arrives as a
      // string; a programmatic set may still hand over a number.
      set(v: string | number) {
        if (this.inTls.reality){
          const s = String(v)
          this.inTls.reality.handshake.server_port = s.length == 0 || Number(s) == 0 ? 443 : parseInt(s)
        }
      }
    },
    short_id: {
      get() { return this.inTls.reality?.short_id ? this.inTls.reality.short_id.join(',') : undefined },
      set(v: string) {
        if (this.inTls.reality){
          this.inTls.reality.short_id = v.length > 0 ? v.split(',') : []
        }
      }
    },
    max_time: {
      get() { return this.inTls?.reality?.max_time_difference ? this.inTls.reality.max_time_difference.replace('m','') : 1 },
      set(v: number) {
        if (this.inTls.reality){
          this.inTls.reality.max_time_difference = v > 0 ? v + 'm' : '1m'
        }
      }
    },
    optionClientAuth: {
      get(): boolean {
        return this.inTls.client_certificate_path != undefined
          || this.inTls.client_certificate != undefined
      },
      set(v: boolean) {
        if (v) {
          this.switchClientTls(true)
        } else {
          this.inTls.client_authentication = undefined
          this.inTls.client_certificate_path = undefined
          this.inTls.client_certificate = undefined
          this.outTls.client_certificate_path = undefined
          this.outTls.client_key_path = undefined
          this.outTls.client_certificate = undefined
          this.outTls.client_key = undefined
          this.inTls.client_certificate_public_key_sha256 = undefined
        }
      }
    },
    optionSNI: {
      get(): boolean { return this.inTls.server_name != undefined },
      set(v:boolean) { this.inTls.server_name = v ? '' : undefined }
    },
    optionALPN: {
      get(): boolean { return this.inTls.alpn != undefined },
      set(v:boolean) { this.inTls.alpn = v ? defaultInTls.alpn : undefined }
    },
    optionMinV: {
      get(): boolean { return this.inTls.min_version != undefined },
      set(v:boolean) { this.inTls.min_version = v ? defaultInTls.min_version : undefined }
    },
    optionMaxV: {
      get(): boolean { return this.inTls.max_version != undefined },
      set(v:boolean) { this.inTls.max_version = v ? defaultInTls.max_version : undefined }
    },
    optionCS: {
      get(): boolean { return this.inTls.cipher_suites != undefined },
      set(v:boolean) { this.inTls.cipher_suites = v ? defaultInTls.cipher_suites : undefined }
    },
    optionFP: {
      get(): boolean { return this.outTls.utls != undefined },
      set(v:boolean) { this.outTls.utls = v ? defaultOutTls.utls : undefined }
    },
    // Written on the server side alone; out_json hands the same value to the
    // client, as it does for the versions and cipher suites.
    handshakeTimeout: {
      get(): number { return parseInt(this.inTls.handshake_timeout?.replace('s', '') ?? '15') || 15 },
      set(v: number) { this.inTls.handshake_timeout = v > 0 ? `${v}s` : '15s' }
    },
    optionHandshake: {
      get(): boolean { return this.inTls.handshake_timeout != undefined },
      set(v: boolean) { this.inTls.handshake_timeout = v ? '15s' : undefined }
    },
    // sing-box refuses spoof_method on its own, so the method only exists
    // while a forged hostname is given.
    spoof: {
      get(): string { return this.outTls.spoof ?? '' },
      set(v: string) {
        this.outTls.spoof = v
        this.outTls.spoof_method = v.length > 0 ? (this.outTls.spoof_method ?? 'wrong-sequence') : undefined
      }
    },
    optionSpoof: {
      get(): boolean { return this.outTls.spoof != undefined },
      set(v: boolean) {
        if (v) {
          this.outTls.spoof = ''
        } else {
          this.outTls.spoof = undefined
          this.outTls.spoof_method = undefined
        }
      }
    },
    optionStore: {
      get(): boolean { return this.inTls.store != undefined },
      set(v:boolean) { this.inTls.store = v ? 'mozilla' : undefined }
    },
    optionKtls: {
      get(): boolean { return this.inTls.kernel_tx != undefined || this.inTls.kernel_rx != undefined },
      set(v:boolean) {
        if (v) {
          this.inTls.kernel_tx = false
          this.inTls.kernel_rx = false
        } else {
          delete this.inTls.kernel_tx
          delete this.inTls.kernel_rx
        }
      }
    },
    optionEch: {
      get(): boolean { return this.outTls.ech != undefined },
      set(v:boolean) { this.outTls.ech = v ? defaultOutTls.ech : undefined }
    },
    optionTime: {
      get(): boolean { return this.inTls?.reality?.max_time_difference != undefined },
      set(v:boolean) { if (this.inTls.reality) this.inTls.reality.max_time_difference = v ? "1m" : undefined }
    }
  },
  watch: {
    visible(v) {
      if (v) {
        this.updateData(this.$props.id)
      }
    },
  },
  methods: {
    switchClientTls(isPath: boolean) {
      if (isPath) {
        this.inTls.client_certificate = undefined
        this.outTls.client_certificate = undefined
        this.outTls.client_key = undefined
        this.inTls.client_certificate_path = []
        this.outTls.client_certificate_path = ''
        this.outTls.client_key_path = ''
      } else {
        this.inTls.client_certificate_path = undefined
        this.outTls.client_certificate_path = undefined
        this.outTls.client_key_path = undefined
        this.inTls.client_certificate = []
        this.outTls.client_certificate = []
        this.outTls.client_key = []
      }
    },
    updateData(id: number) {
      if (id > 0) {
        const newData = <tls>JSON.parse(this.$props.data)
        this.tls = newData
        if (this.tls.server == null) this.tls.server = { enabled: true }
        if (this.tls.client == null) this.tls.client = {}
        this.tlsType = newData.server?.reality == undefined ? 0 : 1
        this.usePath = newData.server?.key == undefined ? 0 : 1
        // A stored config that carries the certificates inline should open on
        // the text side of the toggle.
        this.useClientPath = (newData.server?.client_certificate == undefined
          && newData.client?.client_certificate == undefined
          && newData.client?.client_key == undefined) ? 0 : 1
        this.title = "edit"
      }
      else {
        this.tls = <tls>{ id: 0, name: '', server: {enabled: true}, client: {} }
        this.tlsType = 0
        this.usePath = 0
        this.useClientPath = 0
        this.title = "add"
      }
    },
    changeTlsType(){
      if (this.tlsType) {
        this.tls.server = <iTls>{
          enabled: true,
          reality: { enabled: true, handshake: { server_port: 443 }, short_id: RandomUtil.randomShortId() },
          server_name: ""
        }
        this.tls.client = <oTls>{ reality: { public_key: "" }, utls: defaultOutTls.utls }
      } else {
        this.tls.server = <iTls>{ enabled: true }
        this.tls.client = <oTls>{}
      }
    },
    closeModal() {
      this.updateData(0) // reset
      this.$emit('close')
    },
    async saveChanges() {
      const form = this.$refs.formRef as { validate: () => Promise<unknown> } | undefined
      try {
        await form?.validate()
      } catch {
        return
      }
      this.loading = true
      this.$emit('save', this.tls)
      this.loading = false
    },
    async genSelfSigned(){
      this.loading = true
      const msg = await HttpUtils.get<string[]>('api/keypairs', { k: "tls", o: this.inTls.server_name?? "''" })
      this.loading = false
      if (msg.success) {
        this.inTls.key_path=undefined
        this.inTls.certificate_path=undefined
        this.usePath = 1
        if (msg.obj.length>0){
          let privateKey = <string[]>[]
          let publicKey = <string[]>[]
          let isPrivateKey = false
          let isPublicKey = false

          msg.obj.forEach((line:string) => {
              if (line === "-----BEGIN PRIVATE KEY-----") {
                  isPrivateKey = true
                  isPublicKey = false
                  privateKey.push(line)
              } else if (line === "-----END PRIVATE KEY-----") {
                  isPrivateKey = false
                  privateKey.push(line)
              } else if (line === "-----BEGIN CERTIFICATE-----") {
                  isPublicKey = true
                  isPrivateKey = false
                  publicKey.push(line)
              } else if (line === "-----END CERTIFICATE-----") {
                  isPublicKey = false
                  publicKey.push(line)
              } else if (isPrivateKey) {
                  privateKey.push(line)
              } else if (isPublicKey) {
                  publicKey.push(line)
              }
          })
          this.inTls.key = privateKey?? undefined
          this.inTls.certificate = publicKey?? undefined

        } else {
          push.error({
            message: i18n.global.t('error') + ": " + msg.obj
          })
        }
      }
    },
    async genRealityKey(){
      this.loading = true
      const msg = await HttpUtils.get<string[]>('api/keypairs', { k: "reality" })
      this.loading = false
      if (msg.success) {
        msg.obj.forEach((line:string) => {
          if (this.inTls.reality && this.outTls.reality){
            if (line.startsWith("PrivateKey")){
              this.inTls.reality.private_key = line.substring(12)
            }
            if (line.startsWith("PublicKey")){
              this.outTls.reality.public_key = line.substring(11)
            }
          }
        })
      } else {
        push.error({
          message: i18n.global.t('error') + ": " + msg.obj
        })
      }
    },
    randomSID(){
      this.short_id = RandomUtil.randomShortId().join(',')
    }
  }
}
</script>

<style scoped>
.tls-modal-body {
  max-height: 70vh;
  overflow-y: auto;
  overflow-x: hidden;
}

/* 用 16px 的行间距串起各区块，彻底去掉嵌套卡片 */
.tls-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  /* 滚动条由 .tls-modal-body 承担；表单内容右侧留出空隙，
     避免 ECH 开关、按钮和输入框贴到滚动条上。左侧保持与标题对齐。 */
  padding-inline-end: 24px;
}

.tls-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.tls-form :deep(.ant-form-item) {
  margin-bottom: 0;
}

.tls-divider {
  margin: 0;
}

.tls-subdivider {
  margin: 0;
}

.tls-hint {
  margin-top: 6px;
  font-size: 12px;
  line-height: 18px;
  opacity: 0.65;
}

.tls-subtitle {
  font-size: 13px;
  font-weight: 600;
  line-height: 22px;
  opacity: 0.75;
}

/* 开关组不参与收缩，窄屏时整组换行而不是把“允许不安全”拆成竖排 */
.tls-security-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px 24px;
}

.tls-security-row > * {
  flex: none;
}

.tls-security-row :deep(.ant-space-item) {
  white-space: nowrap;
}

@media (max-width: 575px) {
  .tls-form {
    padding-inline-end: 16px;
  }

  .tls-security-row {
    gap: 12px 16px;
  }
}
</style>

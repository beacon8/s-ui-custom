<template>
  <a-card
    size="small"
    title="Cloudflare Origin CA"
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
      <!-- Either credential works; the API token is the narrower of the two. -->
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('tls.provider.apiToken')"
            style="margin-bottom: 0;"
          >
            <a-input-password
              v-model:value="data.api_token"
              allow-clear
              @change="(e: any) => { if (!e.target?.value) delete data.api_token }"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('tls.provider.originCaKey')"
            style="margin-bottom: 0;"
          >
            <a-input-password
              v-model:value="data.origin_ca_key"
              allow-clear
              @change="(e: any) => { if (!e.target?.value) delete data.origin_ca_key }"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('tls.provider.requestType')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.request_type"
              :options="requestTypes"
              :field-names="{ label: 'title', value: 'value' }"
              allow-clear
              style="width: 100%"
              @change="(v: any) => { if (!v) delete data.request_type }"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('tls.provider.validity')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.requested_validity"
              allow-clear
              style="width: 100%"
              @change="(v: any) => { if (!v) delete data.requested_validity }"
            >
              <a-select-option
                v-for="item in validities"
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
          :md="8"
        >
          <a-form-item
            :label="$t('tls.acme.dataDir')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.data_directory"
              allow-clear
              @change="(e: any) => { if (!e.target?.value) delete data.data_directory }"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row :gutter="[12, 12]">
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
              v-model:value="data.http_client"
              :not-found-content="$t('basic.httpClient.none')"
              allow-clear
              style="width: 100%"
              @change="(v: any) => { if (!v) delete data.http_client }"
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
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { httpClientTags } from '@/plugins/httpClient'
import { certProvider, originCaProvider } from '@/types/tls'

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<certProvider>('data', { required: true })

// The parent holds a union and picks this component with a v-else-if on the
// type, which the template checker cannot follow, so narrow once here. The
// computed returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as originCaProvider)

const requestTypes: { title: string, value: 'origin-rsa' | 'origin-ecc' }[] = [
  { title: 'RSA', value: 'origin-rsa' },
  { title: 'ECC', value: 'origin-ecc' },
]

// The only validities Cloudflare accepts, in days.
const validities: number[] = [7, 30, 90, 365, 730, 1095, 5475]

const httpClients = computed((): string[] => httpClientTags())

const domains = computed({
  get: (): string => data.value.domain ? data.value.domain.join(',') : '',
  set: (v: string) => {
    if (!v.endsWith(',')) {
      data.value.domain = v.length > 0 ? v.split(',') : []
    }
  },
})
</script>

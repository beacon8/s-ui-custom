<template>
  <a-card
    size="small"
    :title="$t('types.hy.realm')"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="16"
      >
        <a-form-item
          :label="$t('types.hy.realmUrl')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="data.server_url"
            placeholder="https://realm.example.com"
          />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.hy.realmId')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.realm_id" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.hy.token')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="token" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="16"
      >
        <a-form-item
          :label="$t('types.hy.stunServers') + ' ' + $t('commaSeparated')"
          style="margin-bottom: 0;"
        >
          <a-input
            v-model:value="stun_servers"
            placeholder="stun.l.google.com:19302"
          />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('rule.ipVer')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="ip_version"
            :options="ipVersions"
            :field-names="{ label: 'title', value: 'value' }"
            style="width: 100%;"
          />
        </a-form-item>
      </a-col>
    </a-row>
    <a-row
      :gutter="[12, 12]"
      align="middle"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-space>
          <a-switch
            v-model:checked="portMapping"
            :disabled="data.ip_version == 6"
          />
          <span>{{ $t('types.hy.portMapping') }}</span>
        </a-space>
      </a-col>
      <template v-if="data.port_mapping">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.hy.timeout')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="portMappingTimeout"
              placeholder="10s"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.hy.lifetime')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="portMappingLifetime"
              placeholder="10m"
            />
          </a-form-item>
        </a-col>
      </template>
    </a-row>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { i18n } from '@/locales'
import { Hysteria2Realm } from '@/types/outbounds'

// The parent owns the realm object and this component edits it in place.
const data = defineModel<Hysteria2Realm>('data', { required: true })

const ipVersions = [
  { title: i18n.global.t('all'), value: 0 },
  { title: 'IPv4', value: 4 },
  { title: 'IPv6', value: 6 },
]

const token = computed({
  get: (): string => data.value.token ?? '',
  set: (v: string) => { data.value.token = v.length > 0 ? v : undefined },
})

// sing-box requires the list, so an empty field keeps an empty one.
const stun_servers = computed({
  get: (): string => data.value.stun_servers?.join(',') ?? '',
  set: (v: string) => { data.value.stun_servers = v.split(',').map(s => s.trim()).filter(s => s.length > 0) },
})

// Port mapping needs IPv4, so restricting the realm to IPv6 turns it off.
const ip_version = computed({
  get: (): number => data.value.ip_version ?? 0,
  set: (v: number) => {
    data.value.ip_version = v > 0 ? v as 4 | 6 : undefined
    if (v == 6) delete data.value.port_mapping
  },
})

const portMapping = computed({
  get: (): boolean => data.value.port_mapping?.enabled ?? false,
  set: (v: boolean) => { data.value.port_mapping = v ? { enabled: true } : undefined },
})

const portMappingTimeout = computed({
  get: (): string => data.value.port_mapping?.timeout ?? '',
  set: (v: string) => { if (data.value.port_mapping) data.value.port_mapping.timeout = v.length > 0 ? v : undefined },
})

const portMappingLifetime = computed({
  get: (): string => data.value.port_mapping?.lifetime ?? '',
  set: (v: string) => { if (data.value.port_mapping) data.value.port_mapping.lifetime = v.length > 0 ? v : undefined },
})
</script>

<template>
  <a-card
    size="small"
    :title="direction != 'out_json' ? 'Snell' : undefined"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.snell.version')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="version"
              :options="versions"
              :field-names="{ label: 'title', value: 'value' }"
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
            :label="$t('types.snell.psk')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.psk">
              <template
                v-if="direction == 'in'"
                #suffix
              >
                <a-button
                  type="text"
                  size="small"
                  @click="generatePsk"
                >
                  <template #icon>
                    <AppIcon name="refresh" />
                  </template>
                </a-button>
              </template>
            </a-input>
          </a-form-item>
        </a-col>
        <a-col
          v-if="direction === 'out'"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.snell.userKey')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.userkey"
              allow-clear
              @change="() => { if (!data.userkey) delete data.userkey }"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <!-- v5 (inbound) and v4 (outbound) carry obfs; v6 carries a mode -->
      <a-row
        v-if="obfsVersion"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.snell.obfsMode')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.obfs_mode"
              :options="obfsModes"
              :field-names="{ label: 'title', value: 'value' }"
              allow-clear
              style="width: 100%"
              @change="(v: any) => { if (v === undefined) delete data.obfs_mode }"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="direction === 'out'"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.snell.obfsHost')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.obfs_host"
              allow-clear
              @change="() => { if (!data.obfs_host) delete data.obfs_host }"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="data.version == 6"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.snell.mode')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.mode"
              :options="modes"
              :field-names="{ label: 'title', value: 'value' }"
              allow-clear
              style="width: 100%"
              @change="(v: any) => { if (v === undefined) delete data.mode }"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="direction === 'out'"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="data.reuse" />
            <span>{{ $t('types.snell.reuse') }}</span>
          </a-space>
        </a-col>
      </a-row>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import RandomUtil from '@/plugins/randomUtil'
import { Inbound } from '@/types/inbounds'
import { Outbound } from '@/types/outbounds'

// One form serves both directions, so it takes the widest shape either side
// can hold: the version numbers differ (v5 in, v4 out) and only the outbound
// side carries the client fields.
interface SnellData {
  version: number
  psk: string
  userkey?: string
  obfs_mode?: 'none' | 'http' | 'tls'
  obfs_host?: string
  mode?: 'default' | 'unshaped' | 'unsafe-raw'
  reuse?: boolean
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound | Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as SnellData)

const props = defineProps<{
  direction: 'in' | 'out' | 'out_json'
}>()

const obfsModes = [
  { title: 'None', value: 'none' },
  { title: 'HTTP', value: 'http' },
  { title: 'TLS', value: 'tls' },
]

const modes = [
  { title: 'Default', value: 'default' },
  { title: 'Unshaped', value: 'unshaped' },
  { title: 'Unsafe Raw', value: 'unsafe-raw' },
]

// Inbounds support v5 and v6, outbounds v4 and v6.
const legacyVersion = computed((): number => props.direction === 'in' ? 5 : 4)

const versions = computed((): { title: string, value: number }[] => [
  { title: 'v' + legacyVersion.value, value: legacyVersion.value },
  { title: 'v6', value: 6 },
])

const obfsVersion = computed((): boolean => data.value.version === legacyVersion.value)

const version = computed({
  get: (): number => data.value.version,
  // The version selects which extra options apply, so drop the ones
  // belonging to the version being left behind.
  set: (v: number) => {
    data.value.version = v
    if (v === 6) {
      delete data.value.obfs_mode
      delete data.value.obfs_host
    } else {
      delete data.value.mode
    }
  },
})

function generatePsk() {
  // sing-box requires 12-255 bytes.
  data.value.psk = RandomUtil.randomSeq(32)
}
</script>

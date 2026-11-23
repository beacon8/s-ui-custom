<template>
  <a-card
    size="small"
    title="SSH"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <template v-if="optionKey">
        <a-row :gutter="[12, 12]">
          <a-col flex="none">
            <a-radio-group
              v-model:value="usePath"
              option-type="button"
              button-style="solid"
            >
              <a-radio-button
                :value="0"
                @click="data.private_key = undefined; data.private_key_path = ''"
              >
                {{ $t('tls.usePath') }}
              </a-radio-button>
              <a-radio-button
                :value="1"
                @click="data.private_key_path = undefined; data.private_key = ''"
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
              :label="$t('tls.keyPath')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="data.private_key_path" />
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
              :label="$t('tls.key')"
              style="margin-bottom: 0;"
            >
              <a-textarea
                v-model:value="data.private_key"
                :rows="3"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="[12, 12]">
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('types.ssh.passphrase')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="data.private_key_passphrase" />
            </a-form-item>
          </a-col>
        </a-row>
      </template>
      <template v-else>
        <a-row :gutter="[12, 12]">
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('types.un')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="data.user" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('types.pw')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="data.password" />
            </a-form-item>
          </a-col>
        </a-row>
      </template>
      <a-row
        v-if="optionHostKey"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.ssh.hostKey')"
            style="margin-bottom: 0;"
          >
            <a-textarea
              v-model:value="host_key"
              :rows="3"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="data.host_key_algorithms != undefined || data.client_version != undefined"
        :gutter="[12, 12]"
      >
        <a-col
          v-if="data.host_key_algorithms != undefined"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.ssh.algorithm') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="algorithms" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="data.client_version != undefined"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.ssh.clientVer')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.client_version" />
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
                <a-switch v-model:checked="optionKey" />
                <span>SSH Key</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionHostKey" />
                <span>{{ $t('types.ssh.hostKey') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionAlgorithms" />
                <span>{{ $t('types.ssh.algorithm') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionVer" />
                <span>{{ $t('types.ssh.clientVer') }}</span>
              </a-space>
            </a-flex>
          </template>
          <a-button>
            {{ $t('types.ssh.options') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Outbound } from '@/types/outbounds'

// Only the SSH outbound fields this form reads or writes. `host_key` holds ''
// between the option being switched on and a value being typed, which is why
// it is not just a list of lines.
interface SshData {
  user?: string
  password?: string
  private_key?: string
  private_key_path?: string
  private_key_passphrase?: string
  host_key?: string[] | ''
  host_key_algorithms?: string[]
  client_version?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as SshData)

const menu = ref(false)
const usePath = ref(0)

const optionKey = computed({
  get: (): boolean => data.value.private_key != undefined || data.value.private_key_path != undefined,
  set: (v: boolean) => {
    usePath.value = 0
    if (v) {
      data.value.private_key_path = ""
      delete data.value.user
      delete data.value.password
    } else {
      delete data.value.private_key_path
      delete data.value.private_key
      delete data.value.private_key_passphrase
    }
  },
})

const optionHostKey = computed({
  get: (): boolean => data.value.host_key != undefined,
  set: (v: boolean) => { data.value.host_key = v ? '' : undefined },
})

const optionAlgorithms = computed({
  get: (): boolean => data.value.host_key_algorithms != undefined,
  set: (v: boolean) => { data.value.host_key_algorithms = v ? [] : undefined },
})

const optionVer = computed({
  get: (): boolean => data.value.client_version != undefined,
  set: (v: boolean) => { data.value.client_version = v ? 'SSH-2.0-OpenSSH_7.4p1' : undefined },
})

const host_key = computed({
  get: (): string => data.value.host_key ? data.value.host_key.join('\n') : '',
  set: (v: string) => { data.value.host_key = v.split('\n') },
})

const algorithms = computed({
  get: (): string => data.value.host_key_algorithms ? data.value.host_key_algorithms.join(',') : '',
  set: (v: string) => { data.value.host_key_algorithms = v.length > 0 ? v.split(',') : undefined },
})
</script>

<template>
  <a-card
    size="small"
    title="ShadowTls"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('version')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="version"
            :disabled="data.id > 0"
            style="width: 100%"
          >
            <a-select-option
              v-for="v in [1, 2, 3]"
              :key="v"
              :value="v"
            >
              {{ v }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col
        v-if="data.password != undefined"
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
      <a-col
        v-if="data.wildcard_sni != undefined"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          label="Wildcard SNI"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.wildcard_sni"
            allow-clear
            style="width: 100%"
          >
            <a-select-option
              v-for="item in ['off', 'authed', 'all']"
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
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.shdwTls.hs')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.handshake.server" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('out.port')"
          style="margin-bottom: 0;"
        >
          <a-input-number
            v-model:value="server_port"
            :min="0"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>
    <Dial :dial="data.handshake" />
    <a-row
      v-if="data.handshake_for_server_name != undefined"
      :gutter="[12, 12]"
      style="margin-top: 12px;"
    >
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('types.shdwTls.addHS')"
          style="margin-bottom: 0;"
        >
          <a-space-compact style="width: 100%">
            <a-input v-model:value="handshake_server" />
            <a-button
              type="primary"
              :disabled="handshake_server == ''"
              @click="addHandshakeServer()"
            >
              <template #icon>
                <AppIcon name="plus" />
              </template>
            </a-button>
          </a-space-compact>
        </a-form-item>
      </a-col>
    </a-row>
    <a-card
      v-for="(value, key) in data.handshake_for_server_name"
      :key="key"
      size="small"
      :title="String(key)"
      style="margin-top: 8px;"
    >
      <template #extra>
        <a-button
          type="text"
          danger
          size="small"
          :aria-label="$t('actions.del')"
          @click="data.handshake_for_server_name ? delete data.handshake_for_server_name[key] : null"
        >
          <template #icon>
            <AppIcon name="delete" />
          </template>
        </a-button>
      </template>
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.shdwTls.hs')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="value.server" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('out.port')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="value.server_port"
              :min="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <Dial :dial="value" />
    </a-card>
  </a-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Inbound } from '@/types/inbounds'
import RandomUtil from '@/plugins/randomUtil'
import Dial from '../Dial.vue'

// The default handshake target and the per-SNI overrides share this shape;
// both are filled in by hand, so they start out empty.
interface ShadowTlsHandshake {
  server?: string
  server_port?: number
}

// Only the ShadowTLS inbound fields this form reads or writes. Which of the
// optional ones apply depends on the version, and the setter below adds and
// removes them as the version changes.
interface ShadowTlsData {
  id: number
  version: 1 | 2 | 3
  password?: string
  handshake: ShadowTlsHandshake
  handshake_for_server_name?: Record<string, ShadowTlsHandshake>
  wildcard_sni?: string
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Inbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as ShadowTlsData)

const handshake_server = ref('')

const version = computed({
  get: (): 1 | 2 | 3 => data.value.version,
  set: (newValue: 1 | 2 | 3) => {
    switch (newValue) {
    case 1:
      delete data.value.password
      delete data.value.handshake_for_server_name
      delete data.value.wildcard_sni
      break
    case 2:
      if (!data.value.password) {
        data.value.password = RandomUtil.randomSeq(16)
      }
      if (!data.value.handshake_for_server_name) {
        data.value.handshake_for_server_name = {}
      }
      delete data.value.wildcard_sni
      break
    case 3:
      delete data.value.password
      if (!data.value.handshake_for_server_name) {
        data.value.handshake_for_server_name = {}
      }
      if (!data.value.wildcard_sni) {
        data.value.wildcard_sni = ""
      }
      break
    }
    data.value.version = newValue
  },
})

const server_port = computed({
  get: (): number => data.value.handshake.server_port ? data.value.handshake.server_port : 443,
  // v-model.number hands back the raw string whenever it cannot be parsed.
  set: (newValue: string | number) => {
    data.value.handshake.server_port = (typeof newValue === 'string' && newValue.length == 0) || newValue == 0 ? 443 : parseInt(String(newValue))
  },
})

// normalise fields for the current version once, instead of inside the getter
version.value = data.value.version

onMounted(() => {
  version.value = data.value.version
})

function addHandshakeServer() {
  const servers = data.value.handshake_for_server_name ??= {}
  servers[handshake_server.value] = {}
  // Clear the input field after adding the server
  handshake_server.value = ''
}
</script>

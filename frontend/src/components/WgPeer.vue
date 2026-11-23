<template>
  <a-row :gutter="[12, 12]">
    <a-col
      :xs="24"
      :sm="16"
    >
      <a-form-item
        :label="$t('types.wg.privKey')"
        style="margin-bottom: 0;"
      >
        <a-input v-model:value="privateKey">
          <template #suffix>
            <AppIcon
              name="key-star"
              style="cursor: pointer;"
              @click="refreshKey"
            />
          </template>
        </a-input>
      </a-form-item>
    </a-col>
    <a-col
      :xs="24"
      :sm="16"
    >
      <a-form-item
        :label="$t('types.wg.pubKey')"
        style="margin-bottom: 0;"
      >
        <a-input v-model:value="publicKey" />
      </a-form-item>
    </a-col>
    <a-col
      :xs="24"
      :sm="16"
    >
      <a-form-item
        :label="$t('types.wg.psk')"
        style="margin-bottom: 0;"
      >
        <a-input v-model:value="data.pre_shared_key" />
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
        :label="$t('out.addr')"
        style="margin-bottom: 0;"
      >
        <a-input v-model:value="address" />
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
          v-model:value="port"
          :min="0"
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
        label="KeepAlive"
        style="margin-bottom: 0;"
      >
        <a-input-number
          v-model:value="keepAlive"
          :min="0"
          :addon-after="$t('date.s')"
          style="width: 100%"
        />
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
    >
      <a-form-item
        :label="$t('types.wg.allowedIp') + ' ' + $t('commaSeparated')"
        style="margin-bottom: 0;"
      >
        <a-input v-model:value="allowed_ips" />
      </a-form-item>
    </a-col>
    <a-col
      :xs="24"
      :sm="12"
    >
      <a-form-item
        :label="'Reserved ' + $t('commaSeparated')"
        style="margin-bottom: 0;"
      >
        <a-input v-model:value="reserved" />
      </a-form-item>
    </a-col>
  </a-row>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { WgPeer } from '@/types/endpoints'

// One entry of the panel's own key store, which keeps the private key beside
// the public key sing-box is configured with.
interface WgKey {
  public_key: string
  private_key: string
}

// The panel-only extras carried beside the endpoint's sing-box options.
interface WgExt {
  keys: WgKey[]
}

// The peer fields this form edits. Emptying one removes it rather than writing
// an empty value, so each is optional here.
interface PeerData {
  address?: string
  port?: number
  public_key: string
  pre_shared_key?: string
  allowed_ips?: string[]
  persistent_keepalive_interval?: number
  reserved?: number[]
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<WgPeer>('data', { required: true })

// An emptied field is removed, which the declared peer shape does not allow,
// so narrow once here. The computed returns the same object, so edits still
// reach the parent.
const data = computed(() => model.value as unknown as PeerData)

const props = defineProps<{
  ext?: WgExt
}>()

const emit = defineEmits<{
  refreshPeerKey: []
}>()

const allowed_ips = computed({
  get: () => data.value.allowed_ips?.join(','),
  set: (v: string) => { data.value.allowed_ips = v.length > 0 ? v.split(',') : undefined },
})

const reserved = computed({
  get: () => data.value.reserved?.join(','),
  set: (v: string) => {
    if (!v.endsWith(',')) {
      data.value.reserved = v.length > 0 ? v.split(',').map(str => parseInt(str, 10)) : undefined
    }
  },
})

const address = computed({
  get: () => data.value.address,
  set: (v: string) => { data.value.address = v.length > 0 ? v : undefined },
})

const port = computed({
  get: () => data.value.port,
  set: (v: number) => { data.value.port = v > 0 ? v : undefined },
})

const keepAlive = computed({
  get: () => data.value.persistent_keepalive_interval ?? 0,
  set: (v: number) => { data.value.persistent_keepalive_interval = v > 0 ? v : undefined },
})

// The private key lives in the endpoint's key store, found by the public key
// the peer is configured with.
const privateKey = computed({
  get: (): string => {
    const keys = props.ext?.keys ?? []
    const indexKeys = keys.findIndex(key => key.public_key == data.value.public_key)
    return indexKeys > -1 ? keys[indexKeys].private_key : ''
  },
  set: (v: string) => {
    const keys = props.ext?.keys ?? []
    const indexKeys = keys.findIndex(key => key.public_key == data.value.public_key)
    // The getter above already allows for a peer the key store does not know,
    // such as one added outside the panel. There is nowhere to put a private
    // key for it, so do not write one rather than throwing on keys[-1].
    if (indexKeys > -1) keys[indexKeys].private_key = v
  },
})

const publicKey = computed({
  get: () => data.value.public_key,
  set: (v: string) => {
    const keys = props.ext?.keys ?? []
    const indexKeys = keys.findIndex(key => key.public_key == data.value.public_key)
    // Same as above: a peer the key store does not know still gets its own
    // public key updated, there is just no stored pair to keep in step.
    if (indexKeys > -1) keys[indexKeys].public_key = v
    data.value.public_key = v
  },
})

function refreshKey() {
  emit('refreshPeerKey')
}
</script>

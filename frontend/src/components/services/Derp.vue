<template>
  <a-card
    size="small"
    title="DERP"
  >
    <a-flex
      vertical
      :gap="12"
    >
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.derp.configPath')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.config_path" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionHome"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('pages.home')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="data.home"
              placeholder="blank | http[s]://example.com:port/path"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionVerifyCE"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('types.derp.verifyClientEndpoint')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="data.verify_client_endpoint"
              mode="multiple"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in tsTags"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
      <template v-if="optionVerifyCU">
        <a-flex
          justify="space-between"
          align="center"
        >
          <strong>{{ $t('types.derp.verifyClientUrl') }}</strong>
          <a-button
            type="primary"
            size="small"
            shape="circle"
            @click="data.verify_client_url?.push({url: ''})"
          >
            <template #icon>
              <AppIcon name="plus" />
            </template>
          </a-button>
        </a-flex>
        <a-card
          v-for="(clientUrl, index) in data.verify_client_url"
          :key="index"
          size="small"
        >
          <a-row
            :gutter="[12, 12]"
            align="middle"
          >
            <a-col flex="none">
              <a-button
                type="text"
                danger
                shape="circle"
                :aria-label="$t('actions.del')"
                @click="data.verify_client_url?.splice(index, 1)"
              >
                <template #icon>
                  <AppIcon name="delete" />
                </template>
              </a-button>
            </a-col>
            <a-col :flex="1">
              <a-flex
                vertical
                :gap="12"
              >
                <a-form-item
                  :label="$t('types.derp.verifyClientUrl')"
                  style="margin-bottom: 0;"
                >
                  <a-input v-model:value="clientUrl.url" />
                </a-form-item>
                <Dial :dial="clientUrl" />
              </a-flex>
            </a-col>
          </a-row>
        </a-card>
      </template>
      <template v-if="optionMesh">
        <a-flex
          justify="space-between"
          align="center"
        >
          <strong>{{ $t('types.derp.meshWith') }}</strong>
          <a-button
            type="primary"
            size="small"
            shape="circle"
            @click="data.mesh_with?.push({tls: {}})"
          >
            <template #icon>
              <AppIcon name="plus" />
            </template>
          </a-button>
        </a-flex>
        <a-card
          v-for="(mesh, index) in data.mesh_with"
          :key="index"
          size="small"
        >
          <a-row
            :gutter="[12, 12]"
            align="middle"
          >
            <a-col flex="none">
              <a-button
                type="text"
                danger
                shape="circle"
                :aria-label="$t('actions.del')"
                @click="data.mesh_with?.splice(index, 1)"
              >
                <template #icon>
                  <AppIcon name="delete" />
                </template>
              </a-button>
            </a-col>
            <a-col :flex="1">
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
                      :label="$t('out.addr')"
                      style="margin-bottom: 0;"
                    >
                      <a-input v-model:value="mesh.server" />
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
                        v-model:value="mesh.server_port"
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
                      :label="$t('transport.host')"
                      style="margin-bottom: 0;"
                    >
                      <a-input v-model:value="mesh.host" />
                    </a-form-item>
                  </a-col>
                </a-row>
                <Dial :dial="mesh" />
                <OutTLS :outbound="mesh" />
              </a-flex>
            </a-col>
          </a-row>
        </a-card>
        <a-row :gutter="[12, 12]">
          <a-col flex="none">
            <a-radio-group
              v-model:value="usePskText"
              option-type="button"
              button-style="solid"
            >
              <a-radio-button
                :value="0"
                @click="delete data.mesh_psk_file"
              >
                {{ $t('types.derp.meshPsk') }}
              </a-radio-button>
              <a-radio-button
                :value="1"
                @click="delete data.mesh_psk"
              >
                {{ $t('types.derp.meshPskFile') }}
              </a-radio-button>
            </a-radio-group>
          </a-col>
        </a-row>
        <a-row
          v-if="usePskText == 1"
          :gutter="[12, 12]"
        >
          <a-col :span="24">
            <a-form-item
              :label="$t('types.derp.meshPskFile')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="data.mesh_psk_file" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-else
          :gutter="[12, 12]"
        >
          <a-col :span="24">
            <a-form-item
              :label="$t('types.derp.meshPsk')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="data.mesh_psk" />
            </a-form-item>
          </a-col>
        </a-row>
      </template>
      <template v-if="optionStun">
        <a-card
          size="small"
          :title="$t('types.derp.stun')"
        >
          <Listen
            :data="stun"
            :in-tags="inTags"
          />
        </a-card>
      </template>
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
                <a-switch v-model:checked="optionVerifyCE" />
                <span>{{ $t('types.derp.verifyClientEndpoint') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionVerifyCU" />
                <span>{{ $t('types.derp.verifyClientUrl') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionHome" />
                <span>{{ $t('pages.home') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionMesh" />
                <span>{{ $t('types.derp.meshWith') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionStun" />
                <span>{{ $t('types.derp.stun') }}</span>
              </a-space>
            </a-flex>
          </template>
          <a-button>
            {{ $t('types.derp.options') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Dial from '@/components/Dial.vue'
import OutTLS from '../tls/OutTLS.vue'
import Listen from '../Listen.vue'
import { Dial as DialOptions } from '@/types/dial'
import { Listen as ListenOptions } from '@/types/inbounds'
import { oTls } from '@/types/tls'
import { Srv } from '@/types/services'

// A client the DERP server verifies over HTTP; Dial.vue edits the dial fields.
interface VerifyClientUrl extends DialOptions {
  url: string
}

// A peer this server meshes with; Dial.vue and OutTLS.vue edit the rest.
interface MeshNode extends DialOptions {
  server?: string
  server_port?: number
  host?: string
  tls?: oTls
}

// The embedded STUN server; Listen.vue edits the listen fields inside it.
interface StunOptions extends Partial<ListenOptions> {
  enabled?: boolean
}

// Only the derp fields this form reads or writes.
interface DerpData {
  config_path?: string
  home?: string
  verify_client_endpoint?: string[]
  verify_client_url?: VerifyClientUrl[]
  mesh_with?: MeshNode[]
  mesh_psk?: string
  mesh_psk_file?: string
  stun?: StunOptions
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Srv>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as DerpData)

defineProps<{
  tsTags: string[]
  inTags: string[]
}>()

// Listen.vue edits the listen options of whatever object it is given, and the
// STUN section only renders once stun exists.
const stun = computed(() => data.value.stun as unknown as Srv)

const menu = ref(false)
const usePskText = ref(data.value.mesh_psk == undefined ? 1 : 0)

const optionVerifyCE = computed({
  get: (): boolean => data.value.verify_client_endpoint != undefined,
  set: (v: boolean) => { data.value.verify_client_endpoint = v ? [] : undefined },
})

const optionVerifyCU = computed({
  get: (): boolean => data.value.verify_client_url != undefined,
  set: (v: boolean) => { data.value.verify_client_url = v ? [{ url: '' }] : undefined },
})

const optionHome = computed({
  get: (): boolean => data.value.home != undefined,
  set: (v: boolean) => { data.value.home = v ? '' : undefined },
})

const optionMesh = computed({
  get: (): boolean => data.value.mesh_with != undefined,
  set: (v: boolean) => {
    if (v) {
      data.value.mesh_with = [{ tls: {} }]
      delete data.value.mesh_psk_file
      data.value.mesh_psk = ''
    } else {
      delete data.value.mesh_with
      delete data.value.mesh_psk_file
      delete data.value.mesh_psk
    }
  },
})

const optionStun = computed({
  get: (): boolean => data.value.stun != undefined,
  set: (v: boolean) => { data.value.stun = v ? { enabled: true } : undefined },
})
</script>

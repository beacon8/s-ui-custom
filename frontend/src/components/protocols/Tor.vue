<template>
  <a-card
    size="small"
    title="Tor"
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
            :label="$t('types.tor.execPath')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.executable_path" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.tor.dataDir')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="data.data_directory" />
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
            :label="$t('types.tor.extArgs') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="extra_args" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-space>
        <span>Torrc</span>
        <a-button
          type="primary"
          size="small"
          shape="circle"
          @click="add_torrc_option"
        >
          <template #icon>
            <AppIcon name="plus" />
          </template>
        </a-button>
      </a-space>
      <a-row
        v-for="(torrc, index) in torrc_options"
        :key="index"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col flex="none">
          <a-button
            type="text"
            danger
            shape="circle"
            :aria-label="$t('actions.del')"
            @click="del_torrc_option(index)"
          >
            <template #icon>
              <AppIcon name="delete" />
            </template>
          </a-button>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('objects.key')"
            style="margin-bottom: 0;"
          >
            <a-input
              :value="torrc.name"
              @update:value="(v: string) => update_key(index, v)"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('objects.value')"
            style="margin-bottom: 0;"
          >
            <a-input
              :value="torrc.value"
              @update:value="(v: string) => update_value(index, v)"
            />
          </a-form-item>
        </a-col>
      </a-row>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Outbound } from '@/types/outbounds'

type torrc_option = {
  name: string
  value: string
}

// Only the Tor outbound fields this form reads or writes. A torrc key may be
// repeated, so its value is either one string or a list of them.
interface TorData {
  executable_path?: string
  data_directory?: string
  extra_args?: string[]
  torrc?: Record<string, string | string[]>
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Outbound>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as TorData)

const torrc_options = computed({
  get: (): torrc_option[] => {
    const options: torrc_option[] = []
    const h = data.value.torrc
    if (h) {
      Object.keys(h).forEach(key => {
        const value = h[key]
        if (Array.isArray(value)){
          value.forEach((v:string) => options.push({ name: key, value: v }))
        } else {
          options.push({ name: key, value: value })
        }
      })
    }
    return options
  },
  set: (v:torrc_option[]) => {
    if (v.length>0) {
      const torrc: Record<string, string | string[]> = {}
      v.forEach((h:torrc_option) => {
        const current = torrc[h.name]
        if (current) {
          if (Array.isArray(current)) {
            current.push(h.value)
          } else {
            torrc[h.name] = [current, h.value]
          }
        } else {
          torrc[h.name] = h.value
        }
      })
      data.value.torrc = torrc
    } else {
      data.value.torrc = undefined
    }
  },
})

const extra_args = computed({
  get: (): string | undefined => data.value.extra_args?.join(','),
  set: (v:string) => { data.value.extra_args = v.length > 0 ? v.split(',') : undefined },
})

function add_torrc_option() {
  torrc_options.value = [...torrc_options.value, {name: "", value: ""}]
}

function del_torrc_option(i:number) {
  const h = torrc_options.value
  h.splice(i,1)
  torrc_options.value = h
}

function update_key(i:number,k:string) {
  const h = torrc_options.value
  h[i].name = k
  torrc_options.value = h
}

function update_value(i:number,v:string) {
  const h = torrc_options.value
  h[i].value = v
  torrc_options.value = h
}
</script>

<template>
  <a-card
    size="small"
    title="Shadowsocks API"
  >
    <template #extra>
      <a-button
        type="primary"
        size="small"
        shape="circle"
        @click="add_server"
      >
        <template #icon>
          <AppIcon name="plus" />
        </template>
      </a-button>
    </template>
    <a-flex
      vertical
      :gap="12"
    >
      <a-row
        v-for="(server, index) in servers"
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
            @click="del_server(index)"
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
            :label="$t('transport.path')"
            style="margin-bottom: 0;"
          >
            <a-input
              :value="server.name"
              @input="(e: any) => update_key(index, e.target?.value ?? '')"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('objects.inbound')"
            style="margin-bottom: 0;"
          >
            <a-select
              :value="server.value"
              style="width: 100%"
              @change="(v: any) => update_value(index, String(v ?? ''))"
            >
              <a-select-option
                v-for="item in ssTags"
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
import { Srv } from '@/types/services'

// One row of the server editor: a path and the shadowsocks inbound tag it maps
// to. The stored form keys by path and allows either one tag or a list, which
// is why the getter flattens and the setter regroups.
type Server = {
  name: string
  value: string
}

// Only the SSM-API service field this form reads or writes.
interface SsmApiData {
  servers?: Record<string, string | string[]>
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Srv>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as SsmApiData)

const props = defineProps<{
  ssTags: string[]
}>()

const servers = computed({
  get: (): Server[] => {
    const servers: Server[] = []
    const h = data.value.servers
    if (h) {
      Object.keys(h).forEach(key => {
        const entry = h[key]
        if (Array.isArray(entry)){
          entry.forEach(v => servers.push({ name: key, value: v }))
        } else {
          servers.push({ name: key, value: entry })
        }
      })
    }
    return servers
  },
  set: (v: Server[]) => {
    if (v.length>0) {
      const servers: Record<string, string | string[]> = {}
      v.forEach(h => {
        const existing = servers[h.name]
        if (existing) {
          if (Array.isArray(existing)) {
            existing.push(h.value)
          } else {
            servers[h.name] = [existing, h.value]
          }
        } else {
          servers[h.name] = h.value
        }
      })
      data.value.servers = servers
    } else {
      data.value.servers = undefined
    }
  }
})

function add_server() {
  servers.value = [...servers.value, {name: "/ss" + servers.value.length, value: props.ssTags[0] || ""}]
}

function del_server(i: number) {
  const h = servers.value
  h.splice(i,1)
  servers.value = h
}

function update_key(i: number, k: string) {
  const h = servers.value
  h[i].name = k
  servers.value = h
}

function update_value(i: number, v: string) {
  const h = servers.value
  h[i].value = v
  servers.value = h
}
</script>

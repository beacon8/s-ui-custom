<template>
  <a-card size="small">
    <template #title>
      <a-space>
        <span>{{ $t('objects.headers') }}</span>
        <a-button
          type="primary"
          size="small"
          shape="circle"
          @click="add_header"
        >
          <template #icon>
            <AppIcon name="plus" />
          </template>
        </a-button>
      </a-space>
    </template>
    <a-flex
      vertical
      gap="small"
    >
      <a-row
        v-for="(header, index) in hdrs"
        :key="index"
        :gutter="[12, 12]"
      >
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
              :value="header.name"
              @update:value="(val: string) => update_key(index, val)"
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
              :value="header.value"
              @update:value="(val: string) => update_value(index, val)"
            >
              <template #suffix>
                <AppIcon
                  :aria-label="$t('actions.del')"
                  name="delete"
                  color="error"
                  style="cursor: pointer;"
                  @click="del_header(index)"
                />
              </template>
            </a-input>
          </a-form-item>
        </a-col>
      </a-row>
    </a-flex>
  </a-card>
</template>

<script lang="ts">
import { PropType } from 'vue'

type Header = {
  name: string
  value: string
}

// The object that owns the headers map. sing-box stores one value per key, or a
// list when a key repeats, which is what the flat rows below are folded into.
interface HeadersData {
  headers?: Record<string, string | string[]>
}

export default {
  props: {
    // Transports, protocols, DNS servers and HTTP clients all bind here, so the
    // prop asks only for an object and the narrowing happens below.
    data: { type: Object as PropType<object>, required: true }
  },
  data() {
    return {}
  },
  computed: {
    // Narrow once, since no caller hands over a type the template checker could
    // follow. The computed returns the same object, so edits reach the parent.
    target(): HeadersData { return <HeadersData>this.$props.data },
    hdrs: {
      get() :Header[] {
        let headers: Header[] = []
        const h = this.target.headers
        if (h) {
          Object.keys(h).forEach(key => {
            const value = h[key]
            if (Array.isArray(value)){
              value.forEach((v:string) => headers.push({ name: key, value: v }))
            } else {
              headers.push({ name: key, value: value })
            }
          })
        }
        return headers
      },
      set(v:Header[]) {
        if (v.length>0) {
          let headers: Record<string, string | string[]> = {}
          v.forEach((h:Header) => {
            const current = headers[h.name]
            if (current) {
              if (Array.isArray(current)) {
                current.push(h.value)
              } else {
                headers[h.name] = [current, h.value]
              }
            } else {
              headers[h.name] = h.value
            }
          })
          this.target.headers = headers
        } else {
          this.target.headers = undefined
        }
      }
    }
  },
  methods: {
    add_header() {
      this.hdrs = [...this.hdrs, {name: "Host", value: ""}]
    },
    del_header(i:number) {
      let h = this.hdrs
      h.splice(i,1)
      this.hdrs = h
    },
    update_key(i:number,k:string) {
      let h = this.hdrs
      h[i].name = k
      this.hdrs = h
    },
    update_value(i:number,v:string) {
      let h = this.hdrs
      h[i].value = v
      this.hdrs = h
    },
  }
}
</script>

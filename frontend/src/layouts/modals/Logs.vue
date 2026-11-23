<template>
  <a-modal
    :open="visible"
    width="90%"
    :style="{ maxWidth: '1200px' }"
    :title="$t('basic.log.title')"
    :footer="null"
    @cancel="model.visible = false"
  >
    <a-form layout="vertical">
      <a-row
        :gutter="[12, 12]"
        align="bottom"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('basic.log.level')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="logLevel"
              :options="logLevels"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
              @change="loadData"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('count')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="logCount"
              style="width: 100%"
              @change="loadData"
            >
              <a-select-option
                v-for="c in [10, 20, 30, 50, 100]"
                :key="c"
                :value="c"
              >
                {{ c }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col flex="none">
          <a-button
            type="default"
            :aria-label="$t('actions.update')"
            :loading="loading"
            @click="loadData"
          >
            <template #icon>
              <AppIcon name="refresh" />
            </template>
          </a-button>
        </a-col>
      </a-row>
    </a-form>
    <!-- Text, not v-html. These are sing-box log lines, and they quote
         values that come from whoever connected: a hostname, an SNI, a
         URL. Rendering them as markup let anyone who could reach a proxy
         port put script into the operator's browser. -->
    <a-card
      size="small"
      style="margin-top: 12px;"
      dir="ltr"
    >
      <pre class="log-lines">{{ lines.join('\n') }}</pre>
    </a-card>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import HttpUtils from '@/plugins/httputil'

// The dialog handle the parent owns and this modal closes.
interface ModalControl {
  visible: boolean
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<ModalControl>('control', { required: true })

const props = defineProps<{
  visible: boolean
}>()

const loading = ref(false)
const lines = ref<string[]>([])
const logLevel = ref('info')
const logLevels = [
  { title: 'DEBUG', value: 'debug' },
  { title: 'INFO', value: 'info' },
  { title: 'WARNING', value: 'warning' },
  { title: 'ERROR', value: 'err' },
]
const logCount = ref(10)

watch(() => props.visible, (v) => {
  lines.value = []
  logLevel.value = 'info'
  logCount.value = 10
  if (v) {
    loadData()
  }
})

async function loadData() {
  loading.value = true
  const data = await HttpUtils.get<string[]>('api/logs',{ c: logCount.value, l: logLevel.value })
  if (data.success) {
    lines.value = data.obj?? []
    loading.value = false
  }
}
</script>

<style scoped>
.log-lines {
  margin: 0;
  padding: .5rem;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-family: monospace;
  font-size: .8rem;
}
</style>

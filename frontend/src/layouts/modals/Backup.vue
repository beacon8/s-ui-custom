<template>
  <a-modal
    :open="visible"
    :width="500"
    :title="$t('main.backup.title')"
    :footer="null"
    @cancel="model.visible = false"
  >
    <a-row :gutter="[12, 12]">
      <a-col :span="24">
        <a-checkbox-group v-model:value="exclude">
          <a-space wrap>
            <a-checkbox value="stats">
              {{ $t('main.backup.exclStats') }}
            </a-checkbox>
            <a-checkbox value="changes">
              {{ $t('main.backup.exclChanges') }}
            </a-checkbox>
          </a-space>
        </a-checkbox-group>
      </a-col>
    </a-row>
    <a-flex
      justify="space-between"
      align="center"
      style="margin-top: 16px;"
    >
      <a-button
        type="primary"
        @click="backup()"
      >
        {{ $t('main.backup.backup') }}
      </a-button>
      <a-button
        type="primary"
        @click="restore()"
      >
        {{ $t('main.backup.restore') }}
      </a-button>
    </a-flex>
    <a-divider />
    <a-button
      type="primary"
      @click="config()"
    >
      {{ $t('main.backup.sbConfig') }}
    </a-button>
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

const exclude = ref<string[]>(["stats", "changes"])

watch(() => props.visible, (v) => {
  if (v) {
    exclude.value = ["stats", "changes"]
  }
})

function backup() {
  const excludeOption = exclude.value.length>0 ? '?exclude=' +exclude.value.join(',') : ''
  window.location.href = 'api/getdb' + excludeOption
}

function config() {
  window.location.href = 'api/singbox-config'
}

function restore() {
  const fileInput = document.createElement('input')
  fileInput.type = 'file'
  fileInput.accept = '.db'

  fileInput.addEventListener('change', async (event: Event) => {
    const inputElement = event.target as HTMLInputElement
    const dbFile = inputElement.files ? inputElement.files[0] : null

    if (dbFile) {
      const formData = new FormData()
      formData.append('db', dbFile)

      model.value.visible = false

      const uploadMsg = await HttpUtils.post('api/importdb', formData, {
          headers: {
              'Content-Type': 'multipart/form-data',
          },
      })

      if (uploadMsg.success) {
        await new Promise(resolve => setTimeout(resolve, 1000))
        location.reload()
      }
    }
  })

  fileInput.click()
}
</script>

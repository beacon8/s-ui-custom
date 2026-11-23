<template>
  <a-modal
    :open="visible"
    :title="title"
    :width="800"
    @cancel="closeModal"
  >
    <div style="max-height: 70vh; overflow-y: auto;">
      <div class="code-editor">
        <div class="line-numbers">
          <span
            v-for="n in lineCount"
            :key="n"
          >{{ n }}</span>
        </div>
        <a-textarea
          ref="textareaRef"
          v-model:value="content"
          :auto-size="{ minRows: 10 }"
          :bordered="false"
          style="font-family: monospace; line-height: 1.5; font-size: 14px; resize: none; white-space: pre; overflow-x: auto;"
          @scroll="syncScroll"
        />
      </div>
    </div>
    <template #footer>
      <a-button
        type="default"
        @click="closeModal"
      >
        {{ $t('actions.close') }}
      </a-button>
      <a-button
        type="primary"
        @click="saveChanges"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import { isDarkTheme } from '@/composables/useThemeSwitcher'

export default {
  props: {
    visible: { type: Boolean, required: true },
    data: { type: String, required: true },
    title: { type: String, required: true }
  },
  emits: ['close', 'save'],
  data() {
    return {
      content: this.$props.data,
    }
  },
  computed: {
    lineCount() {
      return this.content?.split('\n').length
    },
    borderColor() {
      return isDarkTheme.value ? '#424242' : '#d9d9d9'
    },
    surfaceColor() {
      return isDarkTheme.value ? '#1f1f1f' : '#fafafa'
    }
  },
  watch: {
    visible(v) {
      if (v) {
        this.content = this.$props.data
      }
    }
  },
  methods: {
    syncScroll() {
      const textarea = document.querySelector('.code-editor textarea') as HTMLTextAreaElement | null
      const lineNumbers = textarea?.closest('.code-editor')?.querySelector('.line-numbers')
      if (lineNumbers && textarea) {
        lineNumbers.scrollTop = textarea.scrollTop
      }
    },
    closeModal() {
      this.$emit('close')
    },
    saveChanges() {
      this.$emit('save', this.content)
    }
  }
}
</script>

<style scoped>
.code-editor {
  direction: ltr !important;
  display: flex;
  border: 1px solid v-bind(borderColor);
  border-radius: 4px;
  overflow: hidden;
  font-size: 14px;
}

.line-numbers {
  width: 40px;
  background: v-bind(surfaceColor);
  text-align: right;
  padding: 4px 8px 4px 4px;
  line-height: 1.5;
  overflow-y: hidden;
  user-select: none;
  display: flex;
  flex-direction: column;
}

.line-numbers span {
  display: block;
  line-height: 1.5;
  height: 1.5em;
  font-family: monospace;
}
</style>

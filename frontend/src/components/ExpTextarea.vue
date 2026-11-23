<template>
  <a-modal
    :open="visible"
    :title="label"
    :width="620"
    @cancel="closeModal"
  >
    <div style="margin-bottom: 12px;">
      {{ $t('rule.etaHint') }}
    </div>
    <a-textarea
      v-model:value="localText"
      :placeholder="label"
      :rows="16"
      spellcheck="false"
    />
    <div style="text-align: end; margin-top: 4px; opacity: 0.65; font-size: 12px;">
      {{ lineCount }} / {{ $t('count') }}
    </div>
    <template #footer>
      <a-flex
        justify="space-between"
        align="center"
      >
        <a-button
          danger
          type="text"
          @click="resetChanges"
        >
          {{ $t('reset') }}
        </a-button>
        <a-space>
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
        </a-space>
      </a-flex>
    </template>
  </a-modal>
</template>

<script lang="ts">

export default {
  props: {
    visible: { type: Boolean, required: true },
    label: { type: String, required: true },
    content: { type: String, required: true }
  },
  emits: ['update', 'close'],
  data() {
    return {
      dialog: false,
      localText: '',
    }
  },
  computed: {
    lineCount(): number {
      return this.localText.split('\n').filter((l: string) => l.trim().length > 0).length
    },
  },
  watch: {
    visible(v) {
      if (v) {
        this.localText = this.content
      }
    },
  },
  methods: {
    saveChanges() {
      const unique = [
        ...new Set(
          this.localText
            .split('\n')
            .map((l: string) => l.trim())
            .filter((l: string) => l.length > 0)
        ),
      ]
      this.$emit('update', unique)
      this.dialog = false
    },
    resetChanges() {
      this.localText = this.content
    },
    closeModal() {
      this.$emit('close')
    },
  },
}
</script>

<template>
  <a-col flex="none">
    <a-flex
      class="url-test-row"
      align="center"
      :gap="16"
      wrap
    >
      <a-space :size="8">
        <a-switch v-model:checked="enabled" />
        <span class="url-test-label">{{ $t('urlTest.custom') }}</span>
      </a-space>
      <a-flex
        v-if="enabled"
        class="url-test-url"
        align="center"
        :gap="8"
      >
        <span class="url-test-label">{{ $t('urlTest.url') }}</span>
        <a-input
          v-model:value.trim="url"
          class="url-test-input"
          :status="url != '' && !valid ? 'error' : undefined"
          placeholder="https://www.gstatic.com/generate_204"
        />
        <a-tooltip
          v-if="url != '' && !valid"
          :title="$t('urlTest.invalid')"
        >
          <AppIcon
            name="alert-circle-outline"
            class="url-test-error"
          />
        </a-tooltip>
      </a-flex>
    </a-flex>
  </a-col>
</template>

<script setup lang="ts">
import { useUrlTestTarget } from '@/composables/useUrlTestTarget'

const { enabled, url, valid } = useUrlTestTarget()
</script>

<style scoped>
.url-test-label {
  white-space: nowrap;
}

/* 开关组不收缩：窄屏整组换行，避免“自定义测试 URL”被挤成多行 */
.url-test-row > * {
  flex: none;
}

.url-test-input {
  width: clamp(180px, 30vw, 320px);
}

/* 窄屏改为上下两行：开关组一行，URL 输入一行，避免整组溢出 */
@media (max-width: 575px) {
  .url-test-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}

.url-test-error {
  color: var(--ant-color-error, #ff4d4f);
}
</style>

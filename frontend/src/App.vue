<template>
  <a-config-provider
    :theme="antdThemeConfig"
    :locale="antdLocale"
    :direction="isRtl ? 'rtl' : 'ltr'"
  >
    <a-spin
      :spinning="Boolean(loading)"
      :tip="$t('loading')"
      size="large"
    >
      <Message />
      <router-view />
    </a-spin>
  </a-config-provider>
</template>

<script lang="ts" setup>
import Message from '@/components/message.vue'
import { inject, ref, Ref } from 'vue'
import { antdThemeConfig, antdLocale, isRtl } from '@/composables/useThemeSwitcher'

const loading: Ref = inject('loading') ?? ref(false)

// Change page title
document.title = "S-UI " + document.location.hostname
</script>

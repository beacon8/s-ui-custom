<template>
  <a-layout-header style="padding: 0 16px; display: flex; align-items: center; justify-content: space-between;">
    <a-button
      v-if="isMobile"
      type="text"
      :aria-label="$t('a11y.menu')"
      @click="$emit('toggleDrawer')"
    >
      <template #icon>
        <AppIcon name="menu" />
      </template>
    </a-button>
    <span
      v-else
      style="width: 0;"
    />

    <a-typography-title
      :level="4"
      :style="{ margin: 0, flex: 1, textAlign: titleAlign }"
    >
      {{ $t(<string>route.name) }}
    </a-typography-title>

    <a-space align="center">
      <a-tooltip
        v-if="maintenance"
        :title="$t('setting.maintenanceOnHint')"
      >
        <a-tag color="warning">
          <template #icon>
            <AppIcon name="wrench" />
          </template>
          {{ $t('setting.maintenance') }}
        </a-tag>
      </a-tooltip>

      <a-tooltip
        v-if="update"
        :title="$t('update.hint', { current: update.current })"
      >
        <a-tag
          color="processing"
          style="cursor: pointer;"
          @click="Data().updateVisible = true"
        >
          <template #icon>
            <AppIcon name="update" />
          </template>
          {{ $t('update.available', { version: update.latest }) }}
        </a-tag>
      </a-tooltip>

      <a-tooltip :title="$t('donate')">
        <a-button
          type="text"
          shape="circle"
          :aria-label="$t('donate')"
          href="https://donate.alireza0.dev"
          target="_blank"
          rel="noopener"
        >
          <template #icon>
            <AppIcon
              name="heart"
              color="red"
            />
          </template>
        </a-button>
      </a-tooltip>

      <a-dropdown trigger="click">
        <a-button
          type="text"
          shape="circle"
          :aria-label="$t('a11y.language')"
        >
          <template #icon>
            <AppIcon name="translate" />
          </template>
        </a-button>
        <template #overlay>
          <a-menu :selected-keys="[i18nLocale]">
            <a-menu-item
              v-for="lang in languages"
              :key="lang.value"
              @click="changeLocale(lang.value)"
            >
              {{ lang.title }}
            </a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>

      <a-dropdown trigger="click">
        <a-button
          type="text"
          shape="circle"
          :aria-label="$t('a11y.theme')"
        >
          <template #icon>
            <AppIcon name="theme-light-dark" />
          </template>
        </a-button>
        <template #overlay>
          <a-menu :selected-keys="[activeThemeKey]">
            <a-menu-item
              v-for="th in themes"
              :key="th.value"
              @click="changeTheme(th.value)"
            >
              <a-space>
                <AppIcon :name="th.icon" />
                <span>{{ $t(`theme.${th.value}`) }}</span>
              </a-space>
            </a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
    </a-space>
  </a-layout-header>
</template>

<script lang="ts" setup>
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { languages } from '@/locales'
import { useThemeSwitcher } from '@/composables/useThemeSwitcher'
import { computed } from 'vue'
import Data from '@/store/modules/data'

defineProps<{ isMobile: boolean }>()
defineEmits<{ toggleDrawer: [] }>()

const route = useRoute()
const { locale: i18nLocale } = useI18n()
const { themes, changeTheme, isActiveTheme } = useThemeSwitcher()

// Pages can opt into a start-aligned title; every other page stays centered.
const titleAlign = computed((): 'start' | 'center' =>
  String(route.meta.titleAlign ?? '') === 'start' ? 'start' : 'center'
)

const activeThemeKey = computed(() => themes.find(t => isActiveTheme(t.value))?.value ?? 'system')

const changeLocale = (l: string) => {
  i18nLocale.value = l
  localStorage.setItem('locale', l)
  window.location.reload()
}
const maintenance = computed((): boolean => Data().maintenance)
const update = computed(() => Data().update)
</script>

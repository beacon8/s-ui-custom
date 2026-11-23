import { computed, ref } from 'vue'
import { theme as antdTheme } from 'ant-design-vue'
import enUS from 'ant-design-vue/es/locale/en_US'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import zhTW from 'ant-design-vue/es/locale/zh_TW'
import faIR from 'ant-design-vue/es/locale/fa_IR'
import ruRU from 'ant-design-vue/es/locale/ru_RU'
import viVN from 'ant-design-vue/es/locale/vi_VN'
import { i18n } from '@/locales'

export const themes = [
  { value: 'light', icon: 'sun' },
  { value: 'dark', icon: 'moon' },
  { value: 'system', icon: 'laptop' },
]

const storedTheme = ref<string>(localStorage.getItem('theme') ?? 'system')
const systemDark = ref<boolean>(
  typeof window !== 'undefined' && window.matchMedia
    ? window.matchMedia('(prefers-color-scheme: dark)').matches
    : false
)

if (typeof window !== 'undefined' && window.matchMedia) {
  const mq = window.matchMedia('(prefers-color-scheme: dark)')
  mq.addEventListener?.('change', (e) => {
    systemDark.value = e.matches
  })
}

export const isDarkTheme = computed<boolean>(() => {
  if (storedTheme.value === 'dark') return true
  if (storedTheme.value === 'light') return false
  return systemDark.value
})

export const currentThemeName = computed<'light' | 'dark'>(() =>
  isDarkTheme.value ? 'dark' : 'light'
)

export const isRtl = computed<boolean>(() => i18n.global.locale.value === 'fa')

export const antdLocale = computed(() => {
  switch (i18n.global.locale.value) {
    case 'zhHans':
      return zhCN
    case 'zhHant':
      return zhTW
    case 'fa':
      return faIR
    case 'ru':
      return ruRU
    case 'vi':
      return viVN
    default:
      return enUS
  }
})

export const antdThemeConfig = computed(() => ({
  algorithm: isDarkTheme.value ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
  components: {
    Layout: {
      colorBgHeader: isDarkTheme.value ? '#141414' : '#ffffff',
    },
  },
}))

export function useThemeSwitcher() {
  const changeTheme = (th: string) => {
    storedTheme.value = th
    localStorage.setItem('theme', th)
  }

  const isActiveTheme = (th: string) => {
    return storedTheme.value === th
  }

  return {
    themes,
    changeTheme,
    isActiveTheme,
    isDark: isDarkTheme,
    themeName: currentThemeName,
    antdThemeConfig,
    antdLocale,
    isRtl,
  }
}

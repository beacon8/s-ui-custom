<template>
  <a-drawer
    v-if="isMobile"
    :open="displayDrawer"
    placement="left"
    :width="240"
    :body-style="{ padding: 0, display: 'flex', flexDirection: 'column', height: '100%' }"
    @close="$emit('toggleDrawer')"
  >
    <template #title>
      <a-space>
        <a-avatar
          :src="logoUrl"
          size="small"
        />
        <span>S-UI</span>
      </a-space>
    </template>
    <a-menu
      mode="inline"
      :selected-keys="[currentPath]"
      style="flex: 1; border-inline-end: 0;"
      @click="onMenuClick"
    >
      <a-menu-item
        v-for="item in menu"
        :key="item.path"
      >
        <template #icon>
          <AppIcon :name="item.icon" />
        </template>
        {{ $t(item.title) }}
      </a-menu-item>
    </a-menu>
    <a-divider style="margin: 0;" />
    <a-menu
      mode="inline"
      :selectable="false"
      style="border-inline-end: 0;"
    >
      <a-menu-item
        key="logout"
        @click="Logout"
      >
        <template #icon>
          <AppIcon name="logout" />
        </template>
        {{ $t('menu.logout') }}
      </a-menu-item>
    </a-menu>
  </a-drawer>

  <a-layout-sider
    v-else
    v-model:collapsed="collapsed"
    collapsible
    :theme="isDark ? 'dark' : 'light'"
    :width="220"
  >
    <a-flex
      align="center"
      justify="center"
      gap="small"
      style="height: 64px;"
    >
      <a-avatar
        :src="logoUrl"
        size="small"
      />
      <a-typography-text
        v-if="!collapsed"
        strong
      >
        S-UI
      </a-typography-text>
    </a-flex>
    <a-divider style="margin: 0;" />
    <a-menu
      mode="inline"
      :theme="isDark ? 'dark' : 'light'"
      :selected-keys="[currentPath]"
      @click="onMenuClick"
    >
      <a-menu-item
        v-for="item in menu"
        :key="item.path"
      >
        <template #icon>
          <AppIcon :name="item.icon" />
        </template>
        {{ $t(item.title) }}
      </a-menu-item>
      <a-menu-divider />
      <a-menu-item
        key="logout"
        @click="Logout"
      >
        <template #icon>
          <AppIcon name="logout" />
        </template>
        {{ $t('menu.logout') }}
      </a-menu-item>
    </a-menu>
  </a-layout-sider>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import router from '@/router'
import { logout } from '@/plugins/httputil'
import { isDarkTheme } from '@/composables/useThemeSwitcher'
import logoUrl from '@/assets/logo.svg'

const props = defineProps<{ isMobile: boolean, displayDrawer: boolean }>()
const emit = defineEmits<{ toggleDrawer: [] }>()

const collapsed = ref(false)
const isDark = isDarkTheme
const currentPath = computed(() => router.currentRoute.value.path)

const menu = [
  { title: 'pages.home', icon: 'home', path: '/' },
  { title: 'pages.inbounds', icon: 'cloud-download', path: '/inbounds' },
  { title: 'pages.clients', icon: 'account-multiple', path: '/clients' },
  { title: 'pages.outbounds', icon: 'cloud-upload', path: '/outbounds' },
  { title: 'pages.endpoints', icon: 'cloud-tags', path: '/endpoints' },
  { title: 'pages.services', icon: 'server', path: '/services' },
  { title: 'pages.tls', icon: 'certificate', path: '/tls' },
  { title: 'pages.basics', icon: 'application-cog', path: '/basics' },
  { title: 'pages.rules', icon: 'routes', path: '/rules' },
  { title: 'pages.dns', icon: 'dns', path: '/dns' },
  { title: 'pages.admins', icon: 'account-tie', path: '/admins' },
  { title: 'pages.settings', icon: 'cog', path: '/settings' },
]

const onMenuClick = ({ key }: { key: string | number }) => {
  if (key === 'logout') return
  router.push(String(key))
  if (props.isMobile) {
    emit('toggleDrawer')
  }
}

const Logout = async () => {
  logout()
}
</script>

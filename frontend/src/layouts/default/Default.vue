<template>
  <a-layout style="min-height: 100vh;">
    <drawer
      :is-mobile="isMobile"
      :display-drawer="displayDrawer"
      @toggle-drawer="toggleDrawer"
    />
    <a-layout>
      <default-bar
        :is-mobile="isMobile"
        @toggle-drawer="toggleDrawer"
      />
      <default-view />
    </a-layout>
    <update-modal v-model="updateVisible" />
  </a-layout>
</template>

<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue'
import DefaultBar from './AppBar.vue'
import Drawer from './Drawer.vue'
import DefaultView from './View.vue'
import Data from '@/store/modules/data'
import UpdateModal from '@/layouts/modals/Update.vue'
import { storeToRefs } from 'pinia'

const isMobile = ref(typeof window !== 'undefined' ? window.innerWidth < 960 : false)
const displayDrawer = ref(!isMobile.value)

const handleResize = () => {
  const mobile = window.innerWidth < 960
  if (mobile !== isMobile.value) {
    isMobile.value = mobile
    displayDrawer.value = !mobile
  }
}

const toggleDrawer = () => {
  displayDrawer.value = !displayDrawer.value
}

const { updateVisible } = storeToRefs(Data())

onMounted(() => {
  window.addEventListener('resize', handleResize)
  Data().checkUpdate()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

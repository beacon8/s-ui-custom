<template>
  <a-card
    size="small"
    title="TProxy"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <Network :data="inbound" />
      </a-col>
    </a-row>
  </a-card>
</template>

<script lang="ts">
import { PropType } from 'vue'
import Network from '@/components/Network.vue'
import { Inbound } from '@/types/inbounds'

export default {
  components: { Network },
  props: {
    inbound: { type: Object as PropType<Inbound>, required: true }
  },
  data() {
    return {
    }
  }
}
</script>
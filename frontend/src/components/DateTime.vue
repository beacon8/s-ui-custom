<template>
  <a-form-item
    :label="label || $t('date.expiry')"
    style="margin-bottom: 0;"
  >
    <a-date-picker
      :id="elId"
      :value="dateValue"
      :placeholder="placeholder"
      :allow-clear="true"
      :show-time="showTime"
      :show-now="true"
      format="YYYY-MM-DD HH:mm:ss"
      style="width: 100%"
      @change="onChange"
    />
  </a-form-item>
</template>

<script lang="ts">
import dayjs, { type Dayjs } from 'dayjs'

export default {
  props: {
    // A unix timestamp in seconds; 0 means no value / no expiry.
    expiry: { type: Number, required: true },
    label: { type: String, required: false, default: undefined },
    inputId: { type: String, required: false, default: undefined },
    // Client forms pass "unlimited"; stats omit it and keep the Ant placeholder.
    placeholder: { type: String, required: false, default: undefined }
  },
  emits: ['submit'],
  data() {
    return {
      showTime: { format: 'HH:mm:ss' }
    }
  },
  computed: {
    elId() {
      return this.inputId || 'expiry'
    },
    dateValue(): Dayjs | null {
      const seconds = Number(this.expiry) || 0
      return seconds > 0 ? dayjs.unix(seconds) : null
    }
  },
  methods: {
    onChange(value: Dayjs | null) {
      this.$emit('submit', value ? value.unix() : 0)
    }
  }
}
</script>

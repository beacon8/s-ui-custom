<template>
  <a-modal
    :open="visible"
    :width="400"
    :title="$t('admin.changeCred') + ' ' + user.username"
    @cancel="closeModal"
  >
    <a-form
      :model="newData"
      layout="vertical"
    >
      <a-form-item
        :label="$t('admin.oldPass')"
        name="oldPass"
        :rules="[{ required: true, message: $t('login.pwRules') }]"
      >
        <a-input-password v-model:value="newData.oldPass" />
      </a-form-item>
      <a-form-item
        :label="$t('admin.newUname')"
        name="newUsername"
        :rules="[{ required: true, message: $t('login.unRules') }]"
      >
        <a-input v-model:value="newData.newUsername" />
      </a-form-item>
      <a-form-item
        :label="$t('admin.newPass')"
        name="newPass"
        :rules="[{ required: true, message: $t('login.pwRules') }]"
      >
        <a-input-password v-model:value="newData.newPass" />
      </a-form-item>
    </a-form>
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
import { i18n } from '@/locales'
import type { PropType } from 'vue'

// The parent hands over a whole admin row; only these two are read here.
interface AdminUser {
  id?: number
  username?: string
}

export default {
  props: {
    visible: { type: Boolean, required: true },
    user: { type: Object as PropType<AdminUser>, required: true },
  },
  emits: ['close', 'save'],
  data() {
    return {
      newData: {
        id: <number | undefined>0,
        oldPass: "",
        newUsername: "",
        newPass: ""
      },
      usernameRules: [
        (value: string) => {
          if (value?.length > 0) return true
          return i18n.global.t('login.unRules')
        },
      ],
      passwordRules: [
        (value: string) => {
          if (value?.length > 0) return true
          return i18n.global.t('login.pwRules')
        },
      ]
    }
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.resetData()
      }
    },
  },
  methods: {
    resetData() {
      this.newData.id = this.$props.user.id
      this.newData.oldPass = ""
      this.newData.newUsername = ""
      this.newData.newPass = ""
    },
    closeModal() {
      this.resetData() // reset
      this.$emit('close')
    },
    saveChanges() {
      if (this.newData.oldPass == '' || this.newData.newUsername == '' || this.newData.newPass == '') return
      this.$emit('save', this.newData)
    },
  },
}
</script>

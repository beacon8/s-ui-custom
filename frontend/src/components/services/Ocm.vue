<template>
  <a-card
    size="small"
    title="OCM (OpenAI Codex Multiplexer)"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="12"
      >
        <a-form-item
          :label="$t('types.ocm.credentialPath')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.credential_path" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
      >
        <a-form-item
          :label="$t('types.ocm.usagesPath')"
          style="margin-bottom: 0;"
        >
          <a-input v-model:value="data.usages_path" />
        </a-form-item>
      </a-col>
      <a-col
        :xs="24"
        :sm="12"
      >
        <a-form-item
          :label="$t('dial.detourText')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="data.detour"
            style="width: 100%"
          >
            <a-select-option
              v-for="item in outTags"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
    </a-row>
    <a-flex
      align="center"
      :gap="8"
      style="margin: 12px 0 8px;"
    >
      <strong>{{ $t('types.ocm.users') }}</strong>
      <a-button
        type="primary"
        size="small"
        shape="circle"
        @click="addUser"
      >
        <template #icon>
          <AppIcon name="plus" />
        </template>
      </a-button>
    </a-flex>
    <a-card
      v-for="(user, index) in (data.users || [])"
      :key="index"
      size="small"
      style="margin-bottom: 8px;"
    >
      <a-row
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col flex="none">
          <a-button
            type="text"
            danger
            shape="circle"
            :aria-label="$t('actions.del')"
            @click="delUser(index)"
          >
            <template #icon>
              <AppIcon name="delete" />
            </template>
          </a-button>
        </a-col>
        <a-col
          :xs="24"
          :sm="8"
        >
          <a-form-item
            :label="$t('types.ocm.userName')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="user.name" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('types.ocm.userToken')"
            style="margin-bottom: 0;"
          >
            <a-input-password v-model:value="user.token" />
          </a-form-item>
        </a-col>
      </a-row>
    </a-card>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Data from '@/store/modules/data'
import { Srv } from '@/types/services'

// Only the multiplexer fields this form edits.
interface MuxData {
  credential_path?: string
  usages_path?: string
  detour?: string
  users?: { name: string, token: string }[]
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Srv>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as MuxData)

const outTags = computed((): string[] => {
  return [...Data().outbounds?.map((o) => o.tag) ?? [], ...Data().endpoints?.map((e) => e.tag) ?? []]
})

function addUser() {
  if (!data.value.users) data.value.users = []
  data.value.users.push({ name: '', token: '' })
}

function delUser(i: number | string) {
  data.value.users?.splice(Number(i), 1)
}
</script>

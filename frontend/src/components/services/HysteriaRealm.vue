<template>
  <a-card
    size="small"
    title="Hysteria Realm"
  >
    <a-flex
      align="center"
      :gap="8"
      style="margin-bottom: 8px;"
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
          :sm="6"
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
          :sm="10"
        >
          <a-form-item
            :label="$t('types.ocm.userToken')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="user.token">
              <template #suffix>
                <AppIcon
                  :aria-label="$t('actions.generate')"
                  name="refresh"
                  style="cursor: pointer;"
                  @click="user.token = RandomUtil.randomSeq(32)"
                />
              </template>
            </a-input>
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="6"
        >
          <a-form-item
            :label="$t('types.hy.maxRealms')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="user.max_realms"
              :min="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
    </a-card>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import RandomUtil from '@/plugins/randomUtil'
import { HysteriaRealm, Srv } from '@/types/services'

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<Srv>('data', { required: true })

// The parent holds a union and picks this component with a v-if on the type,
// which the template checker cannot follow, so narrow once here. The computed
// returns the same object, so edits still reach the parent.
const data = computed(() => model.value as unknown as HysteriaRealm)

function addUser() {
  if (!data.value.users) data.value.users = []
  data.value.users.push({ name: '', token: RandomUtil.randomSeq(32) })
}

function delUser(i: number | string) {
  data.value.users?.splice(Number(i), 1)
}
</script>

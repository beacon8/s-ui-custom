<template>
  <a-card
    size="small"
    :title="$t('pages.clients')"
  >
    <a-row :gutter="[12, 12]">
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-select
          v-model:value="data.model"
          :options="initUsersModels"
          :field-names="{ label: 'title', value: 'value' }"
          style="width: 100%"
          @change="data.values = []"
        />
      </a-col>
      <a-col
        v-if="data.model == 'group'"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <a-form-item
          :label="$t('client.group')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="groupValues"
            mode="multiple"
            style="width: 100%"
          >
            <a-select-option
              v-for="item in groupNames"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col
        v-if="data.model == 'client'"
        :xs="24"
        :sm="16"
      >
        <a-form-item
          :label="$t('pages.clients')"
          style="margin-bottom: 0;"
        >
          <a-select
            v-model:value="clientValues"
            mode="multiple"
            :options="clientNames"
            :field-names="{ label: 'title', value: 'value' }"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>
  </a-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { i18n } from '@/locales'
import { Client } from '@/types/clients'

// The client selection this form edits: how the clients are picked, and the
// group names or client ids that were picked.
interface UsersSelection {
  model: string
  values: (number | string)[]
}

// The parent owns the object and this component edits it in place, so it is a
// model rather than a plain prop.
const data = defineModel<UsersSelection>('data', { required: true })

const props = defineProps<{
  clients: Client[]
}>()

const initUsersModels = [
  { title: i18n.global.t('none'), value: 'none' },
  { title: i18n.global.t('all'), value: 'all' },
  { title: i18n.global.t('client.group'), value: 'group' },
  { title: i18n.global.t('pages.clients'), value: 'client' },
]

const clientNames = computed(() => props.clients.map(c => { return { title: c.name, value: c.id } }))

const groupNames = computed(() => Array.from(new Set(props.clients.map(c => c.group))))

// The two selects put different things in the same list — group names on one
// side, client ids on the other — so each gets its own view of it.
const groupValues = computed({
  get: () => data.value.values as string[],
  set: (v: string[]) => { data.value.values = v },
})

const clientValues = computed({
  get: () => data.value.values as number[],
  set: (v: number[]) => { data.value.values = v },
})
</script>

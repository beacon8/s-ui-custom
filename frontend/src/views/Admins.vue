<template>
  <AdminModal 
    v-model="editModal.visible"
    :visible="editModal.visible"
    :user="editModal.user"
    @close="closeEditModal"
    @save="saveEditModal"
  />
  <ChangeModal 
    v-model="changesModal.visible"
    :visible="changesModal.visible"
    :admins="users.map((u:any) => u.username)"
    :actor="changesModal.actor"
    @close="closeChangesModal"
  />
  <TokenModal 
    v-model="tokenModal.visible"
    :visible="tokenModal.visible"
    @close="closeTokenModal"
  />
  <a-row
    :gutter="[12, 12]"
    justify="center"
    style="margin-bottom: 12px;"
  >
    <a-col>
      <a-space>
        <a-button
          type="primary"
          @click="showChangesModal('')"
        >
          {{ $t('admin.changes') }}
        </a-button>
        <a-button
          type="primary"
          @click="showTokenModal()"
        >
          {{ $t('admin.api.token') }}
        </a-button>
      </a-space>
    </a-col>
  </a-row>
  <a-row :gutter="[12, 12]">
    <a-col
      v-for="item in <any[]>users"
      :key="item.id"
      :xs="24"
      :sm="8"
      :md="6"
      :lg="4"
    >
      <a-card
        size="small"
        :title="item.username"
      >
        <div style="font-size: 12px; opacity: 0.65; margin-bottom: 8px;">
          {{ $t('admin.lastLogin') }}
        </div>
        <a-row :gutter="[8, 6]">
          <a-col :span="12">
            {{ $t('admin.date') }}
          </a-col>
          <a-col :span="12">
            {{ item.loginDate }}
          </a-col>
          <a-col :span="12">
            {{ $t('admin.time') }}
          </a-col>
          <a-col :span="12">
            {{ item.loginTime }}
          </a-col>
          <a-col :span="12">
            IP
          </a-col>
          <a-col :span="12">
            {{ item.ip }}
          </a-col>
        </a-row>
        <a-divider style="margin: 10px 0 6px;" />
        <a-space :size="4">
          <a-tooltip :title="$t('actions.edit')">
            <a-button
              type="text"
              shape="circle"
              :aria-label="$t('actions.edit')"
              @click="showEditModal(item)"
            >
              <template #icon>
                <AppIcon name="account-edit" />
              </template>
            </a-button>
          </a-tooltip>
          <a-tooltip :title="$t('admin.changes')">
            <a-button
              type="text"
              shape="circle"
              :aria-label="$t('admin.changes')"
              @click="showChangesModal(item.username)"
            >
              <template #icon>
                <AppIcon name="list-box-outline" />
              </template>
            </a-button>
          </a-tooltip>
        </a-space>
      </a-card>
    </a-col>
  </a-row>
</template>

<script lang="ts" setup>
import AdminModal from '@/layouts/modals/Admin.vue'
import ChangeModal  from '@/layouts/modals/Changes.vue'
import TokenModal from '@/layouts/modals/Token.vue'
import { i18n } from '@/locales'
import HttpUtils from '@/plugins/httputil'
import { Ref, ref, inject, onMounted } from 'vue'

const loading:Ref = inject('loading')?? ref(false)

// One admin as api/users answers it. lastLogin is the date, the time and the
// address in a single space separated string.
interface AdminAccount {
  id: number
  username: string
  lastLogin: string
}

// A card on this page: the last login already split into its parts.
interface AdminRow {
  id: number
  username: string
  loginDate: string
  loginTime: string
  ip: string
}

// What the credential modal emits.
interface CredentialChange {
  id?: number
  oldPass: string
  newUsername: string
  newPass: string
}

const users = ref(<AdminRow[]>[])

onMounted(async () => {
  loading.value = true
  await loadData()
  loading.value = false
})

const loadData = async () => {
  loading.value = true
  const msg = await HttpUtils.get<AdminAccount[]>('api/users')
  loading.value = false
  if (msg.success) {
    msg.obj.forEach(u => {
      const lastLogin = u.lastLogin.split(" ")
      const localLastLogin = lastLogin.length > 2 ? dateFormatted(Date.parse(lastLogin[0] + " " + lastLogin[1])) : "- -"
      const loginDateTime = localLastLogin.split(" ")
      users.value.push({
        id: u.id,
        username: u.username,
        loginDate: loginDateTime[0],
        loginTime: loginDateTime[1],
        ip: lastLogin[2]?? "-",
      })
    })
  }
}

const dateFormatted = (dt: number): string => {
  const locale = i18n.global.locale.value.replace('zh', 'zh-')
  const date = new Date(dt)
  return date.toLocaleString(locale)
}

const editModal = ref<{ visible: boolean, user: Partial<AdminRow> }>({
  visible: false,
  user: {},
})

const showEditModal = (user: AdminRow) => {
  editModal.value.user = user
  editModal.value.visible = true
}
const closeEditModal = () => {
  editModal.value.visible = false
  editModal.value.user = {}
}
const saveEditModal = async (data:CredentialChange) => {
  loading.value=true
  const response = await HttpUtils.post('api/changePass',data)
  if(response.success){
    setTimeout(() => {
      loading.value=false
      editModal.value.visible = false
    }, 500)
  } else {
    loading.value=false
  }
}

const changesModal = ref({
  visible: false,
  actor: '',
})
const showChangesModal = (actor: string) => {
  changesModal.value.actor = actor
  changesModal.value.visible = true
}
const closeChangesModal = () => {
  changesModal.value.visible = false
  changesModal.value.actor = ''
}

const tokenModal = ref({
  visible: false,
})
const showTokenModal = () => {
  tokenModal.value.visible = true
}
const closeTokenModal = () => {
  tokenModal.value.visible = false
}
</script>
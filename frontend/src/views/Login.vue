<template>
  <a-row
    justify="center"
    align="middle"
    style="margin-top: 100px; padding: 0 16px;"
  >
    <a-col
      :xs="24"
      :sm="16"
      :md="8"
    >
      <a-card :title="$t('login.title')">
        <a-form
          :model="formState"
          layout="vertical"
          @finish="login"
        >
          <a-form-item
            :label="$t('login.username')"
            name="username"
            :rules="[{ required: true, message: $t('login.unRules') }]"
          >
            <a-input v-model:value="formState.username" />
          </a-form-item>
          <a-form-item
            :label="$t('login.password')"
            name="password"
            :rules="[{ required: true, message: $t('login.pwRules') }]"
          >
            <a-input-password v-model:value="formState.password" />
          </a-form-item>
          <a-button
            :loading="loading"
            html-type="submit"
            type="primary"
            block
          >
            {{ $t('actions.submit') }}
          </a-button>
        </a-form>
        <a-flex
          gap="small"
          align="center"
          style="margin-top: 12px;"
        >
          <a-select
            v-model:value="$i18n.locale"
            :options="languages"
            :field-names="{ label: 'title', value: 'value' }"
            style="flex: 1;"
            @change="(val: any) => changeLocale(val)"
          />
          <a-dropdown :trigger="['click']">
            <a-button
              :aria-label="$t('a11y.theme')"
              type="text"
              shape="circle"
            >
              <template #icon>
                <AppIcon name="theme-light-dark" />
              </template>
            </a-button>
            <template #overlay>
              <a-menu :selected-keys="themes.filter(th => isActiveTheme(th.value)).map(th => th.value)">
                <a-menu-item
                  v-for="th in themes"
                  :key="th.value"
                  @click="changeTheme(th.value)"
                >
                  <template #icon>
                    <AppIcon :name="th.icon" />
                  </template>
                  {{ $t(`theme.${th.value}`) }}
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </a-flex>
      </a-card>
    </a-col>
  </a-row>
</template>

<script lang="ts" setup>
import { reactive, ref } from "vue"
import { i18n, languages } from '@/locales'
import { useRouter } from 'vue-router'
import HttpUtil from '@/plugins/httputil'
import { setAuthenticated } from '@/plugins/auth'
import { useThemeSwitcher } from '@/composables/useThemeSwitcher'

const { themes, changeTheme, isActiveTheme } = useThemeSwitcher()

const formState = reactive({
  username: '',
  password: '',
})

const loading = ref(false)
const router = useRouter()

const login = async () => {
  if (formState.username == '' || formState.password == '') return
  loading.value = true
  const response = await HttpUtil.post('api/login', { user: formState.username, pass: formState.password })
  if (response.success) {
    setAuthenticated()
    setTimeout(() => {
      loading.value = false
      router.push('/')
    }, 500)
  } else {
    loading.value = false
  }
}

const changeLocale = (l: string | null) => {
  const next = l ?? 'en'
  i18n.global.locale.value = next as typeof i18n.global.locale.value
  localStorage.setItem('locale', next)
}
</script>

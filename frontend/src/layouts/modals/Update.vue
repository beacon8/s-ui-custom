<template>
  <a-modal
    v-model:open="visible"
    :width="550"
    :title="$t('update.title')"
    :closable="state != 'restarting'"
    :mask-closable="state != 'restarting'"
    :keyboard="state != 'restarting'"
    :footer="null"
  >
    <template v-if="update">
      <a-space
        align="center"
        wrap
        style="margin-bottom: 16px;"
      >
        <a-tag color="blue">
          v{{ update.current }}
        </a-tag>
        <AppIcon :name="isRtl ? 'arrow-left' : 'arrow-right'" />
        <a-tag color="processing">
          v{{ update.latest }}
        </a-tag>
        <a
          :href="update.url"
          target="_blank"
          rel="noopener"
        >{{ $t('update.notes') }}</a>
      </a-space>

      <template v-if="update.method == 'docker'">
        <a-alert
          type="warning"
          :message="$t('update.dockerHint')"
          show-icon
        />
        <pre class="command">docker compose pull &amp;&amp; docker compose up -d</pre>
      </template>

      <template v-else-if="update.method == 'systemd'">
        <a-alert
          v-if="state == 'idle'"
          type="info"
          :message="$t('update.systemdHint')"
          show-icon
        />
        <div
          v-else-if="state == 'restarting'"
          style="text-align: center; padding: 16px 0;"
        >
          <a-spin size="large" />
          <div style="margin-top: 12px;">
            {{ $t('update.restarting') }}
          </div>
        </div>
        <a-alert
          v-else-if="state == 'timeout'"
          type="error"
          :message="$t('update.timeout')"
          show-icon
        />
      </template>

      <template v-else>
        <a-alert
          type="info"
          :message="$t('update.manualHint')"
          show-icon
        />
        <pre class="command">s-ui update</pre>
      </template>

      <a-flex
        v-if="update.method == 'systemd' && state == 'idle'"
        justify="end"
        style="margin-top: 16px;"
      >
        <a-button
          type="primary"
          :loading="loading"
          @click="upgrade"
        >
          <template #icon>
            <AppIcon name="update" />
          </template>
          {{ $t('update.now') }}
        </a-button>
      </a-flex>
    </template>
  </a-modal>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { i18n } from '@/locales'
import { isRtl } from '@/composables/useThemeSwitcher'
import api from '@/plugins/api'
import HttpUtils from '@/plugins/httputil'
import Data from '@/store/modules/data'
import type { UpdateInfo } from '@/store/modules/data'

const visible = defineModel<boolean>({ required: true })

const update = computed(() => Data().update)
const loading = ref(false)
const state = ref<'idle' | 'restarting' | 'timeout'>('idle')

// How long the download, migration and restart may take before the panel
// gives up waiting and tells the operator to look at the server.
const restartTimeout = 5 * 60 * 1000
let pollId: ReturnType<typeof setInterval> | null = null

const upgrade = async () => {
  if (!update.value || !confirm(i18n.global.t('update.confirm'))) return
  loading.value = true
  const msg = await HttpUtils.post('api/update', null)
  loading.value = false
  if (msg.success) waitForRestart(update.value.current)
}

// Polled with the bare client: while the panel is down every request fails,
// and HttpUtils would raise each failure as an error.
const waitForRestart = (oldVersion: string) => {
  state.value = 'restarting'
  const started = Date.now()
  pollId = setInterval(async () => {
    if (Date.now() - started > restartTimeout) {
      stopPolling()
      state.value = 'timeout'
      return
    }
    try {
      const resp = await api.get<{ success: boolean, obj: UpdateInfo }>('api/checkUpdate')
      if (resp.data?.success && resp.data.obj?.current && resp.data.obj.current != oldVersion) {
        stopPolling()
        window.location.reload()
      }
    } catch {
      // Still restarting.
    }
  }, 3000)
}

const stopPolling = () => {
  if (pollId) {
    clearInterval(pollId)
    pollId = null
  }
}

onBeforeUnmount(stopPolling)
</script>

<style scoped>
.command {
  direction: ltr;
  margin-top: 12px;
  padding: 8px 12px;
  border-radius: 4px;
  background: rgba(128, 128, 128, 0.12);
  white-space: pre-wrap;
  user-select: all;
}
</style>

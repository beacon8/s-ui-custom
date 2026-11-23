<template>
  <a-modal
    :open="visible"
    :width="400"
    title="QrCode"
    :footer="null"
    @cancel="$emit('close')"
  >
    <a-skeleton
      v-if="loading"
      active
      :paragraph="{ rows: 4 }"
    />
    <div
      id="qrcode-modal"
      v-show="!loading"
      style="max-height: 70vh; overflow-y: auto;"
    >
      <a-tabs
        v-model:activeKey="tab"
        centered
      >
        <a-tab-pane
          key="sub"
          :tab="$t('setting.sub')"
        >
          <a-row :gutter="[12, 12]">
            <a-col
              :span="24"
              style="text-align: center;"
            >
              <a-tag style="margin-bottom: 6px;">
                {{ $t('setting.sub') }}
              </a-tag><br>
              <QrcodeVue
                :value="clientSub"
                :size="size"
                :margin="1"
                style="border-radius: 1rem; cursor: copy;"
                @click="copyToClipboard(clientSub)"
              />
            </a-col>
            <a-col
              :span="24"
              style="text-align: center;"
            >
              <a-tag style="margin-bottom: 6px;">
                {{ $t('setting.jsonSub') }}
              </a-tag><br>
              <QrcodeVue
                :value="clientSub + '?format=json'"
                :size="size"
                :margin="1"
                style="border-radius: 1rem; cursor: copy;"
                @click="copyToClipboard(clientSub + '?format=json')"
              />
            </a-col>
            <a-col
              :span="24"
              style="text-align: center;"
            >
              <a-tag style="margin-bottom: 6px;">
                {{ $t('setting.clashSub') }}
              </a-tag><br>
              <QrcodeVue
                :value="clientSub + '?format=clash'"
                :size="size"
                :margin="1"
                style="border-radius: 1rem; cursor: copy;"
                @click="copyToClipboard(clientSub + '?format=clash')"
              />
            </a-col>
            <a-col
              :span="24"
              style="text-align: center;"
            >
              <a-tag style="margin-bottom: 6px;">
                SING-BOX (scan only)
              </a-tag><br>
              <QrcodeVue
                :value="singbox"
                :size="size"
                :margin="1"
                style="border-radius: .8rem; cursor: not-allowed;"
              />
            </a-col>
            <a-col
              :span="24"
              style="text-align: center;"
            >
              <a-tag style="margin-bottom: 6px;">
                HAPP (scan only)
              </a-tag><br>
              <QrcodeVue
                :value="happ"
                :size="size"
                :margin="1"
                style="border-radius: .8rem; cursor: not-allowed;"
              />
            </a-col>
          </a-row>
        </a-tab-pane>
        <a-tab-pane
          key="link"
          :tab="$t('client.links')"
        >
          <a-row :gutter="[12, 12]">
            <a-col
              v-for="(l, index) in clientLinks"
              :key="index"
              :span="24"
              style="text-align: center;"
            >
              <a-tag style="margin-bottom: 6px;">
                {{ l.remark ?? $t('client.' + l.type) }}
              </a-tag><br>
              <QrcodeVue
                :value="l.uri"
                :size="size"
                :margin="1"
                style="border-radius: .5rem; cursor: copy;"
                @click="copyToClipboard(l.uri)"
              />
            </a-col>
          </a-row>
        </a-tab-pane>
      </a-tabs>
    </div>
  </a-modal>
</template>

<script lang="ts">
import QrcodeVue from 'qrcode.vue'
import Data from '@/store/modules/data'
import Clipboard from 'clipboard'
import { i18n } from '@/locales'
import { push } from 'notivue'
import { Client } from '@/types/clients'

export default {
  components: { QrcodeVue },
  props: {
    id: { type: Number, required: true },
    visible: { type: Boolean, required: true },
  },
  emits: ['close'],
  data() {
    return {
      tab: "sub",
      client: <Client>{},
      loading: false,
    }
  },
  computed: {
    clientSub() {
      return Data().subURI + this.client.name
    },
    singbox() {
      const url = Data().subURI + this.client.name + "?format=json"
      return "sing-box://import-remote-profile?url=" +  encodeURIComponent(url) + "#" + this.client.name
    },
    // Happ takes the subscription URL as it is, unencoded, after happ://add/.
    happ() {
      return "happ://add/" + this.clientSub + "?format=happ"
    },
    clientLinks() {
      return this.client.links?? []
    },
    size() {
      if (window.innerWidth > 380) return 300
      if (window.innerWidth > 330) return 280
      return 250
    }
  },
  watch: {
    visible(v) {
      if (v) {
        this.tab = "sub"
        this.load()
      }
    },
  },
  methods: {
    async load() {
      this.loading = true
      const newData = await Data().loadClients(this.$props.id)
      this.client = newData
      this.loading = false
    },
    copyToClipboard(txt:string) {
      const hiddenButton = document.createElement('button')
      hiddenButton.className = 'clipboard-btn'
      document.body.appendChild(hiddenButton)

      const clipboard = new Clipboard('.clipboard-btn', {
        text: () => txt,
        container: document.getElementById('qrcode-modal')?? undefined
      });

      clipboard.on('success', () => {
        clipboard.destroy()
        push.success({
          message: i18n.global.t('success') + ": " + i18n.global.t('copyToClipboard'),
          duration: 5000,
        })
      })

      clipboard.on('error', () => {
        clipboard.destroy()
        push.error({
          message: i18n.global.t('failed') + ": " + i18n.global.t('copyToClipboard'),
          duration: 5000,
        })
      })

      // Perform click on hidden button to trigger copy
      hiddenButton.click()
      document.body.removeChild(hiddenButton)
    }
  }
}
</script>
<template>
  <a-modal
    :open="visible"
    :width="800"
    @cancel="close"
  >
    <template #title>
      <a-flex
        justify="space-between"
        align="center"
        style="padding-right: 24px;"
      >
        <span>{{ $t('rule.import.title') }}</span>
        <a-tag
          v-if="importPreview.length > 0"
          color="blue"
        >
          {{ $t('count') }}: {{ importPreview.length }}
        </a-tag>
      </a-flex>
    </template>
    <div style="max-height: 70vh; overflow-y: auto; padding-right: 4px;">
      <a-form layout="vertical">
        <a-tabs
          v-model:activeKey="tab"
          @change="tabChanged"
        >
          <a-tab-pane
            key="text"
            :tab="$t('rule.import.pasteUrls')"
          >
            <a-alert
              type="info"
              :message="$t('rule.import.urlsHint')"
              show-icon
              style="margin-bottom: 12px;"
            />
            <a-form-item
              label="URLs"
              style="margin-bottom: 12px;"
            >
              <a-textarea
                v-model:value="importRawText"
                :auto-size="{ minRows: 10 }"
                spellcheck="false"
                placeholder="https://github.com/.../geoip-telegram.srs&#10;https://github.com/.../geosite-youtube.srs"
              />
            </a-form-item>
          </a-tab-pane>
          <a-tab-pane
            key="file"
            :tab="$t('rule.import.uploadTxt')"
          >
            <a-alert
              type="info"
              :message="$t('rule.import.fileHint')"
              show-icon
              style="margin-bottom: 12px;"
            />
            <a-upload
              accept=".txt"
              :max-count="1"
              :before-upload="handleBeforeUpload"
              @remove="tabChanged"
            >
              <a-button type="default">
                <template #icon>
                  <AppIcon name="file-document" />
                </template>
                {{ $t('rule.import.selectTxt') }}
              </a-button>
            </a-upload>
          </a-tab-pane>
        </a-tabs>
        <a-row
          :gutter="[12, 12]"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('ruleset.format')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="importFormat"
                style="width: 100%"
              >
                <a-select-option
                  v-for="fmt in ['source', 'binary']"
                  :key="fmt"
                  :value="fmt"
                >
                  {{ fmt }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('objects.outbound')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="importDetour"
                allow-clear
                style="width: 100%"
                @clear="importDetour = ''"
              >
                <a-select-option
                  v-for="tag in outTags"
                  :key="tag"
                  :value="tag"
                >
                  {{ tag }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('ruleset.interval')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="importInterval"
                :addon-after="$t('date.d')"
                :min="0"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <template v-if="importPreview.length > 0">
          <a-divider />
          <a-space style="margin-bottom: 8px;">
            <strong>{{ $t('rule.import.preview') }}</strong>
            <a-badge
              v-if="importPreview.length > 0"
              :count="importPreview.length"
              :number-style="{ backgroundColor: '#52c41a' }"
            />
            <a-tooltip
              v-if="importSkipped > 0"
              :title="$t('rule.import.skipped')"
            >
              <a-badge
                :count="importSkipped"
                :number-style="{ backgroundColor: '#faad14' }"
              />
            </a-tooltip>
          </a-space>
          <a-table
            :data-source="importPreview"
            :pagination="false"
            size="small"
            :row-key="(_record: ImportItem, i?: number) => i ?? 0"
            :custom-row="(record: ImportItem) => ({ style: record.exists ? 'opacity: 0.4' : '' })"
            :columns="[
              { title: $t('objects.tag'), key: 'tag', dataIndex: 'tag' },
              { title: $t('ruleset.format'), key: 'format', dataIndex: 'format' },
              { title: 'URL', key: 'url', dataIndex: 'url' },
              { title: $t('actions.del'), key: 'del', width: 64 },
            ]"
          >
            <template #bodyCell="{ column, record, index }">
              <template v-if="column.key === 'tag'">
                {{ record.tag }}
              </template>
              <template v-else-if="column.key === 'format'">
                {{ record.format }}
              </template>
              <template v-else-if="column.key === 'url'">
                <a-tooltip :title="record.url">
                  <span dir="ltr">.../{{ record.url.split('/').pop() ?? record.url }}</span>
                </a-tooltip>
              </template>
              <template v-else-if="column.key === 'del'">
                <a-button
                  type="text"
                  danger
                  size="small"
                  :aria-label="$t('actions.del')"
                  @click="importPreview.splice(index, 1)"
                >
                  <template #icon>
                    <AppIcon
                      name="delete"
                      color="error"
                    />
                  </template>
                </a-button>
              </template>
            </template>
          </a-table>
        </template>
      </a-form>
    </div>
    <template #footer>
      <a-flex
        justify="space-between"
        align="center"
      >
        <a-button
          type="default"
          :disabled="importRawText.trim().length === 0"
          @click="parseImport"
        >
          <template #icon>
            <AppIcon name="magnify" />
          </template>
          {{ $t('rule.import.parse') }}
        </a-button>
        <a-space>
          <a-button
            type="default"
            @click="close"
          >
            {{ $t('actions.close') }}
          </a-button>
          <a-button
            type="primary"
            :disabled="newCount === 0"
            @click="save"
          >
            {{ $t('actions.save') }}
          </a-button>
        </a-space>
      </a-flex>
    </template>
  </a-modal>
</template>

<script lang="ts">
import { downloadHttpClient } from '@/plugins/httpClient'
import { ruleset } from '@/types/rules'
import type { PropType } from 'vue'
interface ImportItem { tag: string; url: string; format: ruleset['format']; exists: boolean }

export default {
  props: {
    visible: { type: Boolean, required: true },
    outTags: { type: Array as PropType<string[]>, required: true },
    rsTags: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['save', 'close'],
  data() {
    return {
      tab: 'text',
      importRawText: '',
      importFormat: <ruleset['format']>'binary',
      importDetour: '',
      importInterval: 1,
      importPreview: [] as ImportItem[],
    }
  },
  computed: {
    importSkipped(): number {
      return this.importPreview.filter(i => i.exists).length
    },
    newCount(): number {
      return this.importPreview.filter(i => !i.exists).length
    },
  },
  watch: {
    visible(v) {
      if (v) {
        this.tab = 'text'
        this.tabChanged()
      }
    },
  },
  methods: {
    tabChanged() {
      this.importPreview = []
      this.importRawText = ''
    },
    urlToTag(url: string): string {
      try {
        const filename = new URL(url).pathname.split('/').pop() ?? ''
        return filename.replace(/\.[^.]+$/, '')
      } catch {
        const parts = url.split('/')
        return parts[parts.length - 1].replace(/\.[^.]+$/, '') || url
      }
    },
    close() {
      this.$emit('close')
    },
    parseImport() {
      const existingTags = new Set(this.rsTags)
      const seen = new Set<string>()
      this.importPreview = this.importRawText
        .split('\n').map(l => l.trim()).filter(l => l.length > 0 && l.startsWith('http'))
        .filter(url => { if (seen.has(url)) return false; seen.add(url); return true })
        .map(url => ({ tag: this.urlToTag(url), url, format: this.importFormat, exists: existingTags.has(this.urlToTag(url)) }))
    },
    save() {
      const toAdd = this.importPreview.filter(i => !i.exists).map(item => {
        const rs: ruleset = { type: 'remote', tag: item.tag, format: item.format, url: item.url }
        const httpClient = downloadHttpClient(this.importDetour)
        if (httpClient) rs.http_client = httpClient
        if (this.importInterval > 0) rs.update_interval = this.importInterval + 'd'
        return rs
      })
      this.$emit('save', toAdd)
    },
    handleBeforeUpload(file: File) {
      file.text().then(text => {
        this.importRawText = text
        this.parseImport()
      })
      return false
    },
    async onFileUpload(event: Event) {
      const file = (event.target as HTMLInputElement).files?.[0]
      if (!file) return
      this.importRawText = await file.text()
      this.parseImport()
    },
  },
}
</script>

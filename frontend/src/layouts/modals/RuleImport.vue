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
        <span>{{ $t('rule.import.rulesTitle') }}</span>
        <a-tag
          v-if="parsed"
          color="blue"
        >
          {{ parsed.rules?.length ?? 0 }} {{ $t('pages.rules') }} · {{ parsed.rule_set?.length ?? 0 }} {{ $t('rule.ruleset') }}
        </a-tag>
      </a-flex>
    </template>
    <div style="max-height: 70vh; overflow-y: auto; padding-right: 4px;">
      <a-tabs
        v-model:activeKey="tab"
        @change="tabChanged"
      >
        <a-tab-pane
          key="json"
          tab="JSON"
        >
          <a-alert
            type="info"
            :message="$t('rule.import.jsonHint')"
            show-icon
            style="margin-bottom: 12px;"
          />
          <a-form layout="vertical">
            <a-form-item
              label="JSON"
              :validate-status="error ? 'error' : undefined"
              style="margin-bottom: 0;"
            >
              <a-textarea
                v-model:value="rawJson"
                :rows="12"
                spellcheck="false"
              />
            </a-form-item>
          </a-form>
        </a-tab-pane>

        <a-tab-pane
          key="file"
          :tab="$t('rule.import.uploadFile')"
        >
          <a-alert
            type="info"
            :message="$t('rule.import.fileJsonHint')"
            show-icon
            style="margin-bottom: 12px;"
          />
          <a-upload
            accept=".json"
            :max-count="1"
            :before-upload="handleBeforeUpload"
            @remove="tabChanged"
          >
            <a-button type="default">
              <template #icon>
                <AppIcon name="file-code" />
              </template>
              {{ $t('rule.import.selectJson') }}
            </a-button>
          </a-upload>
        </a-tab-pane>

        <a-tab-pane
          key="url"
          :tab="$t('rule.import.fromUrl')"
        >
          <a-alert
            type="info"
            :message="$t('rule.import.urlHint')"
            show-icon
            style="margin-bottom: 12px;"
          />
          <a-form layout="vertical">
            <a-form-item
              label="URL"
              style="margin-bottom: 0;"
            >
              <a-input
                v-model:value="fetchUrl"
                spellcheck="false"
                placeholder="https://raw.githubusercontent.com/.../rules.json"
                @press-enter="fetchFromUrl"
              >
                <template #suffix>
                  <a-button
                    type="text"
                    size="small"
                    :loading="fetching"
                    @click="fetchFromUrl"
                  >
                    <template #icon>
                      <AppIcon name="download" />
                    </template>
                  </a-button>
                </template>
              </a-input>
            </a-form-item>
          </a-form>
        </a-tab-pane>
      </a-tabs>

      <a-alert
        v-if="error"
        type="error"
        :message="error"
        show-icon
        style="margin-top: 12px;"
      />

      <template v-if="parsed">
        <a-divider />

        <a-alert
          v-if="hasConflicts"
          type="warning"
          :message="$t('rule.import.conflict')"
          show-icon
          style="margin-bottom: 12px;"
        >
          <template #description>
            <div>{{ $t('rule.import.conflictDesc', { rules: existingRulesCount, rulesets: existingRulesetsCount }) }}</div>
            <a-radio-group
              v-model:value="mode"
              style="margin-top: 8px;"
            >
              <a-radio value="merge">
                {{ $t('rule.import.merge') }}
              </a-radio>
              <a-radio value="replace">
                {{ $t('rule.import.replace') }}
              </a-radio>
            </a-radio-group>
          </template>
        </a-alert>

        <a-alert
          v-if="parsed.final"
          type="info"
          show-icon
          style="margin-bottom: 12px;"
        >
          <template #message>
            <a-space>
              <span>{{ $t('rule.import.finalOutbound') }}:</span>
              <a-tag>{{ parsed.final }}</a-tag>
            </a-space>
          </template>
          <template #description>
            <a-checkbox v-model:checked="applyFinal">
              {{ $t('rule.import.applyFinal') }}
            </a-checkbox>
          </template>
        </a-alert>

        <a-space style="margin-bottom: 8px;">
          <strong>{{ $t('pages.rules') }}</strong>
          <a-badge
            v-if="(parsed.rules?.length ?? 0) > 0"
            :count="parsed.rules?.length"
            :number-style="{ backgroundColor: '#52c41a' }"
          />
        </a-space>
        <a-table
          v-if="parsed.rules?.length"
          :data-source="parsed.rules"
          :pagination="false"
          size="small"
          style="margin-bottom: 16px;"
          :row-key="(_record: ImportedRule, i?: number) => i ?? 0"
          :columns="[
            { title: '#', key: 'index', width: 56 },
            { title: $t('type'), key: 'type', dataIndex: 'type' },
            { title: $t('admin.action'), key: 'action', dataIndex: 'action' },
            { title: $t('objects.outbound'), key: 'outbound', dataIndex: 'outbound' },
          ]"
        >
          <template #bodyCell="{ column, record, index }">
            <template v-if="column.key === 'index'">
              {{ index + 1 }}
            </template>
            <template v-else-if="column.key === 'type'">
              {{ record.type ?? 'simple' }}
            </template>
            <template v-else-if="column.key === 'action'">
              {{ record.action }}
            </template>
            <template v-else-if="column.key === 'outbound'">
              {{ record.outbound ?? '-' }}
            </template>
          </template>
        </a-table>

        <a-space style="margin-bottom: 8px;">
          <strong>{{ $t('rule.ruleset') }}</strong>
          <a-badge
            v-if="(parsed.rule_set?.length ?? 0) > 0"
            :count="parsed.rule_set?.length"
            :number-style="{ backgroundColor: '#52c41a' }"
          />
          <a-tooltip
            v-if="skippedRulesets > 0"
            :title="$t('rule.import.skipped')"
          >
            <a-badge
              :count="skippedRulesets"
              :number-style="{ backgroundColor: '#faad14' }"
            />
          </a-tooltip>
        </a-space>
        <a-table
          v-if="parsed.rule_set?.length"
          :data-source="parsed.rule_set"
          :pagination="false"
          size="small"
          :row-key="(_record: ruleset, i?: number) => i ?? 0"
          :custom-row="(record: ruleset) => ({ style: mode === 'merge' && existingRulesetTags.includes(record.tag) ? 'opacity: 0.4' : '' })"
          :columns="[
            { title: $t('objects.tag'), key: 'tag', dataIndex: 'tag' },
            { title: $t('ruleset.format'), key: 'format', dataIndex: 'format' },
            { title: $t('type'), key: 'type', dataIndex: 'type' },
            { title: $t('ruleset.interval'), key: 'interval', dataIndex: 'update_interval' },
          ]"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'tag'">
              <span style="font-family: monospace;">{{ record.tag }}</span>
            </template>
            <template v-else-if="column.key === 'format'">
              {{ record.format }}
            </template>
            <template v-else-if="column.key === 'type'">
              {{ record.type }}
            </template>
            <template v-else-if="column.key === 'interval'">
              {{ record.update_interval ?? '-' }}
            </template>
          </template>
        </a-table>
      </template>
    </div>
    <template #footer>
      <a-flex
        justify="space-between"
        align="center"
      >
        <div>
          <a-button
            v-if="tab === 'json'"
            type="default"
            :disabled="rawJson.trim().length === 0"
            @click="parseJson"
          >
            <template #icon>
              <AppIcon name="magnify" />
            </template>
            {{ $t('rule.import.parse') }}
          </a-button>
        </div>
        <a-space>
          <a-button
            type="default"
            @click="close"
          >
            {{ $t('actions.close') }}
          </a-button>
          <a-button
            type="primary"
            :disabled="!parsed"
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
import { logicalRule, ruleset } from '@/types/rules'
import type { PropType } from 'vue'

// A rule as it arrives from an imported config. A logical rule carries every
// field a simple one does plus its own, and nothing here is guaranteed to be
// present, so the preview reads the loosest shape of the two.
type ImportedRule = Partial<logicalRule>

// The piece of a sing-box config this modal imports: either a whole config's
// route section, or a bare object carrying the same lists.
interface RouteBlock {
  rules?: ImportedRule[]
  rule_set?: ruleset[]
  final?: string
}

export default {
  props: {
    visible: { type: Boolean, required: true },
    existingRulesCount: { type: Number, required: true },
    existingRulesetsCount: { type: Number, required: true },
    existingRulesetTags: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['save', 'close'],
  data() {
    return {
      tab: 'json',
      rawJson: '',
      fetchUrl: '',
      fetching: false,
      error: '',
      parsed: null as RouteBlock | null,
      mode: 'merge' as 'merge' | 'replace',
      applyFinal: false,
    }
  },
  computed: {
    hasConflicts(): boolean {
      return this.existingRulesCount > 0 || this.existingRulesetsCount > 0
    },
    skippedRulesets(): number {
      const ruleSets = this.parsed?.rule_set
      if (!ruleSets) return 0
      const existing = new Set(this.existingRulesetTags)
      return ruleSets.filter(rs => existing.has(rs.tag)).length
    },
  },
  watch: {
    visible(v: boolean) {
      if (v) this.reset()
    },
  },
  methods: {
    tabChanged() {
      this.rawJson = ''
      this.fetchUrl = ''
      this.error = ''
      this.parsed = null
      this.mode = this.hasConflicts ? 'merge' : 'replace'
      this.applyFinal = false
    },
    extractRouteBlock(parsed: unknown): RouteBlock | null {
      const obj = parsed as (RouteBlock & { route?: RouteBlock }) | null
      if (obj?.route && (obj.route.rules || obj.route.rule_set)) return obj.route
      if (obj?.rules || obj?.rule_set) return obj
      return null
    },
    setParsed(block: RouteBlock) {
      this.parsed = block
      this.mode = this.hasConflicts ? 'merge' : 'replace'
      this.applyFinal = false
    },
    reset() {
      this.tab = 'json'
      this.tabChanged()
    },
    parseJson() {
      this.error = ''
      this.parsed = null
      try {
        const block = this.extractRouteBlock(JSON.parse(this.rawJson))
        if (!block) {
          this.error = this.$t('rule.import.errNoArrays')
          return
        }
        this.setParsed(block)
      } catch (e: unknown) {
        this.error = this.$t('rule.import.errJsonParse', { message: (<Error>e).message })
      }
    },
    async fetchFromUrl() {
      this.error = ''
      this.parsed = null
      this.fetching = true
      try {
        const resp = await fetch(this.fetchUrl)
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
        const block = this.extractRouteBlock(await resp.json())
        if (!block) this.error = this.$t('rule.import.errNoArraysFetched')
        else this.setParsed(block)
      } catch (e: unknown) {
        this.error = this.$t('rule.import.errFetch', { message: (<Error>e).message })
      } finally {
        this.fetching = false
      }
    },
    handleBeforeUpload(file: File) {
      this.onFileUpload(file)
      return false
    },
    async onFileUpload(files: File | File[] | null) {
      this.error = ''
      this.parsed = null
      const file = Array.isArray(files) ? files[0] : files
      if (!file) {
        this.error = this.$t('rule.import.errNoFile')
        return
      }
      try {
        const block = this.extractRouteBlock(JSON.parse(await file.text()))
        if (!block) {
          this.error = this.$t('rule.import.errNoArraysInFile')
          return
        }
        this.setParsed(block)
      } catch (e: unknown) {
        this.error = this.$t('rule.import.errJsonParse', { message: (<Error>e).message })
        return
      }
    },
    save() {
      if (!this.parsed) return
      this.$emit('save', this.parsed, this.mode, this.applyFinal)
    },
    close() {
      this.$emit('close')
    },
  },
}
</script>

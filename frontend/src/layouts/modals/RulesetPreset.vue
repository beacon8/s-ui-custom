<template>
  <a-modal
    :open="visible"
    :width="600"
    @cancel="$emit('close')"
  >
    <template #title>
      <a-flex
        justify="space-between"
        align="center"
        style="padding-inline-end: 24px;"
      >
        <span>{{ $t('ruleset.preset') }}</span>
        <a-tag
          v-if="preview.length > 0"
          color="blue"
        >
          {{ $t('count') }}: {{ preview.length }}
        </a-tag>
      </a-flex>
    </template>

    <a-form
      layout="vertical"
      style="max-height: 70vh; overflow-y: auto; overflow-x: hidden;"
    >
      <a-row :gutter="[12, 12]">
        <a-col :span="24">
          <a-form-item
            :label="$t('objects.ruleset')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="selected"
              :options="catalogItems"
              :field-names="{ label: 'title', value: 'value' }"
              mode="multiple"
              allow-clear
              style="width: 100%;"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <a-row
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('objects.outbound')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="detour"
              allow-clear
              style="width: 100%;"
              @clear="detour = ''"
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
        >
          <a-form-item
            :label="$t('ruleset.interval')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="interval"
              :min="0"
              :addon-after="$t('date.d')"
              style="width: 100%;"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <a-divider style="margin: 12px 0;" />

      <a-row
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-space>
            <a-switch v-model:checked="addRoute" />
            <span>{{ $t('rule.add') }}</span>
          </a-space>
        </a-col>
        <a-col
          v-if="addRoute"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('ruleset.routeTo')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="routeTo"
              allow-clear
              style="width: 100%;"
              @clear="routeTo = ''"
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
      </a-row>

      <template v-if="preview.length > 0">
        <a-divider style="margin: 12px 0;" />
        <a-space style="margin-bottom: 8px;">
          <span>{{ $t('rule.import.preview') }}</span>
          <a-badge
            :count="newCount"
            :number-style="{ backgroundColor: '#52c41a' }"
            show-zero
          />
          <a-tooltip
            v-if="skipped > 0"
            :title="$t('rule.import.skipped')"
          >
            <a-badge
              :count="skipped"
              :number-style="{ backgroundColor: '#faad14' }"
            />
          </a-tooltip>
        </a-space>
        <a-table
          :data-source="preview"
          :columns="previewColumns"
          :pagination="false"
          size="small"
          row-key="tag"
          :custom-row="customRow"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'url'">
              <a-tooltip :title="record.url">
                <span dir="ltr">.../{{ record.url.split('/').pop() ?? record.url }}</span>
              </a-tooltip>
            </template>
          </template>
        </a-table>
      </template>
    </a-form>

    <template #footer>
      <a-button
        type="default"
        @click="$emit('close')"
      >
        {{ $t('actions.close') }}
      </a-button>
      <a-button
        type="primary"
        :disabled="newCount === 0 && !(addRoute && routeTo && preview.length > 0)"
        @click="save"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import { downloadHttpClient } from '@/plugins/httpClient'
import { CatalogRuleset, geoCatalog, geoList } from '@/plugins/rulesetCatalog'
import { ruleset } from '@/types/rules'
import type { PropType } from 'vue'

// A catalog entry plus whether the config already has a rule-set with that tag.
type PresetItem = CatalogRuleset & { exists: boolean }

export default {
  props: {
    visible: { type: Boolean, required: true },
    outTags: { type: Array as PropType<string[]>, required: true },
    rsTags: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['save', 'close'],
  data() {
    return {
      selected: [] as string[],
      detour: '',
      interval: 1,
      addRoute: false,
      routeTo: '',
      catalogItems: geoList,
    }
  },
  computed: {
    preview(): PresetItem[] {
      const existing = new Set(this.rsTags)
      return this.selected
        .map((tag: string) => geoCatalog.find(g => g.tag == tag))
        .filter((g): g is CatalogRuleset => g != undefined)
        .map(g => ({ ...g, exists: existing.has(g.tag) }))
    },
    previewColumns() {
      return [
        { title: this.$t('objects.tag'), dataIndex: 'tag', key: 'tag' },
        { title: 'URL', dataIndex: 'url', key: 'url' },
      ]
    },
    skipped(): number {
      return this.preview.filter(i => i.exists).length
    },
    newCount(): number {
      return this.preview.filter(i => !i.exists).length
    },
  },
  watch: {
    visible(v) {
      if (v) {
        this.selected = []
        this.detour = ''
        this.interval = 1
        this.addRoute = false
        this.routeTo = ''
      }
    },
  },
  methods: {
    customRow(record: PresetItem) {
      return { style: record.exists ? { opacity: 0.4 } : {} }
    },
    save() {
      const toAdd = this.preview.filter(i => !i.exists).map(item => {
        const rs: ruleset = { type: 'remote', tag: item.tag, format: item.format, url: item.url }
        const httpClient = downloadHttpClient(this.detour)
        if (httpClient) rs.http_client = httpClient
        if (this.interval > 0) rs.update_interval = this.interval + 'd'
        return rs
      })
      // Optionally create one routing rule sending the selected rule-sets
      // (including already existing ones) to the chosen outbound.
      const rule = this.addRoute && this.routeTo && this.preview.length > 0
        ? { rule_set: this.preview.map(i => i.tag), action: 'route', outbound: this.routeTo }
        : null
      this.$emit('save', toAdd, rule)
    },
  },
}
</script>

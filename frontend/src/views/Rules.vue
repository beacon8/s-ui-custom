<template>
  <RuleVue
    v-model="ruleModal.visible"
    :visible="ruleModal.visible"
    :index="ruleModal.index"
    :data="ruleModal.data"
    :clients="clients"
    :in-tags="inboundTags"
    :out-tags="outboundTags"
    :rs-tags="rulesetTags"
    @close="closeRuleModal"
    @save="saveRuleModal"
  />
  <RulesetVue
    v-model="rulesetModal.visible"
    :visible="rulesetModal.visible"
    :index="rulesetModal.index"
    :data="rulesetModal.data"
    :out-tags="outboundTags"
    @close="closeRulesetModal"
    @save="saveRulesetModal"
  />
  <RuleImport
    v-model="importRulesModal.visible"
    :visible="importRulesModal.visible"
    :existing-rules-count="rules.length"
    :existing-rulesets-count="rulesets.length"
    :existing-ruleset-tags="rulesetTags"
    @save="saveImportRule"
    @close="closeImportRule"
  />
  <RulesetImport
    v-model="importRulesetsModal.visible"
    :visible="importRulesetsModal.visible"
    :out-tags="outboundTags"
    :rs-tags="rulesetTags"
    @save="saveImportRulesets"
    @close="closeImportRulesets"
  />
  <RulesetPreset
    v-model="presetRulesetsModal.visible"
    :visible="presetRulesetsModal.visible"
    :out-tags="outboundTags"
    :rs-tags="rulesetTags"
    @save="savePresetRulesets"
    @close="presetRulesetsModal.visible = false"
  />
  <div class="rules-toolbar">
    <div
      class="rules-tabs"
      role="tablist"
    >
      <button
        type="button"
        role="tab"
        class="rules-tab"
        :class="{ active: activeTab === 'rules' }"
        :aria-selected="activeTab === 'rules'"
        @click="activeTab = 'rules'"
      >
        {{ $t('rule.rulesTab') }} ({{ rules.length }})
      </button>
      <button
        type="button"
        role="tab"
        class="rules-tab"
        :class="{ active: activeTab === 'rulesets' }"
        :aria-selected="activeTab === 'rulesets'"
        @click="activeTab = 'rulesets'"
      >
        {{ $t('rule.ruleset') }} ({{ rulesets.length }})
      </button>
    </div>
    <div class="rules-actions">
      <a-button
        type="primary"
        @click="activeTab === 'rules' ? showRuleModal(-1) : showRulesetModal(-1)"
      >
        {{ activeTab === 'rules' ? $t('rule.add') : $t('ruleset.add') }}
      </a-button>
      <a-dropdown
        v-model:open="actionMenu"
        :trigger="['click']"
        placement="bottom"
      >
        <a-button
          :aria-label="$t('a11y.tools')"
          type="text"
          shape="circle"
        >
          <template #icon>
            <AppIcon
              name="tools"
              color="primary"
            />
          </template>
        </a-button>
        <template #overlay>
          <a-menu>
            <a-menu-item
              key="importRule"
              @click="showImportRule(); actionMenu = false"
            >
              <template #icon>
                <AppIcon name="routes" />
              </template>
              {{ $t('rule.import.rulesTitle') }}
            </a-menu-item>
            <a-menu-item
              key="importRulesets"
              @click="showImportRulesets(); actionMenu = false"
            >
              <template #icon>
                <AppIcon name="download-multiple" />
              </template>
              {{ $t('rule.import.title') }}
            </a-menu-item>
            <a-menu-item
              key="presetRulesets"
              @click="presetRulesetsModal.visible = true; actionMenu = false"
            >
              <template #icon>
                <AppIcon name="star-box-multiple" />
              </template>
              {{ $t('ruleset.preset') }}
            </a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
      <a-button
        :type="settingsOpen ? 'primary' : 'default'"
        @click="settingsOpen = !settingsOpen"
      >
        <template #icon>
          <AppIcon name="setting" />
        </template>
        {{ $t('basic.routing.settings') }}
      </a-button>
      <a-button
        type="primary"
        :loading="loading"
        :disabled="stateChange"
        @click="saveConfig"
      >
        {{ $t('actions.save') }}
      </a-button>
    </div>
  </div>

  <div
    v-if="settingsOpen"
    class="route-settings"
  >
    <a-form layout="vertical">
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('basic.routing.defaultOut')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="route.final"
              allow-clear
              style="width: 100%;"
              @change="(v: any) => { if (v === undefined) delete route.final }"
            >
              <a-select-option
                v-for="t in outboundTags"
                :key="t"
                :value="t"
              >
                {{ t }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <!-- Which DNS server resolves the domains outbounds dial. sing-box
             guesses when several servers exist and none is named. -->
          <a-form-item
            :label="$t('basic.routing.defaultDns')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="route.default_domain_resolver"
              allow-clear
              style="width: 100%;"
              @change="(v: any) => { if (v === undefined) delete route.default_domain_resolver }"
            >
              <a-select-option
                v-for="t in dnsTags"
                :key="t"
                :value="t"
              >
                {{ t }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('basic.routing.defaultIf')"
            style="margin-bottom: 0;"
          >
            <a-input
              v-model:value="route.default_interface"
              allow-clear
              @change="(e: any) => { if (!e.target?.value) delete route.default_interface }"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('basic.routing.defaultRm')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="routeMark"
              :min="0"
              style="width: 100%;"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <!-- Which shared client downloads the rule-sets that name none. -->
          <a-form-item
            :label="$t('basic.httpClient.title')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="route.default_http_client"
              allow-clear
              :not-found-content="$t('basic.httpClient.none')"
              style="width: 100%;"
              @change="(v: any) => { if (v === undefined) delete route.default_http_client }"
            >
              <a-select-option
                v-for="t in httpClients"
                :key="t"
                :value="t"
              >
                {{ t }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('basic.routing.autoBind')"
            style="margin-bottom: 0;"
          >
            <div class="auto-bind-control">
              <a-switch v-model:checked="route.auto_detect_interface" />
            </div>
          </a-form-item>
        </a-col>
      </a-row>
    </a-form>
  </div>

  <div class="rules-content">
    <template v-if="activeTab === 'rules'">
      <a-table
        :columns="ruleHeaders"
        :data-source="ruleRows"
        :pagination="false"
        :scroll="{ x: 'max-content' }"
        :row-key="(record: RuleRow) => record.__key"
        :custom-row="ruleRowProps"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'index'">
            {{ record.__index + 1 }}
          </template>
          <template v-else-if="column.key === 'type'">
            {{ record.type !== undefined ? $t('rule.logical') + ' (' + record.mode + ')' : $t('rule.simple') }}
          </template>
          <template v-else-if="column.key === 'action'">
            {{ record.action }}
          </template>
          <template v-else-if="column.key === 'outbound'">
            {{ record.outbound ?? '-' }}
          </template>
          <template v-else-if="column.key === 'conditions'">
            {{ record.rules ? record.rules.length : Object.keys(record).filter(r => !actionKeys.includes(r) && !r.startsWith('__')).length }}
          </template>
          <template v-else-if="column.key === 'invert'">
            {{ $t((record.invert ?? false) ? 'yes' : 'no') }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space size="small">
              <a-tooltip :title="$t('actions.edit')">
                <a-button
                  type="text"
                  shape="circle"
                  :aria-label="$t('actions.edit')"
                  @click="showRuleModal(record.__index)"
                >
                  <template #icon>
                    <AppIcon name="file-edit" />
                  </template>
                </a-button>
              </a-tooltip>
              <a-popconfirm
                :title="$t('confirm')"
                :ok-text="$t('yes')"
                :cancel-text="$t('no')"
                ok-type="danger"
                @confirm="delRule(record.__index)"
              >
                <a-tooltip :title="$t('actions.del')">
                  <a-button
                    type="text"
                    danger
                    shape="circle"
                    :aria-label="$t('actions.del')"
                  >
                    <template #icon>
                      <AppIcon name="file-remove" />
                    </template>
                  </a-button>
                </a-tooltip>
              </a-popconfirm>
              <a-tooltip :title="$t('actions.moveUp')">
                <a-button
                  type="text"
                  shape="circle"
                  :aria-label="$t('actions.moveUp')"
                  :disabled="record.__index === 0"
                  @click="moveRule(record.__index, -1)"
                >
                  <template #icon>
                    <AppIcon name="arrow-up" />
                  </template>
                </a-button>
              </a-tooltip>
              <a-tooltip :title="$t('actions.moveDown')">
                <a-button
                  type="text"
                  shape="circle"
                  :aria-label="$t('actions.moveDown')"
                  :disabled="record.__index === rules.length - 1"
                  @click="moveRule(record.__index, 1)"
                >
                  <template #icon>
                    <AppIcon name="arrow-down" />
                  </template>
                </a-button>
              </a-tooltip>
            </a-space>
          </template>
        </template>
      </a-table>
    </template>
    <template v-else>
      <a-table
        v-if="rulesetRows.length"
        :columns="rulesetHeaders"
        :data-source="rulesetRows"
        :pagination="false"
        :scroll="{ x: 'max-content' }"
        row-key="__key"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'">
            {{ $t('ruleset.' + record.type) }}
          </template>
          <template v-else-if="column.key === 'detour'">
            {{ record.http_client?.detour ?? '-' }}
          </template>
          <template v-else-if="column.key === 'interval'">
            {{ record.update_interval ?? '-' }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space size="small">
              <a-tooltip :title="$t('actions.edit')">
                <a-button
                  type="text"
                  shape="circle"
                  :aria-label="$t('actions.edit')"
                  @click="showRulesetModal(record.__index)"
                >
                  <template #icon>
                    <AppIcon name="file-edit" />
                  </template>
                </a-button>
              </a-tooltip>
              <a-popconfirm
                :title="$t('confirm')"
                :ok-text="$t('yes')"
                :cancel-text="$t('no')"
                ok-type="danger"
                @confirm="delRuleset(record.__index)"
              >
                <a-tooltip :title="$t('actions.del')">
                  <a-button
                    type="text"
                    danger
                    shape="circle"
                    :aria-label="$t('actions.del')"
                  >
                    <template #icon>
                      <AppIcon name="file-remove" />
                    </template>
                  </a-button>
                </a-tooltip>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
      <div
        v-else
        class="ruleset-empty"
      >
        <span>{{ $t('ruleset.empty') }}</span>
        <a-button
          type="link"
          size="small"
          @click="showRulesetModal(-1)"
        >
          {{ $t('ruleset.add') }}
        </a-button>
      </div>
    </template>
  </div>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import { httpClientTags } from '@/plugins/httpClient'
import { computed, ref, onBeforeMount } from 'vue'
import RuleVue from '@/layouts/modals/Rule.vue'
import RulesetVue from '@/layouts/modals/Ruleset.vue'
import RulesetImport from '@/layouts/modals/RulesetImport.vue'
import RulesetPreset from '@/layouts/modals/RulesetPreset.vue'
import RuleImport from '@/layouts/modals/RuleImport.vue'
import { Config } from '@/types/config'
import { actionKeys, logicalRule, rule, ruleset } from '@/types/rules'
import { FindDiff } from '@/plugins/utils'
import { i18n } from '@/locales'

// A routing rule as this page moves it around. The rule modal fills in a whole
// one, but an imported or preset rule carries only the keys it sets, and a
// logical rule carries its own, so nothing here is required.
type RuleEntry = Partial<rule & logicalRule>

// The route block as this page edits it: every field is optional, because the
// template clears them one at a time, and the rules and rule-sets are the
// shapes the modals produce rather than the narrower ones the generated config
// declares.
interface RouteBlock {
  rules?: RuleEntry[]
  rule_set?: ruleset[]
  final?: string
  default_domain_resolver?: string
  default_interface?: string
  default_http_client?: string
  default_mark?: number
  auto_detect_interface?: boolean
}

// What the rule-import modal hands back: the route block it parsed out of a
// pasted config.
interface ImportedRouteBlock {
  rules?: RuleEntry[]
  rule_set?: ruleset[]
  final?: string
}

const oldConfig = ref({})
const loading = ref(false)
const actionMenu = ref(false)
const activeTab = ref<'rules' | 'rulesets'>('rules')
const settingsOpen = ref(false)
const appConfig = computed((): Config => {
  return <Config> Data().config
})

onBeforeMount(async () => {
  loading.value = true
  while (Data().lastLoad == 0) {
    await new Promise(resolve => setTimeout(resolve, 100))
  }
  oldConfig.value = JSON.parse(JSON.stringify(Data().config))
  loading.value = false
})

const routeMark = computed({
  get() { return route.value.default_mark ?? 0 },
  set(v: number | null) {
    if (v && v > 0) route.value.default_mark = v
    else delete appConfig.value.route.default_mark
  }
})

const stateChange = computed(() => FindDiff.deepCompare(appConfig.value, oldConfig.value))

const saveConfig = async () => {
  loading.value = true
  const success = await Data().save("config", "set", appConfig.value)
  if (success) {
    oldConfig.value = JSON.parse(JSON.stringify(Data().config))
    loading.value = false
  }
}

const clients = computed((): string[] => Data().clients.map(c => c.name))
const route = computed((): RouteBlock => (appConfig.value.route as unknown as RouteBlock) ?? {})

const rules = computed((): RuleEntry[] => {
  const data = route.value
  if (!data) return []
  if (!('rules' in data) || !Array.isArray(data.rules)) data.rules = []
  return data.rules
})

const rulesets = computed((): ruleset[] => {
  const data = route.value
  if (!data) return []
  if (!('rule_set' in data) || !Array.isArray(data.rule_set)) data.rule_set = []
  return data.rule_set
})

type RulesetRow = ruleset & { __index: number, __key: string | number }
type RuleRow = RuleEntry & { __index: number, __key: number }

const rulesetRows = computed((): RulesetRow[] => rulesets.value.map((item, index) => ({
  ...item,
  __index: index,
  __key: item.tag ?? index,
})))

const ruleRows = computed((): RuleRow[] => rules.value.map((item, index) => ({
  ...item,
  __index: index,
  __key: index,
})))

const rulesetHeaders = computed(() => [
  { title: i18n.global.t('objects.tag'), dataIndex: 'tag', key: 'tag' },
  { title: i18n.global.t('type'), key: 'type' },
  { title: i18n.global.t('ruleset.format'), dataIndex: 'format', key: 'format' },
  { title: i18n.global.t('objects.outbound'), key: 'detour' },
  { title: i18n.global.t('ruleset.interval'), key: 'interval' },
  { title: i18n.global.t('actions.operations'), key: 'actions' },
])

const ruleHeaders = computed(() => [
  { title: '#', key: 'index', width: 60 },
  { title: i18n.global.t('type'), key: 'type' },
  { title: i18n.global.t('admin.action'), key: 'action' },
  { title: i18n.global.t('objects.outbound'), key: 'outbound' },
  { title: i18n.global.t('rule.matchConditions'), key: 'conditions' },
  { title: i18n.global.t('rule.invert'), key: 'invert' },
  { title: i18n.global.t('actions.operations'), key: 'actions' },
])

const rulesetTags = computed((): string[] => rulesets.value.map(rs => rs.tag))

const httpClients = computed((): string[] => httpClientTags())

const dnsTags = computed((): string[] =>
  (appConfig.value.dns?.servers ?? []).map(s => s.tag).filter((t: string) => t?.length > 0))

const outboundTags = computed((): string[] => [
  ...(Data().outbounds?.map(o => o.tag) ?? []),
  ...(Data().endpoints?.map(e => e.tag) ?? [])
])

const inboundTags = computed((): string[] => [
  ...(Data().inbounds?.map(o => o.tag) ?? []),
  ...(Data().endpoints?.filter(e => e.listen_port > 0).map(e => e.tag) ?? [])
])

const ruleModal = ref({ visible: false, index: -1, data: "" })
const showRuleModal = (index: number) => {
  ruleModal.value.index = index
  ruleModal.value.data = index == -1 ? '' : JSON.stringify(rules.value[index])
  ruleModal.value.visible = true
}
const closeRuleModal = () => { ruleModal.value.visible = false }
const saveRuleModal = (data:RuleEntry) => {
  if (ruleModal.value.index == -1) rules.value.push(data)
  else rules.value[ruleModal.value.index] = data
  ruleModal.value.visible = false
}
const delRule = (index: number) => { rules.value.splice(index, 1) }

const rulesetModal = ref({ visible: false, index: -1, data: "" })
const showRulesetModal = (index: number) => {
  rulesetModal.value.index = index
  rulesetModal.value.data = index == -1 ? '' : JSON.stringify(rulesets.value[index])
  rulesetModal.value.visible = true
}
const closeRulesetModal = () => { rulesetModal.value.visible = false }
const saveRulesetModal = (data:ruleset) => {
  if (rulesetModal.value.index == -1) rulesets.value.push(data)
  else rulesets.value[rulesetModal.value.index] = data
  rulesetModal.value.visible = false
}
const delRuleset = (index: number) => { rulesets.value.splice(index, 1) }

// Dragging is mouse-only: HTML5 drag and drop never starts from a touch in
// most mobile browsers, so the buttons are the way to reorder there (#1283).
const moveRule = (index: number, delta: number) => {
  const target = index + delta
  if (target < 0 || target >= rules.value.length) return
  const [item] = rules.value.splice(index, 1)
  rules.value.splice(target, 0, item)
}

const draggedItemIndex = ref<number | null>(null)
const onDragStart = (index: number) => { draggedItemIndex.value = index }
const onDrop = (index: number) => {
  if (draggedItemIndex.value !== null) {
    const draggedItem = rules.value[draggedItemIndex.value]
    rules.value.splice(draggedItemIndex.value, 1)
    rules.value.splice(index, 0, draggedItem)
    draggedItemIndex.value = null
  }
}

function ruleRowProps(record: RuleRow) {
  return {
    draggable: true,
    onDragstart: () => onDragStart(record.__index),
    onDragover: (event: DragEvent) => event.preventDefault(),
    onDrop: () => onDrop(record.__index),
    style: { cursor: 'move' },
  }
}

const importRulesModal = ref({ visible: false })

function showImportRule() {
  importRulesModal.value.visible = true
}

function closeImportRule() {
  importRulesModal.value.visible = false
}

function saveImportRule(block: ImportedRouteBlock, mode: 'merge' | 'replace', applyFinal: boolean) {
  if (mode === 'replace') {
    route.value.rules = block.rules ?? []
    route.value.rule_set = block.rule_set ?? []
  } else {
    const existingTags = new Set(rulesetTags.value)
    if (block.rules) rules.value.push(...block.rules)
    if (block.rule_set) {
      for (const rs of block.rule_set) {
        if (!existingTags.has(rs.tag)) rulesets.value.push(rs)
      }
    }
  }
  if (applyFinal && block.final) route.value.final = block.final
  importRulesModal.value.visible = false
}

const importRulesetsModal = ref({ visible: false })

function showImportRulesets() {
  importRulesetsModal.value.visible = true
}

function closeImportRulesets() {
  importRulesetsModal.value.visible = false
}

function saveImportRulesets(items: ruleset[]) {
  rulesets.value.push(...items)
  importRulesetsModal.value.visible = false
}

const presetRulesetsModal = ref({ visible: false })

function savePresetRulesets(items: ruleset[], rule: RuleEntry | null) {
  rulesets.value.push(...items)
  if (rule) rules.value.push(rule)
  presetRulesetsModal.value.visible = false
}
</script>


<style scoped>
.rules-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
}

.rules-tabs {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.rules-tab {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  white-space: nowrap;
}

.rules-tab:hover {
  background: rgba(128, 128, 128, 0.1);
}

.rules-tab.active {
  color: #1677ff;
  background: rgba(22, 119, 255, 0.1);
  font-weight: 600;
}

.rules-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.route-settings {
  border-top: 1px solid rgba(128, 128, 128, 0.16);
  padding: 10px 0 2px;
  margin-bottom: 8px;
}

.route-settings .ant-form-item {
  margin-bottom: 0;
}

.auto-bind-control {
  display: flex;
  align-items: center;
  min-height: 32px;
}

.rules-content {
  min-width: 0;
}

.ruleset-empty {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  border: 1px solid rgba(128, 128, 128, 0.16);
  border-radius: 6px;
  font-size: 13px;
  opacity: 0.65;
}

.ruleset-empty .ant-btn {
  height: auto;
  padding: 0;
}

@media (max-width: 575px) {
  .rules-toolbar {
    align-items: flex-start;
  }

  .rules-actions {
    width: 100%;
    margin-left: 0;
  }

  .rules-tab {
    padding: 6px 8px;
  }
}
</style>

<template>
  <ExpTextarea
    v-model="expTextarea.visible"
    :visible="expTextarea.visible"
    :label="expTextarea.title"
    :content="expTextarea.content"
    @update="saveExpTextarea"
    @close="closeExpTextarea"
  />
  <div class="match-section">
    <div class="match-header">
      <span class="section-title">{{ $t('rule.matchConditions') }}</span>
      <a-space
        size="small"
        wrap
      >
        <a-space size="small">
          <a-switch
            v-model:checked="invert"
            size="small"
          />
          <span class="header-label">{{ $t('rule.invert') }}</span>
        </a-space>
        <a-button
          size="small"
          @click="pickerOpen = !pickerOpen"
        >
          <template
            v-if="!pickerOpen"
            #icon
          >
            <AppIcon name="plus" />
          </template>
          {{ pickerOpen ? $t('rule.collapse') : $t('rule.addCondition') }}
        </a-button>
      </a-space>
    </div>

    <div
      v-if="pickerOpen"
      class="condition-picker"
    >
      <a-input
        v-model:value="conditionSearch"
        allow-clear
        class="condition-search"
        :placeholder="$t('rule.searchCondition')"
      >
        <template #prefix>
          <AppIcon name="magnify" />
        </template>
      </a-input>
      <div class="condition-panel">
        <div class="condition-groups">
          <button
            v-for="group in groups"
            :key="group.key"
            type="button"
            class="condition-group"
            :class="{ active: currentGroup === group.key && !conditionSearch.trim() }"
            @click="selectGroup(group.key)"
          >
            {{ $t(group.label) }}
          </button>
        </div>
        <a-select
          :value="currentGroup"
          class="condition-group-select"
          :options="groupOptions"
          @change="selectGroup"
        />
        <div class="condition-list">
          <div
            v-for="def in visibleConditions"
            :key="def.key"
            class="condition-option"
            @click="addCondition(def.key)"
          >
            <div class="condition-option-text">
              <span class="condition-option-name">{{ $t(def.label) }}</span>
              <span
                v-if="conditionSearch.trim()"
                class="condition-group-hint"
              >
                {{ $t(groupLabel(def.group)) }}
              </span>
            </div>
            <span
              v-if="isActive(def.key)"
              class="condition-added"
            >
              ✓ {{ $t('rule.added') }}
            </span>
            <span
              v-else
              class="condition-add"
            >
              ＋
            </span>
          </div>
          <div
            v-if="visibleConditions.length === 0"
            class="match-empty"
          >
            {{ $t('rule.noMatchingConditions') }}
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="activeConditions.length === 0"
      class="match-empty"
    >
      {{ $t('rule.noConditions') }}
    </div>
    <a-row
      v-else
      :gutter="[12, 12]"
    >
      <a-col
        v-for="cond in activeConditions"
        :key="cond.key"
        :xs="24"
        :sm="cond.wide ? 24 : 12"
      >
        <div class="condition-card">
          <div class="condition-head">
            <span class="condition-label">{{ $t(cond.label) }}</span>
            <a-button
              type="text"
              danger
              size="small"
              shape="circle"
              :aria-label="$t('actions.del')"
              @click="removeCondition(cond.key)"
            >
              <template #icon>
                <AppIcon name="close" />
              </template>
            </a-button>
          </div>

          <template v-if="cond.kind === 'tags'">
            <a-select
              :id="conditionInputId(cond.key)"
              :value="asArray(rule[cond.key])"
              mode="tags"
              style="width: 100%"
              @update:value="setCondition(cond.key, $event)"
            >
              <a-select-option
                v-for="item in optionsFor(cond)"
                :key="String(item)"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </template>

          <template v-else-if="cond.kind === 'multi'">
            <a-select
              v-if="cond.key === 'protocol'"
              :id="conditionInputId(cond.key)"
              :value="asArray(rule.protocol)"
              mode="multiple"
              :options="protocols"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
              @update:value="setCondition('protocol', $event)"
            />
            <a-select
              v-else
              :id="conditionInputId(cond.key)"
              :value="asArray(rule.network)"
              mode="multiple"
              style="width: 100%"
              @update:value="setCondition('network', $event)"
            >
              <a-select-option
                v-for="item in ['tcp', 'udp', 'icmp']"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </template>

          <template v-else-if="cond.kind === 'select'">
            <a-select
              :id="conditionInputId(cond.key)"
              :value="rule.ip_version"
              style="width: 100%"
              @update:value="setCondition('ip_version', $event)"
            >
              <a-select-option
                v-for="item in [4, 6]"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </template>

          <template v-else-if="cond.kind === 'switch'">
            <a-switch
              :id="conditionInputId(cond.key)"
              :checked="Boolean(rule[cond.key])"
              @change="setCondition(cond.key, !!$event)"
            />
          </template>

          <template v-else-if="cond.kind === 'ruleSet'">
            <a-select
              :id="conditionInputId(cond.key)"
              :value="asArray(rule.rule_set)"
              mode="tags"
              style="width: 100%"
              @update:value="setCondition('rule_set', $event)"
            >
              <a-select-option
                v-for="item in optionsFor(cond)"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
            <div class="inline-toggle">
              <a-switch
                :checked="Boolean(rule.rule_set_ip_cidr_match_source)"
                size="small"
                @change="setCondition('rule_set_ip_cidr_match_source', !!$event)"
              />
              <span>{{ $t('rule.rulesetMatchSrc') }}</span>
            </div>
          </template>

          <template v-else>
            <a-flex
              align="start"
              gap="small"
              style="min-width: 0;"
            >
              <a-textarea
                :id="conditionInputId(cond.key)"
                :value="lineText(cond.key)"
                :rows="cond.rows ?? 3"
                class="condition-textarea"
                style="resize: vertical"
                @update:value="setText(cond.key, $event)"
              />
              <a-button
                type="text"
                shape="circle"
                :aria-label="$t('rule.expand')"
                @click="openExpTextarea($t(cond.label), cond.key)"
              >
                <template #icon>
                  <AppIcon name="arrow-expand" />
                </template>
              </a-button>
            </a-flex>
          </template>
        </div>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, useId } from 'vue'
import { i18n } from '@/locales'
import ExpTextarea from '@/components/ExpTextarea.vue'
import type { rule as RouteRule } from '@/types/rules'

interface RuleShape {
  [key: string]: unknown
  inbound?: string[]
  auth_user?: string[]
  ip_version?: number
  network?: string[]
  protocol?: string[]
  domain?: string[]
  domain_suffix?: string[]
  domain_keyword?: string[]
  domain_regex?: string[]
  ip_cidr?: string[]
  ip_is_private?: boolean
  port?: number[]
  port_range?: string[]
  source_ip_cidr?: string[]
  source_ip_is_private?: boolean
  source_port?: number[]
  source_port_range?: string[]
  preferred_by?: string[]
  interface_address?: string[]
  network_interface_address?: string[]
  default_interface_address?: string[]
  rule_set?: string[]
  rule_set_ip_cidr_match_source?: boolean
}

type ConditionKind = 'tags' | 'multi' | 'select' | 'switch' | 'ruleSet' | 'text'

interface ConditionDef {
  key: string
  label: string
  kind: ConditionKind
  group: string
  wide?: boolean
  rows?: number
}

const model = defineModel<RouteRule>('rule', { required: true })
const invert = defineModel<boolean>('invert', { default: false })
const rule = computed(() => model.value as unknown as RuleShape)

const props = defineProps<{
  clients?: string[]
  inTags?: string[]
  outTags?: string[]
  rsTags?: string[]
}>()

const protocols = [
  { title: 'HTTP', value: 'http' },
  { title: 'TLS', value: 'tls' },
  { title: 'QUIC', value: 'quic' },
  { title: 'STUN', value: 'stun' },
  { title: 'DNS', value: 'dns' },
  { title: 'BitTorrent', value: 'bittorrent' },
  { title: 'DTLS', value: 'dtls' },
  { title: 'SSH', value: 'ssh' },
  { title: 'RDP', value: 'rdp' },
  { title: 'NTP', value: 'ntp' },
]

const groups = [
  { key: 'inboundUser', label: 'rule.groupInboundUser' },
  { key: 'networkProtocol', label: 'rule.groupNetworkProtocol' },
  { key: 'domain', label: 'rule.groupDomain' },
  { key: 'ipPort', label: 'rule.groupIpPort' },
  { key: 'rulesetInterface', label: 'rule.groupRulesetInterface' },
]

const conditionDefs: ConditionDef[] = [
  { key: 'inbound', group: 'inboundUser', label: 'pages.inbounds', kind: 'tags' },
  { key: 'auth_user', group: 'inboundUser', label: 'pages.clients', kind: 'tags' },
  { key: 'ip_version', group: 'networkProtocol', label: 'rule.ipVer', kind: 'select' },
  { key: 'network', group: 'networkProtocol', label: 'network', kind: 'multi' },
  { key: 'protocol', group: 'networkProtocol', label: 'protocol', kind: 'multi' },
  { key: 'domain', group: 'domain', label: 'rule.domain', kind: 'text' },
  { key: 'domain_suffix', group: 'domain', label: 'rule.domainSufix', kind: 'text' },
  { key: 'domain_keyword', group: 'domain', label: 'rule.domainKw', kind: 'text' },
  { key: 'domain_regex', group: 'domain', label: 'rule.domainRgx', kind: 'text' },
  { key: 'ip_cidr', group: 'ipPort', label: 'rule.ip', kind: 'text' },
  { key: 'ip_is_private', group: 'ipPort', label: 'rule.privateIp', kind: 'switch' },
  { key: 'port', group: 'ipPort', label: 'rule.port', kind: 'text' },
  { key: 'port_range', group: 'ipPort', label: 'rule.portRange', kind: 'text' },
  { key: 'source_ip_cidr', group: 'ipPort', label: 'rule.srcCidr', kind: 'text' },
  { key: 'source_ip_is_private', group: 'ipPort', label: 'rule.srcPrivateIp', kind: 'switch' },
  { key: 'source_port', group: 'ipPort', label: 'rule.srcPort', kind: 'text' },
  { key: 'source_port_range', group: 'ipPort', label: 'rule.srcPortRange', kind: 'text' },
  { key: 'rule_set', group: 'rulesetInterface', label: 'rule.ruleset', kind: 'ruleSet' },
  { key: 'preferred_by', group: 'rulesetInterface', label: 'rule.preferredBy', kind: 'tags' },
  { key: 'interface_address', group: 'rulesetInterface', label: 'rule.interfaceAddr', kind: 'text' },
  { key: 'network_interface_address', group: 'rulesetInterface', label: 'rule.networkInterfaceAddr', kind: 'text' },
  { key: 'default_interface_address', group: 'rulesetInterface', label: 'rule.defaultInterfaceAddr', kind: 'text' },
]

const defsByKey = new Map(conditionDefs.map((def) => [def.key, def]))

const activeConditions = computed(() => conditionDefs.filter((def) => rule.value[def.key] !== undefined))

const pickerOpen = ref(false)
const conditionSearch = ref('')
const currentGroup = ref('inboundUser')
const uid = useId()
const groupOptions = computed(() => groups.map((group) => ({
  value: group.key,
  label: i18n.global.t(group.label),
})))
const visibleConditions = computed(() => {
  const query = conditionSearch.value.trim().toLowerCase()
  if (query) {
    return conditionDefs.filter((def) => {
      const label = i18n.global.t(def.label).toLowerCase()
      return label.includes(query) || def.key.toLowerCase().includes(query)
    })
  }
  return conditionDefs.filter((def) => def.group === currentGroup.value)
})
const expTextarea = ref<{ visible: boolean, title: string, content: string, object: string }>({
  visible: false,
  title: '',
  content: '',
  object: '',
})

function conditionInputId(key: string): string {
  return `rule-condition-${uid}-${key}`
}

function groupLabel(group: string): string {
  return groups.find((item) => item.key === group)?.label ?? ''
}

function selectGroup(group: string) {
  currentGroup.value = group
  conditionSearch.value = ''
}

function asArray(value: unknown): (string | number)[] {
  return Array.isArray(value) ? value as (string | number)[] : []
}

function lineText(key: string): string {
  const value = rule.value[key]
  return Array.isArray(value) ? value.join('\n') : ''
}

function setCondition(key: string, value: unknown) {
  rule.value[key] = value
}

function splitLines(value: string): string[] {
  return value.length > 0 ? value.split('\n').map((line) => line.trim()).filter((line) => line.length > 0) : []
}

function setText(key: string, value: string) {
  if (key === 'port' || key === 'source_port') {
    if (value.endsWith('\n')) return
    const numbers = splitLines(value).map((line) => parseInt(line, 10)).filter((number) => !isNaN(number))
    setCondition(key, numbers)
    return
  }
  setCondition(key, splitLines(value))
}

function optionsFor(def: ConditionDef): (string | number)[] {
  switch (def.key) {
    case 'inbound':
      return props.inTags ?? []
    case 'auth_user':
      return props.clients ?? []
    case 'rule_set':
      return props.rsTags ?? []
    case 'preferred_by':
      return props.outTags?.length ? props.outTags : (props.inTags ?? [])
    default:
      return []
  }
}

function isActive(key: string): boolean {
  return rule.value[key] !== undefined
}

async function addCondition(key: string) {
  const def = defsByKey.get(key)
  if (!def || isActive(key)) return
  switch (def.kind) {
    case 'switch':
      setCondition(key, true)
      break
    case 'select':
      setCondition(key, 4)
      break
    case 'multi':
      setCondition(key, key === 'protocol' ? ['http'] : [])
      break
    case 'ruleSet':
      setCondition(key, [])
      setCondition('rule_set_ip_cidr_match_source', false)
      break
    default:
      setCondition(key, [])
  }
  pickerOpen.value = false
  conditionSearch.value = ''
  await nextTick()
  document.getElementById(conditionInputId(key))?.focus()
}

function removeCondition(key: string) {
  delete rule.value[key]
  if (key === 'rule_set') delete rule.value.rule_set_ip_cidr_match_source
}


function openExpTextarea(title: string, object: string) {
  expTextarea.value.visible = true
  expTextarea.value.title = title
  expTextarea.value.content = lineText(object)
  expTextarea.value.object = object
}

function saveExpTextarea(results: string[]) {
  setText(expTextarea.value.object, results.join('\n'))
  closeExpTextarea()
}

function closeExpTextarea() {
  expTextarea.value.visible = false
  expTextarea.value.title = ''
  expTextarea.value.object = ''
}
</script>

<style scoped>
.match-section {
  border: 1px solid rgba(128, 128, 128, 0.22);
  border-radius: 8px;
  padding: 12px;
  min-width: 0;
}

.match-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.header-label,
.condition-label,
.subrule-head span {
  overflow-wrap: anywhere;
  word-break: break-word;
}

.header-label {
  font-size: 12px;
}

.condition-picker {
  background: rgba(128, 128, 128, 0.05);
  border-radius: 8px;
  padding: 10px;
  margin-bottom: 12px;
}

.condition-search {
  margin-bottom: 8px;
}

.condition-panel {
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr);
  gap: 12px;
  align-items: start;
}

.condition-groups {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.condition-group {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  padding: 7px 8px;
  border-radius: 6px;
  cursor: pointer;
  overflow-wrap: anywhere;
}

.condition-group:hover {
  background: rgba(128, 128, 128, 0.1);
}

.condition-group.active {
  background: rgba(22, 119, 255, 0.1);
  color: #1677ff;
  font-weight: 600;
}

.condition-group-select {
  display: none;
}

.condition-list {
  min-width: 0;
  max-height: 300px;
  overflow-y: auto;
}

.condition-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 7px 8px;
  border-radius: 6px;
  cursor: pointer;
}

.condition-option:hover {
  background: rgba(128, 128, 128, 0.1);
}

.condition-option-text {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.condition-option-name {
  text-align: left;
  overflow-wrap: anywhere;
}

.condition-group-hint {
  font-size: 11px;
  opacity: 0.55;
}

.condition-add {
  color: #1677ff;
  font-weight: 600;
  flex: none;
}

.condition-added {
  color: #52c41a;
  font-size: 12px;
  flex: none;
  white-space: nowrap;
}

.match-empty {
  font-size: 13px;
  opacity: 0.55;
  padding: 2px 0;
  overflow-wrap: anywhere;
}

.condition-card {
  border: 1px solid rgba(128, 128, 128, 0.18);
  border-radius: 6px;
  padding: 8px;
  height: 100%;
  min-width: 0;
}

.condition-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  margin-bottom: 6px;
}

.condition-label {
  font-size: 12px;
  opacity: 0.7;
  min-width: 0;
}

.condition-textarea {
  min-width: 0;
  flex: 1;
}

.inline-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  font-size: 12px;
  overflow-wrap: anywhere;
}

@media (max-width: 575px) {
  .condition-panel {
    grid-template-columns: minmax(0, 1fr);
  }

  .condition-groups {
    display: none;
  }

  .condition-group-select {
    display: block;
    width: 100%;
  }

  .condition-list {
    max-height: none;
  }
}
</style>

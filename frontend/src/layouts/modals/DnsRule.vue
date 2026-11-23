<template>
  <a-modal
    :open="visible"
    :width="800"
    @cancel="closeModal"
  >
    <template #title>
      <a-flex
        justify="space-between"
        align="center"
        style="padding-right: 24px;"
      >
        <span>{{ $t('actions.' + title) + " " + $t('objects.dnsrule') }}</span>
        <DocLink section="dnsRule" />
      </a-flex>
    </template>
    <a-form layout="vertical">
      <a-flex
        justify="space-between"
        align="center"
        style="margin-bottom: 12px;"
      >
        <a-space>
          <a-switch v-model:checked="logical" />
          <span>{{ $t('rule.logical') }}</span>
        </a-space>
        <a-button
          v-if="logical"
          type="primary"
          @click="ruleData.rules.push(<dnsRule>{})"
        >
          {{ $t('actions.add') + " " + $t('objects.rule') }}
        </a-button>
      </a-flex>
      <template v-if="ruleData.type == 'logical'">
        <a-card
          v-for="(r, ruleIndex) in ruleData.rules"
          :key="ruleIndex"
          size="small"
          :title="$t('objects.rule') + ' ' + (Number(ruleIndex) + 1)"
          style="margin-bottom: 8px;"
        >
          <template #extra>
            <a-button
              v-if="ruleData.rules.length > 1"
              :aria-label="$t('actions.del')"
              type="text"
              size="small"
              danger
              @click="ruleData.rules.splice(ruleIndex, 1)"
            >
              <template #icon>
                <AppIcon name="delete" />
              </template>
            </a-button>
          </template>
          <RuleOptions
            :rule="r"
            :clients="clients"
            :in-tags="inTags"
            :rule-sets="ruleSets"
          />
        </a-card>
      </template>
      <RuleOptions
        v-else
        :rule="ruleData.rules[0]"
        :clients="clients"
        :in-tags="inTags"
        :rule-sets="ruleSets"
      />
      <a-row
        :gutter="[12, 12]"
        align="middle"
        style="margin-top: 12px; margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dns.rule.action.title')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="ruleData.action"
              :options="actions"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          v-if="logical"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('rule.mode')"
            style="margin-bottom: 0;"
          >
            <a-auto-complete
              v-model:value="ruleData.mode"
              :options="[{ value: 'and' }, { value: 'or' }]"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="ruleData.invert" />
            <span>{{ $t('rule.invert') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <a-card
        v-if="['route', 'route-options'].includes(ruleData.action ?? '')"
        size="small"
        :title="$t('dns.rule.action.route')"
        style="margin-bottom: 12px;"
      >
        <a-row
          v-if="ruleData.action == 'route'"
          :gutter="[12, 12]"
          style="margin-bottom: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('dns.server')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="ruleData.server"
                style="width: 100%"
              >
                <a-select-option
                  v-for="item in serverTags"
                  :key="item"
                  :value="item"
                >
                  {{ item }}
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
              :label="$t('rule.strategy')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="ruleData.strategy"
                :options="strategies"
                :field-names="{ label: 'title', value: 'value' }"
                allow-clear
                style="width: 100%"
                @change="(v: unknown) => { if (!v) delete ruleData.strategy }"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-space>
              <a-switch v-model:checked="ruleData.disable_cache" />
              <span>{{ $t('dns.disableCache') }}</span>
            </a-space>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('dns.rule.action.rewriteTtl')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="ruleData.rewrite_ttl"
                :min="0"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('dns.rule.action.clientSubnet')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="ruleData.client_subnet" />
            </a-form-item>
          </a-col>
        </a-row>
      </a-card>
      <a-card
        v-if="ruleData.action == 'reject'"
        size="small"
        :title="$t('dns.rule.action.reject')"
        style="margin-bottom: 12px;"
      >
        <a-row
          :gutter="[12, 12]"
          align="middle"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('rule.method')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="ruleData.method"
                :options="[{ title: 'Default', value: 'default' }, { title: 'Drop', value: 'drop' }]"
                :field-names="{ label: 'title', value: 'value' }"
                allow-clear
                style="width: 100%"
                @change="(v: unknown) => { if (!v) delete ruleData.method }"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-space>
              <a-switch v-model:checked="ruleData.no_drop" />
              <span>{{ $t('rule.noDrop') }}</span>
            </a-space>
          </a-col>
        </a-row>
      </a-card>
      <a-card
        v-if="ruleData.action == 'predefined'"
        size="small"
        :title="$t('dns.rule.action.predefined')"
        style="margin-bottom: 12px;"
      >
        <a-row
          :gutter="[12, 12]"
          style="margin-bottom: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="8"
          >
            <a-form-item
              :label="$t('dns.rule.action.rcode')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="ruleData.rcode"
                :options="predefinedRcode"
                :field-names="{ label: 'title', value: 'value' }"
                allow-clear
                style="width: 100%"
                @change="(v: unknown) => { if (!v) delete ruleData.rcode }"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row
          v-if="ruleData.rcode == 'NOERROR'"
          :gutter="[12, 12]"
        >
          <a-col
            :xs="24"
            :sm="16"
          >
            <a-form-item
              :label="$t('dns.rule.action.answer') + ' ' + $t('commaSeparated')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="answer" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="16"
          >
            <a-form-item
              :label="$t('dns.rule.action.ns') + ' ' + $t('commaSeparated')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="ns" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="16"
          >
            <a-form-item
              :label="$t('dns.rule.action.extra') + ' ' + $t('commaSeparated')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="extra" />
            </a-form-item>
          </a-col>
        </a-row>
      </a-card>
    </a-form>
    <template #footer>
      <a-button
        type="default"
        @click="closeModal"
      >
        {{ $t('actions.close') }}
      </a-button>
      <a-button
        type="primary"
        :loading="loading"
        @click="saveChanges"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import { PropType } from 'vue'
import { dnsRule, actionDnsRuleKeys } from '@/types/dns'


import RuleOptions from '@/components/DnsRule.vue'
import DocLink from '@/components/DocLink.vue'
import { i18n } from '@/locales'

// The form assembles a rule field by field and also addresses fields by key
// name at runtime, so it carries an index signature alongside what it reads.
interface DnsRuleFields {
  type?: string
  mode?: string
  rules?: dnsRule[]
  invert?: boolean
  action?: string
  server?: string
  strategy?: string
  disable_cache?: boolean
  rewrite_ttl?: number
  client_subnet?: string
  method?: string
  no_drop?: boolean
  rcode?: string
  answer?: string[]
  ns?: string[]
  extra?: string[]
  [key: string]: unknown
}

// `action` and `invert` stay optional: when an existing simple rule is
// loaded they are copied in field by field, so they are absent for a moment.
type DnsRuleDraft = DnsRuleFields & { type: string, mode: string, rules: dnsRule[] }

export default {
  components: { DocLink, RuleOptions },
  props: {
    visible: { type: Boolean, required: true },
    // A JSON string of the rule being edited, empty when adding a new one.
    data: { type: String, default: '' },
    index: { type: Number, required: true },
    clients: { type: Array as PropType<string[]>, default: () => [] },
    inTags: { type: Array as PropType<string[]>, default: () => [] },
    serverTags: { type: Array as PropType<string[]>, default: () => [] },
    ruleSets: { type: Array as PropType<string[]>, default: () => [] },
  },
  emits: ['close', 'save'],
  data() {
    return {
      title: 'add',
      loading: false,
      ruleData: <DnsRuleDraft>{
        type: 'logical',
        mode: 'and',
        rules: <dnsRule[]>[{}],
        invert: false,
        action: 'route',
        server: 'local',
      },
      actions: [
        { title: i18n.global.t('dns.rule.action.route'), value: 'route'},
        { title: i18n.global.t('dns.rule.action.routeOptions'), value: 'route-options'},
        { title: i18n.global.t('dns.rule.action.reject'), value: 'reject'},
        { title: i18n.global.t('dns.rule.action.predefined'), value: 'predefined'},
      ],
      strategies: [
        { title: 'Prefer IPv4', value: 'prefer_ipv4' },
        { title: 'Prefer IPv6', value: 'prefer_ipv6' },
        { title: 'IPv4 Only', value: 'ipv4_only' },
        { title: 'IPv6 Only', value: 'ipv6_only' },
      ],
      predefinedRcode: [
        { title: i18n.global.t('dns.rule.action.rcodes.noError'), value: 'NOERROR' },
        { title: i18n.global.t('dns.rule.action.rcodes.formerr'), value: 'FORMERR' },
        { title: i18n.global.t('dns.rule.action.rcodes.servFail'), value: 'SERVFAIL' },
        { title: i18n.global.t('dns.rule.action.rcodes.nxDomain'), value: 'NXDOMAIN' },
        { title: i18n.global.t('dns.rule.action.rcodes.notImp'), value: 'NOTIMP' },
        { title: i18n.global.t('dns.rule.action.rcodes.refused'), value: 'REFUSED' },
      ],
    }
  },
  computed: {
    logical: {
      get() { return this.ruleData.type == 'logical' },
      set(v:boolean) {
        this.ruleData.type = v? 'logical' : 'simple'
      }
    },
    answer: {
      get() { return this.ruleData.answer?.length ? this.ruleData.answer.join(',') : "" },
      set(v:string) { this.ruleData.answer = v.length > 0 ? v.split(',') : undefined }
    },
    ns: {
      get() { return this.ruleData.ns?.length ? this.ruleData.ns.join(',') : "" },
      set(v:string) { this.ruleData.ns = v.length > 0 ? v.split(',') : undefined }
    },
    extra: {
      get() { return this.ruleData.extra?.length ? this.ruleData.extra.join(',') : "" },
      set(v:string) { this.ruleData.extra = v.length > 0 ? v.split(',') : undefined }
    },
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.updateData()
      }
    },
  },
  methods: {
    updateData() {
      if (this.$props.index != -1) {
        const newData = JSON.parse(this.$props.data)
        if (newData.type) {
          this.ruleData = newData
        } else {
          this.ruleData = {
            type: 'simple',
            mode: 'and',
            rules: <dnsRule[]>[{}],
          }
          Object.keys(newData).forEach(key => {
            if (actionDnsRuleKeys.includes(key)) {
              this.ruleData[key] = newData[key]
            } else {
              (this.ruleData.rules[0] as Record<string, unknown>)[key] = newData[key]
            }
          })
        }
        this.title = 'edit'
      }
      else {
        this.ruleData = <DnsRuleDraft>{
            type: 'simple',
            mode: 'and',
            rules: <dnsRule[]>[{}],
            invert: false,
            action: 'route',
            server: this.$props.serverTags[0]?? 'local',
          }
        this.title = 'add'
      }
    },
    closeModal() {
      this.$emit('close')
    },
    saveChanges() {
      this.loading = true
      let newRule: DnsRuleFields = {
        action: this.ruleData.action,
        invert: this.ruleData.invert? this.ruleData.invert : undefined,
      }

      // Filter action data
      switch (newRule.action){
        case 'route':
          newRule.server = this.ruleData.server
          newRule.strategy = (this.ruleData.strategy?.length ?? 0) > 0 ? this.ruleData.strategy : undefined
          newRule.disable_cache = this.ruleData.disable_cache? true : undefined
          newRule.rewrite_ttl = (this.ruleData.rewrite_ttl ?? 0) > 0 ? this.ruleData.rewrite_ttl : undefined
          newRule.client_subnet = (this.ruleData.client_subnet?.length ?? 0) > 0 ? this.ruleData.client_subnet : undefined
          break
        case 'route-options':
          newRule.disable_cache = this.ruleData.disable_cache? true : undefined
          newRule.rewrite_ttl = (this.ruleData.rewrite_ttl ?? 0) > 0 ? this.ruleData.rewrite_ttl : undefined
          newRule.client_subnet = (this.ruleData.client_subnet?.length ?? 0) > 0 ? this.ruleData.client_subnet : undefined
          break
        case 'reject':
          newRule.method = (this.ruleData.method?.length ?? 0) > 0 ? this.ruleData.method : undefined
          newRule.no_drop = this.ruleData.no_drop? true : undefined
          break
        case 'predefined':
          newRule.rcode = (this.ruleData.rcode?.length ?? 0) > 0 ? this.ruleData.rcode : undefined
          if (this.ruleData.rcode == 'NOERROR') {
            newRule.answer = this.ruleData.answer
            newRule.ns = this.ruleData.ns
            newRule.extra = this.ruleData.extra
          }
          break
      }

      // Add rules
      if (this.ruleData.type == 'simple'){
        newRule = { ...this.ruleData.rules[0], ...newRule }
      } else {
        newRule.type = 'logical'
        newRule.mode = this.ruleData.mode
        newRule.rules = this.ruleData.rules
      }
      this.$emit('save', newRule)
      this.loading = false
    },
    deleteRule(index:number) {
      this.ruleData.rules.splice(index,1)
    }
  }
}

</script>

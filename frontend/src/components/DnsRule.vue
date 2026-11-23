<template>
  <a-card size="small">
    <a-flex
      vertical
      gap="small"
    >
      <a-row
        v-if="optionInbound || optionClient || optionIPver || optionProtocol"
        :gutter="[12, 12]"
      >
        <a-col
          v-if="optionInbound"
          :xs="24"
        >
          <a-form-item
            :label="$t('pages.inbounds')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="rule.inbound"
              mode="tags"
              style="width: 100%;"
            >
              <a-select-option
                v-for="tag in inTags"
                :key="tag"
                :value="tag"
              >
                {{ tag }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionClient"
          :xs="24"
        >
          <a-form-item
            :label="$t('pages.clients')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="rule.auth_user"
              mode="tags"
              style="width: 100%;"
            >
              <a-select-option
                v-for="c in clients"
                :key="c"
                :value="c"
              >
                {{ c }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionIPver"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('rule.ipVer')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="rule.ip_version"
              style="width: 100%;"
            >
              <a-select-option
                v-for="v in [4, 6]"
                :key="v"
                :value="v"
              >
                {{ v }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-if="optionProtocol"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('protocol')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="rule.protocol"
              mode="tags"
              style="width: 100%;"
            >
              <a-select-option
                v-for="p in ['http', 'tls', 'quic', 'stun', 'dns']"
                :key="p"
                :value="p"
              >
                {{ p }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionDomain"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item style="margin-bottom: 0;">
            <a-select
              v-model:value="domainOption"
              style="width: 100%;"
              @change="updateDomainOption($event as string)"
            >
              <a-select-option
                v-for="k in domainKeys"
                :key="k"
                :value="k"
              >
                {{ k }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.domain != undefined"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.domain') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="domain" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.domain_suffix != undefined"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.domainSufix') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="domain_suffix" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.domain_keyword != undefined"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.domainKw') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="domain_keyword" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.domain_regex != undefined"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.domainRgx') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="domain_regex" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionPort"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item style="margin-bottom: 0;">
            <a-select
              v-model:value="portOption"
              style="width: 100%;"
              @change="updatePortOption($event as string)"
            >
              <a-select-option
                v-for="k in portKeys"
                :key="k"
                :value="k"
              >
                {{ k }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.port != undefined"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.port') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="port" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.port_range != undefined"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.portRange') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="port_range" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionSrcIP"
        :gutter="[12, 12]"
        align="middle"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item style="margin-bottom: 0;">
            <a-select
              v-model:value="srcIPOption"
              style="width: 100%;"
              @change="updateSrcIPOption($event as string)"
            >
              <a-select-option
                v-for="k in srcIPKeys"
                :key="k"
                :value="k"
              >
                {{ k }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.source_ip_cidr != undefined"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.srcCidr') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="source_ip_cidr" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.source_ip_is_private != undefined"
          :xs="24"
          :sm="12"
        >
          <a-space>
            <a-switch v-model:checked="rule.source_ip_is_private" />
            <span>{{ $t('rule.srcPrivateIp') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row
        v-if="optionSrcPort"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item style="margin-bottom: 0;">
            <a-select
              v-model:value="srcPortOption"
              style="width: 100%;"
              @change="updateSrcPortOption($event as string)"
            >
              <a-select-option
                v-for="k in srcPortKeys"
                :key="k"
                :value="k"
              >
                {{ k }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.source_port != undefined"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.srcPort') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="source_port" />
          </a-form-item>
        </a-col>
        <a-col
          v-if="rule.source_port_range != undefined"
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.srcPortRange') + ' ' + $t('commaSeparated')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="source_port_range" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="optionRuleSet"
        :gutter="[12, 12]"
      >
        <a-col
          :xs="24"
          :sm="12"
        >
          <a-form-item
            :label="$t('rule.ruleset')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="rule.rule_set"
              mode="tags"
              style="width: 100%;"
            >
              <a-select-option
                v-for="rs in ruleSets"
                :key="rs"
                :value="rs"
              >
                {{ rs }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
      <a-flex justify="end">
        <a-popover
          v-model:open="menu"
          trigger="click"
          placement="left"
        >
          <template #content>
            <a-flex
              vertical
              gap="small"
            >
              <a-space>
                <a-switch v-model:checked="optionInbound" />
                <span>{{ $t('pages.inbounds') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionClient" />
                <span>{{ $t('pages.clients') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionIPver" />
                <span>{{ $t('rule.ipVer') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionProtocol" />
                <span>{{ $t('protocol') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionDomain" />
                <span>{{ $t('rule.domainRules') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionPort" />
                <span>{{ $t('in.port') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionSrcIP" />
                <span>{{ $t('rule.srcIpRules') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionSrcPort" />
                <span>{{ $t('rule.srcPortRules') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="optionRuleSet" />
                <span>{{ $t('rule.ruleset') }}</span>
              </a-space>
            </a-flex>
          </template>
          <a-button type="default">
            {{ $t('rule.options') }}
          </a-button>
        </a-popover>
      </a-flex>
    </a-flex>
  </a-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { dnsRule as DnsRuleType } from '@/types/dns'

// Only the fields this form reads or writes. The option groups are addressed
// by key name at runtime, so the shape also carries a string index.
interface DnsRuleShape {
  [key: string]: unknown
  inbound?: string[]
  auth_user?: string[]
  ip_version?: number
  protocol?: string[]
  domain?: string[]
  domain_suffix?: string[]
  domain_keyword?: string[]
  domain_regex?: string[]
  port?: number[]
  port_range?: string[]
  source_ip_cidr?: string[]
  source_ip_is_private?: boolean
  source_port?: number[]
  source_port_range?: string[]
  rule_set?: string[]
}

// The parent owns the rule object and this form edits it in place, so it is a
// model rather than a plain prop.
const model = defineModel<DnsRuleType>('rule', { required: true })

// The parent holds a wider rule type than this form touches, so narrow once
// here. The computed returns the same object, so edits still reach the parent.
const rule = computed(() => model.value as unknown as DnsRuleShape)

defineProps<{
  clients?: string[]
  inTags?: string[]
  rsTags?: string[]
  deleteable?: boolean
  ruleSets?: string[]
}>()

const menu = ref(false)
const domainKeys = ['domain', 'domain_suffix', 'domain_keyword', 'domain_regex']
const portKeys = ['port', 'port_range']
const srcIPKeys = ['source_ip_cidr', 'source_ip_is_private']
const srcPortKeys = ['source_port', 'source_port_range']
const domainOption = ref('domain')
const portOption = ref('port')
const srcIPOption = ref('source_ip_cidr')
const srcPortOption = ref('source_port')

const optionInbound = computed({
  get() { return rule.value.inbound != undefined },
  set(v: boolean) { rule.value.inbound = v ? [] : undefined }
})
const optionClient = computed({
  get() { return rule.value.auth_user != undefined },
  set(v: boolean) { rule.value.auth_user = v ? [] : undefined }
})
const optionIPver = computed({
  get() { return rule.value.ip_version != undefined },
  set(v: boolean) { rule.value.ip_version = v ? 4 : undefined }
})
const optionProtocol = computed({
  get() { return rule.value.protocol != undefined },
  set(v: boolean) { rule.value.protocol = v ? ['http'] : undefined }
})
const optionDomain = computed({
  get() { return Object.keys(rule.value).some(r => domainKeys.includes(r)) },
  set(v: boolean) {
    if (v) {
      rule.value.domain = []
    } else {
      domainKeys.forEach(k => delete rule.value[k])
    }
    domainOption.value = 'domain'
  }
})
const optionPort = computed({
  get() { return Object.keys(rule.value).some(r => portKeys.includes(r)) },
  set(v: boolean) {
    if (v) {
      rule.value.port = []
    } else {
      portKeys.forEach(k => delete rule.value[k])
    }
    portOption.value = 'port'
  }
})
const optionSrcIP = computed({
  get() { return Object.keys(rule.value).some(r => srcIPKeys.includes(r)) },
  set(v: boolean) {
    if (v) {
      rule.value.source_ip_cidr = []
    } else {
      srcIPKeys.forEach(k => delete rule.value[k])
    }
    srcIPOption.value = 'source_ip_cidr'
  }
})
const optionSrcPort = computed({
  get() { return Object.keys(rule.value).some(r => srcPortKeys.includes(r)) },
  set(v: boolean) {
    if (v) {
      rule.value.source_port = []
    } else {
      srcPortKeys.forEach(k => delete rule.value[k])
    }
    srcPortOption.value = 'source_port'
  }
})
const optionRuleSet = computed({
  get() { return rule.value.rule_set != undefined },
  set(v: boolean) {
    if (v) {
      rule.value.rule_set = []
    } else {
      delete rule.value.rule_set
    }
  }
})
const domain = computed({
  get() { return rule.value.domain?.join(',') },
  set(v: string) { rule.value.domain = v.length>0 ? v.split(',') : [] }
})
const domain_suffix = computed({
  get() { return rule.value.domain_suffix?.join(',') },
  set(v: string) { rule.value.domain_suffix = v.length>0 ? v.split(',') : [] }
})
const domain_keyword = computed({
  get() { return rule.value.domain_keyword?.join(',') },
  set(v: string) { rule.value.domain_keyword = v.length>0 ? v.split(',') : [] }
})
const domain_regex = computed({
  get() { return rule.value.domain_regex?.join(',') },
  set(v: string) { rule.value.domain_regex = v.length>0 ? v.split(',') : [] }
})
const port = computed({
  get() { return rule.value.port?.join(',') },
  set(v: string) {
    if(!v.endsWith(',')) {
      rule.value.port = v.length > 0 ? v.split(',').map(str => parseInt(str, 10)) : []
    }
  }
})
const port_range = computed({
  get() { return rule.value.port_range?.join(',') },
  set(v: string) { rule.value.port_range = v.length>0 ? v.split(',') : [] }
})
const source_ip_cidr = computed({
  get() { return rule.value.source_ip_cidr?.join(',') },
  set(v: string) { rule.value.source_ip_cidr = v.length>0 ? v.split(',') : [] }
})
const source_port = computed({
  get() { return rule.value.source_port?.join(',') },
  set(v: string) {
    if(!v.endsWith(',')) {
      rule.value.source_port = v.length > 0 ? v.split(',').map(str => parseInt(str, 10)) : []
    }
  }
})
const source_port_range = computed({
  get() { return rule.value.source_port_range?.join(',') },
  set(v: string) { rule.value.source_port_range = v.length>0 ? v.split(',') : [] }
})

onMounted(() => {
  const ruleKeys = Object.keys(rule.value)
  if (optionDomain.value) {
    const enabledOption = domainKeys.filter(k => ruleKeys.includes(k))
    domainOption.value = enabledOption.length>0 ? enabledOption[0] : 'domain'
  }
  if (optionPort.value) {
    const enabledOption = portKeys.filter(k => ruleKeys.includes(k))
    portOption.value = enabledOption.length>0 ? enabledOption[0] : 'port'
  }
  if (optionSrcIP.value) {
    const enabledOption = srcIPKeys.filter(k => ruleKeys.includes(k))
    srcIPOption.value = enabledOption.length>0 ? enabledOption[0] : 'source_ip_cidr'
  }
  if (optionSrcPort.value) {
    const enabledOption = srcPortKeys.filter(k => ruleKeys.includes(k))
    srcPortOption.value = enabledOption.length>0 ? enabledOption[0] : 'source_port'
  }
})

function updateDomainOption(option: string) {
  domainKeys.forEach(k => delete rule.value[k])
  rule.value[option] = []
}
function updatePortOption(option: string) {
  portKeys.forEach(k => delete rule.value[k])
  rule.value[option] = []
}
function updateSrcIPOption(option: string) {
  srcIPKeys.forEach(k => delete rule.value[k])
  rule.value[option] = option == 'source_ip_is_private' ? false : []
}
function updateSrcPortOption(option: string) {
  srcPortKeys.forEach(k => delete rule.value[k])
  rule.value[option] = []
}
</script>

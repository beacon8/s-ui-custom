<template>
  <a-modal
    :open="visible"
    :width="800"
    :body-style="{ maxHeight: 'calc(100vh - 220px)', overflowY: 'auto', overflowX: 'hidden', paddingRight: '12px' }"
    @cancel="closeModal"
  >
    <template #title>
      <a-flex
        justify="space-between"
        align="center"
        style="padding-right: 24px;"
      >
        <span>{{ $t('actions.' + title) + " " + $t('objects.rule') }}</span>
        <DocLink section="rule" />
      </a-flex>
    </template>

    <a-form layout="vertical">
      <a-flex
        class="kind-toggle"
        align="center"
      >
        <a-radio-group
          v-model:value="kind"
          option-type="button"
          button-style="solid"
          size="small"
        >
          <a-radio-button value="simple">
            {{ $t('rule.normalRule') }}
          </a-radio-button>
          <a-radio-button value="logical">
            {{ $t('rule.logicalRule') }}
          </a-radio-button>
        </a-radio-group>
      </a-flex>

      <template v-if="!logical">
        <RuleOptions
          :rule="simpleRule"
          :invert="Boolean(ruleData.invert)"
          :clients="clients"
          :in-tags="inTags"
          :out-tags="outTags"
          :rs-tags="rsTags"
          @update:invert="ruleData.invert = $event"
        />
      </template>

      <template v-else>
        <div class="logical-toolbar">
          <div class="logical-head">
            <span class="section-title">{{ $t('rule.matchConditions') }}</span>
            <a-space
              size="small"
              wrap
            >
              <a-radio-group
                v-model:value="ruleData.mode"
                size="small"
                option-type="button"
              >
                <a-radio-button value="and">
                  {{ $t('rule.matchAll') }}
                </a-radio-button>
                <a-radio-button value="or">
                  {{ $t('rule.matchAny') }}
                </a-radio-button>
              </a-radio-group>
              <a-space size="small">
                <a-switch
                  v-model:checked="ruleData.invert"
                  size="small"
                />
                <span class="header-label">{{ $t('rule.invert') }}</span>
              </a-space>
            </a-space>
          </div>
          <a-button
            type="dashed"
            size="small"
            @click="addSubRule"
          >
            <template #icon>
              <AppIcon name="plus" />
            </template>
            {{ $t('rule.addSubRule') }}
          </a-button>
        </div>

        <div
          v-for="(r, ruleIndex) in ruleData.rules"
          :key="ruleIndex"
          class="subrule"
        >
          <div class="subrule-head">
            <span>{{ $t('rule.subRule') }} {{ Number(ruleIndex) + 1 }}</span>
            <a-button
              v-if="ruleData.rules.length > 1"
              type="text"
              danger
              size="small"
              shape="circle"
              :aria-label="$t('actions.del')"
              @click="removeSubRule(ruleIndex)"
            >
              <template #icon>
                <AppIcon
                  name="delete"
                  color="error"
                />
              </template>
            </a-button>
          </div>
          <RuleOptions
            :rule="r"
            :invert="Boolean(r.invert)"
            :clients="clients"
            :in-tags="inTags"
            :out-tags="outTags"
            :rs-tags="rsTags"
            @update:invert="r.invert = $event"
          />
        </div>
      </template>

      <div class="action-section">
        <div class="section-title">
          {{ $t('rule.actionSection') }}
        </div>
        <a-row :gutter="[12, 12]">
          <a-col
            :xs="24"
            :sm="ruleData.action === 'route' ? 12 : 24"
          >
            <a-form-item
              :label="$t('rule.action')"
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
            v-if="ruleData.action == 'route'"
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('objects.outbound')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="ruleData.outbound"
                style="width: 100%"
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

        <a-row
          v-if="ruleData.action == 'route-options'"
          :gutter="[12, 12]"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('types.direct.overrideAddr')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="ruleData.override_address" />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('types.direct.overridePort')"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="ruleData.override_port"
                :min="0"
                :max="65534"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-space>
              <a-switch v-model:checked="ruleData.udp_disable_domain_unmapping" />
              <span>{{ $t('rule.udpDisableDomainUnmapping') }}</span>
            </a-space>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-space>
              <a-switch v-model:checked="ruleData.udp_connect" />
              <span>{{ $t('rule.udpConnect') }}</span>
            </a-space>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('rule.udpTimeout')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="ruleData.udp_timeout" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row
          v-if="ruleData.action == 'reject'"
          :gutter="[12, 12]"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
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
                @clear="delete ruleData.method"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-space>
              <a-switch v-model:checked="ruleData.no_drop" />
              <span>{{ $t('rule.noDrop') }}</span>
            </a-space>
          </a-col>
        </a-row>

        <a-row
          v-if="ruleData.action == 'sniff'"
          :gutter="[12, 12]"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('rule.sniffer')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="ruleData.sniffer"
                mode="multiple"
                :options="sniffers"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('rule.timeout')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="ruleData.timeout" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row
          v-if="ruleData.action == 'resolve'"
          :gutter="[12, 12]"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
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
                @clear="delete ruleData.strategy"
              />
            </a-form-item>
          </a-col>
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('basic.dns.server')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="ruleData.server" />
            </a-form-item>
          </a-col>
        </a-row>
      </div>
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
import { rule, actionKeys } from '@/types/rules'
import RuleOptions from '@/components/Rule.vue'
import DocLink from '@/components/DocLink.vue'

interface RuleFields {
  type?: string
  mode?: string
  rules?: rule[]
  invert?: boolean
  action?: string
  outbound?: string
  override_address?: string
  override_port?: number
  network_strategy?: string
  fallback_delay?: string
  udp_disable_domain_unmapping?: boolean
  udp_connect?: boolean
  udp_timeout?: string
  sniffer?: string[]
  timeout?: string
  method?: string
  no_drop?: boolean
  server?: string
  strategy?: string
  [key: string]: unknown
}

type RuleDraft = RuleFields & { type: string, mode: string, rules: rule[] }

export default {
  components: { DocLink, RuleOptions },
  props: {
    visible: { type: Boolean, required: true },
    data: { type: String, default: '' },
    index: { type: Number, required: true },
    clients: { type: Array as PropType<string[]>, default: () => [] },
    inTags: { type: Array as PropType<string[]>, default: () => [] },
    outTags: { type: Array as PropType<string[]>, default: () => [] },
    rsTags: { type: Array as PropType<string[]>, default: () => [] },
  },
  emits: ['close', 'save'],
  data() {
    return {
      title: 'add',
      loading: false,
      ruleData: <RuleDraft>{
        type: 'simple',
        mode: 'and',
        rules: <rule[]>[{}],
        invert: false,
        action: 'route',
        outbound: 'direct',
      },
      actions: [
        { title: 'Route', value: 'route' },
        { title: 'Route Options', value: 'route-options' },
        { title: 'Bypass', value: 'bypass' },
        { title: 'Reject', value: 'reject' },
        { title: 'Hijack DNS', value: 'hijack-dns' },
        { title: 'Sniff', value: 'sniff' },
        { title: 'Resolve', value: 'resolve' },
      ],
      sniffers: [
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
      ],
      strategies: [
        { title: 'Prefer IPv4', value: 'prefer_ipv4' },
        { title: 'Prefer IPv6', value: 'prefer_ipv6' },
        { title: 'IPv4 Only', value: 'ipv4_only' },
        { title: 'IPv6 Only', value: 'ipv6_only' },
      ],
    }
  },
  computed: {
    logical(): boolean {
      return this.ruleData.type === 'logical'
    },
    kind: {
      get(): string {
        return this.logical ? 'logical' : 'simple'
      },
      set(value: string) {
        this.ruleData.type = value === 'logical' ? 'logical' : 'simple'
        if (!this.ruleData.rules || this.ruleData.rules.length === 0) {
          this.ruleData.rules = <rule[]>[{}]
        }
      },
    },
    simpleRule(): rule {
      return this.ruleData.rules[0]
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
          this.ruleData.type = newData.type === 'logical' ? 'logical' : 'simple'
          if (!Array.isArray(this.ruleData.rules) || this.ruleData.rules.length === 0) {
            this.ruleData.rules = <rule[]>[{}]
          }
          if (this.ruleData.type === 'logical') {
            this.ruleData.mode = this.ruleData.mode || 'and'
          }
        } else {
          this.ruleData = {
            type: 'simple',
            mode: 'and',
            rules: <rule[]>[{}],
          }
          Object.keys(newData).forEach(key => {
            if (actionKeys.includes(key)) {
              this.ruleData[key] = newData[key]
            } else {
              (this.ruleData.rules[0] as Record<string, unknown>)[key] = newData[key]
            }
          })
        }
        this.title = 'edit'
      } else {
        this.ruleData = <RuleDraft>{
          type: 'simple',
          mode: 'and',
          rules: <rule[]>[{}],
          invert: false,
          action: 'route',
          outbound: this.$props.outTags[0] ?? 'direct',
        }
        this.title = 'add'
      }
    },
    addSubRule() {
      this.ruleData.rules.push(<rule>{})
    },
    removeSubRule(index: number) {
      if (this.ruleData.rules.length > 1) {
        this.ruleData.rules.splice(index, 1)
      }
    },
    closeModal() {
      this.updateData()
      this.$emit('close')
    },
    saveChanges() {
      this.loading = true
      let newRule: RuleFields = {
        action: this.ruleData.action,
        invert: this.ruleData.invert ? true : undefined,
      }

      switch (newRule.action) {
        case 'route':
          newRule.outbound = this.ruleData.outbound
          break
        case 'route-options':
          newRule.override_address = (this.ruleData.override_address?.length ?? 0) > 0 ? this.ruleData.override_address : undefined
          newRule.override_port = (this.ruleData.override_port ?? 0) > 0 ? this.ruleData.override_port : undefined
          newRule.network_strategy = (this.ruleData.network_strategy?.length ?? 0) > 0 ? this.ruleData.network_strategy : undefined
          newRule.fallback_delay = (this.ruleData.fallback_delay?.length ?? 0) > 0 ? this.ruleData.fallback_delay : undefined
          newRule.udp_disable_domain_unmapping = this.ruleData.udp_disable_domain_unmapping ? true : undefined
          newRule.udp_connect = this.ruleData.udp_connect ? true : undefined
          newRule.udp_timeout = (this.ruleData.udp_timeout?.length ?? 0) > 0 ? this.ruleData.udp_timeout : undefined
          break
        case 'reject':
          newRule.method = (this.ruleData.method?.length ?? 0) > 0 ? this.ruleData.method : undefined
          newRule.no_drop = this.ruleData.no_drop ? true : undefined
          break
        case 'sniff':
          newRule.sniffer = (this.ruleData.sniffer?.length ?? 0) > 0 ? this.ruleData.sniffer : undefined
          newRule.timeout = (this.ruleData.timeout?.length ?? 0) > 0 ? this.ruleData.timeout : undefined
          break
        case 'resolve':
          newRule.strategy = (this.ruleData.strategy?.length ?? 0) > 0 ? this.ruleData.strategy : undefined
          newRule.server = (this.ruleData.server?.length ?? 0) > 0 ? this.ruleData.server : undefined
          break
      }

      if (!this.logical) {
        newRule = { ...this.ruleData.rules[0], ...newRule }
      } else {
        newRule.type = 'logical'
        newRule.mode = this.ruleData.mode || 'and'
        newRule.rules = this.ruleData.rules.map((r) => {
          const copy = { ...r } as Record<string, unknown>
          if (!copy.invert) delete copy.invert
          return copy as unknown as rule
        })
      }
      this.$emit('save', newRule)
      this.loading = false
    },
  },
}
</script>

<style scoped>
.kind-toggle {
  margin-bottom: 12px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.header-label,
.logical-head,
.subrule-head span {
  overflow-wrap: anywhere;
  word-break: break-word;
}

.header-label {
  font-size: 12px;
}

.logical-toolbar {
  border: 1px solid rgba(128, 128, 128, 0.22);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
}

.logical-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.subrule {
  border: 1px solid rgba(128, 128, 128, 0.22);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
  min-width: 0;
}

.subrule-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
}

.action-section {
  border-top: 1px solid rgba(128, 128, 128, 0.22);
  padding-top: 12px;
  margin-top: 12px;
}

.action-section > .section-title {
  margin-bottom: 10px;
}

@media (max-width: 575px) {
  .logical-head {
    align-items: flex-start;
  }
}
</style>

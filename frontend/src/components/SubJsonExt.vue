<template>
  <Editor
    v-model="enableEditor"
    :data="settings.subJsonExt"
    :visible="enableEditor"
    :title="$t('editor') + ' - ' + $t('setting.jsonSub')"
    @close="enableEditor = false"
    @save="saveEditor"
  />
  <div class="sub-pane">
    <a-form
      layout="vertical"
      class="sub-form"
    >
      <a-row :gutter="[12, 12]">
        <a-col
          :xs="24"
          :sm="12"
          :md="12"
          :lg="16"
        >
          <a-form-item
            :label="$t('setting.defaultRules')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="selectedDefaultRules"
              :options="defaultRuleItems"
              :field-names="{ label: 'title', value: 'value' }"
              mode="multiple"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="12"
          :lg="8"
        >
          <a-form-item
            :label="$t('setting.routeFinal')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="routeFinal"
              :options="[{ title: 'Proxy', value: 'proxy' }, { title: 'Direct', value: 'direct' }]"
              :field-names="{ label: 'title', value: 'value' }"
              allow-clear
              style="width: 100%"
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
          :md="12"
          :lg="8"
        >
          <a-form-item
            :label="$t('setting.toProxy')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="ruleToProxy"
              :options="geoOptions"
              :field-names="{ label: 'title', value: 'value' }"
              mode="multiple"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="12"
          :lg="8"
        >
          <a-form-item
            :label="$t('setting.toDirect')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="ruleToDirect"
              :options="geoOptions"
              :field-names="{ label: 'title', value: 'value' }"
              mode="multiple"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="12"
          :lg="8"
        >
          <a-form-item
            :label="$t('setting.toBlock')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="ruleToBlock"
              :options="geoOptions"
              :field-names="{ label: 'title', value: 'value' }"
              mode="multiple"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-flex
        class="sub-actions"
        justify="end"
        gap="small"
        style="margin-top: 12px;"
      >
        <a-button
          type="default"
          @click="openEditor"
        >
          {{ $t('editor') }}
        </a-button>
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
                <a-switch v-model:checked="enableLog" />
                <span>{{ $t('basic.log.title') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="enableDns" />
                <span>DNS</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="enableInb" />
                <span>{{ $t('objects.inbound') }}</span>
              </a-space>
              <a-space>
                <a-switch v-model:checked="enableExp" />
                <span>Experimental</span>
              </a-space>
            </a-flex>
          </template>
          <a-button type="default">
            {{ $t('setting.jsonSubOptions') }}
          </a-button>
        </a-popover>
      </a-flex>
      <a-row
        v-if="enableLog"
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('basic.log.level')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="subJsonExt.log!.level"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in levels"
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
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('setting.timestamp')"
            style="margin-bottom: 0;"
          >
            <div class="switch-control">
              <a-switch v-model:checked="subJsonExt.log!.timestamp" />
            </div>
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="enableDns"
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('dns.final')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="subJsonExt.dns!.final"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in dnsTags"
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
          :md="6"
          :lg="4"
        >
          <SimpleDNS
            :data="proxyDns"
            :label="$t('setting.globalDns')"
          />
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <SimpleDNS
            :data="directDns"
            :label="$t('setting.directDns')"
          />
        </a-col>
      </a-row>
      <a-row
        v-if="enableDns"
        :gutter="[12, 12]"
        style="margin-top: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="6"
          :lg="4"
        >
          <a-form-item
            :label="$t('basic.routing.defaultDns')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="subJsonExt.default_domain_resolver"
              allow-clear
              style="width: 100%"
              @change="(v: unknown) => { if (!v) delete subJsonExt.default_domain_resolver }"
            >
              <a-select-option
                v-for="item in dnsTags"
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
          :md="6"
        >
          <a-form-item
            :label="$t('setting.toDirectDns')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="dnsToDirect"
              :options="geositeOptions"
              :field-names="{ label: 'title', value: 'value' }"
              mode="multiple"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <template v-if="enableInb">
        <a-row
          :gutter="[12, 12]"
          style="margin-top: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
            :md="6"
          >
            <a-form-item
              :label="$t('in.addr')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="inbounds[0].address"
                mode="tags"
                style="width: 100%"
              >
                <a-select-option
                  v-for="item in defaultInb[0].address"
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
            :md="6"
            :lg="4"
          >
            <a-form-item
              label="MTU"
              style="margin-bottom: 0;"
            >
              <a-input-number
                v-model:value="inbounds[0].mtu"
                style="width: 100%"
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
            :md="6"
          >
            <a-form-item
              :label="$t('setting.excludePkg')"
              style="margin-bottom: 0;"
            >
              <a-select
                v-model:value="inbounds[0].exclude_package"
                mode="tags"
                style="width: 100%"
              >
                <a-select-option
                  v-for="item in ['ir.mci.ecareapp', 'com.myirancell']"
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
            :md="6"
            :lg="4"
          >
            <a-form-item
              label="Platform HTTP proxy"
              style="margin-bottom: 0;"
            >
              <div class="switch-control">
                <a-switch v-model:checked="platformProxy" />
              </div>
            </a-form-item>
          </a-col>
        </a-row>
      </template>
    </a-form>
  </div>
</template>

<script lang="ts">
import { PropType } from 'vue'
import Editor from './Editor.vue'
import SimpleDNS from './SimpleDNS.vue'
import { push } from 'notivue'
import { i18n } from '@/locales'
import { geoCatalog, geoList, geositeList } from '@/plugins/rulesetCatalog'

// The one setting this form reads and writes, as a string of sing-box JSON.
interface Settings {
  subJsonExt: string
}

// The config below is user-supplied sing-box JSON and carries far more than the
// panel touches. Each shape names only what the form reads or writes; anything
// else the user wrote is parsed, kept and written back untouched.
interface SbRule {
  action?: string
  outbound?: string
  protocol?: string
  clash_mode?: string
  rule_set?: string[]
}

interface SbDnsServer {
  type?: string
  tag?: string
  server?: string
  server_port?: number
  detour?: string
  domain_resolver?: string
}

interface SbDnsRule {
  action?: string
  server?: string
  clash_mode?: string
  source_ip_cidr?: string[]
  rule_set?: string[]
}

interface SbDns {
  servers: SbDnsServer[]
  rules: SbDnsRule[]
  final?: string
  strategy?: string
}

// A rule-set entry is matched against the catalog by tag; the rest of it is
// whatever the catalog or the user put there.
interface SbRuleSet {
  tag: string
}

interface SbInbound {
  type: string
  address?: string[]
  mtu?: number
  auto_route?: boolean
  strict_route?: boolean
  stack?: string
  exclude_package?: string[]
  listen?: string
  listen_port?: number
  users?: string[]
  platform?: {
    http_proxy: {
      enabled: boolean
      server: string
      server_port: number
    }
  }
}

interface SbLog {
  level: string
  timestamp: boolean
}

interface SbExperimental {
  clash_api: {
    external_controller: string
    external_ui: string
    secret: string
    external_ui_download_url: string
    external_ui_download_detour: string
    default_mode: string
  }
  cache_file: {
    enabled: boolean
    store_fakeip: boolean
  }
}

// Every section is optional: the switches in the options menu add and remove
// them, and an empty config is written back as an empty string.
interface SubJson {
  log?: SbLog
  dns?: SbDns
  inbounds?: SbInbound[]
  experimental?: SbExperimental
  rules?: SbRule[]
  rule_set?: SbRuleSet[]
  final?: string
  default_domain_resolver?: string
}

export default {
  components: { Editor, SimpleDNS },
  props: {
    settings: { type: Object as PropType<Settings>, required: true }
  },
  data() {
    return {
      menu: false,
      enableEditor: false,
      subJsonExt: <SubJson>{},
      levels: ["trace", "debug", "info", "warn", "error", "fatal", "panic"],
      defaultRules: [
        { action: "sniff" },
        { clash_mode: "Direct", action: "route", outbound: "direct" },
        { clash_mode: "Global", action: "route", outbound: "proxy" },
      ],
      defaultRuleItems: [
        { title: "Sniff", value: 0 },
        { title: "Clash Direct", value: 1 },
        { title: "Clash Global", value: 2 },
      ],
      defaultLog: {
        "level": "info",
        "timestamp": true
      },
      defaultInb: [
        {
          "type": "tun",
          "address": [
            "172.19.0.1/30",
            "fdfe:dcba:9876::1/126"
          ],
          "mtu": 9000,
          "auto_route": true,
          "strict_route": false,
          "stack": "mixed",
          "exclude_package": [],
          "platform": {
            "http_proxy": {
              "enabled": true,
              "server": "127.0.0.1",
              "server_port": 2080
            }
          }
        },
        {
          "type": "mixed",
          "listen": "127.0.0.1",
          "listen_port": 2080,
          "users": []
        }
      ],
      defaultExp: {
        "clash_api": {
          "external_controller": "127.0.0.1:9090",
          "external_ui": "ui",
          "secret": "",
          "external_ui_download_url": "https://mirror.ghproxy.com/https://github.com/MetaCubeX/Yacd-meta/archive/gh-pages.zip",
          "external_ui_download_detour": "direct",
          "default_mode": "rule"
        },
        "cache_file": {
          "enabled": true,
          "store_fakeip": false
        }
      },
      defaultDns: {
        "servers": [
          {
            "type": "tcp",
            "tag": "proxy-dns",
            "server": "8.8.8.8",
            "server_port": 53,
            "detour": "proxy",
            "domain_resolver": "local-dns",
          },
          { 
            "tag": "direct-dns",
            "type": "local",
          },
          {
            "tag": "local-dns",
            "type": "local",
          }
        ],
        "rules": [
          {
            "clash_mode": "Global",
            "source_ip_cidr": [
              "172.19.0.0/30",
              "fdfe:dcba:9876::1/126"
            ],
            "action": "route",
            "server": "proxy-dns"
          },
          {
            "clash_mode": "Direct",
            "action": "route",
            "server": "direct-dns"
          },
          {
            "source_ip_cidr": [
              "172.19.0.0/30",
              "fdfe:dcba:9876::1/126"
            ],
            "action": "route",
            "server": "proxy-dns"
          },
        ],
        "final": "local-dns",
        "strategy": "prefer_ipv4"
      },
      geositeList: geositeList,
      geoList: geoList,
      geo: geoCatalog,
    }
  },
  computed: {
    enableLog: {
      get() :boolean { return this.subJsonExt?.log != undefined },
      set(v:boolean) {
        if (v) this.subJsonExt.log = this.defaultLog
        else delete this.subJsonExt.log
      }
    },
    enableDns: {
      get() :boolean { return this.subJsonExt?.dns != undefined },
      set(v:boolean) {
        if (v) {
          this.subJsonExt.dns = this.defaultDns
          if (this.rules == undefined) this.subJsonExt.rules = [{ action: 'sniff' }]
          this.subJsonExt.rules!.unshift({ protocol: "dns", action: "hijack-dns" })
        } else {
          delete this.subJsonExt.dns
          const rules = this.subJsonExt?.rules?.filter((r) => r.protocol != "dns") ?? []
          if (rules.length >= 0) this.subJsonExt.rules = rules
          if (this.rules.length == 0) delete this.subJsonExt.rules
        }
      }
    },
    enableInb: {
      get() :boolean { return this.subJsonExt?.inbounds != undefined },
      set(v:boolean) {
        if (v) this.subJsonExt.inbounds = this.defaultInb.slice()
        else delete this.subJsonExt.inbounds
      }
    },
    enableExp: {
      get() :boolean { return this.subJsonExt?.experimental != undefined },
      set(v:boolean) {
        if (v) this.subJsonExt.experimental = this.defaultExp
        else delete this.subJsonExt.experimental
      }
    },
    // Absent until the DNS switch is turned on; every reader guards with ?.
    // before it reaches in.
    dns():SbDns { return <SbDns>this.subJsonExt?.dns?? undefined },
    proxyDns: {
      get() :SbDnsServer { return this.dns?.servers?.findLast((d) => d.tag == "proxy-dns")?? {} },
      set(v:SbDnsServer) { 
        let sIndex = this.dns.servers.findIndex((d) => d.tag == "proxy-dns")
        if (sIndex === -1 || sIndex == undefined) {
          this.dns.servers.push({ ...this.defaultDns.servers[0], ...v })
        } else {
          this.dns.servers[sIndex] = { ...this.defaultDns.servers[0], ...v }
        }
      }
    },
    directDns: {
      get() :SbDnsServer { return this.dns?.servers?.findLast((d) => d.tag == "direct-dns")?? {} },
      set(v:SbDnsServer) {
        const sIndex = this.dns.servers.findIndex((d) => d.tag == "direct-dns")
        if (sIndex === -1 || sIndex == undefined) {
          this.dns.servers.push({ ...this.defaultDns.servers[1], ...v })
        } else {
          this.dns.servers[sIndex] = { ...this.defaultDns.servers[1], ...v }
        }
      },
    },
    dnsTags() { return this.dns?.servers?.map((d) => d.tag) ?? [] },
    final: {
      get() :string { return this.dns.final?? "" },
      set(v:string) { this.dns.final = v.length>0 ? v : undefined }
    },
    dnsToDirect: {
      get() :string[] {
        const ruleIndex = this.dns?.rules?.findIndex((r) => r.server == "direct-dns" && Object.hasOwn(r,'rule_set'))
        return ruleIndex >= 0 ? <string[]>this.dns.rules[ruleIndex].rule_set : []
      },
      set(v:string[]) {
        const ruleIndex = this.dns?.rules?.findIndex((r) => r.server == "direct-dns" && Object.hasOwn(r,'rule_set'))
        if (v.length>0) {
          if (ruleIndex >= 0){
            this.dns.rules[ruleIndex].rule_set = v
          } else {
            this.dns.rules.push({ rule_set: v, action: "route", server: "direct-dns" })
          }
        } else {
          if (ruleIndex != -1) this.dns.rules.splice(ruleIndex,1)
        }
        this.updateRuleSets()
      }
    },
    // Absent until the inbound switch is turned on; every reader guards with ?.
    // before it reaches in.
    inbounds():SbInbound[] { return <SbInbound[]>this.subJsonExt?.inbounds?? undefined },
    platformProxy: {
      get() :boolean { return this.inbounds[0]?.platform != undefined },
      set(v:boolean) { this.subJsonExt.inbounds![0].platform = v ? this.defaultInb[0].platform : undefined }
    },
    // Absent until a rule is added; every reader guards with ?. before it
    // reaches in.
    rules():SbRule[] { return <SbRule[]>this.subJsonExt?.rules?? undefined },
    selectedDefaultRules: {
      get(): number[] {
        return this.defaultRules
          .map((d, i:number) => this.hasDefaultRule(d) ? i : -1)
          .filter((i:number) => i >= 0)
      },
      set(v:number[]) {
        this.defaultRules.forEach((d, i:number) => this.toggleDefaultRule(d, v.includes(i)))
      }
    },
    customRuleSetTags() :string[] {
      return (this.subJsonExt?.rule_set ?? [])
        .filter((rs) => !this.geo.some((g) => g.tag == rs.tag))
        .map((rs) => rs.tag)
    },
    geoOptions() :{ title: string, value: string }[] {
      return [...this.geoList, ...this.customRuleSetTags.map((t:string) => ({ title: t, value: t }))]
    },
    geositeOptions() :{ title: string, value: string }[] {
      return [...this.geositeList, ...this.customRuleSetTags.map((t:string) => ({ title: t, value: t }))]
    },
    routeFinal: {
      get() :string { return this.subJsonExt?.final ?? "" },
      set(v:string|null) {
        if (v && v.length > 0) this.subJsonExt.final = v
        else delete this.subJsonExt.final
      }
    },
    ruleToProxy: {
      get() :string[] {
        const ruleIndex = this.rules?.findIndex((r) => r.outbound == "proxy" && Object.hasOwn(r,'rule_set'))
        return ruleIndex >= 0 ? <string[]>this.rules[ruleIndex].rule_set : []
      },
      set(v:string[]) {
        const ruleIndex = this.rules?.findIndex((r) => r.outbound == "proxy" && Object.hasOwn(r,'rule_set'))
        if (v.length>0) {
          if (ruleIndex >= 0){
            this.rules[ruleIndex].rule_set = v
          } else {
            if (this.rules == undefined) this.subJsonExt.rules = []
            this.rules.push({ rule_set: v, action: "route", outbound: "proxy" })
          }
        } else {
          if (ruleIndex != -1) this.rules.splice(ruleIndex,1)
        }
        this.updateRuleSets()
      }
    },
    ruleToDirect: {
      get() :string[] {
        const ruleIndex = this.rules?.findIndex((r) => r.outbound == "direct" && Object.hasOwn(r,'rule_set'))
        return ruleIndex >= 0 ? <string[]>this.rules[ruleIndex].rule_set : []
      },
      set(v:string[]) {
        const ruleIndex = this.rules?.findIndex((r) => r.outbound == "direct" && Object.hasOwn(r,'rule_set'))
        if (v.length>0) {
          if (ruleIndex >= 0){
            this.rules[ruleIndex].rule_set = v
          } else {
            if (this.rules == undefined) this.subJsonExt.rules = []
            this.rules.push({ rule_set: v, action: "route", outbound: "direct" })
          }
        } else {
          if (ruleIndex != -1) this.rules.splice(ruleIndex,1)
        }
        this.updateRuleSets()
      }
    },
    ruleToBlock: {
      get() :string[] {
        const ruleIndex = this.rules?.findIndex((r) => r.action == "reject" && Object.hasOwn(r,'rule_set'))
        return ruleIndex >= 0 ? <string[]>this.rules[ruleIndex].rule_set : []
      },
      set(v:string[]) {
        const ruleIndex = this.rules?.findIndex((r) => r.action == "reject" && Object.hasOwn(r,'rule_set'))
        if (v.length>0) {
          if (ruleIndex >= 0){
            this.rules[ruleIndex].rule_set = v
          } else {
            if (this.rules == undefined) this.subJsonExt.rules = []
            this.rules.push({ rule_set: v, action: "reject" })
          }
        } else {
          if (ruleIndex != -1) this.rules.splice(ruleIndex,1)
        }
        this.updateRuleSets()
      }
    }
  },
  watch:{
    subJsonExt:{
      handler(v:SubJson) {
        this.$props.settings.subJsonExt = Object.keys(v).length>0 ? JSON.stringify(v, null, 2) : ""
      },
      deep: true
    },
  },
  mounted(){
    this.loadData()
  },
  methods: {
    loadData() {
      if (this.$props.settings?.subJsonExt?.length>0){
        this.subJsonExt = JSON.parse(this.$props.settings.subJsonExt)
      } else {
        this.subJsonExt = <SubJson>{}
      }
    },
    hasDefaultRule(rule:SbRule): boolean {
      return Array.isArray(this.rules) && this.rules.some((r) => JSON.stringify(r) === JSON.stringify(rule))
    },
    toggleDefaultRule(rule:SbRule, v:boolean) {
      if (v) {
        if (!Array.isArray(this.subJsonExt.rules)) this.subJsonExt.rules = []
        if (!this.subJsonExt.rules.some((r) => JSON.stringify(r) === JSON.stringify(rule))) {
          this.subJsonExt.rules.push(JSON.parse(JSON.stringify(rule)))
          this.normalizeRulesOrder()
        }
      } else if (Array.isArray(this.subJsonExt.rules)) {
        this.subJsonExt.rules = this.subJsonExt.rules.filter((r) => JSON.stringify(r) !== JSON.stringify(rule))
        if (this.subJsonExt.rules.length === 0) delete this.subJsonExt.rules
      }
    },
    ruleWeight(r:SbRule): number {
      if (r.protocol == "dns") return 0
      if (r.action == "sniff") return 1
      if (r.clash_mode == "Direct") return 2
      if (r.clash_mode == "Global") return 3
      return 4
    },
    normalizeRulesOrder() {
      if (Array.isArray(this.subJsonExt?.rules)) {
        this.subJsonExt.rules.sort((a,b) => this.ruleWeight(a) - this.ruleWeight(b))
      }
    },
    updateRuleSets(){
      let tags = <string[]>[]
      if (this.dns?.rules?.length>0) this.dns.rules.forEach((r) => { if (r.rule_set) tags.push(...r.rule_set) })
      if (this.rules?.length>0) this.rules.forEach((r) => { if (r.rule_set) tags.push(...r.rule_set) })
      const existing = <SbRuleSet[]>(this.subJsonExt?.rule_set ?? [])
      const byTag = new Map(existing.map((rs) => [rs.tag, rs]))
      const custom = existing.filter((rs) => !this.geo.some((g) => g.tag == rs.tag))
      if (tags.length>0 || custom.length>0){
        this.subJsonExt.rule_set = [
          ...this.geo.filter((g) => tags.includes(g.tag)).map((g) => byTag.get(g.tag) ?? g),
          ...custom,
        ]
      } else {
        delete this.subJsonExt.rule_set
      }
      if (this.rules?.length == 0) delete this.subJsonExt.rules
    },
    openEditor() {
      this.enableEditor = true
    },
    saveEditor(data:string) {
      try {
        this.subJsonExt = JSON.parse(data)
      } catch {
        push.error({
          message: i18n.global.t('failed') + ": " + i18n.global.t('error.invalidData'),
          duration: 5000,
        })
        return
      }
      this.enableEditor = false
    }
  }
}
</script>
<style scoped>
.sub-pane {
  min-width: 0;
}

.sub-form {
  min-width: 0;
  overflow-x: clip;
}

.sub-form .ant-form-item {
  margin-bottom: 0;
}

.switch-control {
  display: flex;
  align-items: center;
  min-height: 32px;
}

.sub-actions {
  width: 100%;
  flex-wrap: wrap;
  margin-top: 12px;
}

@media (max-width: 575px) {
  .sub-actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>

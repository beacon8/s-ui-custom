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
        <span>{{ $t('actions.' + title) + " " + $t('objects.dnsserver') }}</span>
        <DocLink section="dnsServer" />
      </a-flex>
    </template>
    <a-form layout="vertical">
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
            :label="$t('type')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="dnsServer.type"
              :options="dnsTypes"
              :field-names="{ label: 'title', value: 'value' }"
              style="width: 100%"
              @change="changeType($event as string)"
            />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('objects.tag')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dnsServer.tag" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="HasServer.includes(dnsServer.type)"
        :gutter="[12, 12]"
        style="margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('in.addr')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dnsServer.server" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('in.port')"
            style="margin-bottom: 0;"
          >
            <a-input-number
              v-model:value="dnsServer.server_port"
              :min="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="HasHeaders.includes(dnsServer.type)"
        :gutter="[12, 12]"
        style="margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="16"
        >
          <a-form-item
            :label="$t('transport.path')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dnsServer.path" />
          </a-form-item>
        </a-col>
      </a-row>
      <DialVue
        v-if="!WithoutDial.includes(dnsServer.type)"
        :dial="dnsServer"
      />
      <oTlsVue
        v-if="HasTls.includes(dnsServer.type)"
        :outbound="dnsServer"
      />
      <Headers
        v-if="HasHeaders.includes(dnsServer.type)"
        :data="dnsServer"
      />
      <template v-if="dnsServer.type == 'hosts'">
        <a-row
          :gutter="[12, 12]"
          style="margin-bottom: 12px;"
        >
          <a-col
            :xs="24"
            :sm="12"
          >
            <a-form-item
              :label="$t('transport.path') + $t('commaSeparated')"
              style="margin-bottom: 0;"
            >
              <a-input v-model:value="hostsPath" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-card
          size="small"
          title="Predefined"
          style="margin-bottom: 12px;"
        >
          <template #extra>
            <a-button
              type="primary"
              size="small"
              @click="addHostsPredefined"
            >
              <template #icon>
                <AppIcon name="plus" />
              </template>
            </a-button>
          </template>
          <a-row
            v-for="(pd, pdIndex) in hostsPredefined"
            :key="pdIndex"
            :gutter="[12, 12]"
            style="margin-bottom: 8px;"
          >
            <a-col
              :xs="24"
              :sm="12"
              :md="8"
            >
              <a-form-item
                :label="$t('setting.domain')"
                style="margin-bottom: 0;"
              >
                <a-input
                  :value="pd.name"
                  @input="update_pds_key(pdIndex, ($event.target as HTMLInputElement).value)"
                />
              </a-form-item>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
              :md="16"
            >
              <a-form-item
                :label="$t('types.tun.addr') + $t('commaSeparated')"
                style="margin-bottom: 0;"
              >
                <a-flex
                  gap="small"
                  align="center"
                >
                  <a-input
                    :value="pd.value"
                    @input="update_pds_value(pdIndex, ($event.target as HTMLInputElement).value)"
                  />
                  <a-button
                    :aria-label="$t('actions.del')"
                    type="text"
                    danger
                    @click="delHostsPredefined(pdIndex)"
                  >
                    <template #icon>
                      <AppIcon name="delete" />
                    </template>
                  </a-button>
                </a-flex>
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>
      </template>
      <a-row
        v-if="dnsServer.type == 'local'"
        :gutter="[12, 12]"
        style="margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-space>
            <a-switch v-model:checked="dnsServer.prefer_go" />
            <span>{{ $t('dns.local.preferGo') }}</span>
          </a-space>
        </a-col>
      </a-row>
      <a-row
        v-if="dnsServer.type == 'dhcp'"
        :gutter="[12, 12]"
        style="margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('types.tun.ifName')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dnsServer.interface" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="dnsServer.type == 'fakeip'"
        :gutter="[12, 12]"
        style="margin-bottom: 12px;"
      >
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dns.rule.inet4Range')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dnsServer.inet4_range" />
          </a-form-item>
        </a-col>
        <a-col
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('dns.rule.inet6Range')"
            style="margin-bottom: 0;"
          >
            <a-input v-model:value="dnsServer.inet6_range" />
          </a-form-item>
        </a-col>
      </a-row>
      <a-row
        v-if="dnsServer.type == 'tailscale' || dnsServer.type == 'resolved'"
        :gutter="[12, 12]"
        align="middle"
        style="margin-bottom: 12px;"
      >
        <a-col
          v-if="dnsServer.type == 'tailscale'"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('objects.endpoint')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="dnsServer.endpoint"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in tsTags"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col
          v-if="dnsServer.type == 'resolved'"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <a-form-item
            :label="$t('objects.service')"
            style="margin-bottom: 0;"
          >
            <a-select
              v-model:value="dnsServer.service"
              style="width: 100%"
            >
              <a-select-option
                v-for="item in rslvdTags"
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
          <a-space>
            <a-switch v-model:checked="dnsServer.accept_default_resolvers" />
            <span>{{ $t('dns.rule.acceptDefault') }}</span>
          </a-space>
        </a-col>
      </a-row>
    </a-form>
    <template #footer>
      <a-button
        type="default"
        @click="close"
      >
        {{ $t('actions.close') }}
      </a-button>
      <a-button
        type="primary"
        @click="save"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import DialVue from '@/components/Dial.vue'
import DocLink from '@/components/DocLink.vue'
import oTlsVue from '@/components/tls/OutTLS.vue'
import Headers from '@/components/Headers.vue'
import RandomUtil from '@/plugins/randomUtil'
import { PropType } from 'vue'
import { DnsTypes, createDnsServer } from '@/types/dns'

// One row of the predefined-hosts editor: a hostname and its comma-joined
// answers, which the getter flattens from the stored record and the setter
// splits back out.
interface PredefinedHost {
  name: string
  value: string
}
export default {
  components: { DocLink, DialVue, oTlsVue, Headers },
  props: {
    visible: { type: Boolean, required: true },
    // A JSON string of the DNS server being edited, empty when adding one.
    data: { type: String, default: '' },
    index: { type: Number, required: true },
    tsTags: { type: Array as PropType<string[]>, default: () => [] },
    rslvdTags: { type: Array as PropType<string[]>, default: () => [] },
  },
  emits: ['close', 'save'],
  data() {
    return {
      title: "add",
      dnsServer: createDnsServer("local",{tag: "dns-" + RandomUtil.randomSeq(3)}),
      dnsTypes: Object.keys(DnsTypes).map((key,index) => ({title: key, value: Object.values(DnsTypes)[index]})),
      HasServer: [DnsTypes.TCP, DnsTypes.UDP, DnsTypes.TLS, DnsTypes.QUIC, DnsTypes.HTTPS, DnsTypes.HTTP3],
      HasHeaders: [DnsTypes.HTTPS, DnsTypes.HTTP3],
      HasTls: [DnsTypes.TLS, DnsTypes.QUIC, DnsTypes.HTTPS, DnsTypes.HTTP3],
      WithoutDial: [DnsTypes.Hosts, DnsTypes.Tailscale, DnsTypes.FakeIP, DnsTypes.Resolved],
    }
  },
  computed:{
    hostsPath: {
      get() { return this.dnsServer.path },
      set(v: string) {
        this.dnsServer.path = v.length > 0 ? v.split(',').map((item: string) => item.trim()) : undefined
      }
    },
    hostsPredefined: {
      get(): PredefinedHost[] {
        const pds: PredefinedHost[] = []
        const h = this.dnsServer.predefined
        if (h) {
          Object.keys(h).forEach(key => {
            if (Array.isArray(h[key])){
              pds.push({ name: key, value: h[key].join(',') })
            } else {
              pds.push({ name: key, value: h[key] })
            }
          })
        }
        return pds
       },
      set(v: PredefinedHost[]) {
        if (v.length>0) {
          const pds: Record<string, string[]> = {}
          v.forEach(pd => {
            pds[pd.name] = pd.value.split(',').map((item: string) => item.trim())
          })
          this.dnsServer.predefined = pds
        } else {
          this.dnsServer.predefined = undefined
        }
      }
    },
  },
  watch: {
    visible(v) {
      if (v) {
        this.updateData()
      }
    },
  },
  methods: {
    updateData() {
      if (this.$props.index != -1) {
        this.dnsServer = JSON.parse(this.$props.data)
        this.title = 'edit'
      }
      else {
        this.dnsServer = createDnsServer("local",{tag: "dns-" + RandomUtil.randomSeq(3)})
        this.title = 'add'
      }
    },
    changeType(dnsType: string) {
      this.dnsServer = createDnsServer(dnsType,{tag: this.dnsServer.tag})
    },
    close() {
      this.$emit('close')
    },
    save() {
      this.$emit('save', this.dnsServer)
    },
    addHostsPredefined() {
      const newPredefined = { name:'localhost', value: '127.0.0.1,::1' }
      this.hostsPredefined = [...this.hostsPredefined, newPredefined]
    },
    delHostsPredefined(i:number) {
      let pds = this.hostsPredefined
      pds.splice(i,1)
      this.hostsPredefined = pds
    },
    update_pds_key(i:number,k:string) {
      let pds = this.hostsPredefined
      pds[i].name = k
      this.hostsPredefined = pds
    },
    update_pds_value(i:number,v:string) {
      let pds = this.hostsPredefined
      pds[i].value = v
      this.hostsPredefined = pds
    },
  }
}
</script>

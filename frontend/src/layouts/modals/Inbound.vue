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
        <span>{{ $t('actions.' + title) + " " + $t('objects.inbound') }}</span>
        <DocLink
          section="inbound"
          :type="inbound.type"
        />
      </a-flex>
    </template>
    <a-skeleton
      v-if="loading"
      active
      :paragraph="{ rows: 6 }"
    />
    <div
      v-show="!loading"
      style="max-height: 70vh; overflow-y: auto; padding-right: 4px;"
    >
      <a-form layout="vertical">
        <a-row :gutter="[12, 12]">
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
                v-model:value="inbound.type"
                :options="Object.keys(inTypes).map((key, index) => ({ title: key, value: Object.values(inTypes)[index] }))"
                :field-names="{ label: 'title', value: 'value' }"
                style="width: 100%"
                @change="changeType"
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
              <a-input v-model:value="inbound.tag" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-tabs
          v-if="HasInData.includes(inbound.type)"
          v-model:activeKey="side"
          centered
        >
          <a-tab-pane
            key="s"
            :tab="$t('in.sSide')"
          />
          <a-tab-pane
            key="c"
            :tab="$t('in.cSide')"
          />
        </a-tabs>
        <div
          v-show="side === 's'"
          style="margin-top: 10px;"
        >
          <Listen
            v-if="!NoListen.includes(inbound.type)"
            :data="inbound"
            :in-tags="inTags"
          />
          <ProxyProtocolOptions
            v-if="ProxyProtocolTypes.includes(inbound.type)"
            :data="inbound"
          />
          <Direct
            v-if="inbound.type == inTypes.Direct"
            :data="inbound"
          />
          <Shadowsocks
            v-if="inbound.type == inTypes.Shadowsocks"
            direction="in"
            :data="inbound"
          />
          <Snell
            v-if="inbound.type == inTypes.Snell"
            direction="in"
            :data="inbound"
          />
          <Cloudflared
            v-if="inbound.type == inTypes.Cloudflared"
            :data="inbound"
          />
          <Hysteria
            v-if="inbound.type == inTypes.Hysteria"
            direction="in"
            :data="inbound"
          />
          <Hysteria2
            v-if="inbound.type == inTypes.Hysteria2"
            direction="in"
            :data="inbound"
          />
          <Naive
            v-if="inbound.type == inTypes.Naive"
            direction="in"
            :data="inbound"
          />
          <ShadowTls
            v-if="inbound.type == inTypes.ShadowTLS"
            direction="in"
            :data="inbound"
          />
          <Tuic
            v-if="inbound.type == inTypes.TUIC"
            direction="in"
            :data="inbound"
          />
          <Tun
            v-if="inbound.type == inTypes.Tun"
            :data="inbound"
          />
          <AnyTls
            v-if="inbound.type == inTypes.AnyTls"
            :data="inbound"
            direction="in"
          />
          <TProxy
            v-if="inbound.type == inTypes.TProxy"
            :inbound="inbound"
          />
          <Transport
            v-if="Object.hasOwn(inbound,'transport')"
            :data="inbound"
          />
          <Users
            v-if="hasUser"
            :clients="clients"
            :data="initUsers"
          />
          <InTls
            v-if="HasTls.includes(inbound.type)"
            :inbound="inbound"
            :tls-configs="tlsConfigs"
            :tls_id="inbound.tls_id"
          />
          <Multiplex
            v-if="MuxAvailable.includes(inbound.type)"
            direction="in"
            :data="inbound"
          />
        </div>
        <div
          v-if="HasInData.includes(inbound.type)"
          v-show="side === 'c'"
          style="margin-top: 10px;"
        >
          <OutJsonVue
            :in-data="inbound"
            :type="inbound.type"
          />
          <!-- The guard tests the inbound for multiplex support but the form
               edits the client block, which only exists once out_json is
               initialised. -->
          <Multiplex
            v-if="Object.hasOwn(inbound,'multiplex') && inbound.out_json"
            direction="out"
            :data="inbound.out_json"
          />
          <Dial
            v-if="inbound.out_json"
            :dial="inbound.out_json"
            mode="client"
          />
          <a-card
            size="small"
            :title="$t('in.multiDomain')"
            style="margin-top: 12px;"
          >
            <template #extra>
              <a-button
                type="primary"
                size="small"
                @click="add_addr"
              >
                <template #icon>
                  <AppIcon name="plus" />
                </template>
              </a-button>
            </template>
            <template
              v-for="(addr, index) in inbound.addrs"
              :key="index"
            >
              <a-flex
                align="center"
                gap="small"
                style="margin-bottom: 8px;"
              >
                <span>{{ $t('in.addr') }} #{{ index + 1 }}</span>
                <a-button
                  type="text"
                  danger
                  size="small"
                  :aria-label="$t('actions.del')"
                  @click="inbound.addrs?.splice(index, 1)"
                >
                  <template #icon>
                    <AppIcon
                      name="delete"
                      color="error"
                    />
                  </template>
                </a-button>
              </a-flex>
              <a-divider style="margin: 8px 0;" />
              <AddrVue
                :addr="addr"
                :has-tls="HasTls.includes(inbound.type)"
              />
            </template>
          </a-card>
        </div>
      </a-form>
    </div>
    <template #footer>
      <a-button
        type="default"
        @click="closeModal"
      >
        {{ $t('actions.close') }}
      </a-button>
      <!-- validate has been a computed property all along with no caller, so
           an inbound with no tag or a port outside 1-65535 was saved and the
           failure only surfaced when sing-box refused to start. -->
      <a-button
        type="primary"
        :loading="loading"
        :disabled="!validate"
        @click="saveChanges"
      >
        {{ $t('actions.save') }}
      </a-button>
    </template>
  </a-modal>
</template>

<script lang="ts">
import { PropType } from 'vue'
import { InTypes, createInbound, Addr, ShadowTLS } from '@/types/inbounds'
import { tls } from '@/types/tls'
import DocLink from '@/components/DocLink.vue'
import RandomUtil from '@/plugins/randomUtil'
import Dial from '@/components/Dial.vue'
import Listen from '@/components/Listen.vue'
import Direct from '@/components/protocols/Direct.vue'
import Users from '@/components/Users.vue'
import Shadowsocks from '@/components/protocols/Shadowsocks.vue'
import Snell from '@/components/protocols/Snell.vue'
import Cloudflared from '@/components/protocols/Cloudflared.vue'
import Hysteria from '@/components/protocols/Hysteria.vue'
import Hysteria2 from '@/components/protocols/Hysteria2.vue'
import Naive from '@/components/protocols/Naive.vue'
import ShadowTls from '@/components/protocols/ShadowTls.vue'
import Tuic from '@/components/protocols/Tuic.vue'
import Tun from '@/components/protocols/Tun.vue'
import AnyTls from '@/components/protocols/AnyTls.vue'
import InTls from '@/components/tls/InTLS.vue'
import TProxy from '@/components/protocols/TProxy.vue'
import Multiplex from '@/components/Multiplex.vue'
import Transport from '@/components/Transport.vue'
import ProxyProtocolOptions from '@/components/ProxyProtocol.vue'
import AddrVue from '@/components/Addr.vue'
import OutJsonVue from '@/components/OutJson.vue'
import Data from '@/store/modules/data'
export default {
  components: {
    DocLink,
    Listen, InTls, Hysteria2, Naive, Direct, Shadowsocks,
    Users, Hysteria, ShadowTls, TProxy, Multiplex, Tuic, Tun,
    AnyTls, Transport, AddrVue, OutJsonVue, Dial, Snell, Cloudflared,
    ProxyProtocolOptions
  },
  props: {
    visible: { type: Boolean, required: true },
    id: { type: Number, required: true },
    inTags: { type: Array as PropType<string[]>, default: () => [] },
    tlsConfigs: { type: Array as PropType<tls[]>, default: () => [] },
  },
  emits: ['close'],
  data() {
    return {
      inbound: createInbound("direct",{ id:0, "tag": "" }),
      title: "add",
      loading: false,
      side: "s",
      inTypes: InTypes,
      inboundWithUsers: ['mixed', 'socks', 'http', 'shadowsocks', 'vmess', 'trojan', 'naive', 'hysteria', 'shadowtls', 'tuic', 'hysteria2', 'vless', 'anytls', 'snell'],
      initUsers: {
        model: 'none',
        values: <(number | string)[]>[],
      },
      HasInData: [
        InTypes.SOCKS,
        InTypes.HTTP,
        InTypes.Mixed,
        InTypes.Shadowsocks,
        InTypes.VMess,
        InTypes.ShadowTLS,
        InTypes.Trojan,
        InTypes.Hysteria,
        InTypes.VLESS,
        InTypes.AnyTls,
        InTypes.TUIC,
        InTypes.Hysteria2,
        InTypes.Naive,
        InTypes.Snell,
      ],
      HasTls: [
        InTypes.HTTP,
        InTypes.VMess,
        InTypes.Trojan,
        InTypes.Naive,
        InTypes.Hysteria,
        InTypes.TUIC,
        InTypes.Hysteria2,
        InTypes.VLESS,
        InTypes.AnyTls,
      ],
      MuxAvailable: [
        InTypes.VLESS,
        InTypes.VMess,
        InTypes.Trojan,
        InTypes.Shadowsocks,
      ],
      OnlyTLS: [InTypes.Hysteria, InTypes.Hysteria2, InTypes.TUIC, InTypes.Naive, InTypes.AnyTls ],
      // tun brings its own interface options; cloudflared dials out to the
      // Cloudflare edge and sing-box rejects listen fields on it.
      NoListen: [InTypes.Tun, InTypes.Cloudflared],
      // PROXY protocol is read from the TCP listener, so only inbounds with a
      // TCP accept path can use it.
      ProxyProtocolTypes: [
        InTypes.Mixed,
        InTypes.SOCKS,
        InTypes.HTTP,
        InTypes.Shadowsocks,
        InTypes.Snell,
        InTypes.VMess,
        InTypes.Trojan,
        InTypes.Naive,
        InTypes.ShadowTLS,
        InTypes.VLESS,
        InTypes.AnyTls,
      ],
    }
  },
  computed: {
    validate() {
      if (this.inbound == undefined) return false
      if (this.inbound.tag == "") return false
      if (this.inbound.listen_port > 65535 || this.inbound.listen_port < 1) return false
      if (this.OnlyTLS.includes(this.inbound.type) && this.inbound.tls_id == 0) return false
      const proxyProtocol = this.inbound.proxyProtocol
      if (proxyProtocol?.enabled && (proxyProtocol.trustedCidrs?.length ?? 0) === 0) return false
      return true
    },
    clients() {
      return Data().clients?? []
    },
    hasUser() {
      if (this.$props.id > 0) return false
      if (!this.inboundWithUsers.includes(this.inbound.type)) return false
      if (this.inbound.type == InTypes.ShadowTLS && (<ShadowTLS>this.inbound).version < 3 ) return false
      // `managed` is set by the backend on inbounds it owns; not part of the type.
      if ((this.inbound as { managed?: boolean }).managed) return false
      return true
    },
  },
  watch: {
    visible(newValue) {
      if (newValue) {
        this.loading = true
        this.updateData(this.id)
      }
    },
  },
  methods: {
    async loadData(id: number) {
      this.loading = true
      const inboundArray = await Data().loadInbounds([id])
      this.inbound = inboundArray[0]
      if (this.HasInData.includes(this.inbound.type) && this.inbound.out_json == null) {
        this.inbound.out_json = {}
      }
      this.loading = false
    },
    updateData(id: number) {
      if (id > 0) {
        this.loadData(id)
        this.title = "edit"
      }
      else {
        const port = RandomUtil.randomIntRange(10000, 60000)
        this.inbound = createInbound("direct",{ id: 0, tag: "direct-"+port ,listen: "::", listen_port: port })
        if (this.HasInData.includes(this.inbound.type)){
          this.inbound.addrs = []
          this.inbound.out_json = {}
        } else {
          delete this.inbound.addrs
          delete this.inbound.out_json
        }
        this.title = "add"
        this.loading = false
      }
      this.side = "s"
      this.initUsers = {
        model: 'none',
        values: [],
      }
    },
    changeType() {
      if (!this.inbound.listen_port) this.inbound.listen_port = RandomUtil.randomIntRange(10000, 60000)
      const noListen = this.NoListen.includes(this.inbound.type)
      // Tag change only in add inbound
      const tag = this.$props.id > 0
        ? this.inbound.tag
        : this.inbound.type + "-" + (noListen ? RandomUtil.randomSeq(3) : this.inbound.listen_port)
      // Use previous data. Types that do not listen must not carry the listen
      // settings over, sing-box rejects them as unknown fields.
      const prevConfig = noListen
        ? { id: this.inbound.id, tag: tag }
        : { id: this.inbound.id, tag: tag, listen: this.inbound.listen?? "::", listen_port: this.inbound.listen_port }
      this.inbound = createInbound(this.inbound.type, prevConfig)
      if (this.HasInData.includes(this.inbound.type)){
        this.inbound.addrs = []
        this.inbound.out_json = {}
      } else {
        delete this.inbound.addrs
        delete this.inbound.out_json
      }
      this.side = "s"
    },
    add_addr() {
      this.inbound.addrs?.push(<Addr>{ server: location.hostname, server_port: this.inbound.listen_port })
    },
    closeModal() {
      this.updateData(0) // reset
      this.$emit('close')
    },
    async saveChanges() {
      if (!this.$props.visible) return
      // check duplicate tag
      const isDuplicatedTag = Data().checkTag("inbound", this.inbound.id, this.inbound.tag)
      if (isDuplicatedTag) return

      // save data
      this.loading = true
      // Persisted clients always carry an id, and in 'client' mode the
      // selection holds ids rather than group names.
      let clientIds: number[] = []
      if (this.hasUser) {
        switch (this.initUsers.model) {
          case 'all':
            clientIds = this.clients.map(c => c.id as number)
            break
          case 'group':
            clientIds = this.clients.filter(c => this.initUsers.values.includes(c.group)).map(c => c.id as number)
            break
          case 'client':
            clientIds = this.initUsers.values as number[]
        }
      }
      const success = await Data().save("inbounds", this.$props.id == 0 ? "new" : "edit", this.inbound, clientIds)
      if (success) this.closeModal()
      this.loading = false
    },
  }
}
</script>
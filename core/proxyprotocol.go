package core

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/util/proxyproto"

	sbListener "github.com/sagernet/sing-box/common/listener"
	singbuf "github.com/sagernet/sing/common/buf"
	singbufio "github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
)

// ProxyProtocolConfig is the panel-side policy for one inbound. It is never
// written to the sing-box config: only connections whose peer is inside
// Trusted are expected to start with a PROXY header.
type ProxyProtocolConfig struct {
	Enabled       bool
	InboundTag    string
	Trusted       []netip.Prefix
	HeaderTimeout time.Duration
}

const defaultProxyHeaderTimeout = 5 * time.Second

type proxyPolicyStore struct {
	access sync.RWMutex
	byPort map[uint16]ProxyProtocolConfig
}

func newProxyPolicyStore() *proxyPolicyStore {
	return &proxyPolicyStore{byPort: make(map[uint16]ProxyProtocolConfig)}
}

func (s *proxyPolicyStore) Set(policies map[uint16]ProxyProtocolConfig) {
	copied := make(map[uint16]ProxyProtocolConfig, len(policies))
	for port, policy := range policies {
		if port == 0 || !policy.Enabled {
			continue
		}
		copied[port] = policy
	}
	s.access.Lock()
	s.byPort = copied
	s.access.Unlock()
}

func (s *proxyPolicyStore) Get(port uint16) (ProxyProtocolConfig, bool) {
	s.access.RLock()
	policy, loaded := s.byPort[port]
	s.access.RUnlock()
	return policy, loaded
}

// proxyPreprocessor reads the PROXY header for inbounds that opted in and
// whose peer is a trusted relay. Everything else is left untouched: a direct
// client's peer address is already the real client address.
type proxyPreprocessor struct {
	policies *proxyPolicyStore
}

func newProxyPreprocessor(policies *proxyPolicyStore) *proxyPreprocessor {
	return &proxyPreprocessor{policies: policies}
}

func portOf(addr net.Addr) uint16 {
	switch value := addr.(type) {
	case *net.TCPAddr:
		if value.Port < 0 || value.Port > 65535 {
			return 0
		}
		return uint16(value.Port)
	}
	if addr == nil {
		return 0
	}
	parsed, err := netip.ParseAddrPort(addr.String())
	if err != nil {
		return 0
	}
	return parsed.Port()
}

// policyFor reports whether this connection must carry a PROXY header.
func (p *proxyPreprocessor) policyFor(conn net.Conn) (ProxyProtocolConfig, M.Socksaddr, bool) {
	policy, loaded := p.policies.Get(portOf(conn.LocalAddr()))
	if !loaded {
		return ProxyProtocolConfig{}, M.Socksaddr{}, false
	}
	peer := M.SocksaddrFromNet(conn.RemoteAddr())
	if !peer.IsValid() {
		return policy, peer, false
	}
	peerAddr := peer.Addr.Unmap()
	for _, prefix := range policy.Trusted {
		if prefix.Contains(peerAddr) {
			return policy, peer, true
		}
	}
	return policy, peer, false
}

// Match implements listener.ConnectionPreprocessorMatcher, so the transport
// listener can skip the wrapper for direct clients.
func (p *proxyPreprocessor) Match(conn net.Conn) bool {
	_, _, required := p.policyFor(conn)
	return required
}

// Process consumes the PROXY header. It returns an error when the peer is in
// the trusted list but the header is missing, malformed or late: falling back
// to the relay address would silently attribute every such user to the relay.
func (p *proxyPreprocessor) Process(ctx context.Context, conn net.Conn) (context.Context, net.Conn, M.Socksaddr, error) {
	policy, peer, required := p.policyFor(conn)
	if !required {
		return ctx, conn, M.Socksaddr{}, nil
	}
	timeout := policy.HeaderTimeout
	if timeout <= 0 {
		timeout = defaultProxyHeaderTimeout
	}
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	reader := bufio.NewReaderSize(conn, 512)
	header, err := proxyproto.ReadHeader(reader)
	_ = conn.SetReadDeadline(time.Time{})
	if err != nil {
		return ctx, conn, M.Socksaddr{}, fmt.Errorf("inbound %q: PROXY header from %s: %w", policy.InboundTag, peer, err)
	}
	var clientAddr M.Socksaddr
	if header.Command == proxyproto.CommandLocal {
		// A relay health check carries no client. Attribute it to the relay
		// itself; it never authenticates as a user, so it is not counted.
		clientAddr = peer
	} else {
		if !header.Source.IsValid() {
			return ctx, conn, M.Socksaddr{}, fmt.Errorf("inbound %q: PROXY header from %s has no client address", policy.InboundTag, peer)
		}
		clientAddr = M.SocksaddrFrom(header.Source.Unmap(), header.SourcePort)
	}
	if buffered := reader.Buffered(); buffered > 0 {
		buffer := singbuf.NewSize(buffered)
		if _, err = buffer.ReadFullFrom(reader, buffered); err != nil {
			buffer.Release()
			return ctx, conn, M.Socksaddr{}, fmt.Errorf("inbound %q: buffer bytes after PROXY header: %w", policy.InboundTag, err)
		}
		conn = singbufio.NewCachedConn(conn, buffer)
	}
	logger.Debug("inbound ", policy.InboundTag, ": client ", clientAddr, " via relay ", peer)
	return sbListener.ContextWithClientAddr(ctx, clientAddr), conn, clientAddr, nil
}

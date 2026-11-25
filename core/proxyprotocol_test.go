package core

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/netip"
	"testing"
	"time"

	sbListener "github.com/sagernet/sing-box/common/listener"

	M "github.com/sagernet/sing/common/metadata"
)

func tcpPair(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	accepted := make(chan net.Conn, 1)
	acceptErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			acceptErr <- err
			return
		}
		accepted <- conn
	}()
	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	select {
	case server := <-accepted:
		t.Cleanup(func() { _ = server.Close() })
		return client, server
	case err := <-acceptErr:
		t.Fatalf("accept: %v", err)
		return nil, nil
	case <-time.After(5 * time.Second):
		t.Fatal("accept timed out")
		return nil, nil
	}
}

func policyFor(server net.Conn, trusted ...string) *proxyPolicyStore {
	store := newProxyPolicyStore()
	prefixes := make([]netip.Prefix, 0, len(trusted))
	for _, cidr := range trusted {
		prefixes = append(prefixes, netip.MustParsePrefix(cidr))
	}
	store.Set(map[uint16]ProxyProtocolConfig{
		portOf(server.LocalAddr()): {
			Enabled:       true,
			InboundTag:    "relay-in",
			Trusted:       prefixes,
			HeaderTimeout: 2 * time.Second,
		},
	})
	return store
}

func v2TCP4Header(source netip.Addr, sourcePort uint16) []byte {
	payload := make([]byte, 12)
	copy(payload[0:4], source.AsSlice())
	binary.BigEndian.PutUint16(payload[8:10], sourcePort)
	header := []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x21, 0x11, 0, 0}
	binary.BigEndian.PutUint16(header[14:16], uint16(len(payload)))
	return append(header, payload...)
}

func TestProxyPreprocessorReadsV1AndKeepsPayload(t *testing.T) {
	client, server := tcpPair(t)
	preprocessor := newProxyPreprocessor(policyFor(server, "127.0.0.1/32"))
	if !preprocessor.Match(server) {
		t.Fatal("trusted peer did not match")
	}
	go func() {
		_, _ = client.Write([]byte("PROXY TCP4 192.0.2.10 198.51.100.20 54321 443\r\nhello"))
	}()
	ctx, conn, addr, err := preprocessor.Process(context.Background(), server)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if addr.String() != "192.0.2.10:54321" {
		t.Fatalf("client = %v", addr)
	}
	if got := sbListener.ClientAddrFromContext(ctx); got != addr {
		t.Fatalf("context client = %v", got)
	}
	payload := make([]byte, 5)
	if _, err = io.ReadFull(conn, payload); err != nil {
		t.Fatalf("read payload: %v", err)
	}
	if string(payload) != "hello" {
		t.Fatalf("payload = %q", payload)
	}
}

func TestProxyPreprocessorReadsV2(t *testing.T) {
	client, server := tcpPair(t)
	preprocessor := newProxyPreprocessor(policyFor(server, "127.0.0.1/32"))
	go func() {
		_, _ = client.Write(append(v2TCP4Header(netip.MustParseAddr("203.0.113.5"), 60000), []byte("x")...))
	}()
	_, conn, addr, err := preprocessor.Process(context.Background(), server)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if addr.String() != "203.0.113.5:60000" {
		t.Fatalf("client = %v", addr)
	}
	data := make([]byte, 1)
	if _, err = io.ReadFull(conn, data); err != nil || data[0] != 'x' {
		t.Fatalf("payload = %q, err = %v", data, err)
	}
}

func TestProxyPreprocessorRejectsMissingHeader(t *testing.T) {
	client, server := tcpPair(t)
	store := policyFor(server, "127.0.0.1/32")
	for port, policy := range store.byPort {
		policy.HeaderTimeout = 100 * time.Millisecond
		store.byPort[port] = policy
	}
	preprocessor := newProxyPreprocessor(store)
	go func() {
		_, _ = client.Write([]byte("GET / HTTP/1.1\r\n\r\n"))
	}()
	_, _, addr, err := preprocessor.Process(context.Background(), server)
	if err == nil {
		t.Fatal("expected the missing header to be rejected")
	}
	if addr.IsValid() {
		t.Fatalf("unexpected client address %v", addr)
	}
}

func TestProxyPreprocessorIgnoresUntrustedPeerHeader(t *testing.T) {
	client, server := tcpPair(t)
	preprocessor := newProxyPreprocessor(policyFor(server, "10.0.0.0/8"))
	if preprocessor.Match(server) {
		t.Fatal("untrusted peer matched")
	}
	ctx, conn, addr, err := preprocessor.Process(context.Background(), server)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if conn != server || addr.IsValid() || sbListener.ClientAddrFromContext(ctx).IsValid() {
		t.Fatalf("untrusted connection was touched: %v %v", conn, addr)
	}
	// The forged header must stay in the stream for the protocol to reject it.
	go func() {
		_, _ = client.Write([]byte("PROXY TCP4 1.2.3.4 5.6.7.8 1 2\r\n"))
	}()
	head := make([]byte, 6)
	if _, err = io.ReadFull(server, head); err != nil || string(head) != "PROXY " {
		t.Fatalf("head = %q, err = %v", head, err)
	}
}

func TestProxyPreprocessorLocalHeaderUsesPeer(t *testing.T) {
	client, server := tcpPair(t)
	preprocessor := newProxyPreprocessor(policyFor(server, "127.0.0.1/32"))
	header := []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x20, 0x00, 0, 0}
	go func() { _, _ = client.Write(header) }()
	_, _, addr, err := preprocessor.Process(context.Background(), server)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if addr.Addr.String() != "127.0.0.1" {
		t.Fatalf("client = %v", addr)
	}
}

func TestProxyPreprocessorLeavesDisabledPortAlone(t *testing.T) {
	client, server := tcpPair(t)
	_ = client
	preprocessor := newProxyPreprocessor(newProxyPolicyStore())
	if preprocessor.Match(server) {
		t.Fatal("disabled port matched")
	}
	conn, addr, err := processNoHeader(t, preprocessor, server)
	if err != nil || conn != server || addr.IsValid() {
		t.Fatalf("conn = %v, addr = %v, err = %v", conn, addr, err)
	}
}

func processNoHeader(t *testing.T, preprocessor *proxyPreprocessor, conn net.Conn) (net.Conn, M.Socksaddr, error) {
	t.Helper()
	_, processed, addr, err := preprocessor.Process(context.Background(), conn)
	return processed, addr, err
}

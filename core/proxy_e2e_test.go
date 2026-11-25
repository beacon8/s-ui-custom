package core

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/netip"
	"testing"
	"time"
)

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port
}

func startEchoServer(t *testing.T) net.Listener {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("echo listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				_, _ = io.Copy(conn, conn)
			}()
		}
	}()
	return listener
}

func socksNegotiate(t *testing.T, conn net.Conn, user, password string) {
	t.Helper()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	if _, err := conn.Write([]byte{0x05, 0x01, 0x02}); err != nil {
		t.Fatalf("socks greeting: %v", err)
	}
	reply := make([]byte, 2)
	if _, err := io.ReadFull(conn, reply); err != nil {
		t.Fatalf("socks greeting reply: %v", err)
	}
	if reply[0] != 0x05 || reply[1] != 0x02 {
		t.Fatalf("unexpected socks greeting reply: %v", reply)
	}
	auth := []byte{0x01, byte(len(user))}
	auth = append(auth, user...)
	auth = append(auth, byte(len(password)))
	auth = append(auth, password...)
	if _, err := conn.Write(auth); err != nil {
		t.Fatalf("socks auth: %v", err)
	}
	if _, err := io.ReadFull(conn, reply); err != nil {
		t.Fatalf("socks auth reply: %v", err)
	}
	if reply[1] != 0x00 {
		t.Fatalf("socks auth failed: %v", reply)
	}
}

func socksConnect(t *testing.T, conn net.Conn, target *net.TCPAddr) error {
	t.Helper()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	ip := target.IP.To4()
	if ip == nil {
		t.Fatal("test target is not IPv4")
	}
	request := []byte{0x05, 0x01, 0x00, 0x01, ip[0], ip[1], ip[2], ip[3], 0, 0}
	binary.BigEndian.PutUint16(request[8:10], uint16(target.Port))
	if _, err := conn.Write(request); err != nil {
		return err
	}
	reply := make([]byte, 10)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return err
	}
	if reply[1] != 0x00 {
		return fmt.Errorf("socks connect status %d", reply[1])
	}
	return nil
}

func waitForSession(t *testing.T, c *Core, user string) SessionInfo {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if instance := c.GetInstance(); instance != nil {
			for _, session := range instance.SessionTracker().Sessions() {
				if session.User == user {
					return session
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no live session for %q", user)
	return SessionInfo{}
}

func TestProxyProtocolEndToEndSocks(t *testing.T) {
	echo := startEchoServer(t)
	echoAddr := echo.Addr().(*net.TCPAddr)
	port := freeTCPPort(t)
	config := fmt.Sprintf(`{
		"log": {"level": "error"},
		"inbounds": [{
			"type": "socks",
			"tag": "relay-in",
			"listen": "127.0.0.1",
			"listen_port": %d,
			"users": [{"username": "alice", "password": "secret"}]
		}],
		"outbounds": [{"type": "direct", "tag": "direct"}]
	}`, port)

	c := NewCore()
	c.SetProxyProtocolPolicies(map[uint16]ProxyProtocolConfig{
		uint16(port): {
			Enabled:       true,
			InboundTag:    "relay-in",
			Trusted:       []netip.Prefix{netip.MustParsePrefix("127.0.0.1/32")},
			HeaderTimeout: 2 * time.Second,
		},
	})
	c.SetUserIPLimits(map[string]int{"alice": 1})
	if err := c.Start([]byte(config)); err != nil {
		t.Fatalf("start core: %v", err)
	}
	t.Cleanup(func() { _ = c.Stop() })

	address := fmt.Sprintf("127.0.0.1:%d", port)
	first, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("dial first: %v", err)
	}
	defer first.Close()
	if _, err = first.Write([]byte("PROXY TCP4 192.0.2.10 203.0.113.1 50000 443\r\n")); err != nil {
		t.Fatalf("write proxy header: %v", err)
	}
	socksNegotiate(t, first, "alice", "secret")
	if err = socksConnect(t, first, echoAddr); err != nil {
		t.Fatalf("socks connect: %v", err)
	}
	if _, err = first.Write([]byte("ping")); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	echoed := make([]byte, 4)
	if _, err = io.ReadFull(first, echoed); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if string(echoed) != "ping" {
		t.Fatalf("echo = %q", echoed)
	}

	session := waitForSession(t, c, "alice")
	if session.Client != "192.0.2.10:50000" {
		t.Fatalf("client = %q, want 192.0.2.10:50000", session.Client)
	}
	if session.Source == session.Client {
		t.Fatalf("source was overwritten with the client address: %q", session.Source)
	}
	if session.Source[:10] != "127.0.0.1:" {
		t.Fatalf("source = %q, want the loopback peer", session.Source)
	}
	if got := c.GetInstance().SessionTracker().ActiveIPCounts()["alice"]; got != 1 {
		t.Fatalf("active IPs = %d, want 1", got)
	}

	second, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("dial second: %v", err)
	}
	defer second.Close()
	if _, err = second.Write([]byte("PROXY TCP4 198.51.100.20 203.0.113.1 50001 443\r\n")); err != nil {
		t.Fatalf("write second proxy header: %v", err)
	}
	socksNegotiate(t, second, "alice", "secret")
	// The connection may already be closed by the IP limit, in which case the
	// SOCKS reply itself fails. Either way no payload may flow.
	if err = socksConnect(t, second, echoAddr); err == nil {
		_ = second.SetDeadline(time.Now().Add(500 * time.Millisecond))
		_, _ = second.Write([]byte("ping"))
		buffer := make([]byte, 1)
		if _, readErr := second.Read(buffer); readErr == nil {
			t.Fatal("second IP was not cut off")
		}
	}
	if got := c.GetInstance().SessionTracker().ActiveIPCounts()["alice"]; got != 1 {
		t.Fatalf("active IPs after rejection = %d, want 1", got)
	}
	summary := c.GetInstance().SessionTracker().UserIPs("alice")
	if summary.Count != 1 || summary.IPs[0].IP != "192.0.2.10/32" {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestProxyProtocolUntrustedPeerIsNotParsedEndToEnd(t *testing.T) {
	port := freeTCPPort(t)
	config := fmt.Sprintf(`{
		"log": {"level": "error"},
		"inbounds": [{
			"type": "socks",
			"tag": "relay-in",
			"listen": "127.0.0.1",
			"listen_port": %d
		}],
		"outbounds": [{"type": "direct", "tag": "direct"}]
	}`, port)

	c := NewCore()
	c.SetProxyProtocolPolicies(map[uint16]ProxyProtocolConfig{
		uint16(port): {
			Enabled:       true,
			InboundTag:    "relay-in",
			Trusted:       []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			HeaderTimeout: 200 * time.Millisecond,
		},
	})
	if err := c.Start([]byte(config)); err != nil {
		t.Fatalf("start core: %v", err)
	}
	t.Cleanup(func() { _ = c.Stop() })

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	// A forged header from an untrusted peer must reach the SOCKS handshake,
	// which rejects it as an invalid version: "PROXY " starts with 'P', not
	// 0x05, so there is no valid greeting reply.
	_ = conn.SetDeadline(time.Now().Add(time.Second))
	_, _ = conn.Write([]byte("PROXY TCP4 1.2.3.4 5.6.7.8 1 2\r\n"))
	_, _ = conn.Write([]byte{0x05, 0x01, 0x00})
	buffer := make([]byte, 2)
	if n, readErr := io.ReadFull(conn, buffer); readErr == nil && n == 2 && buffer[0] == 0x05 && buffer[1] == 0x00 {
		t.Fatal("the forged PROXY header was consumed as a valid handshake")
	}
}

// headerForwarder accepts local connections, prefixes each with a PROXY header
// and forwards the stream to target, emulating the relay in front of s-ui.
func startHeaderForwarder(t *testing.T, target string, header string) net.Listener {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("forwarder listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			client, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer client.Close()
				upstream, err := net.Dial("tcp", target)
				if err != nil {
					return
				}
				defer upstream.Close()
				if _, err = upstream.Write([]byte(header)); err != nil {
					return
				}
				done := make(chan struct{}, 2)
				go func() {
					_, _ = io.Copy(upstream, client)
					done <- struct{}{}
				}()
				go func() {
					_, _ = io.Copy(client, upstream)
					done <- struct{}{}
				}()
				<-done
			}()
		}
	}()
	return listener
}

type transportCase struct {
	name       string
	transport  string
	serverPort int
	socksPort  int
}

func TestProxyProtocolEndToEndTransports(t *testing.T) {
	cases := []struct {
		name       string
		serverJSON string
		clientJSON string
	}{
		{
			name:       "websocket",
			serverJSON: `"transport": {"type": "ws", "path": "/ws"}`,
			clientJSON: `"transport": {"type": "ws", "path": "/ws"}`,
		},
		{
			name:       "httpupgrade",
			serverJSON: `"transport": {"type": "httpupgrade", "path": "/upgrade"}`,
			clientJSON: `"transport": {"type": "httpupgrade", "path": "/upgrade"}`,
		},
		{
			name:       "http",
			serverJSON: `"transport": {"type": "http", "host": ["example.com"], "path": "/h"}`,
			clientJSON: `"transport": {"type": "http", "host": ["example.com"], "path": "/h"}`,
		},
		{
			name:       "grpc",
			serverJSON: `"transport": {"type": "grpc", "service_name": "GunService"}`,
			clientJSON: `"transport": {"type": "grpc", "service_name": "GunService"}`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			runProxyTransportCase(t, testCase.serverJSON, testCase.clientJSON)
		})
	}
}

func runProxyTransportCase(t *testing.T, serverTransport string, clientTransport string) {
	t.Helper()
	echo := startEchoServer(t)
	echoAddr := echo.Addr().(*net.TCPAddr)
	vlessPort := freeTCPPort(t)
	socksPort := freeTCPPort(t)
	const uuid = "b831381d-6324-4d53-ad4f-8cda48b30811"

	serverConfig := fmt.Sprintf(`{
		"log": {"level": "error"},
		"inbounds": [{
			"type": "vless",
			"tag": "vless-in",
			"listen": "127.0.0.1",
			"listen_port": %d,
			"users": [{"name": "alice", "uuid": "%s"}],
			%s
		}],
		"outbounds": [{"type": "direct", "tag": "direct"}]
	}`, vlessPort, uuid, serverTransport)

	server := NewCore()
	server.SetProxyProtocolPolicies(map[uint16]ProxyProtocolConfig{
		uint16(vlessPort): {
			Enabled:       true,
			InboundTag:    "vless-in",
			Trusted:       []netip.Prefix{netip.MustParsePrefix("127.0.0.1/32")},
			HeaderTimeout: 2 * time.Second,
		},
	})
	if err := server.Start([]byte(serverConfig)); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() { _ = server.Stop() })

	forwarder := startHeaderForwarder(t,
		fmt.Sprintf("127.0.0.1:%d", vlessPort),
		"PROXY TCP4 192.0.2.10 203.0.113.1 50000 443\r\n")
	forwarderPort := forwarder.Addr().(*net.TCPAddr).Port

	clientConfig := fmt.Sprintf(`{
		"log": {"level": "error"},
		"inbounds": [{
			"type": "socks",
			"tag": "socks-in",
			"listen": "127.0.0.1",
			"listen_port": %d
		}],
		"outbounds": [{
			"type": "vless",
			"tag": "vless-out",
			"server": "127.0.0.1",
			"server_port": %d,
			"uuid": "%s",
			%s
		}]
	}`, socksPort, forwarderPort, uuid, clientTransport)

	client := NewCore()
	if err := client.Start([]byte(clientConfig)); err != nil {
		t.Fatalf("start client: %v", err)
	}
	t.Cleanup(func() { _ = client.Stop() })

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", socksPort))
	if err != nil {
		t.Fatalf("dial client socks: %v", err)
	}
	defer conn.Close()
	socksNegotiateNoAuth(t, conn)
	if err = socksConnect(t, conn, echoAddr); err != nil {
		t.Fatalf("socks connect through vless: %v", err)
	}
	if _, err = conn.Write([]byte("ping")); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	echoed := make([]byte, 4)
	if _, err = io.ReadFull(conn, echoed); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if string(echoed) != "ping" {
		t.Fatalf("echo = %q", echoed)
	}

	session := waitForSession(t, server, "alice")
	if session.Client != "192.0.2.10:50000" {
		t.Fatalf("client over transport = %q, want 192.0.2.10:50000", session.Client)
	}
	if session.Source == session.Client {
		t.Fatalf("source was overwritten: %q", session.Source)
	}
	if got := server.GetInstance().SessionTracker().ActiveIPCounts()["alice"]; got != 1 {
		t.Fatalf("active IPs = %d, want 1", got)
	}
}

func socksNegotiateNoAuth(t *testing.T, conn net.Conn) {
	t.Helper()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		t.Fatalf("socks greeting: %v", err)
	}
	reply := make([]byte, 2)
	if _, err := io.ReadFull(conn, reply); err != nil {
		t.Fatalf("socks greeting reply: %v", err)
	}
	if reply[0] != 0x05 || reply[1] != 0x00 {
		t.Fatalf("unexpected socks greeting reply: %v", reply)
	}
}

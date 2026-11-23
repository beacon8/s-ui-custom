//go:build ratee2e

package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	e2eProxyUser = "alice"
	e2eProxyPass = "secret"
)

type tcpSink struct {
	listener net.Listener
	bytes    atomic.Int64
	mu       sync.Mutex
	conns    []net.Conn
	closed   atomic.Bool
}

func startTCPSink(t *testing.T) *tcpSink {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	sink := &tcpSink{listener: listener}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			sink.mu.Lock()
			sink.conns = append(sink.conns, conn)
			sink.mu.Unlock()
			go func(conn net.Conn) {
				defer conn.Close()
				buffer := make([]byte, 32*1024)
				for {
					n, err := conn.Read(buffer)
					if n > 0 {
						sink.bytes.Add(int64(n))
					}
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()
	return sink
}

func (s *tcpSink) Addr() string { return s.listener.Addr().String() }
func (s *tcpSink) Close() {
	if s.closed.Swap(true) {
		return
	}
	_ = s.listener.Close()
	s.mu.Lock()
	for _, conn := range s.conns {
		_ = conn.Close()
	}
	s.mu.Unlock()
}

type tcpSource struct {
	listener net.Listener
	bytes    atomic.Int64
	mu       sync.Mutex
	conns    []net.Conn
	closed   atomic.Bool
}

func startTCPSource(t *testing.T) *tcpSource {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	source := &tcpSource{listener: listener}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			source.mu.Lock()
			source.conns = append(source.conns, conn)
			source.mu.Unlock()
			go func(conn net.Conn) {
				defer conn.Close()
				payload := make([]byte, 32*1024)
				for {
					n, err := conn.Write(payload)
					if n > 0 {
						source.bytes.Add(int64(n))
					}
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()
	return source
}

func (s *tcpSource) Addr() string { return s.listener.Addr().String() }
func (s *tcpSource) Close() {
	if s.closed.Swap(true) {
		return
	}
	_ = s.listener.Close()
	s.mu.Lock()
	for _, conn := range s.conns {
		_ = conn.Close()
	}
	s.mu.Unlock()
}

type tcpEcho struct {
	listener net.Listener
	mu       sync.Mutex
	conns    []net.Conn
	closed   atomic.Bool
}

func startTCPEcho(t *testing.T) *tcpEcho {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	echo := &tcpEcho{listener: listener}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			echo.mu.Lock()
			echo.conns = append(echo.conns, conn)
			echo.mu.Unlock()
			go func(conn net.Conn) {
				defer conn.Close()
				_, _ = io.Copy(conn, conn)
			}(conn)
		}
	}()
	return echo
}

func (s *tcpEcho) Addr() string { return s.listener.Addr().String() }

func (s *tcpEcho) Close() {
	if s.closed.Swap(true) {
		return
	}
	_ = s.listener.Close()
	s.mu.Lock()
	for _, conn := range s.conns {
		_ = conn.Close()
	}
	s.mu.Unlock()
}

type udpSink struct {
	conn  *net.UDPConn
	bytes atomic.Int64
}

func startUDPSink(t *testing.T) *udpSink {
	t.Helper()
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	sink := &udpSink{conn: conn}
	go func() {
		buffer := make([]byte, 64*1024)
		for {
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			sink.bytes.Add(int64(n))
		}
	}()
	return sink
}

func (s *udpSink) Addr() *net.UDPAddr { return s.conn.LocalAddr().(*net.UDPAddr) }
func (s *udpSink) Close()             { _ = s.conn.Close() }

type udpSource struct {
	conn      *net.UDPConn
	bytes     atomic.Int64
	stop      chan struct{}
	mu        sync.Mutex
	remotes   map[string]*net.UDPAddr
	closeOnce sync.Once
}

func startUDPSource(t *testing.T) *udpSource {
	t.Helper()
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	source := &udpSource{conn: conn, stop: make(chan struct{}), remotes: make(map[string]*net.UDPAddr)}
	go func() {
		buffer := make([]byte, 64*1024)
		for {
			_, remote, err := conn.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			source.mu.Lock()
			key := remote.String()
			if _, loaded := source.remotes[key]; !loaded {
				source.remotes[key] = remote
				go source.send(remote)
			}
			source.mu.Unlock()
		}
	}()
	return source
}

func (s *udpSource) Addr() *net.UDPAddr { return s.conn.LocalAddr().(*net.UDPAddr) }

func (s *udpSource) send(remote *net.UDPAddr) {
	payload := make([]byte, 1200)
	for {
		select {
		case <-s.stop:
			return
		default:
		}
		n, err := s.conn.WriteToUDP(payload, remote)
		if n > 0 {
			s.bytes.Add(int64(n))
		}
		if err != nil {
			return
		}
	}
}

func (s *udpSource) Close() {
	s.closeOnce.Do(func() {
		close(s.stop)
		_ = s.conn.Close()
	})
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port
}

func startE2ECore(t *testing.T) (*Core, string, string) {
	t.Helper()
	port1 := freeTCPPort(t)
	port2 := freeTCPPort(t)
	config := fmt.Sprintf(`{
  "log": {"level": "error"},
  "inbounds": [
    {"type":"socks","tag":"socks1","listen":"127.0.0.1","listen_port":%d,"users":[{"username":"alice","password":"secret"},{"username":"bob","password":"secret"},{"username":"carol","password":"secret"}]},
    {"type":"socks","tag":"socks2","listen":"127.0.0.1","listen_port":%d,"users":[{"username":"alice","password":"secret"},{"username":"bob","password":"secret"},{"username":"carol","password":"secret"}]}
  ],
  "outbounds": [{"type":"direct","tag":"direct"}]
}`, port1, port2)
	instance := NewCore()
	if err := instance.Start([]byte(config)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = instance.Stop() })
	return instance, net.JoinHostPort("127.0.0.1", strconv.Itoa(port1)), net.JoinHostPort("127.0.0.1", strconv.Itoa(port2))
}

func socksReadFull(conn net.Conn, n int) ([]byte, error) {
	buffer := make([]byte, n)
	_, err := io.ReadFull(conn, buffer)
	return buffer, err
}

func socksHandshake(conn net.Conn, user, pass string) error {
	if _, err := conn.Write([]byte{5, 1, 2}); err != nil {
		return err
	}
	reply, err := socksReadFull(conn, 2)
	if err != nil {
		return err
	}
	if reply[0] != 5 || reply[1] != 2 {
		return fmt.Errorf("socks greeting reply %v", reply)
	}
	auth := []byte{1, byte(len(user))}
	auth = append(auth, user...)
	auth = append(auth, byte(len(pass)))
	auth = append(auth, pass...)
	if _, err := conn.Write(auth); err != nil {
		return err
	}
	reply, err = socksReadFull(conn, 2)
	if err != nil {
		return err
	}
	if reply[1] != 0 {
		return fmt.Errorf("socks auth reply %v", reply)
	}
	return nil
}

func socksAddress(target string) ([]byte, error) {
	host, portText, err := net.SplitHostPort(target)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("test expects an IPv4 target, got %q", target)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return nil, err
	}
	address := []byte{1}
	address = append(address, ip.To4()...)
	address = append(address, byte(port>>8), byte(port))
	return address, nil
}

func socksDial(proxy, user, pass, target string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", proxy, 5*time.Second)
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err = socksHandshake(conn, user, pass); err != nil {
		_ = conn.Close()
		return nil, err
	}
	address, err := socksAddress(target)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	request := append([]byte{5, 1, 0}, address...)
	if _, err = conn.Write(request); err != nil {
		_ = conn.Close()
		return nil, err
	}
	reply, err := socksReadFull(conn, 4)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if reply[1] != 0 {
		_ = conn.Close()
		return nil, fmt.Errorf("socks connect reply %v", reply)
	}
	addressLength := 0
	switch reply[3] {
	case 1:
		addressLength = 4
	case 4:
		addressLength = 16
	case 3:
		length, err := socksReadFull(conn, 1)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		addressLength = int(length[0])
	default:
		_ = conn.Close()
		return nil, fmt.Errorf("socks connect atyp %d", reply[3])
	}
	if _, err = socksReadFull(conn, addressLength+2); err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = conn.SetDeadline(time.Time{})
	return conn, nil
}

type socksUDPClient struct {
	control net.Conn
	udp     *net.UDPConn
}

func newSocksUDP(proxy, user, pass string) (*socksUDPClient, error) {
	conn, err := net.DialTimeout("tcp", proxy, 5*time.Second)
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err = socksHandshake(conn, user, pass); err != nil {
		_ = conn.Close()
		return nil, err
	}
	request := []byte{5, 3, 0, 1, 0, 0, 0, 0, 0, 0}
	if _, err = conn.Write(request); err != nil {
		_ = conn.Close()
		return nil, err
	}
	reply, err := socksReadFull(conn, 4)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if reply[1] != 0 {
		_ = conn.Close()
		return nil, fmt.Errorf("socks udp reply %v", reply)
	}
	var host string
	switch reply[3] {
	case 1:
		raw, err := socksReadFull(conn, 4)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		host = net.IP(raw).String()
	case 3:
		length, err := socksReadFull(conn, 1)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		raw, err := socksReadFull(conn, int(length[0]))
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		host = string(raw)
	case 4:
		raw, err := socksReadFull(conn, 16)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		host = net.IP(raw).String()
	default:
		_ = conn.Close()
		return nil, fmt.Errorf("socks udp atyp %d", reply[3])
	}
	portBytes, err := socksReadFull(conn, 2)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	port := int(binary.BigEndian.Uint16(portBytes))
	if host == "0.0.0.0" || host == "" {
		host, _, _ = net.SplitHostPort(proxy)
	}
	relay := &net.UDPAddr{IP: net.ParseIP(host), Port: port}
	udp, err := net.DialUDP("udp", nil, relay)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = conn.SetDeadline(time.Time{})
	return &socksUDPClient{control: conn, udp: udp}, nil
}

func (c *socksUDPClient) Close() {
	_ = c.udp.Close()
	_ = c.control.Close()
}

func (c *socksUDPClient) Send(target *net.UDPAddr, payload []byte) error {
	header := []byte{0, 0, 0, 1}
	header = append(header, target.IP.To4()...)
	header = append(header, byte(target.Port>>8), byte(target.Port))
	packet := append(header, payload...)
	_, err := c.udp.Write(packet)
	return err
}

func (c *socksUDPClient) Receive(buffer []byte) (int, error) {
	n, err := c.udp.Read(buffer)
	if err != nil {
		return 0, err
	}
	if n < 10 || buffer[3] != 1 {
		return 0, fmt.Errorf("unexpected socks udp header")
	}
	return n - 10, nil
}

func TestRateLimitE2ESmoke(t *testing.T) {
	for _, tc := range []struct {
		name string
		up   int
		down int
	}{
		{name: "unlimited-0-0"},
		{name: "limited-2-8", up: 2, down: 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core, proxy1, proxy2 := startE2ECore(t)
			core.SetUserRates(map[string]UserRate{"alice": {UpMbps: tc.up, DownMbps: tc.down}})
			tcpSink := startTCPSink(t)
			defer tcpSink.Close()
			tcpSource := startTCPSource(t)
			defer tcpSource.Close()
			udpSink := startUDPSink(t)
			defer udpSink.Close()
			udpSource := startUDPSource(t)
			defer udpSource.Close()

			up, err := socksDial(proxy1, e2eProxyUser, e2eProxyPass, tcpSink.Addr())
			if err != nil {
				t.Fatal(err)
			}
			defer up.Close()
			down, err := socksDial(proxy2, e2eProxyUser, e2eProxyPass, tcpSource.Addr())
			if err != nil {
				t.Fatal(err)
			}
			defer down.Close()

			go func() {
				buffer := make([]byte, 8*1024)
				for i := 0; i < 100; i++ {
					if _, err := up.Write(buffer); err != nil {
						return
					}
				}
			}()
			go func() {
				buffer := make([]byte, 8*1024)
				for i := 0; i < 100; i++ {
					if _, err := down.Read(buffer); err != nil {
						return
					}
				}
			}()

			udpUp, err := newSocksUDP(proxy1, e2eProxyUser, e2eProxyPass)
			if err != nil {
				t.Fatal(err)
			}
			defer udpUp.Close()
			udpDown, err := newSocksUDP(proxy2, e2eProxyUser, e2eProxyPass)
			if err != nil {
				t.Fatal(err)
			}
			defer udpDown.Close()

			payload := bytes.Repeat([]byte("x"), 1200)
			for i := 0; i < 100; i++ {
				_ = udpUp.Send(udpSink.Addr(), payload)
			}
			if err = udpDown.Send(udpSource.Addr(), []byte("start")); err != nil {
				t.Fatal(err)
			}

			var udpDownloaded atomic.Int64
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				buffer := make([]byte, 2048)
				n, err := udpDown.Receive(buffer)
				if err != nil {
					t.Fatal(err)
				}
				udpDownloaded.Add(int64(n))
				if tcpSink.bytes.Load() > 0 && tcpSource.bytes.Load() > 0 && udpSink.bytes.Load() > 0 && udpDownloaded.Load() > 0 {
					break
				}
			}
			if tcpSink.bytes.Load() == 0 {
				t.Fatal("tcp upload did not reach the sink")
			}
			if tcpSource.bytes.Load() == 0 {
				t.Fatal("tcp download source did not send")
			}
			if udpSink.bytes.Load() == 0 {
				t.Fatal("udp upload did not reach the sink")
			}
			if udpDownloaded.Load() == 0 {
				t.Fatal("udp download did not reach the client")
			}

			// A closed target must be reported as a SOCKS failure, not silently
			// dropped or left to time out.
			start := time.Now()
			failed, err := socksDial(proxy1, e2eProxyUser, e2eProxyPass, "127.0.0.1:1")
			if err == nil {
				failed.Close()
				t.Fatal("unreachable target was accepted")
			}
			if isTimeout(err) {
				t.Fatalf("handshake failure timed out: %v", err)
			}
			if elapsed := time.Since(start); elapsed > 5*time.Second {
				t.Fatalf("handshake failure took %v", elapsed)
			}
		})
	}
}

func waitForUserSession(t *testing.T, core *Core, user string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, session := range core.GetInstance().SessionTracker().Sessions() {
			if session.User == user {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no tracked session for user %q", user)
}

func TestRateLimitE2EThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("long rate-limit E2E test")
	}
	core, proxy1, proxy2 := startE2ECore(t)
	core.SetUserRates(map[string]UserRate{"alice": {UpMbps: 2, DownMbps: 8}})

	tcpSink := startTCPSink(t)
	defer tcpSink.Close()
	tcpSource := startTCPSource(t)
	defer tcpSource.Close()
	udpSink := startUDPSink(t)
	defer udpSink.Close()
	udpSource := startUDPSource(t)
	defer udpSource.Close()

	// Upload phase: both inbounds at once, TCP and UDP on each one.
	var uploadTCP []net.Conn
	var uploadUDP []*socksUDPClient
	for _, proxy := range []string{proxy1, proxy2} {
		tcpConn, err := socksDial(proxy, "alice", "secret", tcpSink.Addr())
		if err != nil {
			t.Fatal(err)
		}
		uploadTCP = append(uploadTCP, tcpConn)
		udpConn, err := newSocksUDP(proxy, "alice", "secret")
		if err != nil {
			t.Fatal(err)
		}
		uploadUDP = append(uploadUDP, udpConn)
	}
	uploadDone := make(chan struct{})
	var uploadWG sync.WaitGroup
	for _, conn := range uploadTCP {
		uploadWG.Add(1)
		go func(conn net.Conn) {
			defer uploadWG.Done()
			buffer := make([]byte, 32*1024)
			for {
				select {
				case <-uploadDone:
					return
				default:
				}
				if _, err := conn.Write(buffer); err != nil {
					return
				}
			}
		}(conn)
	}
	for _, conn := range uploadUDP {
		uploadWG.Add(1)
		go func(conn *socksUDPClient) {
			defer uploadWG.Done()
			payload := bytes.Repeat([]byte("u"), 1200)
			for {
				select {
				case <-uploadDone:
					return
				default:
				}
				if err := conn.Send(udpSink.Addr(), payload); err != nil {
					return
				}
			}
		}(conn)
	}
	waitForUserSession(t, core, "alice")
	time.Sleep(5 * time.Second)
	upTCPStart, upUDPStart := tcpSink.bytes.Load(), udpSink.bytes.Load()
	time.Sleep(60 * time.Second)
	upTCPBytes := tcpSink.bytes.Load() - upTCPStart
	upUDPBytes := udpSink.bytes.Load() - upUDPStart
	close(uploadDone)
	for _, conn := range uploadTCP {
		_ = conn.Close()
	}
	for _, conn := range uploadUDP {
		conn.Close()
	}
	uploadWG.Wait()
	uploadMbps := float64(upTCPBytes+upUDPBytes) * 8 / 60 / 1e6
	t.Logf("upload phase (two inbounds, TCP+UDP): %.3f Mbps (tcp=%d udp=%d)", uploadMbps, upTCPBytes, upUDPBytes)
	if upTCPBytes == 0 || upUDPBytes == 0 {
		t.Fatalf("expected both TCP and UDP upload traffic (tcp=%d udp=%d)", upTCPBytes, upUDPBytes)
	}
	if uploadMbps < 1.6 || uploadMbps > 2.4 {
		t.Fatalf("upload outside tolerance: %.3f Mbps", uploadMbps)
	}

	// Download phase: both inbounds at once, TCP and UDP on each one.
	var downloadTCP []net.Conn
	var downloadUDP []*socksUDPClient
	for _, proxy := range []string{proxy1, proxy2} {
		tcpConn, err := socksDial(proxy, "alice", "secret", tcpSource.Addr())
		if err != nil {
			t.Fatal(err)
		}
		downloadTCP = append(downloadTCP, tcpConn)
		udpConn, err := newSocksUDP(proxy, "alice", "secret")
		if err != nil {
			t.Fatal(err)
		}
		downloadUDP = append(downloadUDP, udpConn)
	}
	downloadDone := make(chan struct{})
	var downloadWG sync.WaitGroup
	var tcpDownloaded, udpDownloaded atomic.Int64
	for _, conn := range downloadTCP {
		downloadWG.Add(1)
		go func(conn net.Conn) {
			defer downloadWG.Done()
			buffer := make([]byte, 32*1024)
			for {
				n, err := conn.Read(buffer)
				if n > 0 {
					tcpDownloaded.Add(int64(n))
				}
				if err != nil {
					return
				}
			}
		}(conn)
	}
	for _, conn := range downloadUDP {
		downloadWG.Add(1)
		go func(conn *socksUDPClient) {
			defer downloadWG.Done()
			buffer := make([]byte, 2048)
			for {
				select {
				case <-downloadDone:
					return
				default:
				}
				n, err := conn.Receive(buffer)
				if n > 0 {
					udpDownloaded.Add(int64(n))
				}
				if err != nil {
					return
				}
			}
		}(conn)
	}
	for _, conn := range downloadUDP {
		if err := conn.Send(udpSource.Addr(), []byte("start")); err != nil {
			t.Fatal(err)
		}
	}
	waitForUserSession(t, core, "alice")
	time.Sleep(5 * time.Second)
	downTCPStart, downUDPStart := tcpDownloaded.Load(), udpDownloaded.Load()
	time.Sleep(60 * time.Second)
	downTCPBytes := tcpDownloaded.Load() - downTCPStart
	downUDPBytes := udpDownloaded.Load() - downUDPStart
	close(downloadDone)
	for _, conn := range downloadTCP {
		_ = conn.Close()
	}
	for _, conn := range downloadUDP {
		conn.Close()
	}
	downloadWG.Wait()
	downloadMbps := float64(downTCPBytes+downUDPBytes) * 8 / 60 / 1e6
	t.Logf("download phase (two inbounds, TCP+UDP): %.3f Mbps (tcp=%d udp=%d)", downloadMbps, downTCPBytes, downUDPBytes)
	if downTCPBytes == 0 || downUDPBytes == 0 {
		t.Fatalf("expected both TCP and UDP download traffic (tcp=%d udp=%d)", downTCPBytes, downUDPBytes)
	}
	if downloadMbps < 6.4 || downloadMbps > 9.6 {
		t.Fatalf("download outside tolerance: %.3f Mbps", downloadMbps)
	}

	stats := *core.GetInstance().SessionTracker().GetStats()
	var userTraffic int64
	for _, stat := range stats {
		if stat.Resource == "user" && stat.Tag == "alice" {
			userTraffic += stat.Traffic
		}
	}
	if userTraffic <= 0 {
		t.Fatalf("expected alice traffic in stats, got %d", userTraffic)
	}
	t.Logf("alice traffic in tracker stats across both phases: %d bytes", userTraffic)
}

func TestRateLimitE2EIndependentUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("long rate-limit E2E test")
	}
	core, proxy1, proxy2 := startE2ECore(t)
	core.SetUserRates(map[string]UserRate{
		"alice": {UpMbps: 2, DownMbps: 8},
		"bob":   {UpMbps: 2, DownMbps: 8},
	})
	source := startTCPSource(t)
	defer source.Close()
	aConn, err := socksDial(proxy1, "alice", "secret", source.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer aConn.Close()
	bConn, err := socksDial(proxy2, "bob", "secret", source.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer bConn.Close()
	var aBytes, bBytes atomic.Int64
	go func() {
		buffer := make([]byte, 32*1024)
		for {
			n, err := aConn.Read(buffer)
			if n > 0 {
				aBytes.Add(int64(n))
			}
			if err != nil {
				return
			}
		}
	}()
	go func() {
		buffer := make([]byte, 32*1024)
		for {
			n, err := bConn.Read(buffer)
			if n > 0 {
				bBytes.Add(int64(n))
			}
			if err != nil {
				return
			}
		}
	}()
	waitForUserSession(t, core, "alice")
	waitForUserSession(t, core, "bob")
	time.Sleep(2 * time.Second)
	aStart, bStart := aBytes.Load(), bBytes.Load()
	time.Sleep(10 * time.Second)
	aRate := float64(aBytes.Load()-aStart) * 8 / 10 / 1e6
	bRate := float64(bBytes.Load()-bStart) * 8 / 10 / 1e6
	t.Logf("independent users: alice=%.3f Mbps bob=%.3f Mbps", aRate, bRate)
	if aRate < 6.4 || aRate > 9.6 {
		t.Fatalf("alice outside tolerance: %.3f Mbps", aRate)
	}
	if bRate < 6.4 || bRate > 9.6 {
		t.Fatalf("bob outside tolerance: %.3f Mbps", bRate)
	}
}

func TestRateLimitE2ELiveChange(t *testing.T) {
	if testing.Short() {
		t.Skip("long rate-limit E2E test")
	}
	core, proxy1, _ := startE2ECore(t)
	source := startTCPSource(t)
	defer source.Close()
	conn, err := socksDial(proxy1, "carol", "secret", source.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var received atomic.Int64
	go func() {
		buffer := make([]byte, 32*1024)
		for {
			n, err := conn.Read(buffer)
			if n > 0 {
				received.Add(int64(n))
			}
			if err != nil {
				return
			}
		}
	}()
	waitForUserSession(t, core, "carol")

	measure := func(label string, seconds int) float64 {
		time.Sleep(500 * time.Millisecond)
		start := received.Load()
		time.Sleep(time.Duration(seconds) * time.Second)
		rate := float64(received.Load()-start) * 8 / float64(seconds) / 1e6
		t.Logf("%s: %.3f Mbps", label, rate)
		return rate
	}

	// The connection started unlimited, so the first set is a 0 -> limit update.
	core.SetUserRates(map[string]UserRate{"carol": {DownMbps: 8}})
	fast := measure("rate=8Mbps", 4)
	core.SetUserRates(map[string]UserRate{"carol": {DownMbps: 2}})
	slow := measure("rate=2Mbps", 4)
	core.SetUserRates(map[string]UserRate{"carol": {DownMbps: 0}})
	unlimited := measure("rate=unlimited", 4)

	if fast < 6.4 || fast > 9.6 {
		t.Fatalf("8 Mbps phase outside tolerance: %.3f Mbps", fast)
	}
	if slow < 1.5 || slow > 2.5 {
		t.Fatalf("2 Mbps phase outside tolerance: %.3f Mbps", slow)
	}
	if unlimited < 20 {
		t.Fatalf("unlimited phase was still throttled: %.3f Mbps", unlimited)
	}
}

func TestRateLimitE2EMaxConnsAndHalfClose(t *testing.T) {
	core, proxy1, _ := startE2ECore(t)
	core.SetUserLimits(map[string]int{"alice": 1})
	echo := startTCPEcho(t)
	defer echo.Close()

	first, err := socksDial(proxy1, "alice", "secret", echo.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if _, err = first.Write([]byte("first")); err != nil {
		t.Fatal(err)
	}
	reply := make([]byte, 5)
	if _, err = io.ReadFull(first, reply); err != nil {
		t.Fatal(err)
	}
	if string(reply) != "first" {
		t.Fatalf("first connection echo = %q", reply)
	}

	second, dialErr := socksDial(proxy1, "alice", "secret", echo.Addr())
	if dialErr == nil {
		defer second.Close()
		_ = second.SetDeadline(time.Now().Add(2 * time.Second))
		_, _ = second.Write([]byte("second"))
		_, readErr := second.Read(make([]byte, 1))
		if readErr == nil {
			t.Fatal("second connection transferred data under maxConns=1")
		}
		if isTimeout(readErr) {
			t.Fatalf("second connection was not rejected, it timed out: %v", readErr)
		}
	} else if isTimeout(dialErr) {
		t.Fatalf("second connection dial timed out rather than being rejected: %v", dialErr)
	}

	// Closing the first session must release the slot for a third connection.
	first.Close()
	third, err := socksDial(proxy1, "alice", "secret", echo.Addr())
	if err != nil {
		t.Fatalf("slot was not released after closing the first connection: %v", err)
	}
	defer third.Close()
	if _, err = third.Write([]byte("third")); err != nil {
		t.Fatal(err)
	}
	if _, err = io.ReadFull(third, reply); err != nil {
		t.Fatal(err)
	}
	if string(reply) != "third" {
		t.Fatalf("third connection echo = %q", reply)
	}
	third.Close()
	core.SetUserLimits(map[string]int{"alice": 0})

	// Half-close: the server answers only after EOF.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = io.Copy(io.Discard, conn)
		_, _ = conn.Write([]byte("done"))
	}()
	half, err := socksDial(proxy1, "alice", "secret", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer half.Close()
	tcpConn, ok := half.(*net.TCPConn)
	if !ok {
		t.Fatal("socks connection is not TCP")
	}
	if _, err = half.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	if err = tcpConn.CloseWrite(); err != nil {
		t.Fatal(err)
	}
	_ = half.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, err = io.ReadFull(half, reply[:4]); err != nil {
		t.Fatal(err)
	}
	if string(reply[:4]) != "done" {
		t.Fatalf("unexpected half-close reply %q", reply[:4])
	}
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func TestRateLimitE2EUDPOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("long rate-limit E2E test")
	}
	core, proxy1, proxy2 := startE2ECore(t)
	core.SetUserRates(map[string]UserRate{"alice": {UpMbps: 2, DownMbps: 8}})
	sink := startUDPSink(t)
	defer sink.Close()
	source := startUDPSource(t)
	defer source.Close()
	up, err := newSocksUDP(proxy1, "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer up.Close()
	down, err := newSocksUDP(proxy2, "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer down.Close()
	if err = down.Send(source.Addr(), []byte("start")); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	var downloaded atomic.Int64
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		payload := bytes.Repeat([]byte("u"), 1200)
		for {
			select {
			case <-done:
				return
			default:
			}
			if err := up.Send(sink.Addr(), payload); err != nil {
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		buffer := make([]byte, 2048)
		for {
			n, err := down.Receive(buffer)
			if n > 0 {
				downloaded.Add(int64(n))
			}
			if err != nil {
				return
			}
		}
	}()
	waitForUserSession(t, core, "alice")
	time.Sleep(2 * time.Second)
	upStart, downStart := sink.bytes.Load(), downloaded.Load()
	time.Sleep(10 * time.Second)
	upBytes := sink.bytes.Load() - upStart
	downBytes := downloaded.Load() - downStart
	close(done)
	up.Close()
	down.Close()
	wg.Wait()
	upRate := float64(upBytes) * 8 / 10 / 1e6
	downRate := float64(downBytes) * 8 / 10 / 1e6
	t.Logf("udp only: upload=%.3f Mbps download=%.3f Mbps", upRate, downRate)
	if upRate < 1.5 || upRate > 2.5 {
		t.Fatalf("udp upload outside tolerance: %.3f Mbps", upRate)
	}
	if downRate < 6.0 || downRate > 10.0 {
		t.Fatalf("udp download outside tolerance: %.3f Mbps", downRate)
	}
}

func TestRateLimitE2EQueuedUnlimitedLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("long rate-limit E2E test")
	}
	core, proxy1, proxy2 := startE2ECore(t)
	core.SetUserRates(map[string]UserRate{"carol": {DownMbps: 1}})
	source := startTCPSource(t)
	defer source.Close()

	var conns []net.Conn
	var received atomic.Int64
	for i := 0; i < 4; i++ {
		proxy := proxy1
		if i%2 == 1 {
			proxy = proxy2
		}
		conn, err := socksDial(proxy, "carol", "secret", source.Addr())
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, conn)
		go func(conn net.Conn) {
			buffer := make([]byte, 32*1024)
			for {
				n, err := conn.Read(buffer)
				if n > 0 {
					received.Add(int64(n))
				}
				if err != nil {
					return
				}
			}
		}(conn)
	}
	defer func() {
		for _, conn := range conns {
			_ = conn.Close()
		}
	}()
	waitForUserSession(t, core, "carol")

	time.Sleep(2 * time.Second)
	queuedStart := received.Load()
	time.Sleep(1 * time.Second)
	queuedRate := float64(received.Load()-queuedStart) * 8 / 1e6
	t.Logf("queued rate before release: %.3f Mbps", queuedRate)
	if queuedRate < 0.5 || queuedRate > 1.5 {
		t.Fatalf("four queued connections were not at the 1 Mbps rate: %.3f Mbps", queuedRate)
	}

	start := time.Now()
	core.SetUserRates(map[string]UserRate{"carol": {DownMbps: 0}})
	var latency time.Duration
	lastBytes := received.Load()
	lastTime := time.Now()
	for time.Since(start) < 3*time.Second {
		time.Sleep(100 * time.Millisecond)
		nowBytes := received.Load()
		elapsed := time.Since(lastTime).Seconds()
		rate := float64(nowBytes-lastBytes) * 8 / elapsed / 1e6
		if rate > 20 {
			latency = time.Since(start)
			break
		}
		lastBytes = nowBytes
		lastTime = time.Now()
	}
	if latency == 0 {
		t.Fatal("queued connections did not observe the unlimited change within 3s")
	}
	if latency > 2*time.Second {
		t.Fatalf("unlimited change took %v to take effect", latency)
	}
	t.Logf("queued connections reached >20 Mbps %.1f ms after SetUserRates(0)", float64(latency.Microseconds())/1000)
}

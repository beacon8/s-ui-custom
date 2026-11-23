package core

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"golang.org/x/time/rate"
)

type recordingLimiter struct {
	burst int
	mu    sync.Mutex
	calls []int
	block bool
}

func (l *recordingLimiter) WaitN(ctx context.Context, n int) error {
	l.mu.Lock()
	l.calls = append(l.calls, n)
	block := l.block
	l.mu.Unlock()
	if block {
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}

func (l *recordingLimiter) Burst() int { return l.burst }

func (l *recordingLimiter) recorded() []int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]int(nil), l.calls...)
}

func testRateState(up, down rateWaiter) *userRate {
	state := &userRate{}
	if up != nil {
		state.up.Store(&limiterBox{limiter: up})
	}
	if down != nil {
		state.down.Store(&limiterBox{limiter: down})
	}
	return state
}

func TestWaitNChunksAtBurst(t *testing.T) {
	limiter := &recordingLimiter{burst: 10}
	if err := waitN(context.Background(), func() rateWaiter { return limiter }, 25); err != nil {
		t.Fatal(err)
	}
	got := limiter.recorded()
	want := []int{10, 10, 5}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestWaitNCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitN(ctx, func() rateWaiter { return &recordingLimiter{burst: 10, block: true} }, 5); err == nil {
		t.Fatal("expected the canceled context to stop the wait")
	}
}

func TestRateConnDirectionsAndBuffers(t *testing.T) {
	up := &recordingLimiter{burst: 4}
	down := &recordingLimiter{burst: 4}
	state := testRateState(up, down)
	client, server := net.Pipe()
	defer client.Close()
	conn := &rateConn{ExtendedConn: bufio.NewExtendedConn(server), ctx: context.Background(), rate: state}

	// Read is upload.
	go client.Write([]byte("hello"))
	readBuffer := make([]byte, 8)
	n, err := conn.Read(readBuffer)
	if err != nil || n != 5 {
		t.Fatalf("read: n=%d err=%v", n, err)
	}
	if got := up.recorded(); len(got) != 2 || got[0] != 4 || got[1] != 1 {
		t.Fatalf("upload calls = %v", got)
	}

	// Write is download.
	go io.ReadFull(client, make([]byte, 5))
	if _, err = conn.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if got := down.recorded(); len(got) != 2 || got[0] != 4 || got[1] != 1 {
		t.Fatalf("download calls = %v", got)
	}

	// Buffer-based copies take the same path.
	go client.Write([]byte("hello"))
	buffer := buf.NewSize(8)
	defer buffer.Release()
	if err = conn.ReadBuffer(buffer); err != nil {
		t.Fatal(err)
	}
	if got := up.recorded(); len(got) != 4 {
		t.Fatalf("upload calls after ReadBuffer = %v", got)
	}

	go io.ReadFull(client, make([]byte, 5))
	out := buf.As([]byte("hello"))
	if err = conn.WriteBuffer(out); err != nil {
		t.Fatal(err)
	}
	out.Release()
	if got := down.recorded(); len(got) != 4 {
		t.Fatalf("download calls after WriteBuffer = %v", got)
	}
}

type fakePacketConn struct {
	readData  []byte
	writeData []byte
}

func (c *fakePacketConn) ReadPacket(buffer *buf.Buffer) (M.Socksaddr, error) {
	_, _ = buffer.Write(c.readData)
	return M.Socksaddr{}, nil
}

func (c *fakePacketConn) WritePacket(buffer *buf.Buffer, destination M.Socksaddr) error {
	c.writeData = append(c.writeData, buffer.Bytes()...)
	return nil
}

func (c *fakePacketConn) Close() error                       { return nil }
func (c *fakePacketConn) LocalAddr() net.Addr                { return &net.UDPAddr{} }
func (c *fakePacketConn) SetDeadline(t time.Time) error      { return nil }
func (c *fakePacketConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *fakePacketConn) SetWriteDeadline(t time.Time) error { return nil }

func TestRatePacketConnKeepsPacketWhole(t *testing.T) {
	up := &recordingLimiter{burst: 4}
	down := &recordingLimiter{burst: 4}
	raw := &fakePacketConn{readData: []byte("hello")}
	conn := &ratePacketConn{PacketConn: raw, ctx: context.Background(), rate: testRateState(up, down)}

	buffer := buf.NewSize(8)
	if _, err := conn.ReadPacket(buffer); err != nil {
		t.Fatal(err)
	}
	if got := up.recorded(); len(got) != 2 || got[0] != 4 || got[1] != 1 {
		t.Fatalf("upload calls = %v", got)
	}
	buffer.Release()

	out := buf.As([]byte("hello"))
	if err := conn.WritePacket(out, M.Socksaddr{}); err != nil {
		t.Fatal(err)
	}
	out.Release()
	if got := down.recorded(); len(got) != 2 || got[0] != 4 || got[1] != 1 {
		t.Fatalf("download calls = %v", got)
	}
	if string(raw.writeData) != "hello" {
		t.Fatalf("packet was changed: %q", raw.writeData)
	}
}

func TestSetUserRatesHotUpdateAndSharedState(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.access.Lock()
	state := tracker.rateStateLocked("alice")
	tracker.access.Unlock()
	if state == nil || state.up.Load() != nil || state.down.Load() != nil {
		t.Fatal("expected an unlimited state before the first limit")
	}

	tracker.SetUserRates(map[string]UserRate{"alice": {UpMbps: 2, DownMbps: 8}})
	if state.up.Load() == nil || state.down.Load() == nil {
		t.Fatal("expected both limiters after setting rates")
	}

	tracker.access.Lock()
	shared := tracker.rateStateLocked("alice")
	tracker.access.Unlock()
	if shared != state {
		t.Fatal("a new session must share the existing bucket")
	}

	tracker.SetUserRates(map[string]UserRate{"alice": {UpMbps: 0, DownMbps: 0}})
	if state.up.Load() != nil || state.down.Load() != nil {
		t.Fatal("expected an explicit zero to make the user unlimited")
	}

	tracker.SetUserRates(map[string]UserRate{"alice": {UpMbps: 1, DownMbps: 0}})
	if state.up.Load() == nil || state.down.Load() != nil {
		t.Fatal("expected one-direction rates to leave the other direction unlimited")
	}
}

func TestRateConnShareOneBucketAcrossSessions(t *testing.T) {
	up := &recordingLimiter{burst: 64}
	down := &recordingLimiter{burst: 64}
	state := testRateState(up, down)
	clientA, serverA := net.Pipe()
	defer clientA.Close()
	clientB, serverB := net.Pipe()
	defer clientB.Close()
	a := &rateConn{ExtendedConn: bufio.NewExtendedConn(serverA), ctx: context.Background(), rate: state}
	b := &rateConn{ExtendedConn: bufio.NewExtendedConn(serverB), ctx: context.Background(), rate: state}

	go clientA.Write([]byte("aa"))
	go clientB.Write([]byte("bb"))
	if _, err := a.Read(make([]byte, 2)); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Read(make([]byte, 2)); err != nil {
		t.Fatal(err)
	}
	got := up.recorded()
	if len(got) != 2 || got[0] != 2 || got[1] != 2 {
		t.Fatalf("shared bucket calls = %v", got)
	}
}

func TestNewRateLimiterBoundary(t *testing.T) {
	box, err := newRateLimiter(17179)
	if err != nil || box == nil {
		t.Fatalf("17179 Mbps: box=%v err=%v", box, err)
	}
	if _, err = newRateLimiter(17180); err == nil {
		t.Fatal("17180 Mbps must return a range error")
	}
	if box.limiter.Burst() < 16*1024 || box.limiter.Burst() > 256*1024 {
		t.Fatalf("unexpected burst %d", box.limiter.Burst())
	}
}

func TestRateLimiterInterruptsQueuedWait(t *testing.T) {
	raw := rate.NewLimiter(1, 1)
	if !raw.AllowN(time.Now(), 1) {
		t.Fatal("could not consume the initial token")
	}
	changed := make(chan struct{})
	limiter := &rateLimiter{limiter: raw, changed: changed}
	errCh := make(chan error, 1)
	go func() {
		errCh <- limiter.WaitN(context.Background(), 1)
	}()
	time.Sleep(50 * time.Millisecond)
	start := time.Now()
	close(changed)
	select {
	case err := <-errCh:
		if !errors.Is(err, errRateChanged) {
			t.Fatalf("queued wait error = %v, want errRateChanged", err)
		}
		if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
			t.Fatalf("queued wait took %v to observe the change", elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("queued wait did not observe the rate change")
	}
}

func TestSetUserRatesOnlyChangesModifiedDirection(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.SetUserRates(map[string]UserRate{"a": {UpMbps: 2, DownMbps: 8}})

	tracker.access.Lock()
	state := tracker.userRates["a"]
	tracker.access.Unlock()
	up := state.up.Load()
	down := state.down.Load()
	if up == nil || down == nil {
		t.Fatal("expected both directions to be limited")
	}

	// Saving another user must not refill A's burst.
	tracker.SetUserRates(map[string]UserRate{
		"a": {UpMbps: 2, DownMbps: 8},
		"b": {UpMbps: 1, DownMbps: 1},
	})
	if state.up.Load() != up || state.down.Load() != down {
		t.Fatal("saving B replaced A's unchanged limiters")
	}

	// Changing only download keeps the upload bucket and its tokens.
	tracker.SetUserRates(map[string]UserRate{
		"a": {UpMbps: 2, DownMbps: 4},
		"b": {UpMbps: 1, DownMbps: 1},
	})
	if state.up.Load() != up {
		t.Fatal("download-only change reset the upload bucket")
	}
	newDown := state.down.Load()
	if newDown == down || newDown == nil {
		t.Fatal("download change did not replace the download bucket")
	}

	// Clearing upload to 0 leaves download untouched.
	tracker.SetUserRates(map[string]UserRate{
		"a": {UpMbps: 0, DownMbps: 4},
		"b": {UpMbps: 1, DownMbps: 1},
	})
	if state.up.Load() != nil {
		t.Fatal("upMbps=0 did not clear the upload bucket")
	}
	if state.down.Load() != newDown {
		t.Fatal("upload change reset the download bucket")
	}
}

type handshakeTestConn struct {
	net.Conn
	success     atomic.Int32
	failure     atomic.Int32
	connSuccess atomic.Int32
}

func (c *handshakeTestConn) HandshakeSuccess() error {
	c.success.Add(1)
	return nil
}

func (c *handshakeTestConn) HandshakeFailure(error) error {
	c.failure.Add(1)
	return nil
}

func (c *handshakeTestConn) ConnHandshakeSuccess(net.Conn) error {
	c.connSuccess.Add(1)
	return nil
}

func TestRateConnForwardsHandshakeCallbacks(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	inner := &handshakeTestConn{Conn: server}
	limiter := &recordingLimiter{burst: 64}
	conn := &rateConn{
		ExtendedConn: bufio.NewExtendedConn(inner),
		ctx:          context.Background(),
		rate:         testRateState(limiter, limiter),
	}

	if err := N.ReportConnHandshakeSuccess(conn, client); err != nil {
		t.Fatal(err)
	}
	// ReportHandshakeFailure intentionally returns the original error; only the
	// callback delivery is under test here.
	_ = N.ReportHandshakeFailure(conn, errors.New("dial failed"))
	if got := inner.connSuccess.Load(); got != 1 {
		t.Fatalf("ConnHandshakeSuccess calls = %d, want 1", got)
	}
	if got := inner.failure.Load(); got != 1 {
		t.Fatalf("HandshakeFailure calls = %d, want 1", got)
	}
}

func TestRateConnStillLimitsAfterHandshake(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	inner := &handshakeTestConn{Conn: server}
	limiter := &recordingLimiter{burst: 8}
	conn := &rateConn{
		ExtendedConn: bufio.NewExtendedConn(inner),
		ctx:          context.Background(),
		rate:         testRateState(limiter, limiter),
	}
	if err := N.ReportConnHandshakeSuccess(conn, client); err != nil {
		t.Fatal(err)
	}
	go io.ReadFull(client, make([]byte, 4))
	if _, err := conn.Write([]byte("data")); err != nil {
		t.Fatal(err)
	}
	if got := limiter.recorded(); len(got) != 1 || got[0] != 4 {
		t.Fatalf("limiter calls after handshake = %v", got)
	}
}

func TestSavingAnotherUserDoesNotRefillBurst(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.SetUserRates(map[string]UserRate{"a": {UpMbps: 2, DownMbps: 8}})
	tracker.access.Lock()
	state := tracker.userRates["a"]
	tracker.access.Unlock()
	box := state.up.Load()
	if box == nil {
		t.Fatal("missing upload bucket")
	}
	raw := box.limiter.(*rateLimiter).limiter
	now := time.Now()
	if !raw.AllowN(now, raw.Burst()) {
		t.Fatal("could not drain the upload bucket")
	}
	before := raw.TokensAt(now)

	tracker.SetUserRates(map[string]UserRate{
		"a": {UpMbps: 2, DownMbps: 8},
		"b": {UpMbps: 1, DownMbps: 1},
	})
	if state.up.Load() != box {
		t.Fatal("saving B replaced A's upload bucket")
	}
	tokensAfter := raw.TokensAt(time.Now())
	if tokensAfter > before+1024 {
		t.Fatalf("A's upload burst was refilled: before=%f after=%f", before, tokensAfter)
	}
}

package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/util/rateconv"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-tun"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"golang.org/x/time/rate"
)

type Counter struct {
	read  *atomic.Int64
	write *atomic.Int64
}

// trackerInstanceSeq distinguishes trackers across core restarts. The panel
// uses it to drop a stale speed baseline instead of comparing counters from
// two different boxes.
var trackerInstanceSeq atomic.Uint64

// userTrafficTotal is traffic already drained by GetStats. Live counters are
// added on top when a speed snapshot is taken, so bytes recorded between two
// panel polls never disappear when the 10s stats job runs.
type userTrafficTotal struct {
	up   int64
	down int64
}

// UserTraffic is one user's cumulative bytes in a speed snapshot. Up is what
// the client sent into the proxy, Down is what the proxy sent back.
type UserTraffic struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

// UserSpeedSnapshot is a read-only, non-destructive view of per-user byte
// counters. SampledAt is Unix milliseconds; the panel divides byte deltas by
// the actual interval between two snapshots.
type UserSpeedSnapshot struct {
	Running   bool                   `json:"running"`
	Instance  string                 `json:"instance"`
	SampledAt int64                  `json:"sampledAt"`
	Users     map[string]UserTraffic `json:"users"`
}

// UserRate is a client's bandwidth cap in Mbps. Zero means unlimited.
type UserRate struct {
	UpMbps   int `json:"upMbps"`
	DownMbps int `json:"downMbps"`
}

// rateWaiter is the part of rate.Limiter the forwarding path uses. Tests can
// substitute a fake without waiting on wall-clock time.
type rateWaiter interface {
	WaitN(ctx context.Context, n int) error
	Burst() int
}

type limiterBox struct {
	limiter rateWaiter
	changed chan struct{}
}

// rateLimiter turns rate.Limiter's blocking wait into an interruptible one.
// When a user's rate changes the old box's changed channel is closed, so a
// queued read/write cancels its reservation and retries against the new box
// instead of sleeping out the old delay.
type rateLimiter struct {
	limiter *rate.Limiter
	changed <-chan struct{}
}

func (l *rateLimiter) Burst() int {
	return l.limiter.Burst()
}

func (l *rateLimiter) WaitN(ctx context.Context, n int) error {
	now := time.Now()
	reservation := l.limiter.ReserveN(now, n)
	if !reservation.OK() {
		return errors.New("rate limiter reservation is not possible")
	}
	delay := reservation.DelayFrom(now)
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		reservation.Cancel()
		return ctx.Err()
	case <-l.changed:
		reservation.Cancel()
		return errRateChanged
	}
}

// userRate is shared by every session of one client. The pointers are
// replaced atomically, so changing a rate or switching between limited and
// unlimited reaches connections that are already open.
type userRate struct {
	up   atomic.Pointer[limiterBox]
	down atomic.Pointer[limiterBox]
	// caps are owned by SessionTracker.access. They let SetUserRates tell an
	// unchanged direction from a changed one, so saving another client does
	// not refill this client's burst.
	upMbps   int
	downMbps int
}

var errRateChanged = errors.New("rate limit changed")

const (
	// A single forwarding call is capped so one read/write cannot ask for more
	// tokens than a bucket holds. Packet payloads are kept whole and their
	// tokens are requested in several chunks instead.
	maxRateChunk = 32 * 1024
)

// Session is one routed connection. The same counters it exposes are also
// wired into the aggregate maps below, so draining stats never has to walk the
// session list and traffic of a session that closed mid-cycle is not lost.
type Session struct {
	ID       uint64
	Inbound  string
	User     string
	Outbound string
	Network  string
	Source   M.Socksaddr
	// Client is the verified client address when a trusted relay supplied
	// one; otherwise it is Source.
	Client      M.Socksaddr
	Destination M.Socksaddr
	// Domain is the name the connection is really for: what the client asked
	// for, or what sniffing found when the client only sent an address.
	Domain    string
	Rule      string
	CreatedAt time.Time
	Upload    *atomic.Int64
	Download  *atomic.Int64

	closer io.Closer
	rate   *userRate
	ipKey  netip.Prefix
}

// SessionInfo is the snapshot handed to the API; the live atomics stay inside.
type SessionInfo struct {
	ID          string `json:"id"`
	Inbound     string `json:"inbound,omitempty"`
	User        string `json:"user,omitempty"`
	Outbound    string `json:"outbound,omitempty"`
	Network     string `json:"network,omitempty"`
	Source      string `json:"source,omitempty"`
	Client      string `json:"client,omitempty"`
	Destination string `json:"destination,omitempty"`
	Domain      string `json:"domain,omitempty"`
	Rule        string `json:"rule,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
	Upload      int64  `json:"up"`
	Download    int64  `json:"down"`
}

// userIPState is one distinct client IP of a user. sessions is the number of
// live routed connections using it; since is when the first of them arrived.
type userIPState struct {
	sessions int
	since    time.Time
}

// UserIPInfo is one entry of the per-user IP list.
type UserIPInfo struct {
	IP       string `json:"ip"`
	Sessions int    `json:"sessions"`
	Since    int64  `json:"since"`
}

// UserIPSummary is the response of the per-user IP endpoint.
type UserIPSummary struct {
	Count int          `json:"count"`
	Limit int          `json:"limit"`
	IPs   []UserIPInfo `json:"ips"`
}

var _ adapter.ConnectionTracker = (*SessionTracker)(nil)

// SessionTracker replaces the former StatsTracker/ConnTracker pair: one wrap
// per connection that both counts traffic and keeps the session closable, so a
// user who was just removed from an inbound can actually be cut off.
type SessionTracker struct {
	access     sync.Mutex
	inbounds   map[string]Counter
	outbounds  map[string]Counter
	users      map[string]Counter
	sessions   map[uint64]*Session
	userLimits map[string]int
	userConns  map[string]int
	userRates  map[string]*userRate

	userIPLimits  map[string]int
	userIPs       map[string]map[netip.Prefix]*userIPState
	ipv4PrefixLen int
	ipv6PrefixLen int
	ipLimitDryRun bool

	instanceID string
	userTotals map[string]*userTrafficTotal
	// nextID only has to be unique among live sessions, so it is a counter
	// rather than a UUID: this runs on every connection.
	nextID uint64
}

func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		inbounds:      make(map[string]Counter),
		outbounds:     make(map[string]Counter),
		users:         make(map[string]Counter),
		sessions:      make(map[uint64]*Session),
		userLimits:    make(map[string]int),
		userConns:     make(map[string]int),
		userRates:     make(map[string]*userRate),
		userIPLimits:  make(map[string]int),
		userIPs:       make(map[string]map[netip.Prefix]*userIPState),
		ipv4PrefixLen: 32,
		ipv6PrefixLen: 64,
		instanceID:    fmt.Sprintf("%d-%d", time.Now().UnixNano(), trackerInstanceSeq.Add(1)),
		userTotals:    make(map[string]*userTrafficTotal),
	}
}

// SetUserIPLimits replaces the per-user distinct-IP limits. A limit <= 0
// means unlimited. Lowering a limit never evicts an IP that is already online:
// only a new IP is refused.
func (t *SessionTracker) SetUserIPLimits(limits map[string]int) {
	copied := make(map[string]int, len(limits))
	for user, limit := range limits {
		if user != "" && limit > 0 {
			copied[user] = limit
		}
	}
	t.access.Lock()
	t.userIPLimits = copied
	t.access.Unlock()
}

// SetIPPrefixes sets how addresses are collapsed into one IP. IPv4 defaults to
// /32 and IPv6 to /64 (one subscriber), so a rotating address inside the same
// IPv6 prefix is not counted as a new IP.
func (t *SessionTracker) SetIPPrefixes(ipv4PrefixLen, ipv6PrefixLen int) {
	t.access.Lock()
	defer t.access.Unlock()
	if ipv4PrefixLen > 0 && ipv4PrefixLen <= 32 {
		t.ipv4PrefixLen = ipv4PrefixLen
	}
	if ipv6PrefixLen > 0 && ipv6PrefixLen <= 128 {
		t.ipv6PrefixLen = ipv6PrefixLen
	}
}

// SetIPLimitDryRun only logs a would-be rejection; the connection is admitted
// and the IP is counted.
func (t *SessionTracker) SetIPLimitDryRun(enabled bool) {
	t.access.Lock()
	t.ipLimitDryRun = enabled
	t.access.Unlock()
}

// ActiveIPCounts returns the number of distinct IPs per user with a live
// session, for the client list.
func (t *SessionTracker) ActiveIPCounts() map[string]int {
	t.access.Lock()
	defer t.access.Unlock()
	counts := make(map[string]int, len(t.userIPs))
	for user, ips := range t.userIPs {
		if len(ips) > 0 {
			counts[user] = len(ips)
		}
	}
	return counts
}

// UserIPs lists the distinct IPs of one user, newest first.
func (t *SessionTracker) UserIPs(user string) UserIPSummary {
	t.access.Lock()
	summary := UserIPSummary{Limit: t.userIPLimits[user], IPs: []UserIPInfo{}}
	for prefix, state := range t.userIPs[user] {
		summary.IPs = append(summary.IPs, UserIPInfo{
			IP:       prefix.String(),
			Sessions: state.sessions,
			Since:    state.since.Unix(),
		})
	}
	t.access.Unlock()
	summary.Count = len(summary.IPs)
	sort.Slice(summary.IPs, func(i, j int) bool {
		return summary.IPs[i].Since > summary.IPs[j].Since
	})
	return summary
}

// SetUserLimits replaces the per-user concurrent connection limits.
// A limit <= 0 means unlimited.
func (t *SessionTracker) SetUserLimits(limits map[string]int) {
	t.access.Lock()
	defer t.access.Unlock()
	next := make(map[string]int, len(limits))
	for user, limit := range limits {
		if user != "" && limit > 0 {
			next[user] = limit
		}
	}
	t.userLimits = next
}

// ActiveConns returns the current number of active tracked sessions for a user.
func (t *SessionTracker) ActiveConns(user string) int {
	t.access.Lock()
	defer t.access.Unlock()
	return t.userConns[user]
}

// ActiveConnections returns a snapshot of the active tracked sessions per user.
// The map is a copy: callers can read it after the tracker moves on.
func (t *SessionTracker) ActiveConnections() map[string]int {
	t.access.Lock()
	defer t.access.Unlock()
	active := make(map[string]int, len(t.userConns))
	for user, count := range t.userConns {
		active[user] = count
	}
	return active
}

// SetUserRates replaces the per-user bandwidth caps. It updates the shared
// states in place: existing connections pick the new limits up on their next
// read or write, and a user missing from the map becomes unlimited.
func (t *SessionTracker) SetUserRates(rates map[string]UserRate) {
	t.access.Lock()
	defer t.access.Unlock()
	// Existing users first: a user missing from the map becomes unlimited,
	// while an unchanged user keeps both limiter objects (and their tokens).
	for user, state := range t.userRates {
		t.applyUserRateLocked(state, rates[user])
	}
	for user, caps := range rates {
		if user == "" {
			continue
		}
		state := t.userRates[user]
		if state == nil {
			state = &userRate{}
			t.userRates[user] = state
		}
		t.applyUserRateLocked(state, caps)
	}
}

func (t *SessionTracker) applyUserRateLocked(state *userRate, caps UserRate) {
	if caps.UpMbps < 0 {
		caps.UpMbps = 0
	}
	if caps.DownMbps < 0 {
		caps.DownMbps = 0
	}
	if state.upMbps != caps.UpMbps {
		next, err := newRateLimiter(caps.UpMbps)
		if err == nil {
			old := state.up.Swap(next)
			if old != nil {
				close(old.changed)
			}
			state.upMbps = caps.UpMbps
		}
	}
	if state.downMbps != caps.DownMbps {
		next, err := newRateLimiter(caps.DownMbps)
		if err == nil {
			old := state.down.Swap(next)
			if old != nil {
				close(old.changed)
			}
			state.downMbps = caps.DownMbps
		}
	}
}

// rateStateLocked returns the shared rate state for a user, creating an
// unlimited one when needed. Keeping an entry for every authenticated user is
// what lets a later 0 -> limit change reach connections that started
// unlimited. The caller must hold access.
func (t *SessionTracker) rateStateLocked(user string) *userRate {
	if user == "" {
		return nil
	}
	state := t.userRates[user]
	if state == nil {
		state = &userRate{}
		t.userRates[user] = state
	}
	return state
}

func newRateLimiter(mbps int) (*limiterBox, error) {
	if mbps <= 0 {
		return nil, nil
	}
	bytesPerSecond, err := rateconv.BytesPerSecond(mbps)
	if err != nil {
		return nil, err
	}
	burst := bytesPerSecond / 8
	if burst < 16*1024 {
		burst = 16 * 1024
	}
	if burst > 256*1024 {
		burst = 256 * 1024
	}
	changed := make(chan struct{})
	return &limiterBox{
		limiter: &rateLimiter{
			limiter: rate.NewLimiter(rate.Limit(bytesPerSecond), int(burst)),
			changed: changed,
		},
		changed: changed,
	}, nil
}

// waitN requests n tokens in chunks no larger than the bucket's burst. A UDP
// packet larger than the burst is therefore still sent whole, after several
// waits, rather than split.
func waitN(ctx context.Context, load func() rateWaiter, n int) error {
	for n > 0 {
		limiter := load()
		if limiter == nil {
			return nil
		}
		burst := limiter.Burst()
		if burst < 1 {
			burst = 1
		}
		chunk := n
		if chunk > burst {
			chunk = burst
		}
		if err := limiter.WaitN(ctx, chunk); err != nil {
			if errors.Is(err, errRateChanged) {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return ctxErr
				}
				continue
			}
			return err
		}
		n -= chunk
	}
	return nil
}

// rateConn sits between the raw inbound connection and the traffic counter, so
// every copy path (including buffer-based and spliced ones) passes through it.
type rateConn struct {
	N.ExtendedConn
	ctx  context.Context
	rate *userRate
}

func (c *rateConn) limiter(upload bool) rateWaiter {
	var box *limiterBox
	if upload {
		box = c.rate.up.Load()
	} else {
		box = c.rate.down.Load()
	}
	if box == nil {
		return nil
	}
	return box.limiter
}

func (c *rateConn) wait(upload bool, n int) error {
	if n <= 0 {
		return nil
	}
	return waitN(c.ctx, func() rateWaiter {
		return c.limiter(upload)
	}, n)
}

func (c *rateConn) Read(p []byte) (int, error) {
	if len(p) > maxRateChunk {
		p = p[:maxRateChunk]
	}
	n, err := c.ExtendedConn.Read(p)
	if n > 0 {
		if waitErr := c.wait(true, n); waitErr != nil {
			return n, waitErr
		}
	}
	return n, err
}

func (c *rateConn) ReadBuffer(buffer *buf.Buffer) error {
	if err := c.ExtendedConn.ReadBuffer(buffer); err != nil {
		return err
	}
	if buffer.Len() > 0 {
		return c.wait(true, buffer.Len())
	}
	return nil
}

func (c *rateConn) Write(p []byte) (int, error) {
	total := 0
	for total < len(p) {
		chunk := p[total:]
		if len(chunk) > maxRateChunk {
			chunk = chunk[:maxRateChunk]
		}
		if err := c.wait(false, len(chunk)); err != nil {
			return total, err
		}
		n, err := c.ExtendedConn.Write(chunk)
		total += n
		if err != nil {
			return total, err
		}
		if n == 0 {
			return total, io.ErrShortWrite
		}
	}
	return total, nil
}

func (c *rateConn) WriteBuffer(buffer *buf.Buffer) error {
	if buffer.Len() > 0 {
		if err := c.wait(false, buffer.Len()); err != nil {
			return err
		}
	}
	return c.ExtendedConn.WriteBuffer(buffer)
}

// Forward the optional framing hints so buffer allocation still reserves the
// headroom a protocol needs. Calculate* walks the wrapped chain without
// exposing Upstream() here, so the limiter stays in the unwrap chain.
func (c *rateConn) FrontHeadroom() int {
	return N.CalculateFrontHeadroom(c.ExtendedConn)
}

func (c *rateConn) RearHeadroom() int {
	return N.CalculateRearHeadroom(c.ExtendedConn)
}

// Forward the handshake callbacks through the limiter instead of exposing
// Upstream(). common.Cast stops at rateConn, so the callback is delivered and
// the data copy still has to pass through the limiter below it.
func (c *rateConn) HandshakeSuccess() error {
	return N.ReportHandshakeSuccess(c.ExtendedConn)
}

func (c *rateConn) HandshakeFailure(err error) error {
	return N.ReportHandshakeFailure(c.ExtendedConn, err)
}

func (c *rateConn) ConnHandshakeSuccess(conn net.Conn) error {
	return N.ReportConnHandshakeSuccess(c.ExtendedConn, conn)
}

// ratePacketConn mirrors rateConn for UDP: the payload stays in one packet,
// only the token request is split across the bucket's burst.
type ratePacketConn struct {
	N.PacketConn
	ctx  context.Context
	rate *userRate
}

func (c *ratePacketConn) wait(upload bool, n int) error {
	if n <= 0 {
		return nil
	}
	return waitN(c.ctx, func() rateWaiter {
		var box *limiterBox
		if upload {
			box = c.rate.up.Load()
		} else {
			box = c.rate.down.Load()
		}
		if box == nil {
			return nil
		}
		return box.limiter
	}, n)
}

func (c *ratePacketConn) ReadPacket(buffer *buf.Buffer) (M.Socksaddr, error) {
	destination, err := c.PacketConn.ReadPacket(buffer)
	if err == nil && buffer.Len() > 0 {
		if waitErr := c.wait(true, buffer.Len()); waitErr != nil {
			return destination, waitErr
		}
	}
	return destination, err
}

func (c *ratePacketConn) WritePacket(buffer *buf.Buffer, destination M.Socksaddr) error {
	if buffer.Len() > 0 {
		if err := c.wait(false, buffer.Len()); err != nil {
			return err
		}
	}
	return c.PacketConn.WritePacket(buffer, destination)
}

func (c *ratePacketConn) FrontHeadroom() int {
	return N.CalculateFrontHeadroom(c.PacketConn)
}

func (c *ratePacketConn) RearHeadroom() int {
	return N.CalculateRearHeadroom(c.PacketConn)
}

func (c *ratePacketConn) HandshakeSuccess() error {
	return N.ReportHandshakeSuccess(c.PacketConn)
}

func (c *ratePacketConn) HandshakeFailure(err error) error {
	return N.ReportHandshakeFailure(c.PacketConn, err)
}

func (c *ratePacketConn) PacketConnHandshakeSuccess(conn net.PacketConn) error {
	return N.ReportPacketConnHandshakeSuccess(c.PacketConn, conn)
}

// ipKeyLocked collapses a client address into the prefix used for counting.
// The caller must hold access.
func (t *SessionTracker) ipKeyLocked(addr M.Socksaddr) netip.Prefix {
	if !addr.IsValid() || !addr.Addr.IsValid() {
		return netip.Prefix{}
	}
	ip := addr.Addr.Unmap()
	bits := 32
	prefixLen := t.ipv4PrefixLen
	if ip.Is6() {
		bits = 128
		prefixLen = t.ipv6PrefixLen
	}
	if prefixLen < 0 || prefixLen > bits {
		prefixLen = bits
	}
	return netip.PrefixFrom(ip, prefixLen).Masked()
}

func (t *SessionTracker) trackIPLocked(session *Session) {
	if session.User == "" || !session.ipKey.IsValid() {
		return
	}
	ips := t.userIPs[session.User]
	if ips == nil {
		ips = make(map[netip.Prefix]*userIPState)
		t.userIPs[session.User] = ips
	}
	state := ips[session.ipKey]
	if state == nil {
		state = &userIPState{since: time.Now()}
		ips[session.ipKey] = state
	}
	state.sessions++
}

func (t *SessionTracker) releaseIPLocked(session *Session) {
	if session.User == "" || !session.ipKey.IsValid() {
		return
	}
	ips := t.userIPs[session.User]
	if ips == nil {
		return
	}
	state := ips[session.ipKey]
	if state == nil {
		return
	}
	state.sessions--
	if state.sessions <= 0 {
		delete(ips, session.ipKey)
	}
	if len(ips) == 0 {
		delete(t.userIPs, session.User)
	}
}

func formatIPKeys(ips map[netip.Prefix]*userIPState) string {
	keys := make([]string, 0, len(ips))
	for prefix := range ips {
		keys = append(keys, prefix.String())
	}
	sort.Strings(keys)
	result := ""
	for i, key := range keys {
		if i > 0 {
			result += ", "
		}
		result += key
	}
	return result
}

func (t *SessionTracker) loadOrCreateCounter(obj map[string]Counter, name string) Counter {
	counter, loaded := obj[name]
	if loaded {
		return counter
	}
	counter = Counter{read: &atomic.Int64{}, write: &atomic.Int64{}}
	obj[name] = counter
	return counter
}

// newSession registers the session and returns the counter slices to hand to
// the counting conn: index 0 is the session's own counter, the rest are the
// aggregates GetStats drains.
func (t *SessionTracker) newSession(metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound, network string) (*Session, []*atomic.Int64, []*atomic.Int64) {
	// The router fills RouteRule and RouteOutbound in just before it calls the
	// trackers, so both are reused rather than formatted a second time: a rule's
	// String() is not free and this runs per connection.
	outbound := metadata.RouteOutbound
	if outbound == "" && matchOutbound != nil {
		outbound = matchOutbound.Tag()
	}
	client := metadata.ClientAddr
	if !client.IsValid() {
		client = metadata.Source
	}
	session := &Session{
		Inbound:     metadata.Inbound,
		User:        metadata.User,
		Outbound:    outbound,
		Network:     network,
		Source:      metadata.Source,
		Client:      client,
		Destination: metadata.Destination,
		CreatedAt:   time.Now(),
		Upload:      &atomic.Int64{},
		Download:    &atomic.Int64{},
	}
	// Same precedence as the clash API: what sniffing found wins, because the
	// destination is only a name when the client sent one.
	if metadata.Domain != "" {
		session.Domain = metadata.Domain
	} else {
		session.Domain = metadata.Destination.Fqdn
	}
	session.Rule = metadata.RouteRule
	if session.Rule == "" && matchedRule != nil {
		session.Rule = matchedRule.String()
	}

	readCounter := []*atomic.Int64{session.Upload}
	writeCounter := []*atomic.Int64{session.Download}

	t.access.Lock()
	defer t.access.Unlock()
	if session.User != "" {
		if limit := t.userLimits[session.User]; limit > 0 && t.userConns[session.User] >= limit {
			return nil, nil, nil
		}
		session.ipKey = t.ipKeyLocked(session.Client)
		if limit := t.userIPLimits[session.User]; limit > 0 && session.ipKey.IsValid() {
			ips := t.userIPs[session.User]
			if _, exists := ips[session.ipKey]; !exists && len(ips) >= limit {
				message := []any{"client ", session.User, " IP limit ", limit, ": refusing new IP ", session.ipKey}
				if len(ips) > 0 {
					message = append(message, ", online IPs: ", formatIPKeys(ips))
				}
				if t.ipLimitDryRun {
					logger.Warning(append(message, " (dry run)")...)
				} else {
					logger.Warning(message...)
					return nil, nil, nil
				}
			}
		}
	}
	t.nextID++
	session.ID = t.nextID
	if session.Inbound != "" {
		counter := t.loadOrCreateCounter(t.inbounds, session.Inbound)
		readCounter = append(readCounter, counter.read)
		writeCounter = append(writeCounter, counter.write)
	}
	if session.Outbound != "" {
		counter := t.loadOrCreateCounter(t.outbounds, session.Outbound)
		readCounter = append(readCounter, counter.read)
		writeCounter = append(writeCounter, counter.write)
	}
	if session.User != "" {
		counter := t.loadOrCreateCounter(t.users, session.User)
		readCounter = append(readCounter, counter.read)
		writeCounter = append(writeCounter, counter.write)
		t.userConns[session.User]++
		session.rate = t.rateStateLocked(session.User)
		t.trackIPLocked(session)
	}
	t.sessions[session.ID] = session
	return session, readCounter, writeCounter
}

func (t *SessionTracker) leave(session *Session) {
	t.access.Lock()
	defer t.access.Unlock()
	if _, ok := t.sessions[session.ID]; !ok {
		return
	}
	delete(t.sessions, session.ID)
	if session.User != "" {
		if n := t.userConns[session.User] - 1; n > 0 {
			t.userConns[session.User] = n
		} else {
			delete(t.userConns, session.User)
		}
		t.releaseIPLocked(session)
	}
}

func (t *SessionTracker) RoutedConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound) net.Conn {
	session, readCounter, writeCounter := t.newSession(metadata, matchedRule, matchOutbound, N.NetworkTCP)
	if session == nil {
		_ = conn.Close()
		return conn
	}
	var cancel context.CancelFunc
	extended := bufio.NewExtendedConn(conn)
	if session.rate != nil {
		var rateCtx context.Context
		rateCtx, cancel = context.WithCancel(context.Background())
		extended = &rateConn{ExtendedConn: extended, ctx: rateCtx, rate: session.rate}
	}
	// The limiter goes inside the counter so the copy/splice fast paths still
	// observe it; the counter itself is not replaceable, so unwrapping stops
	// there.
	tracked := &sessionConn{
		ExtendedConn: bufio.NewInt64CounterConn(extended, readCounter, writeCounter),
		tracker:      t,
		session:      session,
		cancel:       cancel,
	}
	session.closer = tracked
	return tracked
}

func (t *SessionTracker) RoutedPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound) N.PacketConn {
	session, readCounter, writeCounter := t.newSession(metadata, matchedRule, matchOutbound, N.NetworkUDP)
	if session == nil {
		_ = conn.Close()
		return conn
	}
	var cancel context.CancelFunc
	extended := conn
	if session.rate != nil {
		var rateCtx context.Context
		rateCtx, cancel = context.WithCancel(context.Background())
		extended = &ratePacketConn{PacketConn: conn, ctx: rateCtx, rate: session.rate}
	}
	tracked := &sessionPacketConn{
		PacketConn: bufio.NewInt64CounterPacketConn(extended, readCounter, nil, writeCounter, nil),
		tracker:    t,
		session:    session,
		cancel:     cancel,
	}
	session.closer = tracked
	return tracked
}

func (t *SessionTracker) RoutedFlow(ctx context.Context, metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound) tun.FlowTracker {
	return nil
}

// GetStats drains the aggregate counters. Destructive: every counter is reset,
// so the caller owns the returned traffic.
func (t *SessionTracker) GetStats() *[]model.Stats {
	t.access.Lock()
	defer t.access.Unlock()

	dt := time.Now().Unix()

	// Emit only directions that actually moved traffic; a zero-traffic row would
	// just bloat the stats table without changing any chart bucket.
	s := []model.Stats{}
	appendStat := func(resource, tag string, down, up int64) {
		if down > 0 {
			s = append(s, model.Stats{DateTime: dt, Resource: resource, Tag: tag, Direction: false, Traffic: down})
		}
		if up > 0 {
			s = append(s, model.Stats{DateTime: dt, Resource: resource, Tag: tag, Direction: true, Traffic: up})
		}
	}

	for inbound, counter := range t.inbounds {
		appendStat("inbound", inbound, counter.write.Swap(0), counter.read.Swap(0))
	}
	for outbound, counter := range t.outbounds {
		appendStat("outbound", outbound, counter.write.Swap(0), counter.read.Swap(0))
	}
	for user, counter := range t.users {
		up := counter.read.Swap(0)
		down := counter.write.Swap(0)
		total := t.userTotals[user]
		if total == nil {
			total = &userTrafficTotal{}
			t.userTotals[user] = total
		}
		total.up += up
		total.down += down
		appendStat("user", user, down, up)
	}
	return &s
}

// UserSpeedSnapshot returns cumulative per-user byte counters without draining
// them. GetStats is destructive, so speed polling must never call it; instead
// this reads drained totals plus the live counters under the same lock. That
// keeps short connections visible between two 10s accounting cycles and lets
// any number of panels read independently.
func (t *SessionTracker) UserSpeedSnapshot(users []string) UserSpeedSnapshot {
	t.access.Lock()
	defer t.access.Unlock()
	snapshot := UserSpeedSnapshot{
		Running:   true,
		Instance:  t.instanceID,
		SampledAt: time.Now().UnixMilli(),
		Users:     make(map[string]UserTraffic, len(users)),
	}
	seen := make(map[string]struct{}, len(users))
	addUser := func(user string) {
		if user == "" {
			return
		}
		if _, loaded := seen[user]; loaded {
			return
		}
		seen[user] = struct{}{}
		var total UserTraffic
		if drained := t.userTotals[user]; drained != nil {
			total.Up = drained.up
			total.Down = drained.down
		}
		if counter, loaded := t.users[user]; loaded {
			total.Up += counter.read.Load()
			total.Down += counter.write.Load()
		}
		snapshot.Users[user] = total
	}
	if len(users) == 0 {
		for user := range t.users {
			addUser(user)
		}
		for user := range t.userTotals {
			addUser(user)
		}
		return snapshot
	}
	for _, user := range users {
		addUser(user)
	}
	return snapshot
}

// Sessions returns a snapshot of the live sessions, newest counters included.
func (t *SessionTracker) Sessions() []SessionInfo {
	// Only the list is taken under the lock. Formatting every address of a
	// busy server takes milliseconds, and holding the tracker that long would
	// stall every connection being opened or closed meanwhile; a session's
	// fields do not change after it is created, and its counters are atomic.
	t.access.Lock()
	live := make([]*Session, 0, len(t.sessions))
	for _, session := range t.sessions {
		live = append(live, session)
	}
	t.access.Unlock()

	sessions := make([]SessionInfo, 0, len(live))
	for _, session := range live {
		sessions = append(sessions, SessionInfo{
			ID:          strconv.FormatUint(session.ID, 10),
			Inbound:     session.Inbound,
			User:        session.User,
			Outbound:    session.Outbound,
			Network:     session.Network,
			Source:      session.Source.String(),
			Client:      session.Client.String(),
			Destination: session.Destination.String(),
			Domain:      session.Domain,
			Rule:        session.Rule,
			CreatedAt:   session.CreatedAt.Unix(),
			Upload:      session.Upload.Load(),
			Download:    session.Download.Load(),
		})
	}
	return sessions
}

// takeSessions removes every session matching the predicate and returns their
// closers. Closing happens outside the lock: a closer runs leave(), which takes
// the same lock.
func (t *SessionTracker) takeSessions(match func(*Session) bool) []io.Closer {
	t.access.Lock()
	var closers []io.Closer
	for id, session := range t.sessions {
		if !match(session) {
			continue
		}
		delete(t.sessions, id)
		if session.User != "" {
			if n := t.userConns[session.User] - 1; n > 0 {
				t.userConns[session.User] = n
			} else {
				delete(t.userConns, session.User)
			}
			t.releaseIPLocked(session)
		}
		if session.closer != nil {
			closers = append(closers, session.closer)
		}
	}
	t.access.Unlock()

	for _, closer := range closers {
		_ = closer.Close()
	}
	return closers
}

// CloseByInbound closes every session of an inbound, for a restart or a config
// change that invalidates all of them.
func (t *SessionTracker) CloseByInbound(inbound string) int {
	return len(t.takeSessions(func(session *Session) bool {
		return session.Inbound == inbound
	}))
}

// CloseByInboundUsers closes the sessions of an inbound whose user is no longer
// in keepUsers, i.e. users just disabled or removed.
func (t *SessionTracker) CloseByInboundUsers(inbound string, keepUsers map[string]struct{}) int {
	return len(t.takeSessions(func(session *Session) bool {
		if session.Inbound != inbound {
			return false
		}
		_, keep := keepUsers[session.User]
		return !keep
	}))
}

// CloseByUser closes every routed connection of a user across all inbounds.
func (t *SessionTracker) CloseByUser(user string) int {
	if user == "" {
		return 0
	}
	return len(t.takeSessions(func(session *Session) bool {
		return session.User == user
	}))
}

type sessionConn struct {
	N.ExtendedConn
	tracker   *SessionTracker
	session   *Session
	cancel    context.CancelFunc
	closeOnce sync.Once
}

func (c *sessionConn) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.closeOnce.Do(func() {
		c.tracker.leave(c.session)
	})
	return c.ExtendedConn.Close()
}

func (c *sessionConn) Upstream() any {
	return c.ExtendedConn
}

func (c *sessionConn) ReaderReplaceable() bool {
	return true
}

func (c *sessionConn) WriterReplaceable() bool {
	return true
}

type sessionPacketConn struct {
	N.PacketConn
	tracker   *SessionTracker
	session   *Session
	cancel    context.CancelFunc
	closeOnce sync.Once
}

func (c *sessionPacketConn) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.closeOnce.Do(func() {
		c.tracker.leave(c.session)
	})
	return c.PacketConn.Close()
}

func (c *sessionPacketConn) Upstream() any {
	return c.PacketConn
}

func (c *sessionPacketConn) ReaderReplaceable() bool {
	return true
}

func (c *sessionPacketConn) WriterReplaceable() bool {
	return true
}

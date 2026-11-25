package core

import (
	"fmt"
	"net/netip"
	"sync"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

func testSession(tracker *SessionTracker, user string, clientAddr string) *Session {
	metadata := adapter.InboundContext{
		User:   user,
		Source: M.ParseSocksaddr("198.51.100.9:40000"),
	}
	if clientAddr != "" {
		metadata.ClientAddr = M.ParseSocksaddr(clientAddr)
	}
	session, _, _ := tracker.newSession(metadata, nil, nil, N.NetworkTCP)
	return session
}

func TestSessionTrackerIPLimitRejectsOnlyNewIPs(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.SetUserIPLimits(map[string]int{"alice": 1})

	first := testSession(tracker, "alice", "192.0.2.10:50000")
	if first == nil {
		t.Fatal("first IP was rejected")
	}
	same := testSession(tracker, "alice", "192.0.2.10:50001")
	if same == nil {
		t.Fatal("another connection from the same IP was rejected")
	}
	other := testSession(tracker, "alice", "192.0.2.11:50002")
	if other != nil {
		t.Fatal("a second IP was admitted under a limit of 1")
	}
	if got := tracker.ActiveIPCounts()["alice"]; got != 1 {
		t.Fatalf("active IPs = %d, want 1", got)
	}

	tracker.leave(first)
	if got := tracker.ActiveIPCounts()["alice"]; got != 1 {
		t.Fatalf("active IPs after one close = %d, want 1", got)
	}
	tracker.leave(same)
	if got := tracker.ActiveIPCounts()["alice"]; got != 0 {
		t.Fatalf("active IPs after all closes = %d, want 0", got)
	}
	if admitted := testSession(tracker, "alice", "192.0.2.11:50003"); admitted == nil {
		t.Fatal("the new IP was not admitted after the old IP closed")
	}
}

func TestSessionTrackerIPv6IsGroupedByPrefix(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.SetUserIPLimits(map[string]int{"alice": 1})

	first := testSession(tracker, "alice", "[2001:db8:1:2::1]:50000")
	if first == nil {
		t.Fatal("first IPv6 address was rejected")
	}
	samePrefix := testSession(tracker, "alice", "[2001:db8:1:2::ffff]:50001")
	if samePrefix == nil {
		t.Fatal("an address in the same /64 was treated as a new IP")
	}
	otherPrefix := testSession(tracker, "alice", "[2001:db8:1:3::1]:50002")
	if otherPrefix != nil {
		t.Fatal("an address in another /64 was admitted")
	}
	summary := tracker.UserIPs("alice")
	if summary.Count != 1 || summary.IPs[0].IP != "2001:db8:1:2::/64" {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestSessionTrackerIPv4PrefixLenCanBeWidened(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.SetIPPrefixes(24, 64)
	tracker.SetUserIPLimits(map[string]int{"alice": 1})
	if session := testSession(tracker, "alice", "192.0.2.10:50000"); session == nil {
		t.Fatal("first address was rejected")
	}
	if session := testSession(tracker, "alice", "192.0.2.250:50001"); session == nil {
		t.Fatal("another address in the same /24 was treated as a new IP")
	}
	if session := testSession(tracker, "alice", "192.0.3.10:50002"); session != nil {
		t.Fatal("an address outside the /24 was admitted")
	}
}

func TestSessionTrackerIPLimitDryRunOnlyLogs(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.SetUserIPLimits(map[string]int{"alice": 1})
	tracker.SetIPLimitDryRun(true)
	if session := testSession(tracker, "alice", "192.0.2.10:50000"); session == nil {
		t.Fatal("first IP was rejected")
	}
	second := testSession(tracker, "alice", "192.0.2.11:50001")
	if second == nil {
		t.Fatal("dry run rejected the second IP")
	}
	if got := tracker.ActiveIPCounts()["alice"]; got != 2 {
		t.Fatalf("active IPs = %d, want 2 in dry run", got)
	}
}

func TestSessionTrackerIPLimitDoesNotConsumeConnectionCount(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.SetUserLimits(map[string]int{"alice": 2})
	tracker.SetUserIPLimits(map[string]int{"alice": 1})
	first := testSession(tracker, "alice", "192.0.2.10:50000")
	if first == nil {
		t.Fatal("first IP was rejected")
	}
	if rejected := testSession(tracker, "alice", "192.0.2.11:50001"); rejected != nil {
		t.Fatal("second IP was admitted")
	}
	if got := tracker.ActiveConns("alice"); got != 1 {
		t.Fatalf("active connections = %d, want 1", got)
	}
	if admitted := testSession(tracker, "alice", "192.0.2.10:50002"); admitted == nil {
		t.Fatal("same-IP connection was blocked after an IP rejection")
	}
}

func TestSessionTrackerUsesSourceWhenRelayIsAbsent(t *testing.T) {
	tracker := NewSessionTracker()
	session := testSession(tracker, "alice", "")
	if session == nil {
		t.Fatal("direct session was rejected")
	}
	if session.Client != session.Source {
		t.Fatalf("client = %v, want source %v", session.Client, session.Source)
	}
	info := tracker.Sessions()
	if len(info) != 1 || info[0].Client != info[0].Source {
		t.Fatalf("unexpected session info: %+v", info)
	}
}

func TestSessionTrackerKeepsSourceAndClientSeparate(t *testing.T) {
	tracker := NewSessionTracker()
	session := testSession(tracker, "alice", "192.0.2.10:50000")
	if session == nil {
		t.Fatal("session was rejected")
	}
	if session.Source.String() != "198.51.100.9:40000" {
		t.Fatalf("source changed: %v", session.Source)
	}
	if session.Client.String() != "192.0.2.10:50000" {
		t.Fatalf("client = %v", session.Client)
	}
}

func TestSessionTrackerReleasesIPsWhenSessionsAreTaken(t *testing.T) {
	tracker := NewSessionTracker()
	for i := 0; i < 3; i++ {
		if session := testSession(tracker, "alice", fmt.Sprintf("192.0.2.%d:5000%d", 10+i, i)); session == nil {
			t.Fatalf("session %d was rejected", i)
		}
	}
	if got := tracker.ActiveIPCounts()["alice"]; got != 3 {
		t.Fatalf("active IPs = %d, want 3", got)
	}
	tracker.CloseByUser("alice")
	if got := tracker.ActiveIPCounts()["alice"]; got != 0 {
		t.Fatalf("active IPs after CloseByUser = %d, want 0", got)
	}
	if got := tracker.ActiveConns("alice"); got != 0 {
		t.Fatalf("active connections after CloseByUser = %d, want 0", got)
	}
}

func TestSessionTrackerUserIPSummary(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.SetUserIPLimits(map[string]int{"alice": 2})
	first := testSession(tracker, "alice", "192.0.2.10:50000")
	if first == nil {
		t.Fatal("session was rejected")
	}
	if same := testSession(tracker, "alice", "192.0.2.10:50001"); same == nil {
		t.Fatal("second session was rejected")
	}
	summary := tracker.UserIPs("alice")
	if summary.Count != 1 || summary.Limit != 2 || len(summary.IPs) != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.IPs[0].Sessions != 2 || summary.IPs[0].Since == 0 {
		t.Fatalf("unexpected IP entry: %+v", summary.IPs[0])
	}
}

func TestSessionTrackerConcurrentIPAccounting(t *testing.T) {
	tracker := NewSessionTracker()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			session := testSession(tracker, "alice", fmt.Sprintf("192.0.2.%d:50000", i%7))
			if session != nil {
				tracker.leave(session)
			}
		}(i)
	}
	wg.Wait()
	if got := tracker.ActiveIPCounts()["alice"]; got != 0 {
		t.Fatalf("active IPs = %d, want 0", got)
	}
	tracker.access.Lock()
	defer tracker.access.Unlock()
	if len(tracker.userIPs) != 0 {
		t.Fatalf("userIPs not empty: %v", tracker.userIPs)
	}
	if len(tracker.sessions) != 0 {
		t.Fatalf("sessions not empty: %d", len(tracker.sessions))
	}
}

func TestSessionTrackerUnlimitedStillCountsIPs(t *testing.T) {
	tracker := NewSessionTracker()
	session := testSession(tracker, "alice", "192.0.2.10:50000")
	if session == nil {
		t.Fatal("session without a limit was rejected")
	}
	if got := tracker.ActiveIPCounts()["alice"]; got != 1 {
		t.Fatalf("active IPs = %d, want 1", got)
	}
}

func TestIPKeyNormalizesMappedIPv4(t *testing.T) {
	tracker := NewSessionTracker()
	tracker.access.Lock()
	defer tracker.access.Unlock()
	key := tracker.ipKeyLocked(M.ParseSocksaddr("[::ffff:192.0.2.10]:443"))
	want := netip.MustParsePrefix("192.0.2.10/32")
	if key != want {
		t.Fatalf("key = %v, want %v", key, want)
	}
}

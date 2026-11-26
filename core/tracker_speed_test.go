package core

import (
	"context"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing/common/buf"
	M "github.com/sagernet/sing/common/metadata"
)

func speedSnapshotFor(t *testing.T, tracker *SessionTracker, users ...string) UserSpeedSnapshot {
	t.Helper()
	snapshot := tracker.UserSpeedSnapshot(users)
	if snapshot.Instance == "" {
		t.Fatal("snapshot has no tracker instance id")
	}
	if snapshot.SampledAt <= 0 {
		t.Fatal("snapshot has no sample time")
	}
	return snapshot
}

func TestUserSpeedSnapshotIsNonDestructive(t *testing.T) {
	tracker := NewSessionTracker()
	if session := testSession(tracker, "alice", "192.0.2.10:50000"); session == nil {
		t.Fatal("session was rejected")
	}
	counter := tracker.users["alice"]
	counter.read.Add(250_000)
	counter.write.Add(25_000)

	first := speedSnapshotFor(t, tracker, "alice", "bob")
	if got := first.Users["alice"]; got.Up != 250_000 || got.Down != 25_000 {
		t.Fatalf("alice first snapshot = %+v", got)
	}
	if got := first.Users["bob"]; got.Up != 0 || got.Down != 0 {
		t.Fatalf("unknown user snapshot = %+v", got)
	}

	// Reading again must not drain anything.
	second := speedSnapshotFor(t, tracker, "alice")
	if got := second.Users["alice"]; got.Up != 250_000 || got.Down != 25_000 {
		t.Fatalf("alice second snapshot = %+v", got)
	}

	// The stats job drains the live counters; the snapshot must still see the
	// same cumulative bytes through the drained totals.
	drained := tracker.GetStats()
	if drained == nil || len(*drained) == 0 {
		t.Fatal("GetStats returned no rows")
	}
	third := speedSnapshotFor(t, tracker, "alice")
	if got := third.Users["alice"]; got.Up != 250_000 || got.Down != 25_000 {
		t.Fatalf("alice after drain = %+v", got)
	}

	// A second drain must not double count the already drained bytes.
	tracker.GetStats()
	fourth := speedSnapshotFor(t, tracker, "alice")
	if got := fourth.Users["alice"]; got.Up != 250_000 || got.Down != 25_000 {
		t.Fatalf("alice after second drain = %+v", got)
	}

	// New live bytes are added on top of the drained total.
	counter.read.Add(1_000)
	fifth := speedSnapshotFor(t, tracker, "alice")
	if got := fifth.Users["alice"]; got.Up != 251_000 || got.Down != 25_000 {
		t.Fatalf("alice after new traffic = %+v", got)
	}
}

func TestUserSpeedSnapshotAllUsersWhenNoFilter(t *testing.T) {
	tracker := NewSessionTracker()
	for _, user := range []string{"alice", "bob"} {
		if session := testSession(tracker, user, "192.0.2.10:50000"); session == nil {
			t.Fatalf("session for %s was rejected", user)
		}
		tracker.users[user].read.Add(10)
	}
	snapshot := speedSnapshotFor(t, tracker)
	if len(snapshot.Users) != 2 {
		t.Fatalf("users = %+v, want alice and bob", snapshot.Users)
	}
}

func TestUserSpeedSnapshotInstanceChangesAcrossTrackers(t *testing.T) {
	first := NewSessionTracker()
	second := NewSessionTracker()
	if first.instanceID == second.instanceID {
		t.Fatalf("tracker instance ids are equal: %q", first.instanceID)
	}
	if speedSnapshotFor(t, first).Instance == speedSnapshotFor(t, second).Instance {
		t.Fatal("snapshots from two trackers share an instance id")
	}
}

func TestUserSpeedSnapshotConcurrentWithDrain(t *testing.T) {
	tracker := NewSessionTracker()
	if session := testSession(tracker, "alice", "192.0.2.10:50000"); session == nil {
		t.Fatal("session was rejected")
	}
	tracker.users["alice"].read.Add(1_000)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 500; i++ {
			tracker.UserSpeedSnapshot([]string{"alice"})
		}
	}()
	for i := 0; i < 100; i++ {
		tracker.GetStats()
	}
	<-done
	snapshot := speedSnapshotFor(t, tracker, "alice")
	if got := snapshot.Users["alice"].Up; got != 1_000 {
		t.Fatalf("up = %d, want 1000", got)
	}
}

func TestUserSpeedSnapshotCountsUDP(t *testing.T) {
	tracker := NewSessionTracker()
	raw := &fakePacketConn{readData: []byte("hello")}
	conn := tracker.RoutedPacketConnection(context.Background(), raw, adapter.InboundContext{
		User:   "alice",
		Source: M.ParseSocksaddr("198.51.100.9:40000"),
	}, nil, nil)
	defer conn.Close()

	out := buf.As(make([]byte, 1000))
	if err := conn.WritePacket(out, M.Socksaddr{}); err != nil {
		t.Fatal(err)
	}
	out.Release()
	in := buf.NewSize(8)
	if _, err := conn.ReadPacket(in); err != nil {
		t.Fatal(err)
	}
	in.Release()

	snapshot := speedSnapshotFor(t, tracker, "alice")
	if got := snapshot.Users["alice"]; got.Up != 5 || got.Down != 1000 {
		t.Fatalf("udp snapshot = %+v, want up=5 down=1000", got)
	}
}

func TestUserSpeedSnapshotKeepsClosedSessionBytes(t *testing.T) {
	tracker := NewSessionTracker()
	session := testSession(tracker, "alice", "192.0.2.10:50000")
	if session == nil {
		t.Fatal("session was rejected")
	}
	tracker.users["alice"].read.Add(4_096)
	tracker.users["alice"].write.Add(2_048)
	tracker.leave(session)

	snapshot := speedSnapshotFor(t, tracker, "alice")
	if got := snapshot.Users["alice"]; got.Up != 4_096 || got.Down != 2_048 {
		t.Fatalf("closed session snapshot = %+v", got)
	}
	tracker.GetStats()
	snapshot = speedSnapshotFor(t, tracker, "alice")
	if got := snapshot.Users["alice"]; got.Up != 4_096 || got.Down != 2_048 {
		t.Fatalf("closed session after drain = %+v", got)
	}
}

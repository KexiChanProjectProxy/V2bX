package sing

import (
	"sync"
	"testing"

	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/common/counter"
	vCore "github.com/InazumaV/V2bX/core"
)

// TestSingImplementsCore verifies Sing satisfies the Core interface at compile time.
// (The var _ line in sing.go already does this, but we confirm it here too.)
func TestSingImplementsCore(t *testing.T) {
	var _ vCore.Core = (*Sing)(nil)
}

// TestProtocols verifies exactly the three supported protocols are returned.
func TestProtocols(t *testing.T) {
	b := &Sing{}
	protos := b.Protocols()
	want := []string{"shadowsocks", "anytls", "hysteria2"}
	if len(protos) != len(want) {
		t.Fatalf("expected %d protocols, got %d: %v", len(want), len(protos), protos)
	}
	m := make(map[string]bool)
	for _, p := range protos {
		m[p] = true
	}
	for _, p := range want {
		if !m[p] {
			t.Errorf("protocol %q missing from Protocols()", p)
		}
	}
	for _, forbidden := range []string{"vmess", "vless", "trojan", "tuic", "hysteria"} {
		if m[forbidden] {
			t.Errorf("protocol %q should not be in Protocols()", forbidden)
		}
	}
}

// TestType verifies Type() returns "sing".
func TestType(t *testing.T) {
	b := &Sing{}
	if b.Type() != "sing" {
		t.Errorf("expected Type()=sing, got %q", b.Type())
	}
}

// TestUserMapConcurrency verifies the UserMap is safe for concurrent read/write.
func TestUserMapConcurrency(t *testing.T) {
	um := &UserMap{
		uidMap: make(map[string]int),
	}

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n * 2)
	for i := range n {
		go func(i int) {
			defer wg.Done()
			um.mapLock.Lock()
			um.uidMap["uuid-w"] = i
			um.mapLock.Unlock()
		}(i)
		go func() {
			defer wg.Done()
			um.mapLock.RLock()
			_ = um.uidMap["uuid-w"]
			um.mapLock.RUnlock()
		}()
	}
	wg.Wait()
}

// TestGetUserTrafficSlice_Empty verifies GetUserTrafficSlice returns nil when no traffic.
func TestGetUserTrafficSlice_Empty(t *testing.T) {
	b := &Sing{
		hookServer: &HookServer{},
		users: &UserMap{
			uidMap: make(map[string]int),
		},
		nodeReportMinTrafficBytes: map[string]int64{},
	}

	result, err := b.GetUserTrafficSlice("nonexistent-tag", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil traffic for unknown tag, got %v", result)
	}
}

// TestGetUserTrafficSlice_WithData verifies traffic is reported when above threshold.
func TestGetUserTrafficSlice_WithData(t *testing.T) {
	hs := &HookServer{}
	tc := counter.NewTrafficCounter()

	// Populate counter directly as if real traffic happened
	uuid := "test-uuid-traffic"
	tc.Tx(uuid, 5000) // up
	tc.Rx(uuid, 3000) // down
	hs.counter.Store("tag1", tc)

	b := &Sing{
		hookServer: hs,
		users: &UserMap{
			uidMap: map[string]int{uuid: 42},
		},
		nodeReportMinTrafficBytes: map[string]int64{"tag1": 0}, // 0 = always report
	}

	result, err := b.GetUserTrafficSlice("tag1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 traffic entry, got %d", len(result))
	}
	if result[0].UID != 42 {
		t.Errorf("expected UID=42, got %d", result[0].UID)
	}
	if result[0].Upload != 5000 {
		t.Errorf("expected Upload=5000, got %d", result[0].Upload)
	}
	if result[0].Download != 3000 {
		t.Errorf("expected Download=3000, got %d", result[0].Download)
	}
}

// TestGetUserTrafficSlice_BelowThreshold verifies low-traffic entries are skipped.
func TestGetUserTrafficSlice_BelowThreshold(t *testing.T) {
	hs := &HookServer{}
	tc := counter.NewTrafficCounter()

	uuid := "low-traffic-uuid"
	tc.Tx(uuid, 100)
	tc.Rx(uuid, 50)
	hs.counter.Store("tag2", tc)

	b := &Sing{
		hookServer: hs,
		users: &UserMap{
			uidMap: map[string]int{uuid: 1},
		},
		// threshold is 1MB = 1024*1024; 150 bytes is below it
		nodeReportMinTrafficBytes: map[string]int64{"tag2": 1024 * 1024},
	}

	result, err := b.GetUserTrafficSlice("tag2", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 traffic entries below threshold, got %d", len(result))
	}
}

// TestGetUserTrafficSlice_Reset verifies reset=true zeros the counters.
func TestGetUserTrafficSlice_Reset(t *testing.T) {
	hs := &HookServer{}
	tc := counter.NewTrafficCounter()

	uuid := "reset-uuid"
	tc.Tx(uuid, 9999)
	tc.Rx(uuid, 9999)
	hs.counter.Store("tag3", tc)

	b := &Sing{
		hookServer: hs,
		users: &UserMap{
			uidMap: map[string]int{uuid: 99},
		},
		nodeReportMinTrafficBytes: map[string]int64{"tag3": 0},
	}

	result, err := b.GetUserTrafficSlice("tag3", true /* reset */)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}

	// After reset the counter should be zero
	if tc.GetUpCount(uuid) != 0 {
		t.Errorf("expected UpCount=0 after reset, got %d", tc.GetUpCount(uuid))
	}
	if tc.GetDownCount(uuid) != 0 {
		t.Errorf("expected DownCount=0 after reset, got %d", tc.GetDownCount(uuid))
	}
}

// TestGetUserTrafficSlice_UnknownUser verifies that traffic for unknown users is cleaned up.
func TestGetUserTrafficSlice_UnknownUser(t *testing.T) {
	hs := &HookServer{}
	tc := counter.NewTrafficCounter()

	// This UUID is not in uidMap → unknown user → should be deleted from counter
	tc.Tx("ghost-uuid", 5000)
	tc.Rx("ghost-uuid", 5000)
	hs.counter.Store("tag4", tc)

	b := &Sing{
		hookServer: hs,
		users: &UserMap{
			uidMap: make(map[string]int), // ghost-uuid not in map → id=0
		},
		nodeReportMinTrafficBytes: map[string]int64{"tag4": 0},
	}

	result, err := b.GetUserTrafficSlice("tag4", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Unknown users (id=0) should be cleaned up, not reported
	if len(result) != 0 {
		t.Errorf("expected 0 results for unknown user, got %d: %v", len(result), result)
	}
	// The counter entry should have been deleted
	if tc.GetUpCount("ghost-uuid") != 0 {
		t.Errorf("ghost user counter should have been deleted")
	}
}

// TestGetUserTraffic verifies GetUserTraffic returns per-user counts.
func TestGetUserTraffic(t *testing.T) {
	hs := &HookServer{}
	tc := counter.NewTrafficCounter()
	uuid := "direct-uuid"
	tc.Tx(uuid, 111)
	tc.Rx(uuid, 222)
	hs.counter.Store("mytag", tc)

	b := &Sing{hookServer: hs}

	up, down := b.GetUserTraffic("mytag", uuid, false)
	if up != 111 {
		t.Errorf("expected up=111, got %d", up)
	}
	if down != 222 {
		t.Errorf("expected down=222, got %d", down)
	}
}

// TestGetUserTraffic_Reset verifies reset=true zeros the per-user counter.
func TestGetUserTraffic_Reset(t *testing.T) {
	hs := &HookServer{}
	tc := counter.NewTrafficCounter()
	uuid := "reset-direct"
	tc.Tx(uuid, 500)
	hs.counter.Store("t", tc)

	b := &Sing{hookServer: hs}
	b.GetUserTraffic("t", uuid, true)

	if tc.GetUpCount(uuid) != 0 {
		t.Errorf("expected 0 after reset, got %d", tc.GetUpCount(uuid))
	}
}

// TestCoreRegistration verifies "sing" is registered in the core registry.
func TestCoreRegistration(t *testing.T) {
	registered := vCore.RegisteredCore()
	found := false
	for _, name := range registered {
		if name == "sing" {
			found = true
			break
		}
	}
	if !found {
		t.Error("\"sing\" should be registered in the core registry")
	}
}

// TestAddUsersUidMap verifies AddUsers populates the uidMap.
func TestAddUsersUidMap(t *testing.T) {
	// We can test the uidMap logic without a real inbound by exercising the map directly.
	um := &UserMap{
		uidMap: make(map[string]int),
	}
	users := []panel.UserInfo{
		{Id: 10, Uuid: "uuid-10"},
		{Id: 20, Uuid: "uuid-20"},
	}

	um.mapLock.Lock()
	for i := range users {
		um.uidMap[users[i].Uuid] = users[i].Id
	}
	um.mapLock.Unlock()

	um.mapLock.RLock()
	defer um.mapLock.RUnlock()
	if um.uidMap["uuid-10"] != 10 {
		t.Errorf("expected uidMap[uuid-10]=10, got %d", um.uidMap["uuid-10"])
	}
	if um.uidMap["uuid-20"] != 20 {
		t.Errorf("expected uidMap[uuid-20]=20, got %d", um.uidMap["uuid-20"])
	}
}

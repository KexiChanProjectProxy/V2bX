package counter

import (
	"sync"
	"testing"
)

func TestTrafficCounterBasic(t *testing.T) {
	c := NewTrafficCounter()

	// Newly created counter returns zeros
	if up := c.GetUpCount("uuid-1"); up != 0 {
		t.Fatalf("expected 0 up, got %d", up)
	}
	if down := c.GetDownCount("uuid-1"); down != 0 {
		t.Fatalf("expected 0 down, got %d", down)
	}
}

func TestTrafficCounterTxRx(t *testing.T) {
	c := NewTrafficCounter()

	c.Tx("user-a", 1000)
	c.Tx("user-a", 500)
	c.Rx("user-a", 2000)

	if up := c.GetUpCount("user-a"); up != 1500 {
		t.Errorf("expected up=1500, got %d", up)
	}
	if down := c.GetDownCount("user-a"); down != 2000 {
		t.Errorf("expected down=2000, got %d", down)
	}
}

func TestTrafficCounterReset(t *testing.T) {
	c := NewTrafficCounter()

	c.Tx("u", 999)
	c.Rx("u", 888)
	c.Reset("u")

	if up := c.GetUpCount("u"); up != 0 {
		t.Errorf("expected 0 after reset, got %d", up)
	}
	if down := c.GetDownCount("u"); down != 0 {
		t.Errorf("expected 0 after reset, got %d", down)
	}
}

func TestTrafficCounterDelete(t *testing.T) {
	c := NewTrafficCounter()

	c.Tx("del-me", 100)
	c.Delete("del-me")

	if up := c.GetUpCount("del-me"); up != 0 {
		t.Errorf("expected 0 after delete, got %d", up)
	}
	if c.Len() != 0 {
		t.Errorf("expected Len()=0 after delete, got %d", c.Len())
	}
}

func TestTrafficCounterLen(t *testing.T) {
	c := NewTrafficCounter()

	for _, uuid := range []string{"u1", "u2", "u3"} {
		c.Tx(uuid, 1)
	}
	if l := c.Len(); l != 3 {
		t.Errorf("expected Len()=3, got %d", l)
	}
	c.Delete("u2")
	if l := c.Len(); l != 2 {
		t.Errorf("expected Len()=2 after delete, got %d", l)
	}
}

func TestTrafficCounterIsolation(t *testing.T) {
	c := NewTrafficCounter()

	c.Tx("alice", 100)
	c.Rx("bob", 200)

	if up := c.GetUpCount("bob"); up != 0 {
		t.Errorf("alice tx should not affect bob up, got %d", up)
	}
	if down := c.GetDownCount("alice"); down != 0 {
		t.Errorf("bob rx should not affect alice down, got %d", down)
	}
}

func TestTrafficCounterConcurrent(t *testing.T) {
	c := NewTrafficCounter()
	const goroutines = 50
	const ops = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			uuid := "user"
			for range ops {
				c.Tx(uuid, i+1)
				c.Rx(uuid, i+1)
			}
		}(i)
	}
	wg.Wait()

	// Just verify no panic and counts are positive
	up := c.GetUpCount("user")
	down := c.GetDownCount("user")
	if up <= 0 || down <= 0 {
		t.Errorf("expected positive traffic after concurrent ops, up=%d down=%d", up, down)
	}
}

func TestGetCounterReturnsSameStorage(t *testing.T) {
	c := NewTrafficCounter()

	s1 := c.GetCounter("x")
	s2 := c.GetCounter("x")
	if s1 != s2 {
		t.Error("GetCounter should return the same *TrafficStorage for the same uuid")
	}
}

func TestTrafficStorageDirect(t *testing.T) {
	s := &TrafficStorage{}
	s.UpCounter.Add(42)
	s.DownCounter.Add(99)

	if s.UpCounter.Load() != 42 {
		t.Errorf("expected UpCounter=42, got %d", s.UpCounter.Load())
	}
	if s.DownCounter.Load() != 99 {
		t.Errorf("expected DownCounter=99, got %d", s.DownCounter.Load())
	}

	s.UpCounter.Store(0)
	s.DownCounter.Store(0)
	if s.UpCounter.Load() != 0 || s.DownCounter.Load() != 0 {
		t.Error("Store(0) should zero the counters")
	}
}

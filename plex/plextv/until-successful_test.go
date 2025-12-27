package plextv

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestUntilSuccessful(t *testing.T) {
	const (
		iterations     = 100
		successfulCall = 75
	)

	var c atomic.Int32
	var once untilSuccessful
	f := func() error {
		current := c.Add(1)
		switch {
		case current < successfulCall:
			return errors.New("not yet")
		case current > successfulCall:
			t.Fatalf("untilSuccessful.Do called after done. c: %d", current)
		}
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(iterations)
	for range iterations {
		go func() {
			defer wg.Done()
			_ = once.Do(f)
		}()
	}
	wg.Wait()
	if calls := c.Load(); calls != successfulCall {
		t.Fatalf("untilSuccessful.Do called %d times, want %d", calls, successfulCall)
	}
}

func TestUntilSuccessfulPanic(t *testing.T) {
	var once untilSuccessful
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("Once.Do did not panic")
			}
		}()
		_ = once.Do(func() error {
			panic("failed")
		})
	}()

	_ = once.Do(func() error {
		t.Fatalf("Once.Do called twice")
		return nil
	})
}

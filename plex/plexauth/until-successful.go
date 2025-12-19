package plexauth

import (
	"sync"
	"sync/atomic"
)

// untilSuccessful is like sync.Once, but only marks the task as done if it succeeds (returns a nil error)
type untilSuccessful struct {
	lock sync.Mutex
	done atomic.Bool
}

func (o *untilSuccessful) Do(f func() error) (err error) {
	if !o.done.Load() {
		err = o.doSlow(f)
	}
	return err
}

func (o *untilSuccessful) doSlow(f func() error) (err error) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if !o.done.Load() {
		defer func() { o.done.Store(err == nil) }()
		err = f()
	}
	return err
}

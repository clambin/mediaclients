package plexauth

import "sync"

// onceSuccessful is like sync.Once, but only marks the task as once if it succeeds (returns a nil error)
type onceSuccessful struct {
	lock sync.Mutex
	done bool
}

func (o *onceSuccessful) Do(f func() error) (err error) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if !o.done {
		err = f()
		o.done = err == nil
	}
	return err
}

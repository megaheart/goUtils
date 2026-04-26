package concurrency

import "sync"

type DoNewest struct {
	latestStopFunc func()
	mx             sync.Mutex
}

func NewDoNewest() *DoNewest {
	return &DoNewest{
		latestStopFunc: nil,
		mx:             sync.Mutex{},
	}
}

func (d *DoNewest) Do(stopFunc func()) {
	d.mx.Lock()
	defer d.mx.Unlock()
	if d.latestStopFunc != nil {
		d.latestStopFunc()
	}
	d.latestStopFunc = stopFunc
}

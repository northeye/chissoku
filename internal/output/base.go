package output

import (
	"context"
	"sync/atomic"
	"time"
)

// Base outputter base struct.
//
// It holds only the latest data and calls the emit function given to start()
// immediately for the first data, and then every Interval seconds.
// If Interval <= 0, every data is emitted as soon as it arrives.
type Base struct {
	// Output interval (sec)
	Interval int `long:"interval" help:"interval (second) for output. default: '60'" default:"60"`

	// the latest data which is not emitted yet
	latest atomic.Pointer[[]byte]
	// notifies arrival of new data
	notify chan struct{}
	// cancels the emitter loop
	cancel context.CancelFunc
	// closed when the emitter loop exits
	done chan struct{}
}

// Output stores the data to be emitted. It never blocks.
func (b *Base) Output(d []byte) {
	b.latest.Store(&d)
	select {
	case b.notify <- struct{}{}:
	default:
	}
}

// Close stops the emitter loop and waits for it to exit.
func (b *Base) Close() {
	if b.cancel == nil {
		return
	}
	b.cancel()
	<-b.done
}

// start runs the emitter loop as a goroutine.
// emit is called on the loop goroutine, so it may block (e.g. network I/O).
func (b *Base) start(ctx context.Context, emit func([]byte)) {
	b.notify = make(chan struct{}, 1)
	b.done = make(chan struct{})
	ctx, b.cancel = context.WithCancel(ctx)
	go b.loop(ctx, emit)
}

func (b *Base) loop(ctx context.Context, emit func([]byte)) {
	defer close(b.done)

	var tick <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-b.notify:
			if tick != nil {
				continue // wait for the next tick
			}
			b.flush(emit)
			if b.Interval > 0 {
				t := time.NewTicker(time.Duration(b.Interval) * time.Second)
				defer t.Stop()
				tick = t.C
			}
		case <-tick:
			b.flush(emit)
		}
	}
}

func (b *Base) flush(emit func([]byte)) {
	if p := b.latest.Swap(nil); p != nil {
		emit(*p)
	}
}

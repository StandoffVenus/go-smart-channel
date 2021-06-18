package safe_channel

import (
	"context"
	"sync"
	"sync/atomic"
)

// Send will attempt to send the provided value.
// If the value couldn't be sent (channel closed)
// then the returned channel will send "false".
// All successful Send() calls will return a channel
// that sends "true".
//
// After reading from the returned channel, it will be
// closed.
type Send func(v interface{}) <-chan bool

// Receive will attempt to receive a value. The returned
// channel will send this value. If the underlying channel
// is closed, so will the returned channel be.
//
// After reading from the returned channel, it will be
// closed.
type Receive func() <-chan interface{}

// Close will safely close the channel it is associated with,
// closing the returned channel once the underlying channel closes.
// It is safe to call Close many times. Any Send() or Receive()
// after Close() is called will result in "failures".
type Close func() <-chan struct{}

// New creates a new channel with no buffer.
// The returned functions will send to, receive from, and close
// the new channel (respectively)
func New() (Send, Receive, Close) {
	return OfSize(0)
}

// OfSize creates a new channel with a buffer of the provided size.
// The returned functions will send to, receive from, and close
// the new channel (respectively)
func OfSize(bufferSize int64) (Send, Receive, Close) {
	const (
		open int32 = iota
		closed
	)

	ch := make(chan interface{}, bufferSize)
	chOpen := open
	ctx, ctxCancel := context.WithCancel(context.Background())
	oncer := sync.Once{}
	waiter := sync.WaitGroup{}

	var receive Receive = func() <-chan interface{} {
		receiver := make(chan interface{}, 1)
		go func() {
			defer close(receiver)
			if v, ok := <-ch; ok {
				receiver <- v
			}
		}()
		return receiver
	}

	var send Send = func(v interface{}) <-chan bool {
		waiter.Add(1)

		sentChan := make(chan bool, 1)
		go func() {
			defer close(sentChan)
			defer waiter.Done()

			if atomic.CompareAndSwapInt32(&chOpen, open, open) {
				select {
				case ch <- v:
					sentChan <- true
				case <-ctx.Done():
					sentChan <- false
				}
			} else {
				sentChan <- false
			}
		}()

		return sentChan
	}

	var earlyClose Close = func() <-chan struct{} {
		done := make(chan struct{})
		oncer.Do(func() {
			go func() {
				atomic.StoreInt32(&chOpen, closed)
				ctxCancel()
				waiter.Wait()
				close(ch)
				close(done)
			}()
		})

		return done
	}

	return send, receive, earlyClose
}

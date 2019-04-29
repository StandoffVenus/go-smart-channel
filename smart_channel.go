package smart_channel

import (
    "sync"
    "sync/atomic"
    "time"
)

const (
    channelOpen uint32 = iota
    channelReleased
    channelClosed = channelReleased
)

// Implementation of ISmartChannel.
type smartChannel struct {
    // The underlying channel.
    channel chan interface{}

    // Keeps track of whether we've closed the channel.
    closed *uint32

    // Used for closing a channel only once.
    once *sync.Once

    // Used to know when we have nothing depending on the channel.
    waiter *sync.WaitGroup
}

// NewSmartChannel creates a new ISmartChannel,
// whose underlying channel has buffer size == buffer.
//
// NewSmartChannel will panic if buffer < 0.
func NewSmartChannel(buffer int) (ISmartChannel) {
    tmp := channelOpen

    return &smartChannel {
        channel: make(chan interface{}, buffer),
        closed: &tmp,
        once: new(sync.Once),
        waiter: new(sync.WaitGroup),
    }
}

// Get() returns a new ISmartChannelReference
func (sc *smartChannel) Get() ISmartChannelReference {
    // Make the new reference
    scr := newSmartChannelReference(sc, atomic.LoadUint32(sc.closed))

    // Increment counter for a reference
    sc.waiter.Add(1)

    // Return the new reference
    return scr
}

// IsReleased() determines if this ISmartChannel is closed,
// returning true if so, false otherwise.
func (sc *smartChannel) IsClosed() bool {
    return atomic.LoadUint32(sc.closed) == channelClosed
}

// send will handle sending data if possible, returning
// true on successful send, false on failure.
//
// If timeout
//  > 0: block until time.After(timeout).
//      return (false, TimeoutError)
//  == 0: block until channel operation is possible.
//      return (true, nil)
//  < 0: panic!
func (sc *smartChannel) send(value interface{}, timeout time.Duration) (bool, *TimeoutError) {
    if sc.IsClosed() {
        return false, nil
    }

    // We're going to send, so we'll wait once
    sc.waiter.Add(1)
    defer sc.waiter.Done()

    // timeout possibilities
    switch {
        case timeout > 0: select {
            case sc.channel <- value:
                return true, nil
            case <-time.After(timeout):
                return false, &TimeoutError { send, timeout }
        }
        case timeout == 0:
            sc.channel <- value
            return true, nil
        default:
            panic("timeout outside logical range (negative)")
    }
}

// receive will handle receiving some data if possible,
// returning true on success, false on failure.
//
// If timeout
//  > 0: block until time.After(timeout).
//      return (<value>, false, TimeoutError)
//  == 0: block until channel operation is possible.
//      return (<value>, true, nil)
//  < 0: panic!
func (sc *smartChannel) receive(timeout time.Duration) (interface{}, bool, *TimeoutError) {
    if sc.IsClosed() {
        return nil, false, nil
    }

    // We're going to receive, so wait once
    sc.waiter.Add(1)
    defer sc.waiter.Done()

    // timeout possibilities
    // The reason we can wait for the channels to receive without
    // needing to check if they're closed (val, ok := ...) is
    // due to checking if the channel is closed above.
    switch {
        case timeout > 0: select {
            // Either we received, or the channel has been closed
            case val := <-sc.channel:
                return val, true, nil
            case <-time.After(timeout):
                return nil, false, &TimeoutError { receive, timeout }
        }
        case timeout == 0:
            val := <-sc.channel
            return val, true, nil
        default:
            panic("timeout outside logical range (negative)")
    }
}

// CALL ONLY ONCE PER REFERENCE.
//
// schedule_release will "release" a reference, closing
// the channel iff
//  1) we want to close on this release
//  2) there are no references left
//  3) there are no items left in the channel
//  4) the channel is not already closed
func (sc *smartChannel) schedule_release(closeWhenPossible bool) (<-chan bool) {
    releaseChannel := make(chan bool)

    // We've released a reference, so we'll free up once
    sc.waiter.Done()

    // Check if anything else is necessary
    if closeWhenPossible && !sc.IsClosed() {
        go func(c chan<- bool) {
            defer close(c)

            // Wait until we can close the underlying channel
            sc.waiter.Wait()

            // Close the channel only once
            sc.once.Do(func() {
                // We want to make sure that we close the channel,
                // *then* reflect the change. If the close somehow
                // panics, we don't want to store an inaccurate state.
                close(sc.channel)

                // Store that we've closed the channel
                atomic.StoreUint32(sc.closed, channelClosed)
            })

            c <- true
        }(releaseChannel)
    } else {
        go func(c chan<- bool) {
            defer close(c);
            c <- false
        }(releaseChannel)
    }

    return releaseChannel
}

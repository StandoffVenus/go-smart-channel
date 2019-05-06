// smart_channel provides the functionality to
// operate on channels safely. The only downside
// of the safety it provides is the loss of types
// on said channels. That said, the consumer of
// the API should be able to determine types when
// needed and so this concern is nullified.
package smart_channel

import ( "time" )

// ISmartChannel is an interface that acts as the
// center for many channels interactions. It maintains
// the channel being operated on inside, but requires
// that no outside user has direct, thread-unsafe
// access to its fields. Hence, it makes it easy to
// pass around a channel between goroutines and suffer
// no multi-threading issues.
//
// NOTE: ISmartChannel supports up to MaxInt64 references.
type ISmartChannel interface {
    // Gets a new reference to the channel,
    // allowing us to operate on it.
    // For each reference created, ISmartChannel
    // keeps track of it to know when it is safe
    // to allow sending, receiving, and closing
    // the channel.
    Get() ISmartChannelReference

    // Returns whether the underlying channel has been
    // closed. If the underlying channel has been closed,
    // all returned ISmartChannelReference's will be
    // unable send or receive, so checking the result of
    // IsClosed() is recommended before calling Get().
    IsClosed() bool
}

// ISmartChannelReference is a wrapper of individual
// channels *operations*. It is rare to create instances
// of these yourself - instead, you should be retrieving
// them from an ISmartChannel (via Get())
type ISmartChannelReference interface {
    // Returns whether
    //  1) Release() has been called
    //  2) The underlying channel is closed
    IsReleased() bool

    // Tells the ISmartChannel that this instance no
    // longer requires the underlying channel to be open,
    // decides whether or not this means to close (if possible)
    // the underlying channel (closeChannel).
    //
    // "Releasing" an ISmartChannelReference will make
    // sending or receiving impossible. Think of it as
    // closing a "child" channel.
    //
    // The returned channel will send true iff the ISmartChannel
    // is open and closeChannel is true.
    // Otherwise, the channel will send false.
    //
    // If any other ISmartChannelReference's have not called
    // Release(), the ISmartChannel will not be closed.
    // For this reason, ALWAYS call Release() on an
    // ISmartChannelReference once you are done using it.
    // Failing to do so WILL cause a memory leak.
    Release(closeChannel bool) <-chan bool

    // These are simply type asserters as functions
    // (I got sick of writing the whole type name)
    AsReceiveOnly() IReceiveOnlySmartChannelReference
    AsSendOnly() ISendOnlySmartChannelReference
    AsBoth() IBothSmartChannelReference
}

// IReceivable has the ability to receive data over
// an ISmartChannel.
type IReceivable interface {
    // Tries to receive a value from the channel.
    // timeout is used to determine how long to
    // wait for the channel operation.
    // 0 means no timeout.
    // If a negative timeout is given, a panic
    // will occur.
    //
    // Returns (nil, false, <TimeoutError) if no value
    // was received.
    // Returns (<value>, true, <TimeoutError>) if a value
    // was received.
    // The TimeoutError will be nil if no timeout occurs.
    TryReceive(timeout time.Duration) (interface{}, bool, *TimeoutError)
}

// ISendable has the ability to send data over an
// ISmartChannel.
type ISendable interface {
    // Tries to send the given value to the channel.
    // timeout is used to determine how long to
    // wait for the channel operation.
    // 0 means no timeout.
    // If a negative timeout is given, a panic
    // will occur.
    //
    // Returns (true, nil) if the value was sent;
    // returns (false, <TimeoutError>) if not.
    // The TimeoutError will be nil if no timeout occurs.
    //
    // Generally, (false, <TimeoutError>) means that
    // this instance is released or the channel is closed.
    TrySend(v interface{}, timeout time.Duration) (bool, *TimeoutError)
}

// IReceiveOnlySmartChannelReference allows an
// ISmartChannelReference receive-only capability.
type IReceiveOnlySmartChannelReference interface {
    ISmartChannelReference
    IReceivable
}

// ISendOnlySmartChannelReference allows an
// ISmartChannelReference send-only capability.
type ISendOnlySmartChannelReference interface {
    ISmartChannelReference
    ISendable
}

// IBothSmartChannelReference gives an ISmartChannelReference
// the ability to send and receive
type IBothSmartChannelReference interface {
    ISmartChannelReference
    ISendable
    IReceivable
}

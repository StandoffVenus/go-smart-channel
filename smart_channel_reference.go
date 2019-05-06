package smart_channel

import (
    "sync"
    "sync/atomic"
    "time"
)

// Creates a new ISmartChannelReference with the given smartChannel pointer.
func newSmartChannelReference(sc *smartChannel, released uint32) ISmartChannelReference {
    return &smartChannelReference {
        sc: sc,
        released: &released,
        once: new(sync.Once),
    }
}

// Implementation of IWriteOnly-, IReadOnly-, and regular ISmartChannelReference.
type smartChannelReference struct {
    sc *smartChannel
    released *uint32
    once *sync.Once
}

// Determines whether the smartChannelReference is released based
// firstly on its released field, then on whatever its smartChannelReference
// returns.
func (scr *smartChannelReference) IsReleased() bool {
    return atomic.LoadUint32(scr.released) == channelReleased ||
        scr.sc.IsClosed()
}

// Releases itself, setting its release flag to channelRelease, then
// returning a channel that holds whether the smartChannelReference
// is closed.
func (scr *smartChannelReference) Release(closeOnLast bool) <-chan bool {
    c := make(chan bool, 1)

    go func(c chan bool) {
        defer close(c)

        if atomic.LoadUint32(scr.released) != channelReleased && closeOnLast {
            scr.once.Do(func() {
                c <- (<-scr.sc.schedule_release(closeOnLast))
                atomic.StoreUint32(scr.released, channelReleased)
            })
        } else {
            c <- false
        }
    }(c)

    return c
}

// Attempts to send a value on the underlying smartChannel, timing out
// if necessary.
func (scr *smartChannelReference) TrySend(v interface{}, timeout time.Duration) (bool, *TimeoutError) {
    return scr.sc.send(v, timeout)
}

// Attempts to receive a value from the underlying smartChannel, timing out
// if necessary.
func (scr *smartChannelReference) TryReceive(timeout time.Duration) (interface{}, bool, *TimeoutError) {
    return scr.sc.receive(timeout)
}

// Returns reference as IReceiveOnlySmartChannelReference.
func (scr *smartChannelReference) AsReceiveOnly() IReceiveOnlySmartChannelReference {
    return scr
}

// Returns reference as ISendOnlySmartChannelReference.
func (scr *smartChannelReference) AsSendOnly() ISendOnlySmartChannelReference {
    return scr
}

// Returns reference as IBothSmartChannelReference.
func (scr *smartChannelReference) AsBoth() IBothSmartChannelReference {
    return scr
}

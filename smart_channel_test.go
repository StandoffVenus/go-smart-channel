package smart_channel

import (
    "reflect"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func Test_NewSmartChannel_Initializes_Correctly(t *testing.T) {
    sc := NewSmartChannel(0).(*smartChannel)

    assert.NotNil(t, sc.channel)
    assert.Equal(t, *sc.closed, channelOpen)
    assert.NotNil(t, sc.once)
    assert.NotNil(t, sc.waiter)
}

func Test_Get_Returns_New_Reference(t *testing.T) {
    sc := NewSmartChannel(0)
    scr := sc.Get()

    assert.NotNil(t, scr)
    assert.False(t, scr.IsReleased())
}

func Test_Get_Creates_A_Closed_Reference_When_Channel_Closed(t *testing.T) {
    sc := NewSmartChannel(0)

    // Wait to close channel
    <-sc.Get().Release(true)

    // Check that channel knows it's closed
    assert.Equal(t, *sc.(*smartChannel).closed, channelClosed)
    assert.True(t, sc.Get().IsReleased()) // Check that reference is released
}

func Test_IsClosed_Is_Accurate(t *testing.T) {
    sc := NewSmartChannel(0)

    assert.False(t, sc.IsClosed())

    // Close from a reference
    <-sc.Get().Release(true)

    assert.True(t, sc.IsClosed())
}

func Test_send_Returns_False_On_Close(t *testing.T) {
    sc := NewSmartChannel(0).(*smartChannel)
    <-sc.Get().Release(true) // Close from a reference

    success, err := sc.send(nil, 0)

    assert.False(t, success)
    assert.Nil(t, err)
}

func Test_send_Causes_Closing_To_Wait(t *testing.T) {
    const duration = time.Second * 3
    const value = "yo"

    if testing.Short() {
        t.Skipf("skipping test to conserve time; could wait for %v",
            duration * 2)
    }

    start := time.Now()
    sc := NewSmartChannel(0).(*smartChannel)

    // This must be done because schedule_release is supposed to be called
    // only from references, which will increment the counter per Get().
    // So, we just directly change it in this case.
    sc.waiter.Add(1)

    // Send on new thread
    go func() {
        // Wait for a fixed amount of time to prove that we are
        // waiting on the send.
        <-time.After(duration)

        // Wait forever
        sc.send(value, 0)
    }()

    // Receive on new thread.
    go sc.receive(0)

    // Wait for send/receive to happen (or, expect
    // that their operations should cause this
    // channel to send after a wait)
    // We have to sleep a little to make sure that
    // our goroutines have time to schedule sends/receives.
    time.Sleep(time.Second * 1)
    <-sc.schedule_release(true)

    // The time we waited should be > duration
    end := time.Now()
    if end.Sub(start) < duration {
        t.Errorf("Wait did not last long enough.\n\t(expected: %o >= %o)",
            end.Sub(start),
            duration)
    }

    // Make sure we ended up closing
    assert.True(t, sc.IsClosed())
}

func Test_send_Times_Out_If_Supplied_Timeout(t *testing.T) {
    const timeout = time.Second * 1

    if testing.Short() {
        t.Skipf("Waiting for send to timeout could add %v", timeout * 2)
    }

    sc := NewSmartChannel(0).(*smartChannel)

    // In the case send doesn't timeout, we don't
    // want our test to take forever.
    timer := time.After(timeout * 2)

    sendResults := WrapToChannel(sc.send, "nonZero", timeout)

    // We'll block here because it must timeout
    select {
        case success := <-sendResults:                          // Returns will be
            assert.False(t, success.(bool))                     // 1) success (bool)
            assert.NotNil(t, (<-sendResults).(*TimeoutError))   // 2) timeout (TimeoutError)
        case <-timer:
            t.Fatalf("Should have timed out by now. Over %v has passed.", timeout)
    }
}

// Helper that lets us wrap functions so that we will
// only need to wait on a channel.
func WrapToChannel(f interface{}, args ...interface{}) (<-chan interface{}) {
    c := make(chan interface{}, 0)

    go func(returns chan interface{}) {
        defer close(c)

        params := []reflect.Value{}
        for _, v := range args {
            params = append(params, reflect.ValueOf(v))
        }

        results := reflect.ValueOf(f).Call(params)

        for _, v := range results {
            returns <- v.Interface()
        }
    }(c)

    return c
}

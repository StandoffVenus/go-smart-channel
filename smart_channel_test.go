package smart_channel

import (
    "math"
    "math/rand"
    "testing"
    "time"
    "sync"
    "sync/atomic"
    "github.com/stretchr/testify/assert"
)

const Buffer int = 0

func getSmartChannel() *smartChannel {
    return NewSmartChannel(Buffer).(*smartChannel)
}

func TestGetAddsOneReference(t *testing.T) {
    s := getSmartChannel()
    result := s.Get()

    asserter := assert.New(t)

    asserter.Equal(
        int64(1),
        *s.references,
        "A call to Get() should increase the reference count by 1.")

    asserter.NotNil(result, "Get() should not return nil.")
}

func TestGetIsConcurrentSafe(t *testing.T) {
    s := getSmartChannel()
    asserter := assert.New(t)

    // Threads are bound to attempt writes at the same time
    // with this high a count
    const GoroutineCount = 1000
    group := new(sync.WaitGroup)
    group.Add(GoroutineCount)

    for i := 0; i < GoroutineCount; i++ {
        go func() {
            defer group.Done()
            time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
            _ = s.Get()
        }()
    }

    // Wait for goroutines
    group.Wait()

    asserter.Equal(
        int64(GoroutineCount),
        atomic.LoadInt64(s.references),
        "All calls to Get() should result in the same reference number.")
}

func TestGetPanicsOnOverflow(t *testing.T) {
    s := getSmartChannel()
    *s.references = math.MaxInt64

    assert.Panics(t,
        func() { s.Get() },
        "Get() should panic when reference count overflows.")
}

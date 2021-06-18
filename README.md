[![Tests](https://github.com/StandoffVenus/safe-channel/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/StandoffVenus/safe-channel/actions/workflows/go.yml)

# Safe Channel

`safe_channel` is a prackage that provides a few abstractions over channels to help with thread-safety between multiple Goroutines trying to operate on a channel. It solves the "send-on-closed" issue.

## Example

```go
func makeRequest(id int) ([]int, error) {
    // Make some HTTP request that takes awhile
}

func f(ids []int) (int, error) {
    // Buffered channel for sake of example
    send, recv, closeCh := safe_channel.OfSize(len(ids))
    defer closeCh() // Should call channel closer once done with it

    errChans := make([]<-chan error, 0, len(ids))
    for _, id := range ids {
        errChan := make(chan error, 1)
        errChans = append(errChans, errChan)

        go func(id int) {
            defer close(errChan)

            s, err := makeRequest(id)
            if err != nil {
                errChan <- err
            } else {
                for _, v := range s { send(v) }
            }
        }(id)
    }

    if err := <-safe_channel.FirstError(errChans); err != nil {
        // There could be sends happening when we
        // exit this func, but it's safe to close
        // the channel because of safe_channel.
        return 0, err
    }

    // recv() will return a channel that waits on
    // a single value before closing first (unless
    // there is no value - in which case it closes
    // first)
    sum := 0
    for i, open := <-recv(); open {
        sum += i
    }

    return sum, nil
}
```
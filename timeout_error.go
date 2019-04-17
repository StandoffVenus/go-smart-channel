package smart_channel

import (
    "fmt"
    "time"
)

// channelOperation is simply a couple of strings
// describing what channel operation was occurring
// before timeout
type channelOperation string
const (
    send channelOperation = "Send"
    receive               = "Receive"
)

// TimeoutError is simply a struct that represents the
// resulting error of a timeout.
type TimeoutError struct {
    operation channelOperation
    timeoutLength time.Duration
}

func (err *TimeoutError) Error() string {
    return fmt.Sprintf(
        "The attempted operation (%v) timed out after %v.",
        err.operation,
        err.timeoutLength,
    )
}


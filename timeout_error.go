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

// Error will return a string for this error in the format
//  "The attempted operation (<operation>) timed out after <time>."
// where <operation> is the operation on the channel (send or receive)
// and <time> is how long the specified timeout was.
func (err *TimeoutError) Error() string {
    return fmt.Sprintf(
        "The attempted operation (%v) timed out after %v.",
        err.operation,
        err.timeoutLength,
    )
}


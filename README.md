[![Tests](https://github.com/StandoffVenus/safe-channel/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/StandoffVenus/safe-channel/actions/workflows/go.yml)

# Safe Channel

`safe_channel` is a prackage that provides a few abstractions over channels to help with thread-safety between multiple Goroutines trying to operate on a channel. It solves the "send-on-closed" issue.

## Example
See [this](https://goplay.space/#vLN7D8S1wFU) working example to see what `safe_channel` can do.
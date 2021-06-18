package safe_channel

import "sync"

// FirstError will read from each channel simultaneously,
// closing the returned channel after all error channels
// are read from. The first error to be read is sent over
// the returned channel. It is valid for no provided channel
// to send an error.
//
// Note: This method expects that each channels either
// closes or sends a value. While failing to do so won't
// panic or cause an error, every Goroutine waiting for
// a provided channel to close will hang indefinitely,
// causing memory leaks.
func FirstError(chans ...<-chan error) <-chan error {
	ch := make(chan error, 1)
	oncer := sync.Once{}
	waiter := sync.WaitGroup{}

	for _, errChan := range chans {
		waiter.Add(1)
		go func(c <-chan error) {
			defer waiter.Done()
			if err := <-c; err != nil {
				oncer.Do(func() { ch <- err })
			}
		}(errChan)
	}

	go func() {
		waiter.Wait()
		close(ch)
	}()

	return ch
}

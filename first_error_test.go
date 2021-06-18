package safe_channel_test

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	safe_channel "github.com/standoffvenus/safe-channel"
	"github.com/stretchr/testify/assert"
)

func TestFirstErrorNoChannelWithError(t *testing.T) {
	errChanCount := 100
	errChans := make([]<-chan error, 0, 100)
	for i := 0; i < errChanCount; i++ {
		ec := make(chan error)
		close(ec)
		errChans = append(errChans, ec)
	}

	assert.NoError(t, <-safe_channel.FirstError(errChans...))
}

func TestFirstErrorAllChannelsError(t *testing.T) {
	errChanCount := 100
	errChans := make([]<-chan error, 0, 100)
	for i := 0; i < errChanCount; i++ {
		ec := make(chan error, 1)
		ec <- errors.New("error")
		close(ec)
		errChans = append(errChans, ec)
	}

	assert.Error(t, <-safe_channel.FirstError(errChans...))
}

func TestFirstErrorOneChannelError(t *testing.T) {
	errChanCount := 100
	errChans := make([]<-chan error, 0, 100)
	for i := 0; i < errChanCount-1; i++ {
		ec := make(chan error)
		close(ec)
		errChans = append(errChans, ec)
	}

	ec := make(chan error, 1)
	ec <- errors.New("error")
	close(ec)
	errChans = append(errChans, ec)

	assert.Error(t, <-safe_channel.FirstError(errChans...))
}

func TestFirstErrorSomeChannelError(t *testing.T) {
	errChanCount := 100
	sentErrCount := rand.Intn(errChanCount-2) + 2 // At least 2
	errChans := make([]<-chan error, 0, 100)
	for i := 0; i < sentErrCount; i++ {
		ec := make(chan error, 1)
		ec <- errors.New("error")
		close(ec)
		errChans = append(errChans, ec)
	}

	for i := 0; i < errChanCount-sentErrCount; i++ {
		ec := make(chan error)
		close(ec)
		errChans = append(errChans, ec)
	}

	assert.Error(t, <-safe_channel.FirstError(errChans...))
}

func BenchmarkFirstError(b *testing.B) {
	errChanCount := b.N
	errChans := make([]<-chan error, 0, errChanCount)
	for i := 0; i < errChanCount; i++ {
		ec := make(chan error, 1)
		ec <- errors.New("error")
		close(ec)
		errChans = append(errChans, ec)
	}

	<-safe_channel.FirstError(errChans...)
}

func ExampleFirstError() {
	errChans := make([]<-chan error, 0, 3)

	for i := 0; i < cap(errChans); i++ {
		ec := make(chan error)
		errChans = append(errChans, ec)
		go func(x int) {
			// Wait x milliseconds
			<-time.After(time.Millisecond * time.Duration(x))

			// Send an error where for loop index is the message
			ec <- errors.New(strconv.FormatInt(int64(x), 10))
		}(i)
	}

	fmt.Println(<-safe_channel.FirstError(errChans...))
	// Output: 0
}

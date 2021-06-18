package safe_channel_test

import (
	"errors"
	"math/rand"
	"testing"

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

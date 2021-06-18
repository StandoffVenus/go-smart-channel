package safe_channel_test

import (
	"fmt"
	"testing"
	"time"

	safe_channel "github.com/standoffvenus/safe-channel"
	"github.com/stretchr/testify/assert"
)

func TestSendTrueOnSuccess(t *testing.T) {
	send, recv, close := safe_channel.New()
	defer close()
	go func() { <-recv() }() // Prevent deadlock

	assert.True(t, <-send(0))
}

func TestSendFalseOnClose(t *testing.T) {
	send, _, close := safe_channel.New()
	<-close()

	assert.False(t, <-send(0))
}

func TestSendIsUnblockedOnClose(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	send, _, close := safe_channel.New()
	go func() {
		<-time.After(time.Second)
		<-close()
	}()

	before := time.Now()
	assert.False(t, <-send(0))
	assert.GreaterOrEqual(t, time.Since(before), int64(time.Second))
}

func TestSendChannelClosedAfterFirstReceive(t *testing.T) {
	send, recv, close := safe_channel.New()
	defer close()
	go func() { recv() }()

	sentChan := send(0)
	assert.True(t, <-sentChan)
	assert.False(t, <-sentChan)
}

// Run this with the -timeout flag.
func TestSendIsNotABlockingFunc(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if _, hasDeadline := t.Deadline(); !hasDeadline {
		t.Fatal("Test must be run with deadline to ensure functionality")
	}

	send, _, close := safe_channel.New()
	defer close()

	send(4)
}

func TestReceiveValueOnSuccess(t *testing.T) {
	send, recv, close := safe_channel.New()
	defer close()

	send(0x42)

	assert.Equal(t, 0x42, <-recv())
}

func TestReceiveNoValueOnClose(t *testing.T) {
	send, recv, close := safe_channel.New()
	<-close()

	send(0x42)

	v, open := <-recv()
	assert.Nil(t, v)
	assert.False(t, open)
}

func TestReceiveChannelClosedAfterFirstCall(t *testing.T) {
	send, recv, close := safe_channel.New()
	defer close()

	go func() { send(0x42) }()

	receiver := recv()
	assert.Equal(t, 0x42, <-receiver)

	v, open := <-receiver
	assert.Nil(t, v)
	assert.False(t, open)
}

func ExampleNew() {
	send, recv, close := safe_channel.New()
	defer close()

	go func() {
		<-send(0x42)
		<-send(0x999)
	}()

	firstValue, firstOpen := <-recv()
	fmt.Println(firstValue)
	fmt.Println(firstOpen)
	// Output: 66

	// Don't care what the channel is doing; we
	// can just close it.
	// We can also call it multiple times from
	// everywhere.
	go func() { close() }()

	<-close()

	secondValue, secondOpen := <-recv()
	fmt.Println(secondValue)
	fmt.Println(secondOpen)
	// Output:
	// 66
	// true
	// <nil>
	// false
}

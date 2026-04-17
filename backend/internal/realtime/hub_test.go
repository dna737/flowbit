package realtime

import (
	"context"
	"testing"
	"time"
)

func TestHubBroadcastAndUnregister(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub()
	go hub.Run(ctx)

	clientOne := &Client{send: make(chan []byte, 2)}
	clientTwo := &Client{send: make(chan []byte, 2)}

	hub.Register(clientOne)
	hub.Register(clientTwo)

	hub.Broadcast([]byte(`{"status":"running"}`))

	assertReceives(t, clientOne.send, `{"status":"running"}`)
	assertReceives(t, clientTwo.send, `{"status":"running"}`)

	hub.Unregister(clientOne)
	hub.Broadcast([]byte(`{"status":"succeeded"}`))

	assertClosed(t, clientOne.send)
	assertReceives(t, clientTwo.send, `{"status":"succeeded"}`)
}

func TestHubDropsSlowClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub()
	go hub.Run(ctx)

	slow := &Client{send: make(chan []byte, 1)}
	fast := &Client{send: make(chan []byte, 2)}

	hub.Register(slow)
	hub.Register(fast)

	hub.Broadcast([]byte(`{"status":"retrying"}`))
	hub.Broadcast([]byte(`{"status":"failed"}`))

	assertReceives(t, fast.send, `{"status":"retrying"}`)
	assertReceives(t, fast.send, `{"status":"failed"}`)
	assertReceives(t, slow.send, `{"status":"retrying"}`)
	assertClosed(t, slow.send)
}

func assertReceives(t *testing.T, ch <-chan []byte, want string) {
	t.Helper()

	select {
	case got := <-ch:
		if string(got) != want {
			t.Fatalf("want %s got %s", want, string(got))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for payload")
	}
}

func assertClosed(t *testing.T, ch <-chan []byte) {
	t.Helper()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for close")
	}
}

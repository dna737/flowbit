package realtime

import (
	"context"
	"testing"
	"time"
)

func TestHubBroadcastsOnlyToOwningUser(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub()
	go hub.Run(ctx)

	alice1 := &Client{userID: "alice", send: make(chan []byte, 2)}
	alice2 := &Client{userID: "alice", send: make(chan []byte, 2)}
	bob := &Client{userID: "bob", send: make(chan []byte, 2)}

	hub.Register(alice1)
	hub.Register(alice2)
	hub.Register(bob)

	hub.BroadcastToUser("alice", []byte(`{"status":"running"}`))

	assertReceives(t, alice1.send, `{"status":"running"}`)
	assertReceives(t, alice2.send, `{"status":"running"}`)
	assertNoReceive(t, bob.send)

	hub.Unregister(alice1)
	hub.BroadcastToUser("alice", []byte(`{"status":"succeeded"}`))

	assertClosed(t, alice1.send)
	assertReceives(t, alice2.send, `{"status":"succeeded"}`)
	assertNoReceive(t, bob.send)
}

func TestHubDropsSlowClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub()
	go hub.Run(ctx)

	slow := &Client{userID: "alice", send: make(chan []byte, 1)}
	fast := &Client{userID: "alice", send: make(chan []byte, 2)}

	hub.Register(slow)
	hub.Register(fast)

	hub.BroadcastToUser("alice", []byte(`{"status":"retrying"}`))
	hub.BroadcastToUser("alice", []byte(`{"status":"failed"}`))

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

func assertNoReceive(t *testing.T, ch <-chan []byte) {
	t.Helper()

	select {
	case got, ok := <-ch:
		if ok {
			t.Fatalf("did not expect payload, got %q", string(got))
		}
		t.Fatal("did not expect channel close")
	case <-time.After(50 * time.Millisecond):
		// no message — good
	}
}

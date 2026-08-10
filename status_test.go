package main

import (
	"errors"
	"io"
	"net"
	"testing"
)

// serveOne accepts a single connection on a fresh 127.0.0.1 listener, hands it
// to handleConnection, and returns the client side plus a channel closed once
// the daemon side is done.
func serveOne(t *testing.T) (net.Conn, chan struct{}) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	done := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			close(done)
			return
		}
		handleConnection(conn, io.Discard)
		close(done)
	}()

	client, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })

	return client, done
}

func TestHandleConnectionStatusQuery(t *testing.T) {
	origIdle := hidIdleSeconds
	hidIdleSeconds = func() (int64, error) { return 42, nil }
	t.Cleanup(func() { hidIdleSeconds = origIdle })

	opened := make(chan string, 1)
	origOpenURL := openURL
	openURL = func(line string) (string, error) { opened <- line; return "", nil }
	t.Cleanup(func() { openURL = origOpenURL })

	client, done := serveOne(t)

	if _, err := client.Write([]byte(`{"query":"status"}` + "\n")); err != nil {
		t.Fatal(err)
	}

	// One read of the whole response: the daemon writes a single line and then
	// closes, so what arrives before EOF is the complete answer.
	b, err := io.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	<-done

	if want := `{"status":"ok","hid_idle_seconds":42}` + "\n"; string(b) != want {
		t.Errorf("response = %q, want %q", string(b), want)
	}

	select {
	case line := <-opened:
		t.Errorf("a status query must not open anything, but openURL got %q", line)
	default:
	}
}

func TestHandleConnectionStatusQueryUnreadableIdleAnswersNothing(t *testing.T) {
	origIdle := hidIdleSeconds
	hidIdleSeconds = func() (int64, error) { return 0, errors.New("no ioreg here") }
	t.Cleanup(func() { hidIdleSeconds = origIdle })

	opened := make(chan string, 1)
	origOpenURL := openURL
	openURL = func(line string) (string, error) { opened <- line; return "", nil }
	t.Cleanup(func() { openURL = origOpenURL })

	client, done := serveOne(t)

	if _, err := client.Write([]byte(`{"query":"status"}` + "\n")); err != nil {
		t.Fatal(err)
	}

	b, err := io.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	<-done

	if len(b) != 0 {
		t.Errorf("response = %q, want nothing when the idle time is unreadable", string(b))
	}

	select {
	case line := <-opened:
		t.Errorf("a status query must not open anything, but openURL got %q", line)
	default:
	}
}

func TestHandleConnectionOpenRequestStillOpensAlongsideTheQuery(t *testing.T) {
	origIdle := hidIdleSeconds
	hidIdleSeconds = func() (int64, error) {
		t.Fatal("an open request must not read the HID idle time")
		return 0, nil
	}
	t.Cleanup(func() { hidIdleSeconds = origIdle })

	opened := make(chan string, 1)
	origOpenURL := openURL
	openURL = func(line string) (string, error) { opened <- line; return "", nil }
	t.Cleanup(func() { openURL = origOpenURL })

	client, done := serveOne(t)

	if _, err := client.Write([]byte(`{"url":"https://example.com/a"}` + "\n")); err != nil {
		t.Fatal(err)
	}
	<-done

	select {
	case line := <-opened:
		if line != "https://example.com/a" {
			t.Errorf("openURL got %q, want the request URL", line)
		}
	default:
		t.Error("expected the open request to reach openURL")
	}
}

package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"path/filepath"
	"testing"
)

func TestOpenerOptionsValidate(t *testing.T) {
	tt := []struct {
		test        string
		o           *OpenerOptions
		expectedErr string
	}{
		{
			"unix domain socket can be used",
			&OpenerOptions{
				Network: "unix",
				Address: filepath.Join("/", "tmp", fmt.Sprintf("%03d", rand.Intn(1000)), "opener.sock"),
			},
			"",
		},
		{
			"tcp can be used",
			&OpenerOptions{
				Network: "tcp",
				Address: "127.0.0.1:8888",
			},
			"",
		},
		{
			"udp cannot be used",
			&OpenerOptions{
				Network: "udp",
				Address: "127.0.0.1:8888",
			},
			"allowed network are: unix,tcp",
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			err := tc.o.Validate()
			if err == nil {
				if tc.expectedErr != "" {
					t.Errorf("expect err nil, but actual %q", err)
				}
			} else {
				if tc.expectedErr != err.Error() {
					t.Errorf("expect err %q, but actual %q", tc.expectedErr, err)
				}
			}
		})
	}
}

func TestHandleConnection(t *testing.T) {
	tt := []struct {
		test        string
		openURLFunc func(string) (string, error)
		data        string
		err         error
	}{
		{
			"Say nothing when successful",
			func(line string) (string, error) {
				return "pong\n", nil
			},
			"",
			io.EOF,
		},
		{
			"Sending back the logs when failure",
			func(line string) (string, error) {
				return "pong\n", errors.New("exit status 1")
			},
			"pong\n",
			nil,
		},
	}

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	defer ln.Close()

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			openURL = tc.openURLFunc

			go func() {
				conn, _ := ln.Accept()
				go handleConnection(conn, io.Discard)
			}()

			client, err := net.Dial("tcp", ln.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			if _, err := client.Write([]byte("ping\n")); err != nil {
				t.Fatal(err)
			}

			buf := make([]byte, 1024)
			n, err := client.Read(buf)
			data := string(buf[:n])
			if tc.data != data {
				t.Errorf("expect %q, but actual %q", tc.data, data)
			}

			if tc.err != err {
				t.Errorf("expect %v, but actual %v", tc.err, err)
			}
		})
	}
}

// A liveness probe (perch's opener-bridge, smart-open's health check) connects
// and closes without writing a URL. The daemon must not translate that into an
// open of "" — on macOS `open ""` opens a Finder window on the daemon's machine.
func TestHandleConnectionEmptyProbe(t *testing.T) {
	tt := []struct {
		test string
		send func(client net.Conn)
	}{
		{
			"connect-and-close probe must not open",
			func(client net.Conn) { client.Close() },
		},
		{
			"bare newline must not open",
			func(client net.Conn) {
				client.Write([]byte("\n"))
				client.Close()
			},
		},
		{
			"empty-url JSON must not open",
			func(client net.Conn) {
				client.Write([]byte(`{"url":""}` + "\n"))
				client.Close()
			},
		},
	}

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	defer ln.Close()

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			opened := make(chan string, 1)
			openURL = func(line string) (string, error) {
				opened <- line
				return "", nil
			}

			done := make(chan struct{})
			go func() {
				conn, _ := ln.Accept()
				handleConnection(conn, io.Discard)
				close(done)
			}()

			client, err := net.Dial("tcp", ln.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			tc.send(client)
			<-done

			select {
			case line := <-opened:
				t.Errorf("expected no open, but openURL was called with %q", line)
			default:
			}
		})
	}
}

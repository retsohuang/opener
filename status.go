package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"regexp"
	"strconv"
)

// statusResponse is the single line the daemon writes back to a status query.
// hid_idle_seconds is the daemon host's own idle time, so a sender can tell
// whether anyone is at this machine before routing a URL to it.
type statusResponse struct {
	Status         string `json:"status"`
	HIDIdleSeconds int64  `json:"hid_idle_seconds"`
}

// hidIdleRe matches the HIDIdleTime property ioreg prints for IOHIDSystem,
// in nanoseconds since the last human input device event on this host.
var hidIdleRe = regexp.MustCompile(`"HIDIdleTime" = ([0-9]+)`)

// hidIdleSeconds reports this host's HID idle time in whole seconds. It shells
// out to ioreg rather than linking IOKit so the release builds stay CGO-free.
// On a host without ioreg or without the property it fails, and the caller
// answers nothing rather than inventing an idle time.
var hidIdleSeconds = func() (int64, error) {
	out, err := exec.Command("ioreg", "-c", "IOHIDSystem").Output()
	if err != nil {
		return 0, err
	}

	m := hidIdleRe.FindSubmatch(out)
	if m == nil {
		return 0, errors.New("no HIDIdleTime property in ioreg output")
	}

	ns, err := strconv.ParseInt(string(m[1]), 10, 64)
	if err != nil {
		return 0, err
	}

	return ns / 1000000000, nil
}

// writeStatusResponse answers a status query with exactly one JSON line; the
// caller closes the connection afterwards. When the idle time cannot be read
// the failure is logged locally and nothing is written, so the sender's strict
// validation sees an invalid response and falls back on its own terms.
func writeStatusResponse(conn net.Conn, errOut io.Writer) {
	idle, err := hidIdleSeconds()
	if err != nil {
		fmt.Fprintf(errOut, "failed to read HID idle time: %v\n", err)
		return
	}

	b, err := json.Marshal(statusResponse{Status: "ok", HIDIdleSeconds: idle})
	if err != nil {
		fmt.Fprintf(errOut, "failed to encode status response: %v\n", err)
		return
	}

	if _, err := conn.Write(append(b, '\n')); err != nil {
		fmt.Fprintf(errOut, "failed to send status response to client: %v\n", err)
	}
}

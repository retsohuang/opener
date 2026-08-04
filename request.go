package main

import "encoding/json"

// openRequest is the structured payload the socket accepts. A line that
// fails to parse as JSON — or parses but carries no url — is treated as a
// bare URL instead, exactly as upstream did, so an un-upgraded sender still
// works against this upgraded receiver.
type openRequest struct {
	URL                string `json:"url"`
	ChromeProfileEmail string `json:"chrome_profile_email,omitempty"`
}

func parseOpenRequest(line string) openRequest {
	var req openRequest
	if err := json.Unmarshal([]byte(line), &req); err == nil && req.URL != "" {
		return req
	}
	return openRequest{URL: line}
}

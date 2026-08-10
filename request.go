package main

import "encoding/json"

// openRequest is the structured payload the socket accepts. A line that
// fails to parse as JSON is treated as a bare URL instead, exactly as
// upstream did, so an un-upgraded sender still works against this upgraded
// receiver. A line that parses as JSON is trusted as-is: an empty url in a
// valid JSON object means "nothing to open", never a bare-URL fallback.
type openRequest struct {
	URL                string `json:"url"`
	ChromeProfileEmail string `json:"chrome_profile_email,omitempty"`
}

func parseOpenRequest(line string) openRequest {
	var req openRequest
	if err := json.Unmarshal([]byte(line), &req); err == nil {
		return req
	}
	return openRequest{URL: line}
}

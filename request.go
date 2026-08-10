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
	Query              string `json:"query,omitempty"`
}

// isStatusQuery reports whether the line asked this daemon about itself
// rather than for an open: a JSON object carrying query=status and no url.
// A bare URL line always parses into URL, so it can never look like a query.
func (r openRequest) isStatusQuery() bool {
	return r.Query == "status" && r.URL == ""
}

func parseOpenRequest(line string) openRequest {
	var req openRequest
	if err := json.Unmarshal([]byte(line), &req); err == nil {
		return req
	}
	return openRequest{URL: line}
}

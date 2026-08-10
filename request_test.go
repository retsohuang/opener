package main

import "testing"

func TestParseOpenRequest(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		wantURL  string
		wantMail string
	}{
		{
			name:     "json line with both fields",
			line:     `{"url":"https://example.com/a","chrome_profile_email":"retso.huang@ikala.ai"}`,
			wantURL:  "https://example.com/a",
			wantMail: "retso.huang@ikala.ai",
		},
		{
			name:     "json line with url only",
			line:     `{"url":"https://example.com/b"}`,
			wantURL:  "https://example.com/b",
			wantMail: "",
		},
		{
			name:     "bare url line",
			line:     "https://example.com/c",
			wantURL:  "https://example.com/c",
			wantMail: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := parseOpenRequest(tc.line)
			if req.URL != tc.wantURL {
				t.Errorf("URL = %q, want %q", req.URL, tc.wantURL)
			}
			if req.ChromeProfileEmail != tc.wantMail {
				t.Errorf("ChromeProfileEmail = %q, want %q", req.ChromeProfileEmail, tc.wantMail)
			}
		})
	}
}

func TestParseOpenRequestStatusQuery(t *testing.T) {
	cases := []struct {
		name  string
		line  string
		query bool
	}{
		{
			name:  "status query",
			line:  `{"query":"status"}`,
			query: true,
		},
		{
			name:  "another query value is not a status query",
			line:  `{"query":"whatever"}`,
			query: false,
		},
		{
			name:  "a query alongside a url stays an open request",
			line:  `{"query":"status","url":"https://example.com/a"}`,
			query: false,
		},
		{
			name:  "an open request is not a status query",
			line:  `{"url":"https://example.com/a"}`,
			query: false,
		},
		{
			name:  "a bare url line is not a status query",
			line:  "https://example.com/c",
			query: false,
		},
		{
			name:  "a bare line that mentions status is not a status query",
			line:  "status",
			query: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseOpenRequest(tc.line).isStatusQuery(); got != tc.query {
				t.Errorf("isStatusQuery() = %v, want %v", got, tc.query)
			}
		})
	}
}

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

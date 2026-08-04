package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// newOpenCmd is the one-shot CLI mode: a single open that reuses
// openRequestURL, the exact resolution and launch code the daemon path
// uses, so profile behavior cannot drift between the two.
func newOpenCmd(errOut io.Writer) *cobra.Command {
	var url string
	var chromeProfileEmail string

	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open a single URL, optionally targeting a Chrome profile by account email, then exit",
		RunE: func(_ *cobra.Command, _ []string) error {
			logs, err := openRequestURL(openRequest{URL: url, ChromeProfileEmail: chromeProfileEmail})
			if logs != "" {
				fmt.Fprint(errOut, logs)
			}
			return err
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "URL to open (required)")
	cmd.Flags().StringVar(&chromeProfileEmail, "chrome-profile-email", "", "Account email to resolve a Chrome profile for (optional)")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

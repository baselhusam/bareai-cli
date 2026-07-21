package cli

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/version"
)

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	GOOS    string `json:"goos"`
	GOARCH  string `json:"goarch"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := versionInfo{
			Version: version.Version,
			Commit:  version.Commit,
			Date:    version.Date,
			GOOS:    runtime.GOOS,
			GOARCH:  runtime.GOARCH,
		}

		if opts.JSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(info)
		}

		_, err := fmt.Fprintf(cmd.OutOrStdout(), "bareai %s (%s, %s) %s/%s\n",
			info.Version, info.Commit, info.Date, info.GOOS, info.GOARCH)
		return err
	},
}

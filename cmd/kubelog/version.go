package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/didhd/kubelog/pkg"
	goVersion "go.hein.dev/go-version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var (
	date       = ""
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long: `Print the version information.

Examples:
  # Print the version information
  kubelog version
	`,
		Run: func(cmd *cobra.Command, args []string) {
			resp := goVersion.FuncWithOutput(false, pkg.Version, pkg.CommitID, date, "json")
			fmt.Print(resp)
			return
		},
	}
)

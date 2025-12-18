package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set at build time via ldflags
var (
	version = "1.0.0-dev.0"
	hash    = "local"
)

type VersionInfo struct {
	Version string
	Hash    string
}

func GetVersion() VersionInfo {
	return VersionInfo{
		Version: version,
		Hash:    hash,
	}
}

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			v := GetVersion()
			fmt.Printf("%s (%s)\n", v.Version, v.Hash)
		},
	}
}

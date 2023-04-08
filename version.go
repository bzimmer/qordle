package qordle

import (
	"runtime"

	"github.com/urfave/cli/v2"
)

var (
	// buildVersion of the package
	buildVersion = "development"
	// buildTime of the package
	buildTime = "now"
	// buildCommit of the package
	buildCommit = "snapshot"
	// buildBuilder of the package
	buildBuilder = "local"
)

func CommandVersion() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Show the version information of the binary",
		Action: func(c *cli.Context) error {
			return Runtime(c).Encoder.Encode(map[string]string{
				"version":  buildVersion,
				"datetime": buildTime,
				"commit":   buildCommit,
				"builder":  buildBuilder,
				"go":       runtime.Version(),
			})
		},
	}
}

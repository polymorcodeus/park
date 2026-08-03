// Command park is a standalone parking-lot for markdown notes, organized
// as IPAA — Inbox / Projects / Areas / Archive, a PARA variant.

package main

import (
	_ "embed"
	"strings"

	"github.com/polymorcodeus/park/cmd/park"
)

//go:embed VERSION
var versionFile string

// version and buildTime are set by GoReleaser via ldflags at build time.
var (
	version   string
	buildTime string
)

// use embedded VERSION file for local `go install`d version
func init() {
	if version == "" {
		version = strings.TrimSpace(versionFile)
	}
}

func main() {
	cmd.SetVersion(version)
	cmd.SetBuildTime(buildTime)
	cmd.Main()
}

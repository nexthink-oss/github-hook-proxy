package main

import (
	"fmt"
	"os"

	"github.com/nexthink-oss/github-hook-proxy/cmd"
)

var (
	version = "snapshot"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cmd := cmd.New()
	cmd.Version = fmt.Sprintf("%s-%s (built %s)", version, commit, date)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

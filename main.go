package main

import (
	"github.com/lorislab/changelog/cmd"
)

var (
	// Used for flags.
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(cmd.BuildVersion{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
}

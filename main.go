package main

import (
	"os"

	"github.com/ThomasCrouzet/inframap-d2/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

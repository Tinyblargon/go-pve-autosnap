package main

import (
	"go-pve-autosnap/internal/cli"
	_ "go-pve-autosnap/internal/cli/command"
	"log"
	"os"
)

var version string // Set by build script

func main() {
	if err := cli.Execute(version); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

package main

import (
	"os"

	"github.com/DNSGeek/filebrowser/v3/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"vchan.in/docker-health/cmd"
)

var Version string

func main() {
	Version = "1.0.0"
	cmd.Server(Version)
}

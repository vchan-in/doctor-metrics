package main

import (
	"vchan.in/doctor-metrics/cmd"
)

var Version string

func main() {
	Version = "1.2.1"
	cmd.Server(Version)
}

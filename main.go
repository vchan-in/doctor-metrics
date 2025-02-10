package main

import (
	"vchan.in/doctor-metrics/cmd"
)

var Version string

func main() {
	Version = "1.2.0"
	cmd.Server(Version)
}

//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

var common = []string{
	"-in", "./data/access.log",
	"-access-log",
	"-k", "20",
	"-tick", "1m",
	"-window", "1h",
	"-json-timestamp-layout", "02/Jan/2006:15:04:05 -0700",
	"-view-split", "30",
	"-stats",
	"-stats-window", "256",
	"-alt-screen=false",
}

var modes = map[string][]string{
	"recommended": {
		"-replay", "-replay-speed", "500", "-replay-max-sleep", "10ms",
		"-plot-fps", "15", "-items-fps", "2", "-item-counts-fps", "2",
		"-search", "-full-refresh", "3s", "-partial-size", "30",
	},
	"fast": {
		"-plot-fps", "5", "-items-fps", "1", "-item-counts-fps", "0",
		"-search=false", "-full-refresh", "0",
	},
}

func main() {
	mode := "recommended"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	extra, ok := modes[mode]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown mode %q (use \"recommended\" or \"fast\")\n", mode)
		os.Exit(1)
	}

	build := exec.Command("go", "build", "-mod=vendor", "-o", "./logspeed.exe", "./program")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		os.Exit(1)
	}

	run := exec.Command("./logspeed.exe", append(common, extra...)...)
	run.Stdin = os.Stdin
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	if err := run.Run(); err != nil {
		os.Exit(1)
	}
}

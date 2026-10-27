package main

import (
	"fmt"
	"runtime/debug"
)

// version is empty unless set with -ldflags "-X main.version=...".
var version string

func main() {
	fmt.Printf("version: %q\n", version)

	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("no build info")
		return
	}
	fmt.Println("go:", info.GoVersion)
	fmt.Println("path:", info.Path)
	fmt.Println("module:", info.Main.Path, info.Main.Version)
	for _, s := range info.Settings {
		if s.Key == "GOOS" || s.Key == "GOARCH" || s.Key == "-ldflags" {
			fmt.Printf("%s=%s\n", s.Key, s.Value)
		}
	}
}

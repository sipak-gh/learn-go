//go:build !windows

package main

import "runtime"

func platform() string { return "not Windows (" + runtime.GOOS + ")" }

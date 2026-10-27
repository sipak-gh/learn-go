package main

// The _windows suffix in the file name is an implicit build
// constraint: this file is only compiled when GOOS is windows.
func platform() string { return "Windows" }

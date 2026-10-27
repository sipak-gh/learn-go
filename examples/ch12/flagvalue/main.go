package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

// listFlag collects the values of a flag that may be repeated.
type listFlag []string

func (l *listFlag) String() string {
	if l == nil { // the flag package may call String on a zero value
		return ""
	}
	return strings.Join(*l, ",")
}

func (l *listFlag) Set(s string) error {
	*l = append(*l, s)
	return nil
}

// level is an enum flag that accepts only three words.
type level int

const (
	low level = iota
	medium
	high
)

var levelNames = []string{"low", "medium", "high"}

func (l level) String() string { return levelNames[l] }

func (l *level) Set(s string) error {
	for i, name := range levelNames {
		if s == name {
			*l = level(i)
			return nil
		}
	}
	return errors.New("must be low, medium, or high")
}

// Compile-time checks: the pointer types implement flag.Value.
var (
	_ flag.Value = (*listFlag)(nil)
	_ flag.Value = (*level)(nil)
)

func main() {
	var tags listFlag
	lvl := medium // the default
	flag.Var(&tags, "tag", "add a `tag` (repeatable)")
	flag.Var(&lvl, "level", "low, medium, or high")

	// Func is a shortcut: a flag that calls a function with each value.
	var upper []string
	flag.Func("up", "add `text` in upper case", func(s string) error {
		upper = append(upper, strings.ToUpper(s))
		return nil
	})
	flag.Parse()

	fmt.Printf("tags=%q level=%v up=%q\n", tags, lvl, upper)
}

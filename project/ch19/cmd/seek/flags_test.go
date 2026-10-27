package main

import (
	"errors"
	"flag"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestParseArgs(t *testing.T) {
	t.Setenv("SEEK_COLOR", "auto") // don't depend on the user's environment
	cpus := runtime.NumCPU()
	tests := []struct {
		name string
		args []string
		want options
	}{
		{"defaults", []string{"fox"},
			options{pattern: "fox", paths: []string{"."}, workers: cpus}},
		{"short flags", []string{"-i", "-n", "-c", "fox", "a", "b"},
			options{pattern: "fox", paths: []string{"a", "b"}, ignoreCase: true,
				lineNumbers: true, count: true, workers: cpus}},
		{"long flags", []string{"--invert-match", "--max-count", "2", "--color=never", "fox"},
			options{pattern: "fox", paths: []string{"."}, invert: true, maxCount: 2,
				color: colorNever, workers: cpus}},
		{"values", []string{"-j", "1", "--timeout", "1m30s", "--glob", "*.go", "fox"},
			options{pattern: "fox", paths: []string{"."}, workers: 1,
				timeout: 90 * time.Second, glob: "*.go"}},
		{"double dash", []string{"--", "-v", "-"},
			options{pattern: "-v", paths: []string{"-"}, workers: cpus}},
		{"flags stop at pattern", []string{"fox", "-n"},
			options{pattern: "fox", paths: []string{"-n"}, workers: cpus}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs(%q): %v", tt.args, err)
			}
			// reflect.DeepEqual compares any two values field by field,
			// following slices; == can't compare structs holding slices.
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseArgs(%q)\n got %+v\nwant %+v", tt.args, got, tt.want)
			}
		})
	}
}

func TestParseArgsErrors(t *testing.T) {
	t.Setenv("SEEK_COLOR", "auto")
	tests := []struct {
		args []string
		want string
	}{
		{nil, "missing PATTERN"},
		{[]string{"-x", "fox"}, "flag provided but not defined: -x"},
		{[]string{"-in", "fox"}, "flag provided but not defined: -in"},
		{[]string{"-m", "two", "fox"}, `invalid value "two" for flag -m: parse error`},
		{[]string{"-m", "-1", "fox"}, "invalid value for -m: -1 is negative"},
		{[]string{"-j", "0", "fox"}, "invalid value for -j: must be at least 1"},
		{[]string{"--color", "pink", "fox"},
			`invalid value "pink" for flag -color: must be "auto", "always", or "never"`},
		{[]string{"--glob", "[", "fox"}, `invalid --glob pattern "[": syntax error in pattern`},
		{[]string{"--timeout", "-1s", "fox"}, "invalid value for --timeout: -1s is negative"},
	}
	for _, tt := range tests {
		_, err := parseArgs(tt.args)
		var usageErr *usageError
		if !errors.As(err, &usageErr) || err.Error() != tt.want {
			t.Errorf("parseArgs(%q) error = %v, want usage error %q", tt.args, err, tt.want)
		}
	}
}

func TestParseArgsHelp(t *testing.T) {
	for _, arg := range []string{"-h", "--help"} {
		if _, err := parseArgs([]string{arg}); !errors.Is(err, flag.ErrHelp) {
			t.Errorf("parseArgs(%q) error = %v, want flag.ErrHelp", arg, err)
		}
	}
}

func TestParseArgsEnv(t *testing.T) {
	// t.Setenv restores the old value when the test ends. Because the
	// environment is shared by the whole program, a test that uses it
	// can't call t.Parallel.
	t.Setenv("SEEK_COLOR", "always")
	opts, err := parseArgs([]string{"fox"})
	if err != nil || opts.color != colorAlways {
		t.Errorf("got color %v, error %v; want always", opts.color, err)
	}
	opts, err = parseArgs([]string{"--color", "never", "fox"})
	if err != nil || opts.color != colorNever {
		t.Errorf("flag should override $SEEK_COLOR: got %v, error %v", opts.color, err)
	}

	t.Setenv("SEEK_COLOR", "pink")
	want := `invalid SEEK_COLOR "pink": must be "auto", "always", or "never"`
	if _, err := parseArgs([]string{"fox"}); err == nil || err.Error() != want {
		t.Errorf("got error %v, want %q", err, want)
	}
}

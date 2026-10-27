package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// go test ./cmd/seek -update rewrites the golden files with the current
// output. Review the changes with git diff before committing them!
var update = flag.Bool("update", false, "update golden files")

// TestRun runs the whole program, as the shell would, against the
// fixtures in the project's testdata directory. The expected output of
// each case is kept in a "golden file": cmd/seek/testdata/NAME.golden.
func TestRun(t *testing.T) {
	t.Setenv("SEEK_COLOR", "auto")
	golden, err := filepath.Abs("testdata") // before t.Chdir below
	if err != nil {
		t.Fatal(err)
	}
	// Run from the project root so the output shows paths like
	// testdata/poem.txt, just as in the book. t.Chdir changes back when
	// the test ends.
	t.Chdir("../..")

	tests := []struct {
		name     string
		args     []string
		stdin    string
		wantCode int
	}{
		{"fox", []string{"fox", "testdata"}, "", exitOK},
		{"ignore-case-numbers", []string{"-i", "-n", "fox", "testdata"}, "", exitOK},
		{"count-hidden", []string{"-c", "--hidden", "fox", "testdata"}, "", exitOK},
		{"files-with-matches", []string{"-l", "-i", "fox", "testdata"}, "", exitOK},
		{"regex-glob", []string{"-E", "--glob", "*.txt", "qu?ick|lazy", "testdata"}, "", exitOK},
		{"json", []string{"--json", "-i", "fox", "testdata/notes.md"}, "", exitOK},
		{"color", []string{"--color", "always", "-i", "fox", "testdata/poem.txt"}, "", exitOK},
		{"stdin-invert", []string{"-v", "fox", "-"}, "a fox\na dog\n", exitOK},
		{"no-match", []string{"zebra", "testdata"}, "", exitNoMatch},
		{"missing-file", []string{"fox", "testdata/poem.txt", "nope.txt"}, "", exitError},
		{"usage", []string{"-x", "fox"}, "", exitError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(t.Context(), tt.args, strings.NewReader(tt.stdin), &stdout, &stderr)
			if code != tt.wantCode {
				t.Errorf("exit code %d, want %d", code, tt.wantCode)
			}
			got := stdout.String()
			if stderr.Len() > 0 {
				got += "-- stderr --\n" + stderr.String()
			}
			checkGolden(t, filepath.Join(golden, tt.name+".golden"), got)
		})
	}
}

// checkGolden compares got with the contents of the file at path, or
// with -update, writes got to the file.
func checkGolden(t *testing.T, path, got string) {
	t.Helper()
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err) // run with -update to create it
	}
	if got != string(want) {
		t.Errorf("output differs from %s\n--- got ---\n%s--- want ---\n%s", path, got, want)
	}
}

// TestRunCanceled checks that a canceled search reports why and fails.
func TestRunCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var stdout, stderr bytes.Buffer
	code := run(ctx, []string{"fox", "-"}, strings.NewReader("a fox\n"), &stdout, &stderr)
	if code != exitError || stderr.String() != "seek: context canceled\n" {
		t.Errorf("got code %d, stderr %q; want 2 and a cancellation message", code, stderr.String())
	}
}

// TestRunWorkers checks that the output doesn't depend on how many
// files are searched at once.
func TestRunWorkers(t *testing.T) {
	t.Chdir("../..")
	var want string
	for _, j := range []string{"1", "2", "8"} {
		var stdout, stderr bytes.Buffer
		run(t.Context(), []string{"-j", j, "-n", "--hidden", "o", "testdata"}, nil, &stdout, &stderr)
		if j == "1" {
			want = stdout.String()
		} else if stdout.String() != want {
			t.Errorf("-j %s output differs from -j 1:\n%s", j, stdout.String())
		}
	}
}

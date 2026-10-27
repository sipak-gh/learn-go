package files

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile creates a file holding text in a fresh temporary directory
// and returns its path.
func writeFile(t *testing.T, text string) string {
	t.Helper() // report failures at the caller's line
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCountLines(t *testing.T) {
	t.Cleanup(func() { t.Log("cleanup runs after the test and its subtests") })

	path := writeFile(t, "one\ntwo\nthree\n")
	n, err := CountLines(path)
	if err != nil {
		t.Fatal(err) // stop: n is meaningless if err != nil
	}
	if n != 3 {
		t.Errorf("CountLines = %d, want 3", n)
	}
	t.Log("checked", path)
}

func TestCountLinesTestdata(t *testing.T) {
	// go test runs in the package's directory, so testdata/ is right here.
	n, err := CountLines(filepath.Join("testdata", "poem.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Errorf("CountLines = %d, want 4", n)
	}
}

func TestCountLinesMissing(t *testing.T) {
	_, err := CountLines(filepath.Join(t.TempDir(), "missing.txt"))
	if !os.IsNotExist(err) {
		t.Errorf("got error %v, want a not-exist error", err)
	}
}

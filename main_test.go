package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "no argument", args: nil, want: exitUsage},
		{name: "too many arguments", args: []string{"a", "b"}, want: exitUsage},
		{name: "unknown flag", args: []string{"--nope", "a"}, want: exitUsage},
		{name: "unknown format", args: []string{"--format", "xml", "a"}, want: exitUsage},
		{name: "negative depth", args: []string{"--depth", "-1", "a"}, want: exitUsage},
		{name: "help", args: []string{"--help"}, want: exitOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if got := run(tc.args, &stdout, &stderr); got != tc.want {
				t.Errorf("run(%v) = %d, want %d (stderr: %s)", tc.args, got, tc.want, stderr.String())
			}
		})
	}
}

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run([]string{"--version"}, &stdout, &stderr); got != exitOK {
		t.Fatalf("run(--version) = %d, want %d", got, exitOK)
	}
	if !strings.HasPrefix(stdout.String(), "elftree ") {
		t.Errorf("unexpected version output: %q", stdout.String())
	}
}

func TestRunNonElfFile(t *testing.T) {
	name := filepath.Join(t.TempDir(), "not-an-elf.txt")
	if err := os.WriteFile(name, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if got := run([]string{name}, &stdout, &stderr); got != exitError {
		t.Fatalf("run(%s) = %d, want %d", name, got, exitError)
	}
	if !strings.Contains(stderr.String(), "elftree:") {
		t.Errorf("missing error message: %q", stderr.String())
	}
}

// findSystemBinary returns a dynamically linked ELF binary of the host, if any.
func findSystemBinary(t *testing.T) string {
	t.Helper()

	if runtime.GOOS != "linux" {
		t.Skip("no ELF binary available on this platform")
	}
	for _, name := range []string{"/bin/ls", "/usr/bin/ls", "/bin/cat"} {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}
	t.Skip("no system binary found")
	return ""
}

func TestRunFormats(t *testing.T) {
	binary := findSystemBinary(t)

	for _, format := range []string{"tree", "flat", "json"} {
		t.Run(format, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if got := run([]string{"--format", format, "--path", binary}, &stdout, &stderr); got != exitOK {
				t.Fatalf("run(%s) = %d, want %d (stderr: %s)", format, got, exitOK, stderr.String())
			}
			if stdout.Len() == 0 {
				t.Error("no output produced")
			}
			if format == "json" && !strings.HasPrefix(strings.TrimSpace(stdout.String()), "{") {
				t.Errorf("unexpected JSON output: %q", stdout.String())
			}
		})
	}
}

func TestRunVerbose(t *testing.T) {
	binary := findSystemBinary(t)

	var stdout, stderr bytes.Buffer
	args := []string{"-v", "--segments", "--sections", "--dynamic", "--symbols", binary}
	if got := run(args, &stdout, &stderr); got != exitOK {
		t.Fatalf("run(%v) = %d, want %d (stderr: %s)", args, got, exitOK, stderr.String())
	}
	for _, want := range []string{"total dependency:", "Program headers:", "Section headers:", "Dynamic info:", "Dynamic symbols:"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("output does not contain %q", want)
		}
	}
}

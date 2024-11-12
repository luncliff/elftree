package main

import (
	"os"
	"runtime"
	"testing"
)

func TestLibraryDirectoryExists(t *testing.T) {
	if len(deflib) == 0 {
		t.Error("Default library directories are required")
	}
	for i := range deflib {
		folder := deflib[i]
		_, err := os.Stat(folder)
		if err != nil && os.IsExist(err) {
			continue
		}
		t.Log(folder, err)
	}
}

func TestReadLdSoConfBypassInWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
	}
	paths := make([]string, 0)
	outputs := readLdSoConf("", paths)
	if len(paths) != len(outputs) {
		t.FailNow()
	}
}

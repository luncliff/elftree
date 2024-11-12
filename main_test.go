package main

import (
	"os"
	"testing"
)

func TestLibraryDirectoryExists(t *testing.T) {
	if len(deflib) == 0 {
		t.Error("Default library directories are required")
	}
	for i := range deflib {
		folder := deflib[i]
		if _, err := os.Stat(folder); os.IsNotExist(err) {
			t.Log("not exist:", folder)
		}
	}
}

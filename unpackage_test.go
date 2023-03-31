package unitypackage

import (
	"os"
	"path/filepath"
	"testing"
)

func getTestDataPath() string {
	return filepath.Join(".", "tests", "test-data")
}

func TestPackageExtract(t *testing.T) {
	dir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	unityPath := filepath.Join(getTestDataPath(), "test.unitypackage")

	err = UnPackage(unityPath, dir)
	if err != nil {
		t.Fatalf("Failed to unpackage unitypackage: %v", err)
		return
	}

	got := isDir(dir)
	want := true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	got = isDir(filepath.Join(dir, "Assets"))
	want = true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	got = !isDir(filepath.Join(dir, "Assets", "test.txt"))
	want = true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	data, err := os.ReadFile(filepath.Join(dir, "Assets", "test.txt"))
	if string(data) != "testing" {
		t.Errorf("got %v, wanted %v", string(data), "testing")
	}
}

func TestPackageExtractWithLeadingDots(t *testing.T) {
	dir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	unityPath := filepath.Join(getTestDataPath(), "testLeadingDots.unitypackage")

	err = UnPackage(unityPath, dir)
	if err != nil {
		t.Fatalf("Failed to unpackage unitypackage: %v", err)
		return
	}

	got := isDir(dir)
	want := true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	got = isDir(filepath.Join(dir, "Assets"))
	want = true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	got = !isDir(filepath.Join(dir, "Assets", "test.txt"))
	want = true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	txtPath := filepath.Join(dir, "Assets", "test.txt")
	data, err := os.ReadFile(txtPath)
	if err != nil {
		t.Errorf("Failed to read file %s, err %v", txtPath, err)
		return
	}

	if string(data) != "testing" {
		t.Errorf("got %v, wanted %v", string(data), "testing")
	}
}

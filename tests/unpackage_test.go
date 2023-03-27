package tests

import (
	"go-unitypackage"
	"os"
	"path/filepath"
	"testing"
)

func isDir(path string) bool {
	dirInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return dirInfo.IsDir()
}

func TestPackageExtract(t *testing.T) {
	dir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	unityPath := filepath.Join(".", "test-data", "test.unitypackage")

	err = main.UnPackage(unityPath, dir)
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

	unityPath := filepath.Join(".", "test-data", "testLeadingDots.unitypackage")

	err = main.UnPackage(unityPath, dir)
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

func TestPackageExtractWithUnicodePath(t *testing.T) {
	dir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	unityPath := filepath.Join(".", "test-data", "testo.unitypackage")

	err = main.UnPackage(unityPath, dir)
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

	got = !isDir(filepath.Join(dir, "Assets", "テスト.txt"))
	want = true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	txtPath := filepath.Join(dir, "Assets", "テスト.txt")
	gotData, err := os.ReadFile(txtPath)
	if err != nil {
		t.Errorf("Failed to read file %s, err %v", txtPath, err)
		return
	}
	wantData := "テスト, but with katakana!"
	if string(gotData) != wantData {
		t.Errorf("got %v, wanted %v", string(gotData), wantData)
	}
}

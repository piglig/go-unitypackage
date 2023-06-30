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
	if !got {
		t.Errorf("dir %s got %v", dir, got)
	}

	assetsDir := filepath.Join(dir, "Assets")
	got = isDir(assetsDir)
	if !got {
		t.Errorf("dir %s got %v", assetsDir, got)
	}

	txtFilePath := filepath.Join(dir, "Assets", "test.txt")
	got = !isDir(txtFilePath)
	if !got {
		t.Errorf("file %s got %v", txtFilePath, got)
	}

	data, err := os.ReadFile(txtFilePath)
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
	if !got {
		t.Errorf("dir %s got %v", dir, got)
	}

	assetsDir := filepath.Join(dir, "Assets")
	got = isDir(assetsDir)
	if !got {
		t.Errorf("dir %s got %v", assetsDir, got)
	}

	txtFilePath := filepath.Join(dir, "Assets", "test.txt")
	got = !isDir(txtFilePath)
	if !got {
		t.Errorf("file %s got %v", txtFilePath, got)
	}

	data, err := os.ReadFile(txtFilePath)
	if err != nil {
		t.Errorf("Failed to read file %s, err %v", txtFilePath, err)
		return
	}

	if string(data) != "testing" {
		t.Errorf("got %v, wanted %v", string(data), "testing")
	}
}

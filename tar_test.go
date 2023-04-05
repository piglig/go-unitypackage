package unitypackage

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestTarHeader(t *testing.T) {
	// create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// create a temporary file for testing
	tmpFile, err := os.CreateTemp(tmpDir, "test")
	if err != nil {
		t.Fatal(err)
	}

	// write some content to the temporary file
	content := []byte("Hello, world!")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}

	// create a tar.Writer and write the header for the temporary file
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err = tarHeader(tmpDir, tmpFile.Name(), tw)
	if err != nil {
		t.Fatalf("tarHeader(%q, %q, tw) failed with %v", tmpDir, tmpFile.Name(), err)
	}
	defer tw.Close()

	// verify the tar header
	tarReader := tar.NewReader(&buf)
	header, err := tarReader.Next()
	if err != nil {
		t.Fatalf("tarReader.Next() failed with %v", err)
	}
	if header.Name != filepath.Base(tmpFile.Name()) {
		t.Errorf("header.Name=%q, want %q", header.Name, filepath.Base(tmpFile.Name()))
	}
	if header.Size != int64(len(content)) {
		t.Errorf("header.Size=%d, want %d", header.Size, len(content))
	}
}

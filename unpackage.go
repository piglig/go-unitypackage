// Package utils implements utility unitypackage for unpackage or package
package unitypackage

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func UnPackage(packagePath, outputPath string) error {
	md5Dir, err := os.MkdirTemp("", "md5")
	if err != nil {
		return err
	}
	defer os.RemoveAll(md5Dir)

	tempDir, err := extractAll(packagePath, md5Dir)
	if err != nil {
		return fmt.Errorf("extractAll %w", err)
	}

	dirs, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			pathNameFilePath := filepath.Join(tempDir, dir.Name())
			assetFilePath := pathNameFilePath
			if runtime.GOOS == "windows" {
				pathNameFilePath = filepath.Join(pathNameFilePath, "pathname")
			} else {
				pathNameFilePath = filepath.Join(pathNameFilePath, "pathname")
			}

			if _, err = os.Stat(pathNameFilePath); err != nil {
				continue
			}

			if runtime.GOOS == "windows" {
				assetFilePath = filepath.Join(assetFilePath, "asset")
			} else {
				assetFilePath = filepath.Join(assetFilePath, "asset")
			}
			if _, err = os.Stat(assetFilePath); err != nil {
				continue
			}

			pathNameByte, err := os.ReadFile(pathNameFilePath)
			if err != nil {
				return fmt.Errorf("UnPackage os.ReadFile %w", err)

			}

			pathName := strings.TrimSpace(string(pathNameByte))
			if runtime.GOOS == "windows" {
				pathName = regexp.MustCompile(`[>:"|?*]`).ReplaceAllString(pathName, "_")
			}

			outputFile := filepath.Join(outputPath, pathName)
			err = os.MkdirAll(filepath.Dir(outputFile), 0777)
			if err != nil {
				return fmt.Errorf("UnPackage os.MkdirAll %w", err)
			}

			if err = copyFile(assetFilePath, outputFile); err != nil {
				return fmt.Errorf("UnPackage copyFile %w", err)
			}
		}
	}

	assetDir := getAssetsRootPath(outputPath)
	return preProcessFilesInPath(assetDir, "./")
}

// extractAll takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func extractAll(unityPackagePath, outputPath string) (output string, err error) {
	unityPackage, err := os.Open(unityPackagePath)
	if err != nil {
		return "", err
	}
	defer unityPackage.Close()

	gzr, err := gzip.NewReader(unityPackage)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		switch {
		// if no more files are found return
		case err == io.EOF:
			return outputPath, nil
		case err != nil:
			return "", fmt.Errorf("extractAll err %w", err)
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		header.Name = filepath.Clean(header.Name)
		target := header.Name
		absFlag := filepath.IsAbs(header.Name)
		if !absFlag {
			absFlag = false
			target = filepath.Join(outputPath, header.Name)
		}
		target = filepath.Clean(target)

		log.Println("target:", target, "header: ", header.Name)

		// check the file type
		switch header.Typeflag {
		// if it's a dir and doesn't exist create it
		case tar.TypeDir:
			if err = os.MkdirAll(target, 0755); err != nil {
				return "", fmt.Errorf("extractAll tar.TypeDir %w", err)
			}
		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return "", fmt.Errorf("extractAll tar.TypeReg %w", err)
			}

			if absFlag {
				tempOutput := outputPath

				fileBase := filepath.Base(target)
				dir := filepath.Dir(target)
				upDir := filepath.Dir(dir)
				upDir = strings.ReplaceAll(dir, upDir, "")
				outputDir := filepath.Join(tempOutput, upDir)
				tempOutput = filepath.Join(outputDir, fileBase)

				if err = os.MkdirAll(outputDir, 0755); err != nil {
					return "", fmt.Errorf("extractAll tar.TypeReg os.MkdirAll %w", err)
				}

				if err = copyFile(target, tempOutput); err != nil {
					return "", fmt.Errorf("extractAll tar.TypeReg copyFile %w", err)

				}
			} else {
				// copy over contents
				if _, err := io.Copy(f, tr); err != nil {
					f.Close()
					return "", fmt.Errorf("extractAll tar.TypeReg io.Copy %w", err)
				}
			}

			f.Close()
		}
	}
}

package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// UnPackage extract unitypackage content
// packagePath is unitypackage file path, output is extract content storage path,
// tempPath is temp path for deal with unitypackage
func UnPackage(packagePath, output, tempPath string) error {
	packagePath = filepath.Clean(packagePath)
	output = filepath.Clean(output)
	tempPath = filepath.Clean(tempPath)

	os.RemoveAll(output)
	if err := os.MkdirAll(output, 0777); err != nil {
		return err
	}

	os.RemoveAll(tempPath)
	if err := os.MkdirAll(tempPath, 0777); err != nil {
		return err
	}

	tempDir, err := extractAll(packagePath, tempPath)
	if err != nil {
		fmt.Println("000000 err", err)
		return err
	}

	fmt.Println(tempDir)

	dirs, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			pathNameFilePath := filepath.Join(tempDir, dir.Name(), "pathname")
			if _, err = os.Stat(pathNameFilePath); err != nil {
				continue
			}

			assetFilePath := filepath.Join(tempDir, dir.Name(), "asset")
			if _, err = os.Stat(assetFilePath); err != nil {
				continue
			}

			pathNameByte, err := ioutil.ReadFile(pathNameFilePath)
			if err != nil {
				fmt.Println("11111111111 err", err)
				return err
			}

			pathName := string(pathNameByte)
			pathName = strings.TrimSpace(pathName)
			if runtime.GOOS == "windows" {
				var re = regexp.MustCompile(`[>:"|?*]`)
				pathName = re.ReplaceAllString(pathName, "_")
			}

			outputParent := filepath.Join(output, pathName)
			fp := filepath.Dir(outputParent)
			if _, err = os.Stat(fp); err != nil {
				if os.IsNotExist(err) {
					err = os.MkdirAll(fp, 0777)
					if err != nil {
						fmt.Println("222222222222 err", err)
						return err
					}
				}
			}

			err = MoveFile(assetFilePath, outputParent)
			if err != nil {
				fmt.Println("33333333 err", err)
				return err
			}
		}
	}

	return nil
}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func extractAll(unityPackagePath, outputPath string) (string, error) {
	//tempDir, err := ioutil.TempDir("", "temp-unpackages")
	//if err != nil {
	//	return "", err
	//}
	unityPackage, err := os.Open(unityPackagePath)
	if err != nil {
		return "", err
	}

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

		// return any other error
		case err != nil:
			fmt.Println("1111111", err)
			return "", err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := header.Name
		absFlag := true
		if !filepath.IsAbs(header.Name) {
			absFlag = false
			target = filepath.Join(outputPath, header.Name)
		}
		target = filepath.Clean(target)

		fmt.Println("target:", target, "header: ", header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					fmt.Println("2222222", err)
					return "", err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				fmt.Println("333333", err)
				return "", err
			}

			if absFlag {
				tempOutput := outputPath

				fileBase := filepath.Base(target)
				dir := filepath.Dir(target)
				upDir := filepath.Dir(dir)
				upDir = strings.ReplaceAll(dir, upDir, "")
				outputDir := filepath.Join(tempOutput, upDir)
				tempOutput = filepath.Join(outputDir, fileBase)

				if _, err := os.Stat(outputDir); err != nil {
					if err := os.MkdirAll(outputDir, 0755); err != nil {
						fmt.Println("77777", err)
						return "", err
					}
				}
				if err = CopyFile(target, tempOutput); err != nil {
					fmt.Println(err)
					return "", err
				}
			} else {
				// copy over contents
				if _, err := io.Copy(f, tr); err != nil {
					fmt.Println("44444444", err)
					return "", err
				}
			}

			f.Close()
		}
	}
}

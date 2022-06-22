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

func UnPackage(packagePath, output string) error {
	//packagePath = filepath.Clean(packagePath)
	// output = filepath.Clean(output)
	// tempPath = filepath.Clean(tempPath)
	//
	// os.RemoveAll(output)
	// if err := os.MkdirAll(output, 0777); err != nil {
	// 	return err
	// }
	//
	// os.RemoveAll(tempPath)
	// if err := os.MkdirAll(tempPath, 0777); err != nil {
	// 	return err
	// }

	md5Dir, err := ioutil.TempDir("", "md5")
	if err != nil {
		fmt.Println(err)
		return err
	}

	tempDir, err := extractAll(packagePath, md5Dir)
	if err != nil {
		fmt.Println("extractAll err", err)
		return err
	}

	dirs, err := ioutil.ReadDir(tempDir)
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

			pathNameByte, err := ioutil.ReadFile(pathNameFilePath)
			if err != nil {
				fmt.Println("UnPackage ioutil.ReadFile err", err)
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
						fmt.Println("UnPackage os.MkdirAll err", err)
						return err
					}
				}
			}

			if err = MoveFile(assetFilePath, outputParent); err != nil {
				fmt.Println("UnPackage MoveFile err", err)
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
			fmt.Println("extractAll err", err)
			return "", err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		header.Name = filepath.Clean(header.Name)
		target := header.Name
		absFlag := true
		if !filepath.IsAbs(header.Name) {
			absFlag = false
			target = filepath.Join(outputPath, header.Name)
		}
		target = filepath.Clean(target)

		fmt.Println("target:", target, "header: ", header.Name)

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err = os.Stat(target); err != nil {
				if err = os.MkdirAll(target, 0755); err != nil {
					fmt.Println("extractAll tar.TypeDir", err)
					return "", err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				fmt.Println("extractAll tar.TypeReg", err)
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

				if _, err = os.Stat(outputDir); err != nil {
					if err = os.MkdirAll(outputDir, 0755); err != nil {
						fmt.Println("extractAll tar.TypeReg os.MkdirAll", err)
						return "", err
					}
				}
				if err = copyFile(target, tempOutput); err != nil {
					fmt.Println("extractAll tar.TypeReg CopyFile", err)
					return "", err
				}
			} else {
				// copy over contents
				if _, err := io.Copy(f, tr); err != nil {
					fmt.Println("extractAll tar.TypeReg io.Copy", err)
					return "", err
				}
			}

			f.Close()
		}
	}
}

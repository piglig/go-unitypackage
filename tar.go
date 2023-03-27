package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func tarHeader(basePath, path string, tw *tar.Writer) error {
	// 获取文件或目录信息
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	relativePath, err := filepath.Rel(basePath, path)
	if err != nil {
		return err
	}

	if len(relativePath) > 0 {
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		hdr.Name = relativePath
		if err = tw.WriteHeader(hdr); err != nil {
			return err
		}
	}

	return nil
}

func tarGz(outFilePath string, inPath string) error {
	// file write
	fw, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer fw.Close()

	// gzip write
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err = tarHeader(inPath, inPath, tw); err != nil {
		return err
	}

	if err = iterDirectory(inPath, inPath, tw); err != nil {
		return err
	}
	return err
}

func tarGzWrite(basePath, filePath string, tw *tar.Writer, fi os.FileInfo) error {
	relativePath, err := filepath.Rel(filePath, basePath)
	if err != nil {
		return err
	}

	fr, err := os.Open(basePath)
	if err != nil {
		return err
	}
	defer fr.Close()

	h := new(tar.Header)
	h.Name = relativePath
	h.Size = fi.Size()
	h.Mode = int64(fi.Mode())
	h.ModTime = fi.ModTime()

	err = tw.WriteHeader(h)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, fr)
	if err != nil {
		return err
	}
	return nil
}

func iterDirectory(rootPath, dirPath string, tw *tar.Writer) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, f := range files {
		curPath := filepath.Join(dirPath, f.Name())

		if f.IsDir() {
			if err = tarHeader(rootPath, curPath, tw); err != nil {
				return err
			}
			if err = iterDirectory(rootPath, curPath, tw); err != nil {
				return err
			}
		} else {
			if err = tarGzWrite(curPath, rootPath, tw, f); err != nil {
				return err
			}
		}
	}

	return nil
}

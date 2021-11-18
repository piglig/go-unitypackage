package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MetaFile struct {
	Guid     string `yaml:"guid"`
	Path     string `yaml:"-"`
	MetaPath string `yaml:"-"`
}

func PreprocessAssets(assetsRoot string) {
	preProcessFilesInPath(assetsRoot, "./")
}

func preProcessFilesInPath(assetsRoot, relativePath string) error {
	assetPath := filepath.Join(assetsRoot, relativePath)

	fmt.Println("Process files in directory: ", assetPath)

	// 处理根目录
	if err := processFile(assetsRoot, relativePath); err != nil {
		return err
	}

	// 处理内容
	dirs, err := ioutil.ReadDir(assetPath)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		childRelativePath := filepath.Join(relativePath, dir.Name())
		childAbsPath := filepath.Join(assetsRoot, childRelativePath)

		info, err := os.Stat(childAbsPath)

		if err != nil {
			return err
		}

		if info.IsDir() {
			preProcessFilesInPath(assetsRoot, childRelativePath)
		} else {
			if strings.HasSuffix(assetsRoot, childRelativePath) {
				continue
			} else {
				processFile(assetsRoot, childRelativePath)
			}
		}
	}

	return nil
}

func processFile(assetsRoot, relativeFilePath string) error {
	fmt.Println("processFile: ", relativeFilePath)
	fullFilePath := filepath.Join(assetsRoot, relativeFilePath)

	fullMetaFilePath := getAssetMetaPath(fullFilePath)

	if _, err := os.Stat(fullMetaFilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			rst := generateMetafile(fullMetaFilePath, relativeFilePath)
			if rst != nil {
				return rst
			}
		} else {
			return err
		}
	}

	return nil
}

func getAssetMetaPath(filePath string) string {
	if strings.HasSuffix(filePath, ".meta") {
		return filePath
	}
	return filePath + ".meta"
}

func generateMetafile(fullMetaFilePath, relativeFilePath string) error {
	fmt.Println("generateMetaFile at", fullMetaFilePath)

	contents, err := getMetafileTemplatePath()
	if err != nil {
		return err
	}

	contents = strings.ReplaceAll(contents, "{guid}", getDeterministicGuid(relativeFilePath))
	contents = strings.ReplaceAll(contents, "{timeCreated}", fmt.Sprintf("%d", time.Now().Unix()))

	err = ioutil.WriteFile(fullMetaFilePath, []byte(contents), 0755)
	if err != nil {
		return err
	}

	return nil
}

func getMetafileTemplatePath() (string, error) {
	content, err := os.ReadFile("./utils/metafile_template.yaml")
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func getDeterministicGuid(relativeFilePath string) string {
	hash := md5.Sum([]byte(relativeFilePath))
	return hex.EncodeToString(hash[:])
}

func GeneratePackage(assetsRoot, outputPath, tempMd5Path, tempPath string) error {
	assetsRoot = filepath.Clean(assetsRoot)
	outputPath = filepath.Clean(outputPath)
	tempMd5Path = filepath.Clean(tempMd5Path)
	tempPath = filepath.Clean(tempPath)
	os.RemoveAll(tempPath)
	if err := os.MkdirAll(tempPath, 0777); err != nil {
		return err
	}

	assets, err := collectAssetsInPath(assetsRoot)
	if err != nil {
		return err
	}

	localBaseName := filepath.Base(assetsRoot)

	for _, asset := range assets {
		if strings.HasSuffix(asset.Path, ".meta") {
			continue
		}

		fmt.Println("GeneratePackage asset", asset.Path)
		assetDir := filepath.Join(tempPath, asset.Guid)
		assetPath := filepath.Join(assetDir, "asset")
		metaPath := filepath.Join(assetDir, "asset.meta")
		pathNamePath := filepath.Join(assetDir, "pathname")
		pathNameLocal := strings.ReplaceAll(asset.Path, assetsRoot, "")
		pathNameLocal = strings.ReplaceAll(pathNameLocal, ".meta", "")

		if strings.HasPrefix(pathNameLocal, "/") {
			pathNameLocal = strings.TrimPrefix(pathNameLocal, "/")
		}

		// 用 unity 相对路径 Assets/... 代替根路径
		const DefaultUnityRootPath = "Assets/"
		pathNameLocal = filepath.Join(localBaseName, pathNameLocal)
		pathNameLocal = strings.ReplaceAll(pathNameLocal, "\\", "/")

		if isDir(assetDir) {
			err = os.RemoveAll(assetDir)
			if err != nil {
				return err
			}
		}

		err = os.MkdirAll(assetDir, 0777)
		if err != nil {
			return err
		}

		if !isDir(asset.Path) {
			// 拷贝素材
			if err = copyFile(asset.Path, assetPath); err != nil {
				return err
			}
		}

		// 拷贝元文件
		if err = copyFile(asset.MetaPath, metaPath); err != nil {
			return err
		}

		fmt.Println("pathNamePath", pathNamePath, "content:", pathNameLocal)
		if err = os.WriteFile(pathNamePath, []byte(pathNameLocal), 0777); err != nil {
			return err
		}
	}

	if err = TarGz(outputPath, tempPath); err != nil {
		return err
	}

	return nil
}

func TarGzWrite(_path string, tw *tar.Writer, fi os.FileInfo) error {
	fr, err := os.Open(_path)
	if err != nil {
		return err
	}
	defer fr.Close()

	h := new(tar.Header)
	h.Name = _path
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

func IterDirectory(dirPath string, tw *tar.Writer) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()
	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		curPath := dirPath + "/" + fi.Name()

		if fi.IsDir() {
			if err = IterDirectory(curPath, tw); err != nil {
				return err
			}
		} else {
			if err = TarGzWrite(curPath, tw, fi); err != nil {
				return err
			}
		}
	}

	return nil
}

func TarGz(outFilePath string, inPath string) error {
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

	IterDirectory(inPath, tw)
	return err
}

func isDir(path string) bool {
	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	if info.IsDir() {
		return true
	} else {
		return false
	}
}

func collectAssetsInPath(assetPath string) ([]MetaFile, error) {
	assets := make([]MetaFile, 0)

	assetRef, err := getAssetReference(assetPath)
	if err == nil && assetRef != nil {
		fmt.Println("Adding asset ref ", assetRef)
		assets = append(assets, *assetRef)
	}

	dirs, err := ioutil.ReadDir(assetPath)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(assetPath, dir.Name())
		fmt.Println(fullPath)

		info, err := os.Stat(fullPath)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			subAssets, err := collectAssetsInPath(fullPath)
			if err != nil {
				return nil, err
			}
			for _, as := range subAssets {
				assets = append(assets, as)
			}
		} else {
			assetRef, err = getAssetReference(fullPath)
			if err == nil && assetRef != nil {
				fmt.Println("Adding asset ref ", assetRef)
				assets = append(assets, *assetRef)
			}
		}
	}
	return assets, nil
}

func getAssetReference(filePath string) (*MetaFile, error) {
	metaFilePath := getAssetMetaPath(filePath)

	info, err := os.Stat(metaFilePath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, nil
	} else {
		content, err := ioutil.ReadFile(metaFilePath)
		if err != nil {
			return nil, err
		}

		res := MetaFile{
			Path:     filePath,
			MetaPath: metaFilePath,
		}
		err = yaml.Unmarshal(content, &res)
		if err != nil {
			return nil, err
		}

		return &res, nil
	}
}

// copyFile 文件拷贝
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type MetaFile struct {
	Guid     string `yaml:"guid"`
	Path     string `yaml:"-"`
	MetaPath string `yaml:"-"`
}

func PreprocessAssets(assetsRoot string) error {
	assetDir := filepath.Join(assetsRoot, "Assets")
	return preProcessFilesInPath(assetDir, "./")
}

func preProcessFilesInPath(assetsRoot, relativePath string) error {
	assetPath := filepath.Join(assetsRoot, relativePath)

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
			if err = preProcessFilesInPath(assetsRoot, childRelativePath); err != nil {
				return err
			}
		} else {
			if strings.HasSuffix(assetsRoot, childRelativePath) {
				continue
			} else {
				if err = processFile(assetsRoot, childRelativePath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func processFile(assetsRoot, relativeFilePath string) error {
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

//go:embed metafile_template.yaml
var data []byte

func generateMetafile(fullMetaFilePath, relativeFilePath string) error {
	temp := make([]byte, len(data))
	copy(temp, data)

	contents := string(temp)

	contents = strings.ReplaceAll(contents, "{guid}", getDeterministicGuid(relativeFilePath))

	err := ioutil.WriteFile(fullMetaFilePath, []byte(contents), 0755)
	if err != nil {
		return err
	}

	return nil
}

func getDeterministicGuid(relativeFilePath string) string {
	hash := md5.Sum([]byte(relativeFilePath))
	return hex.EncodeToString(hash[:])
}

func GeneratePackage(assetsRoot, outputPath string) error {
	assetsRoot = filepath.Join(assetsRoot, "Assets")
	outputPath = filepath.Clean(outputPath)
	os.Remove(outputPath)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0777); err != nil {
		return err
	}

	assets, err := collectAssetsInPath(assetsRoot)
	if err != nil {
		return err
	}

	localBaseName := filepath.Base(assetsRoot)

	tempDir, err := ioutil.TempDir("", "temp")
	if err != nil {
		return err
	}

	for _, asset := range assets {
		if strings.HasSuffix(asset.Path, ".meta") {
			continue
		}

		assetDir := filepath.Join(tempDir, asset.Guid)
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

		if err = os.WriteFile(pathNamePath, []byte(pathNameLocal), 0777); err != nil {
			return err
		}
	}

	if err = TarGz(outputPath, tempDir); err != nil {
		return err
	}

	return nil
}

func TarGzWrite(_path, rootPath string, tw *tar.Writer, fi os.FileInfo) error {
	_path = filepath.Clean(_path)
	rootPath = filepath.Clean(rootPath)
	relativePath, err := filepath.Rel(rootPath, _path)
	if err != nil {
		return err
	}

	fr, err := os.Open(_path)
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

func IterDirectory(rootPath, dirPath string, tw *tar.Writer) error {
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
			if err = tarHeader(rootPath, curPath, tw); err != nil {
				return err
			}
			if err = IterDirectory(rootPath, curPath, tw); err != nil {
				return err
			}
		} else {
			if err = TarGzWrite(curPath, rootPath, tw, fi); err != nil {
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

	//re, err := filepath.Rel(inPath, inPath)
	//if err != nil {
	//	return err
	//}
	//fmt.Println(re)

	if err = tarHeader(inPath, inPath, tw); err != nil {
		return err
	}

	if err = IterDirectory(inPath, inPath, tw); err != nil {
		return err
	}
	return err
}

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
		assets = append(assets, *assetRef)
	}

	dirs, err := ioutil.ReadDir(assetPath)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(assetPath, dir.Name())

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

// copyFile the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
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

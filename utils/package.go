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
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type MetaFile struct {
	Guid     string `yaml:"guid"`
	Path     string `yaml:"-"`
	MetaPath string `yaml:"-"`
}

// PreprocessAssets preprocesses the assets at the given assets root directory.
func PreprocessAssets(assetsRoot string) error {
	assetDir := GetAssetsRootPath(assetsRoot)
	return preProcessFilesInPath(assetDir, "./")
}

func GetAssetsRootPath(path string) string {
	return filepath.Join(path, "Assets")
}

func preProcessFilesInPath(assetsRoot, relPath string) error {
	assetPath := filepath.Join(assetsRoot, relPath)
	err := filepath.Walk(assetPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == assetPath {
			return processFile(assetsRoot, relPath)
		}

		if info.IsDir() {
			return nil
		}

		childRelativePath, err := filepath.Rel(assetsRoot, path)
		if err != nil {
			return err
		}

		if strings.HasSuffix(assetsRoot, path) {
			return nil
		}

		return processFile(assetsRoot, childRelativePath)
	})

	return err
}

func processFile(assetsRoot, relativeFilePath string) error {
	fullFilePath := filepath.Join(assetsRoot, relativeFilePath)
	fullMetaFilePath := getAssetMetaPath(fullFilePath)

	_, err := os.Stat(fullMetaFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if errors.Is(err, os.ErrNotExist) {
		return generateMetafile(fullMetaFilePath, relativeFilePath)
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
var metafileTemplate []byte

// // generateMetafile generates a metafile for the given file.
func generateMetafile(fullMetaFilePath, relativeFilePath string) error {
	metafile := make([]byte, len(metafileTemplate))
	copy(metafile, metafileTemplate)

	contents := string(metafile)

	contents = strings.ReplaceAll(contents, "{guid}", getDeterministicGuid(relativeFilePath))

	err := os.WriteFile(fullMetaFilePath, []byte(contents), 0755)
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
	assetsRoot = GetAssetsRootPath(assetsRoot)
	outputPath = filepath.Clean(outputPath)

	if err := os.RemoveAll(outputPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0777); err != nil {
		return err
	}

	assets, err := collectAssetsInPath(assetsRoot)
	if err != nil {
		return err
	}

	localBaseName := filepath.Base(assetsRoot)

	tempDir, err := os.MkdirTemp("", "temp")
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

	err := filepath.Walk(assetPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			assetRef, err := getAssetReference(path)
			if err == nil && assetRef != nil {
				assets = append(assets, *assetRef)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
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
		content, err := os.ReadFile(metaFilePath)
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

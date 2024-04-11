package unitypackage

import (
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

// metaFile the asset metafile
type metaFile struct {
	Guid     string `yaml:"guid"` // a unique identifier for the asset
	Path     string `yaml:"-"`    // the asset path
	MetaPath string `yaml:"-"`    // the asset metafile path
}

const DefaultUnityRootPath = "Assets"

// getAssetsRootPath get Assets path from unpackage path
func getAssetsRootPath(path string) string {
	return filepath.Join(path, DefaultUnityRootPath)
}

// preProcessFilesInPath preprocesses the assets at the given assets root directory.
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

	return os.WriteFile(fullMetaFilePath, []byte(contents), 0755)
}

func getDeterministicGuid(relativeFilePath string) string {
	hash := md5.Sum([]byte(relativeFilePath))
	return hex.EncodeToString(hash[:])
}

func GeneratePackage(assetsRoot, outputPath string) error {
	assetsRoot = getAssetsRootPath(assetsRoot)
	outputPath = filepath.Clean(outputPath)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0777); err != nil {
		return err
	}

	assets, err := collectAssetsInPath(assetsRoot)
	if err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp("", "temp")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	localBaseName := filepath.Base(assetsRoot)
	for _, asset := range assets {
		if strings.HasSuffix(asset.Path, ".meta") {
			continue
		}

		assetDir := filepath.Join(tempDir, asset.Guid)
		assetPath, metaPath, pathNamePath := prepareAssetPaths(assetDir)

		pathNameLocal := strings.ReplaceAll(asset.Path, assetsRoot, "")
		pathNameLocal = strings.ReplaceAll(pathNameLocal, ".meta", "")
		pathNameLocal = strings.TrimPrefix(pathNameLocal, "/")
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
			if err = copyFile(asset.Path, assetPath); err != nil {
				return err
			}
		}

		if err = copyFile(asset.MetaPath, metaPath); err != nil {
			return err
		}

		if err = os.WriteFile(pathNamePath, []byte(pathNameLocal), 0777); err != nil {
			return err
		}
	}

	if err = tarGz(outputPath, tempDir); err != nil {
		return err
	}

	return nil
}

// prepareAssetPaths prepares file paths for the asset's content.
func prepareAssetPaths(assetDir string) (assetPath, metaPath, pathNamePath string) {
	assetPath = filepath.Join(assetDir, "asset")
	metaPath = filepath.Join(assetDir, "asset.meta")
	pathNamePath = filepath.Join(assetDir, "pathname")
	return
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func collectAssetsInPath(assetPath string) ([]metaFile, error) {
	assets := make([]metaFile, 0)

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

func getAssetReference(filePath string) (*metaFile, error) {
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

		res := metaFile{
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

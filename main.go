package main

import (
	"fmt"
	"go-unitypackage/utils"
	"io/ioutil"
	"path/filepath"
)

func main() {
	fileName := "D://test_unity//rw_en_resouce_1009.unitypackage"

	unpackageDir, err := ioutil.TempDir("", "unpackage")
	if err != nil {
		fmt.Println(err)
		return
	}

	packageDir, err := ioutil.TempDir("", "package")
	if err != nil {
		fmt.Println(err)
		return
	}

	tempDir, err := ioutil.TempDir("", "temp")
	if err != nil {
		fmt.Println(err)
		return
	}

	md5Dir, err := ioutil.TempDir("", "md5")
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = utils.UnPackage(fileName,
		unpackageDir, md5Dir); err != nil {
		fmt.Println(err)
		return
	}

	assetDir := filepath.Join(unpackageDir, "Assets")
	if err = utils.PreprocessAssets(assetDir); err != nil {
		fmt.Println(err)
		return
	}

	name := filepath.Base(fileName)
	output := packageDir + name
	if err = utils.GeneratePackage(assetDir,
		output, tempDir); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(output)
}

package main

import (
	"fmt"
	"go-unitypackage/utils"
	"os"
	"path/filepath"
)

func main() {
	fileName := "D://test_unity//111111.unitypackage"

	unpackageDir, err := os.MkdirTemp("", "unpackage")
	if err != nil {
		fmt.Println(err)
		return
	}

	packageDir, err := os.MkdirTemp("", "package")
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = utils.UnPackage(fileName,
		unpackageDir); err != nil {
		fmt.Println(err)
		return
	}

	if err = utils.PreprocessAssets(unpackageDir); err != nil {
		fmt.Println(err)
		return
	}

	name := filepath.Base(fileName)
	output := packageDir + name
	if err = utils.GeneratePackage(unpackageDir,
		output); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(output)
}

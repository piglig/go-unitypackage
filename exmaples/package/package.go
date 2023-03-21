package main

import (
	"fmt"
	"go-unitypackage/utils"
	"path/filepath"
)

func main() {
	unityFile := "D:\\test_unity\\test.unitypackage"
	unpackageOutputPath := "D:\\test_unity"
	//if err := utils.UnPackage(unityFile,
	//	unpackageOutputPath); err != nil {
	//	return
	//}

	//if err = utils.PreprocessAssets(unpackageDir); err != nil {
	//	fmt.Println(err)
	//	return
	//}

	name := filepath.Base(unityFile)
	output := "D:\\test_unity\\" + name
	if err := utils.GeneratePackage(unpackageOutputPath,
		output); err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println(output)
}

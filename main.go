package main

import (
	"fmt"
	"go-unitypackage/utils"
)

func main() {
	err := utils.UnPackage("D://test_unity//package//bard_en_test.unitypackage",
		"D://test_unity//temp_unpackage//", "D://test_unity//temp//")
	if err != nil {
		fmt.Println(err)
	}

	utils.PreprocessAssets("D://test_unity//temp_unpackage//Assets//")
	if err := utils.GeneratePackage("D://test_unity//temp_unpackage//Assets//",
		"D://test_unity//package//bard_en_test.unitypackage",
		"D://test_unity//temp_unpackage//", "D://test_unity//temp_package//"); err != nil {
		fmt.Println(err)
		return
	}
}

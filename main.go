/*
Go-unitypackage compress and decompress .unitypackage

# Given a file path, it compress or decompress .unitypackage

Usage:

	go run main.go [flags] [path]

The flags are:

	-p
		The .unitypackage file path
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {

	fileName := flag.String("p", "", ".unitypackage file path")
	flag.Parse()
	//"D://test_unity//111111.unitypackage"
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

	if err = UnPackage(*fileName,
		unpackageDir); err != nil {
		fmt.Println(err)
		return
	}

	name := filepath.Base(*fileName)
	output := packageDir + name
	if err = GeneratePackage(unpackageDir,
		output); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(output)
}

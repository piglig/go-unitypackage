# go-unitypackage
[![GoDoc](https://godoc.org/github.com/piglig/go-unitypackage?status.svg)](https://pkg.go.dev/github.com/piglig/go-unitypackage)
[![Go Report Card](https://goreportcard.com/badge/github.com/piglig/go-unitypackage)](https://goreportcard.com/report/github.com/piglig/go-unitypackage)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](https://opensource.org/licenses/MIT)

## Overview
pack and unpack *.unitypackage files, with Golang

## Installation
```go
go get github.com/piglig/go-unitypackage
```

## How to use
### UnPackage
```go
package main

import "github.com/piglig/go-unitypackage"

func main() {
	// the unitypackage path
	packagePath := "D:\\test_unity\\test.unitypackage"
	// the output assets path
	assetsPath := "D:\\test_unity"
	// Unpackage command will extract content from unitypackage.
	err := unitypackage.UnPackage(packagePath, assetsPath)
	if err != nil {
		return
	}
}
```
### Package
```Golang
package main

import "github.com/piglig/go-unitypackage"

func main() {
	// the assets directory
	assetsPath := "D:\\test_unity"
	// the output unitypackage path
	packagePath := "D:\\test_unity\\package\\output.unitypackage"
	// GeneratePackage command will generate a package from assets directory
	err := unitypackage.GeneratePackage(assetsPath, packagePath)
	if err != nil {
		return
	}
}
```

## Command Line Access
```Golang
go run main.go -p "D://test_unity//test.unitypackage"
```

## License
See the [LICENSE](LICENSE) file for license rights and limitations (MIT).
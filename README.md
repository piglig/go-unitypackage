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

## Command Line Tool
**go-unitypackage** command line tool to package or unpackage the unitypackage.
### Installation
In order to use the tool, compile it using the following command
```shell
go install github.com/piglig/go-unitypackage/tools/go-unitypackage@latest
```

### Usage
```shell
go-unitypackage [options] [arguments]
```

#### Options
- `-act`: compress or decompress unitypackage (default "decompress")
#### Arguments
- `-src`: compress or decompress the source directory
- `-dest`: compress or decompress the destination directory
### Example
```shell
$:/usr/local/test_unity# ls
test.unitypackage

$:/usr/local/test_unity# go-unitypackage -src=test.unitypackage -dest=.
2023/04/04 10:59:55 target: /tmp/md5136041829/afd150e418c353f49b2bd5b68d0c43fa header:  afd150e418c353f49b2bd5b68d0c43fa
2023/04/04 10:59:55 target: /tmp/md5136041829/afd150e418c353f49b2bd5b68d0c43fa/asset header:  afd150e418c353f49b2bd5b68d0c43fa/asset
2023/04/04 10:59:55 target: /tmp/md5136041829/afd150e418c353f49b2bd5b68d0c43fa/asset.meta header:  afd150e418c353f49b2bd5b68d0c43fa/asset.meta
2023/04/04 10:59:55 target: /tmp/md5136041829/afd150e418c353f49b2bd5b68d0c43fa/pathname header:  afd150e418c353f49b2bd5b68d0c43fa/pathname

$:/usr/local/test_unity# ls
Assets  Assets.meta test.unitypackage

$:/usr/local/test_unity# mkdir package

$:/usr/local/test_unity# go-unitypackage -act=compress -src=. -dest=package/out.unitypackage

$:/usr/local/test_unity/package# ls -l
total 4
-rw-r--r-- 1 root root 340 Apr  4 11:03 out.unitypackage
```
## License
See the [LICENSE](LICENSE) file for license rights and limitations (MIT).
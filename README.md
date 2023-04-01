# go-unitypackage
[![GoDoc](https://godoc.org/github.com/piglig/go-unitypackage?status.svg)](https://pkg.go.dev/github.com/piglig/go-unitypackage)
[![Go Report Card](https://goreportcard.com/badge/github.com/piglig/go-unitypackage)](https://goreportcard.com/report/github.com/piglig/go-unitypackage)

## Overview
pack and unpack *.unitypackage files, with Golang

## Installation
```go
go get github.com/piglig/go-unitypackage
```

## How to use
### UnPackage
Given the following setup:
```
D://test_unity//package//test.unitypackage
```

The command will extract content from unitypackage.

```Golang
err := utils.UnPackage("D://test_unity//package//test.unitypackage",
    "D://test_unity//temp_unpackage//", "D://test_unity//temp//")
if err != nil {
    fmt.Println(err)
}
```

### Package
Given the following setup:
```
D://test_unity//temp_unpackage//Assets//
D://test_unity//temp_unpackage//Assets//code.dll
D://test_unity//temp_unpackage//Assets//object.prefab
D://test_unity//temp_unpackage//Assets//code.dll.mbd
```

The command will generate a package that installs the content of "D://test_unity//temp_unpackage//Assets//" into "Assets/content/".
It uses the last folder name in the path as the containing folder for the assets.

```Golang
utils.PreprocessAssets("D://test_unity//temp_unpackage//Assets//")
if err := utils.GeneratePackage("D://test_unity//temp_unpackage//Assets//",
    "D://test_unity//package//test.unitypackage",
    "D://test_unity//temp_unpackage//", "D://test_unity//temp_package//"); err != nil {
    fmt.Println(err)
    return
}
```

## Command Line Access
```Golang
go run main.go -p "D://test_unity//test.unitypackage"
```

## License
See the [LICENSE](LICENSE) file for license rights and limitations (MIT).
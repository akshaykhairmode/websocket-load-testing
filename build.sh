set -x

rm -rf bin/*

tempPath="bin/temp"

zipBinary() {
    zip -j $1 $tempPath/*
    rm -rf $tempPath
}

buildX64() {
    GOOS=$1 GOARCH=amd64 go build -o $tempPath/websocket-client-$1-x64 client/main.go
    GOOS=$1 GOARCH=amd64 go build -o $tempPath/websocket-server-$1-x64 server/main.go
}

buildX32() {
    GOOS=$1 GOARCH=386 go build -o $tempPath/websocket-client-$1-x32 client/main.go
    GOOS=$1 GOARCH=386 go build -o $tempPath/websocket-server-$1-x32 server/main.go
}

#LINUX
buildX64 "linux"
zipBinary "bin/linux_x64.zip"

buildX32 "linux"
zipBinary "bin/linux_x32.zip"

#Windows
buildX64 "windows"
zipBinary "bin/windows_x64.zip"

buildX32 "windows"
zipBinary "bin/windows_x32.zip"

#Mac
buildX64 "darwin"
zipBinary "bin/darwin_x64.zip"
setup:
	mkdir -p bin

build_win32:
	CGO_ENABLED="0" \
	GOOS="windows" \
	GOARCH="386" \
	go build -o bin/vss_x86_32.exe cmd/example/*

build_win64:
	CGO_ENABLED="0" \
	GOOS="windows" \
	GOARCH="amd64" \
	go build -o bin/vss_x86_64.exe cmd/example/*

build_arm64:
	CGO_ENABLED="0" \
	GOOS="windows" \
	GOARCH="arm64" \
	go build -o bin/vss_arm64.exe cmd/example/*

build:
	CGO_ENABLED="0" \
	GOOS="windows" \
    go build -o bin/ cmd/example/*

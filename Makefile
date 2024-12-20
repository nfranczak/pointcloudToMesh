mac:
	rm -rf bin
	GOOS=darwin GOARCH=arm64 go build -o bin/darwin-arm64/pc-to-mesh

linux:
	rm -rf bin
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/pc-to-mesh
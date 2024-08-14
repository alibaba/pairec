## 定义通用变量
BUILD=go build -ldflags '-s -w -extldflags'

.PHONY : setup  clean pairecmd
pairecmd: setup
	GOARCH=amd64 GOOS=linux ${BUILD} -o pairecmd .
	GOARCH=amd64 GOOS=darwin ${BUILD} -o pairecmdmac .
	GOARCH=arm64 GOOS=darwin ${BUILD} -o pairecmdmacarm .
	GOARCH=amd64 GOOS=windows ${BUILD} -o pairecmdwin .
	-mv pairecmd pairecmdwin pairecmdmac pairecmdmacarm bin/

setup:
	if [ ! -d bin ]; then\
		mkdir bin; \
	fi

clean:
	-rm -rf build
	-rm -rf bin

build:
	go build -o /Users/admin/execFile/pairec-kafka main.go

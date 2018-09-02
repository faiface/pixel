OUT := amidakuji
ASSET_TARGET := glossary/asset.go
ASSET_SOURCE_DIR := assets
VERSION := $(shell git describe --always --long)
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: build build_windows

${ASSET_TARGET}: 
	go-bindata -o "${ASSET_TARGET}" -pkg "glossary" -prefix "${ASSET_SOURCE_DIR}" ${ASSET_SOURCE_DIR}/emoji ${ASSET_SOURCE_DIR}/karaoke ${ASSET_SOURCE_DIR}/NanumBarunGothic.ttf

build: ${ASSET_TARGET}
	go build -i -v -o ${OUT} -ldflags "-w -s -X main.version=${VERSION}"

build_windows: ${ASSET_TARGET}
	go build -i -v -o ${OUT}.exe -ldflags "-w -s -X main.version=${VERSION} -H windowsgui"

run: build
	./${OUT}

test:
	@go test -short ${PKG_LIST}

vet:
	@go vet -copylocks=false ${PKG_LIST}

vet_annoying:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

#static: vet lint
#	go build -i -v -o ${OUT}-${VERSION} -ldflags "-extldflags \"-static\" -w -s -X main.version=${VERSION}"

#static_windows: vet lint
#	go build -i -v -o ${OUT}-${VERSION}.exe -ldflags "-extldflags \"-static\" -w -s -X main.version=${VERSION} -H windowsgui"

clean:
	-@rm ${ASSET_TARGET} ${OUT} ${OUT}.exe #${OUT}-*

.PHONY: build build_windows run vet vet_annoying lint clean

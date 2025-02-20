GO_FILES := $(shell find . -name '*.go')
TARGET := treesitter-go-ts-perf

.PHONY: all build test clean submodule-init

all: build

build: build-O0 build-O2 build-O3

build-O0:
	go build -o $(TARGET)-O0 $(GO_FILES)

build-O2:
	CGO_CFLAGS="-O2" go build -o $(TARGET)-O2 $(GO_FILES)

build-O3:
	CGO_CFLAGS="-O3" go build -o $(TARGET)-O3 $(GO_FILES)

test: build
	hyperfine --warmup 1 --runs 10 "./$(TARGET)-O0 angular Babylon.js" "./$(TARGET)-O2 angular Babylon.js" "./$(TARGET)-O3 angular Babylon.js"

clean:
	rm -f $(TARGET)-O0 $(TARGET)-O2 $(TARGET)-O3

submodule-init:
	git submodule update --init --recursive --depth 1

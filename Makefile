UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    TARGET := commandcode.dylib
else ifeq ($(OS),Windows_NT)
    TARGET := commandcode.dll
else
    TARGET := commandcode.so
endif

.PHONY: all build test clean lint

all: build

build:
	CGO_ENABLED=1 go build -buildmode=c-shared -o $(TARGET) main.go

test:
	go test -v -race ./...

clean:
	rm -f commandcode.dylib commandcode.so commandcode.dll commandcode.h

lint:
	go vet ./...

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    TARGET := commandcode.dylib
else ifeq ($(OS),Windows_NT)
    TARGET := commandcode.dll
else
    TARGET := commandcode.so
endif

.PHONY: all build test clean lint pagecheck

all: build pagecheck

build:
	CGO_ENABLED=1 go build -buildmode=c-shared -o $(TARGET) main.go

# Extract embedded JS from quota_page.go and syntax-check it with node.
# Guards against parse-time SyntaxErrors (e.g. duplicate const) that break
# the whole resource page; Go substring tests cannot catch these.
pagecheck:
	@node scripts/pagecheck.js

test:
	go test -v -race ./...
	$(MAKE) pagecheck

clean:
	rm -f commandcode.dylib commandcode.so commandcode.dll commandcode.h

lint:
	go vet ./...

# get the repo root and output path
REPO_ROOT:=${CURDIR}
OUT_DIR=$(REPO_ROOT)/bin
# record the source commit in the binary, overridable
COMMIT?=$(shell git rev-parse HEAD 2>/dev/null)

# used for building the binary
BINARY_NAME?=action
PKG:=github.com/curtbushko/commit-status-action
BUILD_LD_FLAGS:=-X=$(PKG).gitCommit=$(COMMIT)
BUILD_FLAGS?=-trimpath -ldflags="-buildid= -w $(BUILD_LD_FLAGS)"

default: build

.PHONY: build
build:
	go build -v -o "$(OUT_DIR)/$(BINARY_NAME)" $(BUILD_FLAGS)

.PHONY: clean
clean:
	rm -rf "$(OUT_DIR)/"

.PHONY: test 
test:
	go test -v ./.

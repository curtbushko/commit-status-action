# get the repo root and output path
REPO_ROOT:=${CURDIR}
BIN_PATH=$(REPO_ROOT)/bin
IMAGE_TAG="curtbushko/commit-status-action"
# record the source commit in the binary, overridable.
COMMIT?=$(shell git rev-parse HEAD 2>/dev/null)

# used for building the binary.
BIN_NAME?=action
PKG:=github.com/curtbushko/commit-status-action
BUILD_LD_FLAGS:=-X=$(PKG).gitCommit=$(COMMIT)
BUILD_FLAGS?=-trimpath -buildvcs=false -ldflags="-buildid= -w $(BUILD_LD_FLAGS)"

default: build

.PHONY: build
build:
	CGO_ENABLED=0 go build -v -o "$(BIN_PATH)/$(BIN_NAME)" $(BUILD_FLAGS)

.PHONY: clean
clean:
	rm -rf "$(BIN_PATH)/"

.PHONY: test
test:
	go test -v ./.

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE_TAG):latest .

.PHONY: lint
lint:
	golangci-lint run -c ./.golangci.yml

.PHONY: github-lint
github-lint: action-lint super-linter

.PHONY: action-lint
action-lint:
	docker run --rm -v $(CURDIR):/repo \
	--workdir /repo \
	rhysd/actionlint:latest \
	-color \
	-ignore 'SC2129' \
	-verbose

.PHONY: super-linter
super-linter:
	echo "Ignore long line warnings as we cannot pass the yaml rules into the container"
	docker run --rm -v $(CURDIR):/tmp/lint \
	-e RUN_LOCAL=true \
	-e USE_FILE_ALGORITHM=true \
	-e VALIDATE_ALL_CODEBASE=false \
	-e VALIDATE_YAML=true \
	github/super-linter:latest

## mqbridge Makefile:
## Translate messages from one message queue system to another one
#:

SHELL          = /bin/sh

# -----------------------------------------------------------------------------
# Build config

GO          ?= go
SOURCES      = $(shell find . -maxdepth 3 -mindepth 1 -name '*.go'  -printf '%p\n')
SRCU = $(wildcard *.go) $(wildcard */*.go) $(wildcard */*/*.go)
PLUGIN_DIRS  = $(shell go list -f '{{.Dir}}' ./plugins/...)
PLUGIN_NAMES = $(notdir $(PLUGIN_DIRS))
PLUGINS      = $(PLUGIN_NAMES:%=%.so)

BUILD_DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
APP_VERSION   ?= $(shell git describe --tags --always)
GOLANG_VERSION = 1.15.5-alpine3.12

OS            ?= linux
ARCH          ?= amd64
ALLARCH       ?= "linux/amd64 linux/386 darwin/amd64"
DIRDIST       ?= dist

# -----------------------------------------------------------------------------
# Docker image config

# application name, docker-compose prefix
PRG           ?= $(shell basename $$PWD)

# Hardcoded in docker-compose.yml service name
DC_SERVICE    ?= app

# Generated docker image
DC_IMAGE      ?= ghcr.io/lekovr/mqbridge

# docker-compose image version
DC_VER        ?= latest

# docker app for change inside containers
DOCKER_BIN    ?= docker

# docker app log files directory
LOG_DIR       ?= ./log

# -----------------------------------------------------------------------------
# App config

# Docker container port
SERVER_PORT   ?= 8080

# -----------------------------------------------------------------------------

.PHONY: all doc gen build-standalone coverage cov-html build test lint fmt vet vendor up down docker-docker docker-clean

# default: show target list
all: help

# ------------------------------------------------------------------------------
## Compile operations
#:

## Run lint
lint:
	@golint ./...
	@golangci-lint run ./...

## Run vet
vet:
	$(GO) vet ./...

## Run tests
test: coverage.out

coverage.out: $(SOURCES)
	$(GO) test -tags test -covermode=atomic -coverprofile=$@ ./...

## Show package coverage in html (make cov-html PKG=counter)
cov-html: coverage.out
	$(GO) tool cover -html=coverage.out

## Build app
build: $(PRG)

## Build webtail command
$(PRG): $(SOURCES)
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -v -o $@ -ldflags \
	  "-X main.built=$(BUILD_DATE) -X main.version=$(APP_VERSION)" ./cmd/$@

plugin-off:
	[ ! -f .plugins ] || exit 0
	for n in $(PLUGIN_NAMES) ; do \
	  sed -i "s/package main/package $$n/" \
	    plugins/$$n/$$n.go plugins/$$n/*_test.go ; \
	done
	rm .plugins

plugin-on:
	[ -f .plugins ] || exit 0
	for n in $(PLUGIN_NAMES) ; do \
	  sed -i "s/package $$n/package main/" \ 
	  plugins/$$n/$$n.go plugins/$$n/*_test.go ; \
	done
	touch .plugins

## Build app
plugin-build: $(PLUGINS)
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -v -o $@ -tags plugin -ldflags \
	  "-X main.built=$(BUILD_DATE) -X main.version=$(APP_VERSION)" ./cmd/$@

plugin-test: $(SOURCES) $(PLUGINS)
	$(GO) test -tags test,plugin -covermode=atomic -coverprofile=coverage.out ./...

example.so: plugins/example/*.go
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=plugin -o $@ $<

file.so: plugins/file/*.go
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=plugin -o $@ $<

nats.so: plugins/nats/*.go
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=plugin -o $@ $<

pg.so: plugins/pg/*.go
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=plugin -o $@ $<


## Build like docker image from scratch
build-standalone: lint vet test
	GOOS=linux CGO_ENABLED=0 $(GO) build -a -v -o $(PRG) -ldflags \
	  "-X main.built=$(BUILD_DATE) -X main.version=$(APP_VERSION)" ./cmd/$(PRG)

## build and run in foreground
run: build
	./$(PRG) --debug

doc:
	@echo "Open http://localhost:6060/pkg/LeKovr/webtail"
	@godoc -http=:6060

# ------------------------------------------------------------------------------
## Prepare distros
#:

## build app for all platforms
buildall: lint vet
	@echo "*** $@ ***" ; \
	  for a in "$(ALLARCH)" ; do \
	    echo "** $${a%/*} $${a#*/}" ; \
	    P=$(PRG)-$${a%/*}_$${a#*/} ; \
	    GOOS=$${a%/*} GOARCH=$${a#*/} $(GO) build -o $$P -ldflags \
	      "-X main.built=$(BUILD_DATE) -X main.version=$(APP_VERSION)" ./cmd/$(PRG) ; \
	  done

## create disro files
dist: clean buildall
	@echo "*** $@ ***"
	@[ -d $(DIRDIST) ] || mkdir $(DIRDIST)
	@sha256sum $(PRG)-* > $(DIRDIST)/SHA256SUMS ; \
	  for a in "$(ALLARCH)" ; do \
	    echo "** $${a%/*} $${a#*/}" ; \
	    P=$(PRG)-$${a%/*}_$${a#*/} ; \
	    zip "$(DIRDIST)/$$P.zip" "$$P" README.md README.ru.md screenshot.png; \
        rm "$$P" ; \
	  done


## clean generated files
clean:
	@echo "*** $@ ***" ; \
	  for a in "$(ALLARCH)" ; do \
	    P=$(PRG)_$${a%/*}_$${a#*/} ; \
	    [ -f $$P ] && rm $$P || true ; \
	  done
	@[ -d $(DIRDIST) ] && rm -rf $(DIRDIST) || true
	@[ -f $(PRG) ] && rm -f $(PRG) || true
	@[ ! -f coverage.out ] || rm coverage.out

# ------------------------------------------------------------------------------
## Docker operations
#:

## Start service in container
up:
up: CMD="up -d $(DC_SERVICE)"
up: dc

## Stop service
down:
down: CMD="rm -f -s $(DC_SERVICE)"
down: dc

## Build docker image
docker-build: CMD="build --force-rm $(DC_SERVICE)"
docker-build: dc

## Remove docker image & temp files
docker-clean:
	[ "$$($(DOCKER_BIN) images -q $(DC_IMAGE) 2> /dev/null)" = "" ] || $(DOCKER_BIN) rmi $(DC_IMAGE)

# ------------------------------------------------------------------------------

# $$PWD usage allows host directory mounts in child containers
# Thish works if path is the same for host, docker, docker-compose and child container
## run $(CMD) via docker-compose
dc: docker-compose.yml
	@$(DOCKER_BIN) run --rm  -i \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $$PWD:$$PWD -w $$PWD \
  --env=DC_IMAGE=$(DC_IMAGE) \
  --env=GOLANG_VERSION=$(GOLANG_VERSION) \
  docker/compose:$(DC_VER) \
  -p $(PRG) \
  "$(CMD)"

# ------------------------------------------------------------------------------
## Other
#:

## Update docs at pkg.go.dev
update-godoc:
	vf=$(APP_VERSION) ; v=$${vf%%-*} ; echo "Update for $$v..." ; \
	curl 'https://proxy.golang.org/github.com/!le!kovr/mqbridge/@v/'$$v'.info'

## Update latest docker image tag at ghcr.io
update-ghcr:
	vf=$(APP_VERSION) ; vs=$${vf%%-*} ; v=$${vs#v} ; echo "Update for $$v..." ; \
	docker pull ghcr.io/lekovr/mqbridge:$$v && \
	docker tag ghcr.io/lekovr/mqbridge:$$v ghcr.io/lekovr/mqbridge:latest && \
	docker push ghcr.io/lekovr/mqbridge:latest


# This code handles group header and target comment with one or two lines only
## list Makefile targets
## (this is default target)
help:
	@grep -A 1 -h "^## " $(MAKEFILE_LIST) \
  | sed -E 's/^--$$// ; /./{H;$$!d} ; x ; s/^\n## ([^\n]+)\n(## (.+)\n)*(.+):(.*)$$/"    " "\4" "\1" "\3"/' \
  | sed -E 's/^"    " "#" "(.+)" "(.*)"$$/"" "" "" ""\n"\1 \2" "" "" ""/' \
  | xargs printf "%s\033[36m%-15s\033[0m %s %s\n"

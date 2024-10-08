## mqbridge Makefile:
## Stream messages from PG/NATS/File channel to another PG/NATS/File channel
#:

SHELL          = /bin/bash

# -----------------------------------------------------------------------------
# Build config

GO          ?= go
SOURCES      = $(shell find . -maxdepth 4 -mindepth 1 -path ./var -prune -o -name '*.go')
PLUGIN_DIRS  = $(shell go list -f '{{.Dir}}' ./plugins/...)
PLUGIN_NAMES = $(notdir $(PLUGIN_DIRS))
USE_PLUGINS  = $(shell test -f .plugins && echo yes)
ifeq ($(USE_PLUGINS),yes)
  PLUGINS      = $(PLUGIN_NAMES:%=%.so)
  BUILD_ARG    = -tags plugin
  TEST_TAGS    = test,plugin
else
  TEST_TAGS    = test
endif

APP_VERSION   ?= $(shell git describe --tags --always)
# Last project tag (used in `make changelog`)
RELEASE       ?= $(shell git describe --tags --abbrev=0 --always)

GOLANG_VERSION ?= 1.22.3-alpine3.20

OS            ?= linux
ARCH          ?= amd64
ALLARCH       ?= "linux/amd64 linux/386 darwin/amd64"
DIRDIST       ?= dist

# Path to golang package docs
GODOC_REPO    ?= github.com/!le!kovr/$(PRG)
# Path to docker image registry
DOCKER_IMAGE  ?= ghcr.io/lekovr/$(PRG)

# -----------------------------------------------------------------------------
# Docker image config

# application name, docker-compose prefix
PRG           ?= $(shell basename $$PWD)

# Hardcoded in docker-compose.yml service name
DC_SERVICE    ?= app

# docker app for change inside containers
DOCKER_BIN     ?= docker compose

# Ports for docker tests
TEST_NATS_PORT ?= 34222
TEST_PG_PORT   ?= 35432

TEST_PG_PASS   ?= secret
DB_CONTAINER   ?= $(PRG)-pg-1

export
# -----------------------------------------------------------------------------

.PHONY: all doc gen build-standalone coverage cov-html build test lint fmt vet vendor up down docker-docker docker-clean

# default: show target list
all: help

# ------------------------------------------------------------------------------
## Compile operations
#:

## run `golint` and `golangci-lint`
lint:
	@golint ./...
	@golangci-lint run ./...

## run `go vet`
vet:
	$(GO) vet ./...

## run tests
test: clean coverage.out

coverage.out: $(SOURCES) $(PLUGINS)
	$(GO) test -tags $(TEST_TAGS)$(TEST_TAGS_MORE) -covermode=atomic -coverprofile=$@ ./...

.PHONY: .docker-wait

# Wait for postgresql container start
.docker-wait:
	@echo -n "Checking PG is ready..." ; \
	until [[ `docker inspect -f "{{.State.Health.Status}}" $${DB_CONTAINER:?Must be set}` == healthy ]] ; do sleep 1 ; echo -n "." ; done
	@echo "Ok"

## run tests that use services from docker-compose.yml
test-docker: CMD=up -d nats pg
test-docker: dc .docker-wait .dockertest

.dockertest: export TEST_DSN_PG=postgres://mqbridge:$(TEST_PG_PASS)@localhost:$(TEST_PG_PORT)/mqbridge_test?sslmode=disable
.dockertest: export TEST_DSN_NATS=nats://localhost:$(TEST_NATS_PORT)
.dockertest: coverage.out
	@$(MAKE) -s down

## run tests that run docker themselves
test-docker-self: TEST_TAGS_MORE=,docker
test-docker-self: coverage.out

## show package coverage in html (make cov-html PKG=counter)
cov-html: coverage.out
	$(GO) tool cover -html=coverage.out

## build app
build: $(PRG)

# build webtail command
$(PRG): $(SOURCES) $(PLUGINS)
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -v -o $@ $(BUILD_ARG) -ldflags \
	  "-X main.version=$(APP_VERSION)" ./cmd/$@

## build and run in foreground
run: build
	./$(PRG) --debug

## Changes from last tag
changelog:
	@echo Changes since $(RELEASE)
	@echo
	@git log $(RELEASE)..@ --pretty=format:"* %s"


# ------------------------------------------------------------------------------
## Plugin support
#:

## enable plugin mode
## (this command changes source files)
plugin-on:
	@echo -n "Enable plugin mode.."
ifneq ($(USE_PLUGINS),yes)
	@for n in $(PLUGIN_NAMES) ; do \
	  sed -i "s/package $$n/package main/" plugins/$$n/$$n.go plugins/$$n/*_test.go ; \
	done
	@touch .plugins
	@echo "done."
else
	@echo "done already"
endif

## disable plugin mode
## (this command changes source files back)
plugin-off:
	@echo -n "Disable plugin mode.."
ifeq ($(USE_PLUGINS),yes)
	@for n in $(PLUGIN_NAMES) ; do \
	  sed -i "s/package main/package $$n/" plugins/$$n/$$n.go plugins/$$n/*_test.go ; \
	done
	@rm .plugins
	@echo "done."
else
	@echo "done already"
endif

example.so: plugins/example/*.go
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=plugin -o $@ $<

file.so: plugins/file/*.go
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=plugin -o $@ $<

nats.so: plugins/nats/*.go
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=plugin -o $@ $<

pg.so: plugins/pg/*.go
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=plugin -o $@ $<

# ------------------------------------------------------------------------------
## Prepare distros
#:

## build like docker image from scratch
build-standalone: lint vet test
	GOOS=linux CGO_ENABLED=0 $(GO) build -a -v -o $(PRG) -ldflags \
	  "-X main.version=$(APP_VERSION)" ./cmd/$(PRG)


## build app for all platforms
buildall: lint vet
	@echo "*** $@ ***" ; \
	  for a in "$(ALLARCH)" ; do \
	    echo "** $${a%/*} $${a#*/}" ; \
	    P=$(PRG)-$${a%/*}_$${a#*/} ; \
	    GOOS=$${a%/*} GOARCH=$${a#*/} $(GO) build -o $$P -ldflags \
	      "-X main.version=$(APP_VERSION)" ./cmd/$(PRG) ; \
	  done

## create disro files
dist: clean buildall
	@echo "*** $@ ***"
	@[ -d $(DIRDIST) ] || mkdir $(DIRDIST)
	@sha256sum $(PRG)-* > $(DIRDIST)/SHA256SUMS ; \
	  for a in "$(ALLARCH)" ; do \
	    echo "** $${a%/*} $${a#*/}" ; \
	    P=$(PRG)-$${a%/*}_$${a#*/} ; \
	    zip "$(DIRDIST)/$$P.zip" "$$P" README.md draw.io.dia.png; \
        rm "$$P" ; \
	  done

# ------------------------------------------------------------------------------
## Docker operations
#:

## start service in container
up:
up: CMD=up -d $(DC_SERVICE)
up: dc

## stop service
down:
down: CMD=rm -f -s
down: dc

## build docker image
docker-build: CMD=build --force-rm $(DC_SERVICE)
docker-build: dc

## remove docker image & temp files
docker-clean:
	[ "$$($(DOCKER_BIN) images -q $(DC_IMAGE) 2> /dev/null)" = "" ] || $(DOCKER_BIN) rmi $(DC_IMAGE)

# ------------------------------------------------------------------------------

# $$PWD usage allows host directory mounts in child containers
# Thish works if path is the same for host, docker, docker-compose and child container
## run $(CMD) via docker-compose
dc: docker-compose.yml
	@$(DOCKER_BIN) --project-directory $$PWD -p $(PRG) $(CMD)

# ------------------------------------------------------------------------------
## Other
#:

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
	@for f in $(PLUGINS) ; do [ ! -f $$f ] || rm $$f ; done

## update docs at pkg.go.dev
godoc:
	vf=$(APP_VERSION) ; v=$${vf%%-*} ; echo "Update for $$v..." ; \
	curl 'https://proxy.golang.org/$(GODOC_REPO)/@v/'$$v'.info'

## update latest docker image tag at ghcr.io
ghcr:
	vf=$(APP_VERSION) ; vs=$${vf%%-*} ; v=$${vs#v} ; echo "Update for $$v..." ; \
	docker pull $(DOCKER_IMAGE):$$v && \
	docker tag $(DOCKER_IMAGE):$$v $(DOCKER_IMAGE):latest && \
	docker push $(DOCKER_IMAGE):latest

# This code handles group header and target comment with one or two lines only
## list Makefile targets
## (this is default target)
help:
	@grep -A 1 -h "^## " $(MAKEFILE_LIST) \
  | sed -E 's/^--$$// ; /./{H;$$!d} ; x ; s/^\n## ([^\n]+)\n(## (.+)\n)*(.+):(.*)$$/"    " "\4" "\1" "\3"/' \
  | sed -E 's/^"    " "#" "(.+)" "(.*)"$$/"" "" "" ""\n"\1 \2" "" "" ""/' \
  | xargs printf "%s\033[36m%-15s\033[0m %s %s\n"

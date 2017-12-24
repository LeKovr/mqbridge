# exam project makefile

SHELL          = /bin/bash

# -----------------------------------------------------------------------------
# Docker image config

# Container prefix
DC_PROJECT    ?= mqbridge

# Generated docker image
DC_IMAGE      ?= lekovr/mqbridge:0.1

# docker/compose version
DC_VER        ?= 1.14.0

.PHONY: all help up down dc

##
## Available targets are:
##

# default: show target list
all: help

# ------------------------------------------------------------------------------
# Docker part
# ------------------------------------------------------------------------------

## Start service in container
up:
up: CMD=up -d
up: dc

## Stop service
down:
down: CMD=rm -f -s
down: dc

# ------------------------------------------------------------------------------

# $$PWD используется для того, чтобы текущий каталог был доступен в контейнере по тому же пути
# и относительные тома новых контейнеров могли его использовать
## run docker-compose
dc: docker-compose.yml
	@docker run --rm  -i \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $$PWD:$$PWD \
  -w $$PWD \
  --env=DC_IMAGE1=$(DC_IMAGE) \
  docker/compose:$(DC_VER) \
  -p $(DC_PROJECT) \
  $(CMD)

## Show available make targets
help:
	@grep -A 1 "^##" Makefile | less

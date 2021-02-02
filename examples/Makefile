## mqbridge usage examples
#:

SHELL       = /bin/sh

# -----------------------------------------------------------------------------
# App config

BR_IN      ?= src.txt
BR_OUT     ?= dst.txt

# -----------------------------------------------------------------------------
# Docker image config

# mqbridge docker image
IMAGE      ?= ghcr.io/lekovr/mqbridge
IMAGE_VER  ?= latest

# docker/compose
DC         ?= docker/compose:latest
# Container prefix
DC_PROJECT ?= mqbridge

# PG user
USER_NAME  ?= mqbridge
USER_PASS  ?= $(shell < /dev/urandom tr -dc A-Za-z0-9 2>/dev/null | head -c14; echo)

# ------------------------------------------------------------------------------

.PHONY: all help up down dc

# default: show target list
all: help

# -----------------------------------------------------------------------------
## Tests
#:

# Default test
SERVICES=mqbr-file

## test file -> file bridge
test-file:
test-file: show

## test file -> nats -> file bridges
test-nats: SERVICES=mqbr-nats
test-nats: show

## test file -> pg -> file bridges
test-pg: SERVICES=mqbr-pg
test-pg: show

# Internal targets

show: hi up fill cat bye

hi:
	@echo "** Clear $(BR_IN) and $(BR_OUT) **"
	@echo -n "" > $(BR_IN)
	@echo -n "" > $(BR_OUT)

fill:
	@echo "** Fill $(BR_IN) **"
	@seq 1000 1005 >> $(BR_IN)
	@echo "** Cat $(BR_IN) **"
	@cat $(BR_IN)

cat:
	@echo "** Cat $(BR_OUT) **"
	@while ! test -s $(BR_OUT) ; do sleep 1 ; done
	@cat $(BR_OUT)

bye:
	@echo "** Shutdown services **"
	@$(MAKE) -s down

# -----------------------------------------------------------------------------
## Docker-compose commands
#:

## start service(s)
up:
up: CMD=up -d $(SERVICES)
up: dc

## stop service(s)
down:
down: CMD=rm -f -s
down: dc

# ------------------------------------------------------------------------------

# $$PWD usage allows host directory mounts in child containers
# Thish works if path is the same for host, docker, docker-compose and child container
## run $(CMD) via docker-compose
dc: docker-compose.yml
	@docker run --rm  -i \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $$PWD:$$PWD -w $$PWD \
  --env=IMAGE=$(IMAGE) --env=IMAGE_VER=$(IMAGE_VER) \
  --env=BR_IN=$(BR_IN) --env=BR_OUT=$(BR_OUT) \
  --env=USER_NAME=$(USER_NAME) --env=USER_PASS=$(USER_PASS) \
  $(DC) -p $(DC_PROJECT) $(CMD)

# ------------------------------------------------------------------------------
## Other
#:

## remove temp files
clean:
	@[ ! -f $(BR_IN) ] || rm $(BR_IN)
	@[ ! -f $(BR_OUT) ] || rm -f $(BR_OUT)

# This code handles group header and target comment with one or two lines only
## list Makefile targets
## (this is default target)
help:
	@grep -A 1 -h "^## " $(MAKEFILE_LIST) \
  | sed -E 's/^--$$// ; /./{H;$$!d} ; x ; s/^\n## ([^\n]+)\n(## (.+)\n)*(.+):(.*)$$/"    " "\4" "\1" "\3"/' \
  | sed -E 's/^"    " "#" "(.+)" "(.*)"$$/"" "" "" ""\n"\1 \2" "" "" ""/' \
  | xargs printf "%s\033[36m%-15s\033[0m %s %s\n"

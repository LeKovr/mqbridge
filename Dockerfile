ARG GOLANG_VERSION

FROM golang:$GOLANG_VERSION as builder

# git used for app version fetch
RUN apk add --no-cache git build-base
# gcc g++

WORKDIR /opt/app

# Cached layer
COPY ./go.mod ./go.sum ./
RUN go mod download

# Sources dependent layer
COPY ./ ./
#RUN go generate ./cmd/webtail/...
RUN for plug in example file nats pg ; do \
    go build -buildmode=plugin -o ${plug}.so@ ./plugins/${plug} ; \
  done
RUN go test -tags test -covermode=atomic -coverprofile=coverage.out ./...
RUN go build -ldflags "-X main.version=`git describe --tags --always`" -a ./cmd/mqbridge

# GOOS=linux GOARCH=amd64 
FROM alpine:3.12

WORKDIR /
COPY --from=builder /opt/app/mqbridge .
COPY --from=builder /opt/app/*.so .
ENTRYPOINT ["/mqbridge"]

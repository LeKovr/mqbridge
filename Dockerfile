ARG GOLANG_VERSION

FROM golang:$GOLANG_VERSION as builder

ARG TARGETARCH

WORKDIR /opt/app

# Cached layer
COPY ./go.mod ./go.sum ./

RUN echo "Build for arch $TARGETARCH"

# Sources dependent layer
COPY ./ ./
RUN CGO_ENABLED=0 go test -tags test -covermode=atomic -coverprofile=coverage.out ./...
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=`git describe --tags --always`" -a ./cmd/mqbridge

FROM scratch

WORKDIR /
COPY --from=builder /opt/app/mqbridge .
ENTRYPOINT ["/mqbridge"]

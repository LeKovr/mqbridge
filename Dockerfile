
# cloud.docker.com do not use ARG, we do now use hooks
# ARG golang_version
# FROM golang:$golang_version

FROM golang:1.9.2-alpine3.6

MAINTAINER Alexey Kovrizhkin <lekovr+docker@gmail.com>

# alpine does not have these apps
RUN apk add --no-cache make bash git curl

WORKDIR /go/src/github.com/LeKovr/mqbridge
COPY *.go ./
COPY plugins plugins
COPY types types
COPY Makefile .
COPY glide.* ./

#RUN go get -u github.com/golang/lint/golint
RUN make vendor
RUN make build-standalone

FROM scratch

WORKDIR /
COPY --from=0 /go/src/github.com/LeKovr/mqbridge/mqbridge .

#EXPOSE 8080
ENTRYPOINT ["/mqbridge"]

FROM golang:1.14

RUN mkdir -p /go/src/github.com/http-status-code-server
WORKDIR /go/src/github.com/http-status-code-server
COPY server.go .
RUN go build -ldflags "-linkmode external -extldflags -static" -a server.go

FROM scratch
COPY --from=0 /go/src/github.com/http-status-code-server/server /server

USER 65534:65534
CMD ["/server"]

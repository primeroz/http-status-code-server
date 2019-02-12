FROM golang:1.11

RUN mkdir -p /go/src/github.com/404-server
WORKDIR /go/src/github.com/404-server
COPY server.go .
RUN go build -ldflags "-linkmode external -extldflags -static" -a server.go
#RUN go build server.go

FROM scratch
COPY --from=0 /go/src/github.com/404-server/server /server

USER 65534:65534
CMD ["/server"]

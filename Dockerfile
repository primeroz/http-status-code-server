FROM golang:1.13

RUN mkdir -p /go/src/github.com/http-status-code-server
WORKDIR /go/src/github.com/http-status-code-server
COPY server.go .
RUN go build -ldflags "-s -w" -a server.go
#RUN go build server.go

FROM scratch
COPY --from=0 /go/src/github.com/http-status-code-server/server /server

USER 65534:65534
CMD ["/server"]

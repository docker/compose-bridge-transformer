FROM golang:1.21 AS builder
WORKDIR $GOPATH/src/github.com/docker/transform
COPY .. .
RUN go build -o /go/bin/transform

FROM scratch as transformer
COPY --from=builder /go/bin/transform /transform
CMD ["/transform"]

FROM transformer
COPY templates /templates

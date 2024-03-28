FROM golang:1.21 AS builder
WORKDIR $GOPATH/src/github.com/docker/transform
COPY .. .
RUN go build -o /go/bin/transform

FROM scratch as transformer
LABEL com.docker.compose.bridge=transformation
COPY --from=builder /go/bin/transform /transform
CMD ["/transform"]

FROM transformer
LABEL com.docker.compose.bridge=transformation
COPY templates /templates

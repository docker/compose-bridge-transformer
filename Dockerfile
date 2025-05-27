FROM --platform=${BUILDPLATFORM} golang:1.21 AS builder
WORKDIR $GOPATH/src/github.com/docker/compose-bridge-transformer
COPY . .
RUN go build -o /go/bin/transform

FROM scratch AS transformer
LABEL com.docker.compose.bridge=transformation
COPY --from=builder /go/bin/transform /transform
CMD ["/transform"]

FROM transformer AS kubernetes
LABEL com.docker.compose.bridge=transformation
COPY templates /templates

FROM transformer AS helm
LABEL com.docker.compose.bridge=transformation
COPY helm-templates /templates
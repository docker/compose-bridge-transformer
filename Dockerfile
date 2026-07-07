FROM --platform=${BUILDPLATFORM} tonistiigi/xx:1.9.0 AS xx

FROM --platform=${BUILDPLATFORM} golang:1.24.5 AS builder
COPY --from=xx / /
RUN xx-verify --setup
WORKDIR $GOPATH/src/github.com/docker/compose-bridge-transformer
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS TARGETARCH TARGETVARIANT
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    go build -o /go/bin/transform && \
    xx-verify --static /go/bin/transform

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
ARG LICENSE_AGREEMENT
FROM --platform=${BUILDPLATFORM} golang:1.21 AS builder
WORKDIR $GOPATH/src/github.com/docker/transform
COPY .. .
RUN go build -o /go/bin/transform

FROM scratch AS transformer
LABEL com.docker.compose.bridge=transformation
COPY --from=builder /go/bin/transform /transform
COPY --from=license LICENSE LICENSE
CMD ["/transform"]

FROM transformer
ENV LICENSE_AGREEMENT=${LICENSE_AGREEMENT}
LABEL com.docker.compose.bridge=transformation
COPY templates /templates

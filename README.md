# Compose Bridge Transformer Templates

This repository contains the default Go templates used by the Docker Compose team to generate transformer images for Docker Desktop's Kubernetes cluster integration.

## Overview

The repository provides:

- **Base Transformer Image**: A minimal image containing the core transformation binary.
- **Kubernetes Transformer Image**: Includes templates for generating Kubernetes manifests from Compose files.
- **Helm Charts Transformer Image**: Includes templates for generating Helm charts from Compose files.

## Structure

- `templates/`: Go templates for Kubernetes manifests.
- `helm-templates/`: Go templates for Helm charts.
- `Dockerfile`: Multi-stage build to produce the transformer images.

## Usage

build the transfomer binary
```shell
go build -o /go/bin/transform
```

build the transformer base image for local architecture
```shell
docker bake -f docker-bake.hcl transformer_local
```

build the transformer base image for all architectures
```shell
docker bake -f docker-bake.hcl transformer_all
```

build the kubernetes transformer image for local architecture
```shell
docker bake -f docker-bake.hcl kubernetes_local
```

build the kubernetes transformer image for all architectures
```shell
docker bake -f docker-bake.hcl kubernetes_all
```

build the helm transformer image for local architecture
```shell
docker bake -f docker-bake.hcl helm_local
```

build the helm transformer image for all architectures
```shell
docker bake -f docker-bake.hcl helm_all
```

## Docker Desktop
These templates are used internally by Docker Desktop to enable seamless conversion of Compose files to Kubernetes and deploy them into the internal Kubernetes cluster.

## License

Licensed under the [Apache License 2.0](LICENSE).

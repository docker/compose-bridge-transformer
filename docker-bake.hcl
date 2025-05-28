
# Special target: https://github.com/docker/metadata-action#bake-definition
target "meta-helper" {}

group "default" {
  targets = ["transformer", "kubernetes", "helm"]
}

target "_all-platforms" {
  platforms = [
    "linux/386",
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le",
  ]
}

target "transformer" {
  inherits = ["meta-helper"]
  target = "transformer"
}

target "transformer_all" {
  inherits = ["transformer", "_all-platforms"]
}

target "transformer_local" {
    inherits = ["transformer"]
    tags = ["docker/compose-bridge-transformer"]
}

target "kubernetes" {
  inherits = ["meta-helper"]
  target = "kubernetes"
}

target "kubernetes_all" {
  inherits = ["kubernetes", "_all-platforms"]
}

target "kubernetes_local" {
    inherits = ["kubernetes"]
    tags = ["docker/compose-bridge-kubernetes"]
}

target "helm" {
  inherits = ["meta-helper"]
  target = "helm"
}

target "helm_all" {
  inherits = ["helm", "_all-platforms"]
}

target "helm_local" {
    inherits = ["helm"]
    tags = ["docker/compose-bridge-helm"]
}
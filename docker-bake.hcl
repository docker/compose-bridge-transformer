
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
  tags = ["docker/compose-bridge-transformer"]
}

target "transformer_all" {
  inherits = ["transformer", "_all-platforms"]
}

target "kubernetes" {
  inherits = ["meta-helper"]
  target = "kubernetes"
  tags = ["docker/compose-bridge-kubernetes"]
}

target "kubernetes_all" {
  inherits = ["kubernetes", "_all-platforms"]
}

target "helm" {
  inherits = ["meta-helper"]
  target = "helm"
  tags = ["docker/compose-bridge-helm"]
}

target "helm_all" {
  inherits = ["helm", "_all-platforms"]
}
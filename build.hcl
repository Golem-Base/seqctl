variable "REGISTRY" {
  default = "docker.io"
}

variable "REPOSITORY" {
  default = "golemnetwork"
}

variable "IMAGE" {
  default = "seqctl"
}

variable "TAG" {
  default = "latest"
}

variable "PLATFORMS" {
  default = ["linux/amd64"]
}

target "default" {
  context    = "."
  dockerfile = "Dockerfile"
  tags       = ["${REGISTRY}/${REPOSITORY}/${IMAGE}:${TAG}"]
  platforms  = PLATFORMS
}

target "dev" {
  context    = "."
  dockerfile = "Dockerfile"
  tags       = ["${REGISTRY}/${REPOSITORY}/${IMAGE}:dev"]
  platforms  = PLATFORMS
}

target "release" {
  context    = "."
  dockerfile = "Dockerfile"
  tags = [
    "${REGISTRY}/${REPOSITORY}/${IMAGE}:${TAG}",
    "${REGISTRY}/${REPOSITORY}/${IMAGE}:latest"
  ]
  platforms = PLATFORMS
}

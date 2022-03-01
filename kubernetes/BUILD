load("@io_bazel_rules_k8s//k8s:object.bzl", "k8s_object")

k8s_object(
  name = "password",
  kind = "deployment",
  cluster="cluster.anthony.bible",
  # A template of a Kubernetes Deployment object yaml.
  template = ":password-exchange.deployment.yaml",

  # An optional collection of docker_build images to publish
  # when this target is bazel run.  The digest of the published
  # image is substituted as a part of the resolution process.
  images = {
    "ghcr.io/anthony-bible/passwordexchange-container-prod": "//app:app"
  },
)

k8s_object(
  name = "encryption",
  kind = "deployment",
  cluster="cluster.anthony.bible",
  # A template of a Kubernetes Deployment object yaml.
  template = ":encryption.deployment.yaml",

  # An optional collection of docker_build images to publish
  # when this target is bazel run.  The digest of the published
  # image is substituted as a part of the resolution process.
  images = {
    "ghcr.io/anthony-bible/passwordexchange-encryption": "//app/encryption:encryption"
  },
)


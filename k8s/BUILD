load("@io_bazel_rules_k8s//k8s:object.bzl", "k8s_object")
load("@io_bazel_rules_k8s//k8s:objects.bzl", "k8s_objects")

k8s_object(
    name = "password",
    cluster = "cluster.anthony.bible",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    images = {
        "ghcr.io/anthony-bible/passwordexchange-container-prod": "//app:app",
    },
    kind = "deployment",
    # A template of a Kubernetes Deployment object yaml.
    template = ":password-exchange.deployment.yaml",
)

k8s_object(
    name = "encryption",
    cluster = "cluster.anthony.bible",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    images = {
        "ghcr.io/anthony-bible/passwordexchange-encryption": "//app/encryption:encryption",
    },
    kind = "deployment",
    # A template of a Kubernetes Deployment object yaml.
    template = ":encryption.deployment.yaml",
)

k8s_object(
    name = "database",
    cluster = "cluster.anthony.bible",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    images = {
        "ghcr.io/anthony-bible/passwordexchange-database": "//app/database:database",
    },
    kind = "deployment",
    # A template of a Kubernetes Deployment object yaml.
    template = ":database.deployment.yaml",
)

k8s_object(
    name = "slackbot",
    cluster = "cluster.anthony.bible",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    images = {
        "ghcr.io/anthony-bible/passwordexchange-slackbot": "//slackbot:slackbot_bin",
    },
    kind = "deployment",
    # A template of a Kubernetes Deployment object yaml.
    template = ":slackbot.deployment.yaml",
)


##CREATE ALL deployments
k8s_objects(
    name = "deployments",
    objects = [
        ":password",
        ":slackbot",
        ":encryption",
        ":database",
    ]
)
load("@io_bazel_rules_k8s//k8s:object.bzl", "k8s_object")
load("@io_bazel_rules_k8s//k8s:objects.bzl", "k8s_objects")

k8s_object(
    name = "password",
    cluster = "cluster.local",
    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    images = {
        "ghcr.io/anthony-bible/passwordexchange-container-prod": "//app:app",
    },
    kind = "deployment",
    substitutions = {
        "%{VERSION}": "{VERSION}",
        "%{PHASE}": "{PHASE}",
    },
    # A template of a Kubernetes Deployment object yaml.
    template = ":password-exchange.deployment.yaml",
)

k8s_object(
    name = "password-service",
    cluster = "cluster.local",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    kind = "service",
    substitutions = {
        "%{PHASE}": "{PHASE}",
    },
    # A template of a Kubernetes Deployment object yaml.
    template = ":password-exchange.service.yaml",
)

k8s_object(
    name = "encryption",
    cluster = "cluster.local",
    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    images = {
        "ghcr.io/anthony-bible/passwordexchange-encryption": "//app/encryption:encryption",
    },
    kind = "deployment",
    substitutions = {
        "%{VERSION}": "{VERSION}",
        "%{PHASE}": "{PHASE}",
    },
    # A template of a Kubernetes Deployment object yaml.
    template = ":encryption.deployment.yaml",
)

k8s_object(
    name = "encryption-service",
    cluster = "cluster.local",
    kind = "service",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    substitutions = {
        "%{PHASE}": "{PHASE}",
    },

    # A template of a Kubernetes Deployment object yaml.
    template = ":encryption.service.yaml",
)

k8s_object(
    name = "database",
    cluster = "cluster.local",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    images = {
        "ghcr.io/anthony-bible/passwordexchange-database": "//app/database:database",
    },
    kind = "deployment",
    substitutions = {
        "%{VERSION}": "{VERSION}",
        "%{PHASE}": "{PHASE}",
    },
    # A template of a Kubernetes Deployment object yaml.
    template = ":database.deployment.yaml",
)

k8s_object(
    name = "database-service",
    cluster = "cluster.local",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    kind = "service",
    substitutions = {
        "%{PHASE}": "{PHASE}",
    },
    # A template of a Kubernetes Deployment object yaml.
    template = ":database.service.yaml",
)

k8s_object(
    name = "slackbot",
    cluster = "cluster.local",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    images = {
        "ghcr.io/anthony-bible/passwordexchange-slackbot": "//slackbot:slackbot_bin",
    },
    kind = "deployment",
    substitutions = {
        "%{VERSION}": "{VERSION}",
        "%{PHASE}": "{PHASE}",
    },
    # A template of a Kubernetes Deployment object yaml.
    template = ":slackbot.deployment.yaml",
)

k8s_object(
    name = "slackbot-service",
    cluster = "cluster.local",

    # An optional collection of docker_build images to publish
    # when this target is bazel run.  The digest of the published
    # image is substituted as a part of the resolution process.
    kind = "service",

    # A template of a Kubernetes Deployment object yaml.
    template = ":slackbot.service.yaml",
)

#CREATE ALL Services
k8s_objects(
    name = "services",
    objects = [
        ":password-service",
        ":slackbot-service",
        ":encryption-service",
        ":database-service",
    ],
)

##CREATE ALL deployments
k8s_objects(
    name = "deployments",
    objects = [
        ":password",
        ":slackbot",
        ":encryption",
        ":database",
    ],
)

k8s_objects(
    name = "deployments-and-services",
    objects = [
        ":deployments",
        ":services",
    ],
)

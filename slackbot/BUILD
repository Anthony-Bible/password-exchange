load("@rules_python//python:defs.bzl", "py_library")
load("@io_bazel_rules_docker//container:container.bzl", "container_image")
load("@io_bazel_rules_docker//python3:image.bzl", "py3_image")
load("@pip//:requirements.bzl", "all_whl_requirements")
load("@rules_python//gazelle/manifest:defs.bzl", "gazelle_python_manifest")
load("@rules_python//gazelle/modules_mapping:def.bzl", "modules_mapping")

# This rule fetches the metadata for python packages we depend on. That data is
# required for the gazelle_python_manifest rule to update our manifest file.
modules_mapping(
    name = "modules_map",
    wheels = all_whl_requirements,
)

# Gazelle python extension needs a manifest file mapping from
# an import to the installed package that provides it.
# This macro produces two targets:
# - //:gazelle_python_manifest.update can be used with `bazel run`
#   to recalculate the manifest
# - //:gazelle_python_manifest.test is a test target ensuring that
#   the manifest doesn't need to be updated
gazelle_python_manifest(
    name = "gazelle_python_manifest",
    modules_mapping = ":modules_map",
    # This is what we called our `pip_install` rule, where third-party
    # python libraries are loaded in BUILD files.
    pip_repository_name = "pip",
    # When using pip_parse instead of pip_install, set the following.
    # pip_repository_incremental = True,
    # This should point to wherever we declare our python dependencies
    # (the same as what we passed to the modules_mapping rule in WORKSPACE)
    requirements = ":requirements.txt",
)

#
#py_library(
#    name = "slackbot_lib",
#    srcs = ["app.py"],
#    importpath = "github.com/Anthony-Bible/password-exchange/app",
#    visibility = ["//visibility:private"],
#    deps = [
#        "//app/commons",
#        "//app/email",
#        "//app/message",
#        "//protos:databasepb",
#        "//protos:encryptionpb",
#        "@com_github_gin_gonic_gin//:gin",
#        "@com_github_rs_xid//:xid",
#        "@com_github_rs_zerolog//log",
#        "@org_golang_google_grpc//:go_default_library",
#    ],
#)
#
#container_image(
#    name = "forms_image",
#    base = "@alpine_linux_amd64//image",
#    data_path = "/",
#    files = ["//app/templates:tarred_file"],
#    tars = ["//app/templates:tarred_file"],
#)
#
#go_image(
#    name = "app",
#    base = ":forms_image",
#    basename = "app",
#    embed = [":app_lib"],
#    pure = "on",
#    visibility = ["//visibility:public"],
#)

py_library(
    name = "slackbot",
    srcs = [
        "program.py",
        "encryptionClient.py",
        "database.py",
    ],
    srcs_version = "PY3",
    imports = [".."],
    visibility = ["//:__subpackages__"],
    deps = [
        # "//protos:database_py_pb2",
        "//protos:databasepb_grpc_py_library",
        # "//protos:encryption_py_pb2",
        "//protos:encryptionpb_grpc_py_library",
        "@pip//pypi__flask",
        "@pip//pypi__grpcio",
        "@pip//pypi__protobuf",
        "@pip//pypi__slack_bolt",
        "@pip//pypi__slack_sdk",
    ],
)
container_image(
    name = "python38_base_image",
    base = "@python3_linux_amd64//image",
    # slim-buster image places python3 under /usr/local/bin, but the host
    # toolchain used by py3_image might use /usr/bin instead.
    symlinks = {
        "/usr/bin/python": "/usr/local/bin/python",
        "/usr/bin/python3": "/usr/local/bin/python3",
    },
)
py3_image(
    name = "slackbot_bin",
    srcs = ["program.py"],
    srcs_version = "PY3",
    main = "program.py",
    base = ":python38_base_image",
    deps = [":slackbot"],
    visibility = ["//visibility:public"],
)

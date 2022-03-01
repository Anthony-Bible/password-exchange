load("@rules_pkg//:pkg.bzl", "pkg_tar")

filegroup(
    name = "template_files",
    srcs = glob([
        "**",
    ]),
)
pkg_tar(
    name = "tarred_file",
    srcs = [":template_files"],
    package_dir = "/templates",
    mode = "0o644",
    visibility = ["//visibility:public"],
    extension = "tgz",
)


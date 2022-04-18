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
    extension = "tgz",
    strip_prefix = ".",
    mode = "0o644",
    package_dir = "/templates",
    visibility = ["//visibility:public"],
)

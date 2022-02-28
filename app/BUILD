load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")
load("@io_bazel_rules_docker//container:container.bzl", "container_image")
load("@rules_pkg//:pkg.bzl", "pkg_tar")


load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/Anthony-Bible/password-exchange/app
gazelle(name = "gazelle")

#go_library(
#    name = "project_lib",
#    srcs = ["forms.go"],
#    importpath = "github.com/Anthony-Bible/password-exchange/app/email",
#    visibility = ["//visibility:private"],
#    deps = [
#        "//protos:encryptionpb",
#        "@com_github_anthony_bible_password_exchange_app//commons:go_default_library",
#        "@com_github_anthony_bible_password_exchange_app//email:go_default_library",
#        "@com_github_anthony_bible_password_exchange_app//message:go_default_library",
#        "@com_github_gin_gonic_gin//:go_default_library",
#        "@com_github_rs_xid//:go_default_library",
#        "@com_github_rs_zerolog//log:go_default_library",
#        "@org_golang_google_grpc//:go_default_library",
#    ],
#)
go_library(
    name = "app_lib",
    srcs = ["forms.go"],
    importpath = "github.com/Anthony-Bible/password-exchange/app",
    visibility = ["//visibility:private"],
    deps = [
        "//app/commons",
        "//app/database",
        "//app/email",
        "//app/message",
        "//app/protos:encryptionpb",
        "@com_github_gin_gonic_gin//:gin",
        "@com_github_rs_xid//:xid",
        "@com_github_rs_zerolog//log",
        "@org_golang_google_grpc//:go_default_library",
    ],
)


container_image(
    name = "forms_image",
    base = "@alpine_linux_amd64//image",
    tars = ["//app/templates:tarred_file"],
    files= ["//app/templates:tarred_file"],
    data_path = "/",

)

go_image(
    name="app",
    basename = "app",
    embed = [":app_lib"],
    visibility = ["//visibility:public"],
    base = ":forms_image",
    pure= "on",
)

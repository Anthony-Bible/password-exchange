load("@rules_python//python:defs.bzl", "py_runtime", "py_runtime_pair")

constraint_value(
    name = "dont_run_in_container",
    constraint_setting = "@io_bazel_rules_docker//platforms:containerized",
)

platform(
    name = "local_config_platform_but_dont_run_in_container",
    constraint_values = [":dont_run_in_container"],
    parents = ["@local_config_platform//:host"],
)

#py_runtime(
#    name = "python3_runtime",
#    files = ["@python310_resolved_interpreter//:files"],
#    interpreter = "@python310_resolved_interpreter//:python_bin",
#    python_version = "PY3",
#    visibility = ["//visibility:public"],
#)
py_runtime(
    name = "default_container_py3_runtime",
    interpreter_path = "/usr/local/bin/python3",
    python_version = "PY3",
)

py_runtime(
    name = "python3_runtime",
    files = ["@python38//:files"],
    interpreter = "@python38//:python3",
    python_version = "PY3",
    visibility = ["//visibility:public"],
)

# A toolchain to run python outputs inside a container.
# If you are using a custom base for py_image which has python tools in a
# different location, you must register that toolchain prior to the

py_runtime_pair(
    name = "py_runtime_pair",
    py2_runtime = None,
    py3_runtime = ":python3_runtime",
)

py_runtime_pair(
    name = "default_container_py_runtime_pair",
    py3_runtime = ":default_container_py3_runtime",
)

toolchain(
    name = "my_py_toolchain",
    exec_compatible_with = [":dont_run_in_container"],
    toolchain = ":py_runtime_pair",
    toolchain_type = "@bazel_tools//tools/python:toolchain_type",
)

# registration of this one in @io_bazel _rules_docker//python:image.bzl
toolchain(
    name = "container_py_toolchain",
    exec_compatible_with = [
        "@io_bazel_rules_docker//platforms:run_in_container",
    ],
    toolchain = ":default_container_py_runtime_pair",
    toolchain_type = "@bazel_tools//tools/python:toolchain_type",
)

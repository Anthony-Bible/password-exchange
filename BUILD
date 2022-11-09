load("@rules_python//gazelle:def.bzl", "GAZELLE_PYTHON_RUNTIME_DEPS")
load("@bazel_gazelle//:def.bzl", "gazelle", "gazelle_binary")

# gazelle:proto package
# gazelle:proto_group go_package
## Go ##
# gazelle:proto_plugin protoc-gen-go implementation golang:protobuf:protoc-gen-go
# gazelle:proto_rule proto_go_library implementation stackb:rules_proto:proto_go_library
# gazelle:proto_rule proto_go_library deps @org_golang_google_protobuf//reflect/protoreflect
# gazelle:proto_rule proto_go_library deps @org_golang_google_protobuf//runtime/protoimpl
# gazelle:proto_rule proto_go_library visibility //visibility:public
# gazelle:proto_language go plugin protoc-gen-go
# gazelle:proto_language go rule proto_compile
## Hold off on proto_go_library until we upgrade the deps of this repo
## gazelle:proto_language go rule proto_go_library
## Python ##
# gazelle:proto_plugin python implementation builtin:python
# gazelle:proto_plugin protoc-gen-grpc-python implementation grpc:grpc:protoc-gen-grpc-python
# gazelle:proto_rule proto_python_library implementation stackb:rules_proto:proto_py_library
# gazelle:proto_rule proto_python_library deps @com_google_protobuf//:protobuf_python
# gazelle:proto_rule proto_python_library visibility //visibility:public
# gazelle:proto_rule grpc_py_library implementation stackb:rules_proto:grpc_py_library
# gazelle:proto_rule proto_compile implementation stackb:rules_proto:proto_compile
# TODO: add grpc_py_library deps
# gazelle:proto_rule grpc_py_library visibility //visibility:public
# gazelle:proto_language python plugin python
# gazelle:proto_language python plugin protoc-gen-grpc-python
# gazelle:proto_language python rule proto_compile
# gazelle:proto_language python rule proto_python_library
# gazelle:proto_language python rule grpc_py_library
##Python pip ##
#gazelle:resolve py encryption_py_pb2 //protos:encryption_py_pb2
#gazelle:resolve py encryption_py_pb2_grpc //protos:encryption_py_pb2_grpc
#gazelle:resolve py protos //protos:encryption_pb2_grpc
#gazelle:resolve py database_py_pb2 //protos:database_py_pb2
#gazelle:resolve py database_py_pb2_grpc //protos:database_py_pb2_grpc

# gazelle:prefix github.com/Anthony-Bible/password-exchange/
# --- show debugging output ---
# gazelle:log_level debug

# --- show summary of total time on .Info ---
# gazelle:progress true

gazelle_binary(
    name = "gazelle_debug",
    languages = [
        "@rules_python//gazelle",
        "@bazel_gazelle//language/go",
        "@bazel_gazelle//language/proto",
        # must be after the proto extension (order matters)
        "@build_stack_rules_proto//language/protobuf",
        "@build_stack_bazel_gazelle_debug//language/debug",
    ],
)

gazelle(
    name = "gazelle",
    data = GAZELLE_PYTHON_RUNTIME_DEPS,
    gazelle = "//:gazelle_debug",
)

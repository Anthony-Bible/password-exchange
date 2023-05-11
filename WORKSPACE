load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "099a9fb96a376ccbbb7d291ed4ecbdfd42f6bc822ab77ae6f1b5cb9e914e94fa",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.35.0/rules_go-v0.35.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.35.0/rules_go-v0.35.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "501deb3d5695ab658e82f6f6f549ba681ea3ca2a5fb7911154b5aa45596183fa",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.26.0/bazel-gazelle-v0.26.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.26.0/bazel-gazelle-v0.26.0.tar.gz",
    ],
)

http_archive(
    name = "build_stack_rules_proto",
    sha256 = "733bdc9267a90404d48668853025bb6c660d7a23e38f819335b03865fe2bee89",
    strip_prefix = "rules_proto-36ceb79a987a6de33768c8bdb08d22b516a7e32e",
    urls = ["https://github.com/stackb/rules_proto/archive/36ceb79a987a6de33768c8bdb08d22b516a7e32e.tar.gz"],
)

# Branch: master
# Commit: 66315dd31d70f2e03d7dfaa310f4d549be6522e4
# Date: 2021-10-25 18:13:58 +0000 UTC
# URL: https://github.com/stackb/bazel-gazelle-debug/commit/66315dd31d70f2e03d7dfaa310f4d549be6522e4
#
# Update ws name
# Size: 14361 (14 kB)
http_archive(
    name = "build_stack_bazel_gazelle_debug",
    sha256 = "c4e8d487a06792b0086651c44cc0795d816d4b53f21e84f11b6bb4cb1746df5e",
    strip_prefix = "bazel-gazelle-debug-2518a88cb0908f1896ff50b506e87d880a9695cb",
    urls = ["https://github.com/stackb/bazel-gazelle-debug/archive/2518a88cb0908f1896ff50b506e87d880a9695cb.tar.gz"],
)

http_archive(
    name = "bazel_skylib",
    sha256 = "74d544d96f4a5bb630d465ca8bbcfe231e3594e5aae57e1edbf17a6eb3ca2506",
    urls = [
        "https://github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
    ],
)

http_archive(
    name = "rules_proto_grpc",
    sha256 = "507e38c8d95c7efa4f3b1c0595a8e8f139c885cb41a76cab7e20e4e67ae87731",
    strip_prefix = "rules_proto_grpc-4.1.1",
    urls = ["https://github.com/rules-proto-grpc/rules_proto_grpc/archive/4.1.1.tar.gz"],
)

# Fetch official Python rules for Bazel
#INSTALL PYTHON RULES
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "rules_python",
    sha256 = "cdf6b84084aad8f10bf20b46b77cb48d83c319ebe6458a18e9d2cebf57807cdd",
    strip_prefix = "rules_python-0.8.1",
    url = "https://github.com/bazelbuild/rules_python/archive/refs/tags/0.8.1.tar.gz",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("@rules_python//gazelle:deps.bzl", _py_gazelle_deps = "gazelle_deps")
load("@rules_proto_grpc//:repositories.bzl", "rules_proto_grpc_toolchains")
load("@build_stack_rules_proto//:go_deps.bzl", "gazelle_protobuf_extension_go_deps")
load("@build_stack_rules_proto//deps:protobuf_core_deps.bzl", "protobuf_core_deps")

go_repository(
    name = "com_github_cespare_xxhash_v2",
    importpath = "github.com/cespare/xxhash/v2",
    sum = "h1:YRXhKfTDauu4ajMg1TPgFO5jnlC2HCbmLXMcTG5cbYE=",
    version = "v2.1.2",
)

go_repository(
    name = "com_github_alecthomas_template",
    importpath = "github.com/alecthomas/template",
    sum = "h1:JYp7IbQjafoB+tBA3gMyHYHrpOtNuDiK/uB5uXxq5wM=",
    version = "v0.0.0-20190718012654-fb15b899a751",
)

go_repository(
    name = "com_github_alecthomas_units",
    importpath = "github.com/alecthomas/units",
    sum = "h1:UQZhZ2O0vMHr2cI+DC1Mbh0TJxzA3RcLoMsFw+aXw7E=",
    version = "v0.0.0-20190924025748-f65c72e2690d",
)

go_repository(
    name = "com_github_beorn7_perks",
    importpath = "github.com/beorn7/perks",
    sum = "h1:VlbKKnNfV8bJzeqoa4cOKqO6bYr3WgKZxO8Z16+hsOM=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_circonus_labs_circonus_gometrics",
    importpath = "github.com/circonus-labs/circonus-gometrics",
    sum = "h1:C29Ae4G5GtYyYMm1aztcyj/J5ckgJm2zwdDajFbx1NY=",
    version = "v2.3.1+incompatible",
)

go_repository(
    name = "com_github_circonus_labs_circonusllhist",
    importpath = "github.com/circonus-labs/circonusllhist",
    sum = "h1:TJH+oke8D16535+jHExHj4nQvzlZrj7ug5D7I/orNUA=",
    version = "v0.1.3",
)

go_repository(
    name = "com_github_datadog_datadog_go",
    importpath = "github.com/DataDog/datadog-go",
    sum = "h1:qSG2N4FghB1He/r2mFrWKCaL7dXCilEuNEeAn20fdD4=",
    version = "v3.2.0+incompatible",
)

go_repository(
    name = "com_github_dustin_go_humanize",
    importpath = "github.com/dustin/go-humanize",
    sum = "h1:VSnTsYCnlFHaM2/igO1h6X3HA71jcobQuxemgkq4zYo=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_frankban_quicktest",
    importpath = "github.com/frankban/quicktest",
    sum = "h1:FJKSZTDHjyhriyC81FLQ0LY93eSai0ZyR/ZIkd3ZUKE=",
    version = "v1.14.3",
)

go_repository(
    name = "com_github_go_kit_kit",
    importpath = "github.com/go-kit/kit",
    sum = "h1:e4o3o3IsBfAKQh5Qbbiqyfu97Ku7jrO/JbohvztANh4=",
    version = "v0.12.0",
)

go_repository(
    name = "com_github_go_kit_log",
    importpath = "github.com/go-kit/log",
    sum = "h1:7i2K3eKTos3Vc0enKCfnVcgHh2olr/MyfboYq7cAcFw=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_go_logfmt_logfmt",
    importpath = "github.com/go-logfmt/logfmt",
    sum = "h1:otpy5pqBCBZ1ng9RQ0dPu4PN7ba75Y/aA+UpowDyNVA=",
    version = "v0.5.1",
)

go_repository(
    name = "com_github_go_stack_stack",
    importpath = "github.com/go-stack/stack",
    sum = "h1:MmJCxYVKTJ0SplGKqFVX3SBnmaUhODHZrrFF6jMbpZk=",
    version = "v1.6.0",
)

go_repository(
    name = "com_github_golang_snappy",
    importpath = "github.com/golang/snappy",
    sum = "h1:K9KHZbXKpGydfDN0aZrsoHpLJlZsBrGMFWbgLDGnPZk=",
    version = "v0.0.0-20170215233205-553a64147049",
)

go_repository(
    name = "com_github_googleapis_google_cloud_go_testing",
    importpath = "github.com/googleapis/google-cloud-go-testing",
    sum = "h1:tlyzajkF3030q6M8SvmJSemC9DTHL/xaMa18b65+JM4=",
    version = "v0.0.0-20200911160855-bcd43fbb19e8",
)

go_repository(
    name = "com_github_grpc_ecosystem_go_grpc_prometheus",
    importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
    sum = "h1:Ovs26xHkKqVztRpIrF/92BcuyuQ/YW4NSIpoGtfXNho=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_hashicorp_go_hclog",
    importpath = "github.com/hashicorp/go-hclog",
    sum = "h1:K4ev2ib4LdQETX5cSZBG0DVLk1jwGqSPXBjdah3veNs=",
    version = "v0.16.2",
)

go_repository(
    name = "com_github_hashicorp_go_retryablehttp",
    importpath = "github.com/hashicorp/go-retryablehttp",
    sum = "h1:QlWt0KvWT0lq8MFppF9tsJGF+ynG7ztc2KIPhzRGk7s=",
    version = "v0.5.3",
)

go_repository(
    name = "com_github_jpillora_backoff",
    importpath = "github.com/jpillora/backoff",
    sum = "h1:uvFg412JmmHBHw7iwprIxkPMI+sGQ4kzOWsMeHnm2EA=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_julienschmidt_httprouter",
    importpath = "github.com/julienschmidt/httprouter",
    sum = "h1:U0609e9tgbseu3rBINet9P48AI/D3oJs4dN7jwJOQ1U=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_konsorten_go_windows_terminal_sequences",
    importpath = "github.com/konsorten/go-windows-terminal-sequences",
    sum = "h1:CE8S1cTafDpPvMhIxNJKvHsGVBgn1xWYf1NbHQhywc8=",
    version = "v1.0.3",
)

go_repository(
    name = "com_github_kr_logfmt",
    importpath = "github.com/kr/logfmt",
    sum = "h1:T+h1c/A9Gawja4Y9mFVWj2vyii2bbUNDw3kt9VxK2EY=",
    version = "v0.0.0-20140226030751-b84e30acd515",
)

go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
    sum = "h1:4hp9jkHxhMHkqkrB3Ix0jegS5sx/RkqARlsWZ6pIwiU=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_mwitkow_go_conntrack",
    importpath = "github.com/mwitkow/go-conntrack",
    sum = "h1:KUppIJq7/+SVif2QVs3tOP0zanoHgBEVAwHxUSIzRqU=",
    version = "v0.0.0-20190716064945-2f068394615f",
)

go_repository(
    name = "com_github_pelletier_go_toml_v2",
    importpath = "github.com/pelletier/go-toml/v2",
    sum = "h1:ipoSadvV8oGUjnUbMub59IDPPwfxF694nG/jwbMiyQg=",
    version = "v2.0.5",
)

go_repository(
    name = "com_github_prometheus_client_golang",
    importpath = "github.com/prometheus/client_golang",
    sum = "h1:HNkLOAEQMIDv/K+04rukrLx6ch7msSRwf3/SASFAGtQ=",
    version = "v1.11.0",
)

go_repository(
    name = "com_github_prometheus_common",
    importpath = "github.com/prometheus/common",
    sum = "h1:JEkYlQnpzrzQFxi6gnukFPdQ+ac82oRhzMcIduJu/Ug=",
    version = "v0.30.0",
)

go_repository(
    name = "com_github_prometheus_procfs",
    importpath = "github.com/prometheus/procfs",
    sum = "h1:4jVXhlkAyzOScmCkXBTOLRLTz8EeU+eyjrwB/EPq0VU=",
    version = "v0.7.3",
)

go_repository(
    name = "com_github_sagikazarmark_crypt",
    importpath = "github.com/sagikazarmark/crypt",
    sum = "h1:xtk0uUHVWVsRBdEUGYBym4CXbcllXky2M7Qlwsf8C0Y=",
    version = "v0.8.0",
)

go_repository(
    name = "com_github_sirupsen_logrus",
    importpath = "github.com/sirupsen/logrus",
    sum = "h1:dJKuHgqk1NNQlqoA6BTlM1Wf9DOH3NBjQyu0h9+AZZE=",
    version = "v1.8.1",
)

go_repository(
    name = "com_github_tv42_httpunix",
    importpath = "github.com/tv42/httpunix",
    sum = "h1:G3dpKMzFDjgEh2q1Z7zUUtKa8ViPtH+ocF0bE0g00O8=",
    version = "v0.0.0-20150427012821-b75d8614f926",
)

go_repository(
    name = "com_google_cloud_go_compute",
    importpath = "cloud.google.com/go/compute",
    sum = "h1:gKVJMEyqV5c/UnpzjjQbo3Rjvvqpr9B1DFSbJC4OXr0=",
    version = "v1.12.1",
)

go_repository(
    name = "in_gopkg_alecthomas_kingpin_v2",
    importpath = "gopkg.in/alecthomas/kingpin.v2",
    sum = "h1:jMFz6MfLP0/4fUyZle81rXUoxOBFi19VUFKVDOQfozc=",
    version = "v2.2.6",
)

go_repository(
    name = "io_etcd_go_etcd_client_v3",
    importpath = "go.etcd.io/etcd/client/v3",
    sum = "h1:62Eh0XOro+rDwkrypAGDfgmNh5Joq+z+W9HZdlXMzek=",
    version = "v3.5.0",
)

go_repository(
    name = "io_k8s_sigs_yaml",
    importpath = "sigs.k8s.io/yaml",
    sum = "h1:kr/MCeFWJWTwyaHoR9c8EjH9OumOmoF9YGiZd7lFm/Q=",
    version = "v1.2.0",
)

go_repository(
    name = "org_golang_google_grpc_cmd_protoc_gen_go_grpc",
    importpath = "google.golang.org/grpc/cmd/protoc-gen-go-grpc",
    sum = "h1:M1YKkFIboKNieVO5DLUEVzQfGwJD30Nv2jfUgzb5UcE=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_rabbitmq_amqp091_go",
    importpath = "github.com/rabbitmq/amqp091-go",
    sum = "h1:VouyHPBu1CrKyJVfteGknGOGCzmOz0zcv/tONLkb7rg=",
    version = "v1.5.0",
)

go_repository(
    name = "org_uber_go_goleak",
    importpath = "go.uber.org/goleak",
    sum = "h1:gZAh5/EyT/HQwlpkCy6wTpqfH9H8Lz8zbm3dZh+OyzA=",
    version = "v1.1.12",
)

go_repository(
    name = "com_github_cpuguy83_go_md2man_v2",
    importpath = "github.com/cpuguy83/go-md2man/v2",
    sum = "h1:p1EgwI/C7NhT0JmVkwCD2ZBK8j4aeHQX2pMHHBfMQ6w=",
    version = "v2.0.2",
)

go_repository(
    name = "com_github_inconshreveable_mousetrap",
    importpath = "github.com/inconshreveable/mousetrap",
    sum = "h1:U3uMjPSQEBMNp1lFxmllqCPM6P5u/Xq7Pgzkat/bFNc=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_russross_blackfriday_v2",
    importpath = "github.com/russross/blackfriday/v2",
    sum = "h1:JIOH55/0cWyOuilr9/qlrm0BSXldqnqwMsf35Ld67mk=",
    version = "v2.1.0",
)

go_repository(
    name = "com_github_spf13_cobra",
    importpath = "github.com/spf13/cobra",
    sum = "h1:o94oiPyS4KD1mPy2fmcYYHHfCxLqYjJOhGsCHFZtEzA=",
    version = "v1.6.1",
)

go_repository(
    name = "com_github_streadway_amqp",
    importpath = "github.com/streadway/amqp",
    sum = "h1:kuuDrUJFZL1QYL9hUNuCxNObNzB0bV/ZG5jV3RWAQgo=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_afex_hystrix_go",
    importpath = "github.com/afex/hystrix-go",
    sum = "h1:rFw4nCn9iMW+Vajsk51NtYIcwSTkXr+JGrMd36kTDJw=",
    version = "v0.0.0-20180502004556-fa1af6a1f4f5",
)

go_repository(
    name = "com_github_aws_aws_sdk_go",
    importpath = "github.com/aws/aws-sdk-go",
    sum = "h1:QN1nsY27ssD/JmW4s83qmSb+uL6DG4GmCDzjmJB4xUI=",
    version = "v1.40.45",
)

go_repository(
    name = "com_github_aws_aws_sdk_go_v2",
    importpath = "github.com/aws/aws-sdk-go-v2",
    sum = "h1:ZbovGV/qo40nrOJ4q8G33AGICzaPI45FHQWJ9650pF4=",
    version = "v1.9.1",
)

go_repository(
    name = "com_github_aws_aws_sdk_go_v2_service_cloudwatch",
    importpath = "github.com/aws/aws-sdk-go-v2/service/cloudwatch",
    sum = "h1:w/fPGB0t5rWwA43mux4e9ozFSH5zF1moQemlA131PWc=",
    version = "v1.8.1",
)

go_repository(
    name = "com_github_aws_smithy_go",
    importpath = "github.com/aws/smithy-go",
    sum = "h1:AEwwwXQZtUwP5Mz506FeXXrKBe0jA8gVM+1gEcSRooc=",
    version = "v1.8.0",
)

go_repository(
    name = "com_github_bradfitz_gomemcache",
    importpath = "github.com/bradfitz/gomemcache",
    sum = "h1:7IjN4QP3c38xhg6wz8R3YjoU+6S9e7xBc0DAVLLIpHE=",
    version = "v0.0.0-20170208213004-1952afaa557d",
)

go_repository(
    name = "com_github_casbin_casbin_v2",
    importpath = "github.com/casbin/casbin/v2",
    sum = "h1:/poEwPSovi4bTOcP752/CsTQiRz2xycyVKFG7GUhbDw=",
    version = "v2.37.0",
)

go_repository(
    name = "com_github_cenkalti_backoff_v4",
    importpath = "github.com/cenkalti/backoff/v4",
    sum = "h1:G2HAfAmvm/GcKan2oOQpBXOd2tT2G57ZnZGWa1PxPBQ=",
    version = "v4.1.1",
)

go_repository(
    name = "com_github_clbanning_mxj",
    importpath = "github.com/clbanning/mxj",
    sum = "h1:HuhwZtbyvyOw+3Z1AowPkU87JkJUSv751ELWaiTpj8I=",
    version = "v1.8.4",
)

go_repository(
    name = "com_github_edsrzf_mmap_go",
    importpath = "github.com/edsrzf/mmap-go",
    sum = "h1:CEBF7HpRnUCSJgGUb5h1Gm7e3VkmVDrR8lvWVLtrOFw=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_franela_goreq",
    importpath = "github.com/franela/goreq",
    sum = "h1:a9ENSRDFBUPkJ5lCgVZh26+ZbGyoVJG7yb5SSzF5H54=",
    version = "v0.0.0-20171204163338-bcd34c9993f8",
)

go_repository(
    name = "com_github_garyburd_redigo",
    importpath = "github.com/garyburd/redigo",
    sum = "h1:Sk0u0gIncQaQD23zAoAZs2DNi2u2l5UTLi4CmCBL5v8=",
    version = "v1.1.1-0.20170914051019-70e1b1943d4f",
)

go_repository(
    name = "com_github_gin_contrib_gzip",
    importpath = "github.com/gin-contrib/gzip",
    sum = "h1:ezvKOL6jH+jlzdHNE4h9h8q8uMpDQjyl0NN0Jd7jozc=",
    version = "v0.0.1",
)

go_repository(
    name = "com_github_go_openapi_jsonpointer",
    importpath = "github.com/go-openapi/jsonpointer",
    sum = "h1:nH6xp8XdXHx8dqveo0ZuJBluCO2qGrPbDNZ0dwoRHP0=",
    version = "v0.17.0",
)

go_repository(
    name = "com_github_go_openapi_jsonreference",
    importpath = "github.com/go-openapi/jsonreference",
    sum = "h1:BqWKpV1dFd+AuiKlgtddwVIFQsuMpxfBDBHGfM2yNpk=",
    version = "v0.19.0",
)

go_repository(
    name = "com_github_go_openapi_spec",
    importpath = "github.com/go-openapi/spec",
    sum = "h1:A4SZ6IWh3lnjH0rG0Z5lkxazMGBECtrZcbyYQi+64k4=",
    version = "v0.19.0",
)

go_repository(
    name = "com_github_go_openapi_swag",
    importpath = "github.com/go-openapi/swag",
    sum = "h1:iqrgMg7Q7SvtbWLlltPrkMs0UBJI6oTSs79JFRUi880=",
    version = "v0.17.0",
)

go_repository(
    name = "com_github_go_zookeeper_zk",
    importpath = "github.com/go-zookeeper/zk",
    sum = "h1:4mx0EYENAdX/B/rbunjlt5+4RTA/a9SMHBRuSKdGxPM=",
    version = "v1.0.2",
)

go_repository(
    name = "com_github_goccy_go_json",
    importpath = "github.com/goccy/go-json",
    sum = "h1:IcB+Aqpx/iMHu5Yooh7jEzJk1JZ7Pjtmys2ukPr7EeM=",
    version = "v0.9.7",
)

go_repository(
    name = "com_github_golang_gddo",
    importpath = "github.com/golang/gddo",
    sum = "h1:fHbEOLXdSMRJrzJwQkKixEg3Uq2RIcNQF0Y7iiQ+gRk=",
    version = "v0.0.0-20200310004957-95ce5a452273",
)

go_repository(
    name = "com_github_golang_jwt_jwt_v4",
    importpath = "github.com/golang-jwt/jwt/v4",
    sum = "h1:RAqyYixv1p7uEnocuy8P1nru5wprCh/MH2BIlW5z5/o=",
    version = "v4.0.0",
)

go_repository(
    name = "com_github_golang_lint",
    importpath = "github.com/golang/lint",
    sum = "h1:ior8LN6127GsA53E9mD9nH/oP/LVbJplmLH5V8o+/Uk=",
    version = "v0.0.0-20170918230701-e5d664eb928e",
)

go_repository(
    name = "com_github_googleapis_enterprise_certificate_proxy",
    importpath = "github.com/googleapis/enterprise-certificate-proxy",
    sum = "h1:y8Yozv7SZtlU//QXbezB6QkpuE6jMD2/gfzk4AftXjs=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_googleapis_gax_go",
    importpath = "github.com/googleapis/gax-go",
    sum = "h1:j0GKcs05QVmm7yesiZq2+9cxHkNK9YM6zKx4D2qucQU=",
    version = "v2.0.0+incompatible",
)

go_repository(
    name = "com_github_gregjones_httpcache",
    importpath = "github.com/gregjones/httpcache",
    sum = "h1:vM1v1UTa2Ny7gGhGhzR4CdX2MPyisKM/fXoCZTunV6c=",
    version = "v0.0.0-20170920190843-316c5e0ff04e",
)

go_repository(
    name = "com_github_hdrhistogram_hdrhistogram_go",
    importpath = "github.com/HdrHistogram/hdrhistogram-go",
    sum = "h1:5IcZpTvzydCQeHzK4Ef/D5rrSqwxob0t8PQPMybUNFM=",
    version = "v1.1.2",
)

go_repository(
    name = "com_github_hudl_fargo",
    importpath = "github.com/hudl/fargo",
    sum = "h1:ZDDILMbB37UlAVLlWcJ2Iz1XuahZZTDZfdCKeclfq2s=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_inconshreveable_log15",
    importpath = "github.com/inconshreveable/log15",
    sum = "h1:g/SJtZVYc1cxSB8lgrgqeOlIdi4MhqNNHYRAC8y+g4c=",
    version = "v0.0.0-20170622235902-74a0988b5f80",
)

go_repository(
    name = "com_github_influxdata_influxdb1_client",
    importpath = "github.com/influxdata/influxdb1-client",
    sum = "h1:HqW4xhhynfjrtEiiSGcQUd6vrK23iMam1FO8rI7mwig=",
    version = "v0.0.0-20200827194710-b269163b24ab",
)

go_repository(
    name = "com_github_jmespath_go_jmespath",
    importpath = "github.com/jmespath/go-jmespath",
    sum = "h1:BEgLn5cpjn8UN1mAw4NjwDrS35OdebyEtFe+9YPoQUg=",
    version = "v0.4.0",
)

go_repository(
    name = "com_github_klauspost_compress",
    importpath = "github.com/klauspost/compress",
    sum = "h1:P76CopJELS0TiO2mebmnzgWaajssP/EszplttgQxcgc=",
    version = "v1.13.6",
)

go_repository(
    name = "com_github_knetic_govaluate",
    importpath = "github.com/Knetic/govaluate",
    sum = "h1:1G1pk05UrOh0NlF1oeaaix1x8XzrfjIDK47TY0Zehcw=",
    version = "v3.0.1-0.20171022003610-9aa49832a739+incompatible",
)

go_repository(
    name = "com_github_mailru_easyjson",
    importpath = "github.com/mailru/easyjson",
    sum = "h1:2gxZ0XQIU/5z3Z3bUBu+FXuk2pFbkN6tcwi/pjyaDic=",
    version = "v0.0.0-20180823135443-60711f1a8329",
)

go_repository(
    name = "com_github_minio_highwayhash",
    importpath = "github.com/minio/highwayhash",
    sum = "h1:Aak5U0nElisjDCfPSG79Tgzkn2gl66NxOMspRrKnA/g=",
    version = "v1.0.2",
)

go_repository(
    name = "com_github_nats_io_jwt_v2",
    importpath = "github.com/nats-io/jwt/v2",
    sum = "h1:i/O6cmIsjpcQyWDYNcq2JyZ3/VTF8SJ4JWluI5OhpvI=",
    version = "v2.0.3",
)

go_repository(
    name = "com_github_nats_io_nats_go",
    importpath = "github.com/nats-io/nats.go",
    sum = "h1:+0ndxwUPz3CmQ2vjbXdkC1fo3FdiOQDim4gl3Mge8Qo=",
    version = "v1.12.1",
)

go_repository(
    name = "com_github_nats_io_nats_server_v2",
    importpath = "github.com/nats-io/nats-server/v2",
    sum = "h1:wsnVaaXH9VRSg+A2MVg5Q727/CqxnmPLGFQ3YZYKTQg=",
    version = "v2.5.0",
)

go_repository(
    name = "com_github_nats_io_nkeys",
    importpath = "github.com/nats-io/nkeys",
    sum = "h1:cgM5tL53EvYRU+2YLXIK0G2mJtK12Ft9oeooSZMA2G8=",
    version = "v0.3.0",
)

go_repository(
    name = "com_github_nats_io_nuid",
    importpath = "github.com/nats-io/nuid",
    sum = "h1:5iA8DT8V7q8WK2EScv2padNa/rTESc1KdnPw4TC2paw=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_op_go_logging",
    importpath = "github.com/op/go-logging",
    sum = "h1:lDH9UUVJtmYCjyT0CI4q8xvlXPxeZ0gYCVvWbmPlp88=",
    version = "v0.0.0-20160315200505-970db520ece7",
)

go_repository(
    name = "com_github_opentracing_opentracing_go",
    importpath = "github.com/opentracing/opentracing-go",
    sum = "h1:uEJPy/1a5RIPAJ0Ov+OIO8OxWu77jEv+1B0VhjKrZUs=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_openzipkin_zipkin_go",
    importpath = "github.com/openzipkin/zipkin-go",
    sum = "h1:UwtQQx2pyPIgWYHRg+epgdx1/HnBQTgN3/oIYEJTQzU=",
    version = "v0.2.5",
)

go_repository(
    name = "com_github_p768lwy3_gin_server_timing",
    importpath = "github.com/p768lwy3/gin-server-timing",
    sum = "h1:7dGPkdjqVYklnFVWsWx/10A8n0WJfftkGSe7xMz4CFU=",
    version = "v0.0.0-20200316080543-ab69795cf847",
)

go_repository(
    name = "com_github_performancecopilot_speed_v4",
    importpath = "github.com/performancecopilot/speed/v4",
    sum = "h1:VxEDCmdkfbQYDlcr/GC9YoN9PQ6p8ulk9xVsepYy9ZY=",
    version = "v4.0.0",
)

go_repository(
    name = "com_github_puerkitobio_purell",
    importpath = "github.com/PuerkitoBio/purell",
    sum = "h1:rmGxhojJlM0tuKtfdvliR84CFHljx9ag64t2xmVkjK4=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_puerkitobio_urlesc",
    importpath = "github.com/PuerkitoBio/urlesc",
    sum = "h1:d+Bc7a5rLufV/sSk/8dngufqelfh6jnri85riMAaF/M=",
    version = "v0.0.0-20170810143723-de5bf2ad4578",
)

go_repository(
    name = "com_github_sony_gobreaker",
    importpath = "github.com/sony/gobreaker",
    sum = "h1:oMnRNZXX5j85zso6xCPRNPtmAycat+WcoKbklScLDgQ=",
    version = "v0.4.1",
)

go_repository(
    name = "com_github_streadway_handy",
    importpath = "github.com/streadway/handy",
    sum = "h1:mOtuXaRAbVZsxAHVdPR3IjfmN8T1h2iczJLynhLybf8=",
    version = "v0.0.0-20200128134331-0f66f006fb2e",
)

go_repository(
    name = "com_github_swaggo_files",
    importpath = "github.com/swaggo/files",
    sum = "h1:PyYN9JH5jY9j6av01SpfRMb+1DWg/i3MbGOKPxJ2wjM=",
    version = "v0.0.0-20190704085106-630677cd5c14",
)

go_repository(
    name = "com_github_swaggo_gin_swagger",
    importpath = "github.com/swaggo/gin-swagger",
    sum = "h1:YskZXEiv51fjOMTsXrOetAjrMDfFaXD79PEoQBOe2W0=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_swaggo_swag",
    importpath = "github.com/swaggo/swag",
    sum = "h1:2Agm8I4K5qb00620mHq0VJ05/KT4FtmALPIcQR9lEZM=",
    version = "v1.5.1",
)

go_repository(
    name = "com_github_urfave_cli",
    importpath = "github.com/urfave/cli",
    sum = "h1:fDqGv3UG/4jbVl/QkFwEdddtEDjh/5Ov6X+0B/3bPaw=",
    version = "v1.20.0",
)

go_repository(
    name = "com_github_vividcortex_gohistogram",
    importpath = "github.com/VividCortex/gohistogram",
    sum = "h1:6+hBz+qvs0JOrrNhhmR7lFxo5sINxBCGXrdtl/UvroE=",
    version = "v1.0.0",
)

go_repository(
    name = "com_google_cloud_go_aiplatform",
    importpath = "cloud.google.com/go/aiplatform",
    sum = "h1:QqHZT1IMldf/daXoSnkJWBIqGBsw50X+xP6HSVzLRPo=",
    version = "v1.24.0",
)

go_repository(
    name = "com_google_cloud_go_analytics",
    importpath = "cloud.google.com/go/analytics",
    sum = "h1:NKw6PpQi6V1O+KsjuTd+bhip9d0REYu4NevC45vtGp8=",
    version = "v0.12.0",
)

go_repository(
    name = "com_google_cloud_go_area120",
    importpath = "cloud.google.com/go/area120",
    sum = "h1:TCMhwWEWhCn8d44/Zs7UCICTWje9j3HuV6nVGMjdpYw=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_artifactregistry",
    importpath = "cloud.google.com/go/artifactregistry",
    sum = "h1:9yKYCozdh29v7QMx3QBuksZGtPNICFb5SVnyNvkKRGg=",
    version = "v1.7.0",
)

go_repository(
    name = "com_google_cloud_go_asset",
    importpath = "cloud.google.com/go/asset",
    sum = "h1:qzYOcI6u4CD+0R1E8rWbrqs04fISCcg2YYxW8yBAqFM=",
    version = "v1.8.0",
)

go_repository(
    name = "com_google_cloud_go_assuredworkloads",
    importpath = "cloud.google.com/go/assuredworkloads",
    sum = "h1:IYhjgcgwb5TIAhC0aWQGGOqBnP0c2xijgMGf1iJRs50=",
    version = "v1.7.0",
)

go_repository(
    name = "com_google_cloud_go_automl",
    importpath = "cloud.google.com/go/automl",
    sum = "h1:U+kHmeKGXgBvTlrecPJhwkItWaIpIscG5DUpQxBQZZg=",
    version = "v1.6.0",
)

go_repository(
    name = "com_google_cloud_go_billing",
    importpath = "cloud.google.com/go/billing",
    sum = "h1:4RESn+mA7eGPBr5eQ4B/hbkHNivzYHbgRWpdlNeNjiE=",
    version = "v1.5.0",
)

go_repository(
    name = "com_google_cloud_go_binaryauthorization",
    importpath = "cloud.google.com/go/binaryauthorization",
    sum = "h1:5F7dowxGuYQlX3LjfjH/sKf+IvI1TsItTw0sDZmoec4=",
    version = "v1.2.0",
)

go_repository(
    name = "com_google_cloud_go_cloudtasks",
    importpath = "cloud.google.com/go/cloudtasks",
    sum = "h1:IL5W4fh6dAq9x1mO+4evrWCISOmPJegdaO0hZRZmWNE=",
    version = "v1.6.0",
)

go_repository(
    name = "com_google_cloud_go_compute_metadata",
    importpath = "cloud.google.com/go/compute/metadata",
    sum = "h1:efOwf5ymceDhK6PKMnnrTHP4pppY5L22mle96M1yP48=",
    version = "v0.2.1",
)

go_repository(
    name = "com_google_cloud_go_containeranalysis",
    importpath = "cloud.google.com/go/containeranalysis",
    sum = "h1:2824iym832ljKdVpCBnpqm5K94YT/uHTVhNF+dRTXPI=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_datacatalog",
    importpath = "cloud.google.com/go/datacatalog",
    sum = "h1:xzXGAE2fAuMh+ksODKr9nRv9ega1vHjFwRqMA8tRrVE=",
    version = "v1.6.0",
)

go_repository(
    name = "com_google_cloud_go_dataflow",
    importpath = "cloud.google.com/go/dataflow",
    sum = "h1:CW3541Fm7KPTyZjJdnX6NtaGXYFn5XbFC5UcjgALKvU=",
    version = "v0.7.0",
)

go_repository(
    name = "com_google_cloud_go_dataform",
    importpath = "cloud.google.com/go/dataform",
    sum = "h1:fnwkyzCVcPI/TmBheGgpmK2h+hWUIDHcZBincHRhrQ0=",
    version = "v0.4.0",
)

go_repository(
    name = "com_google_cloud_go_datalabeling",
    importpath = "cloud.google.com/go/datalabeling",
    sum = "h1:dp8jOF21n/7jwgo/uuA0RN8hvLcKO4q6s/yvwevs2ZM=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_dataqna",
    importpath = "cloud.google.com/go/dataqna",
    sum = "h1:gx9jr41ytcA3dXkbbd409euEaWtofCVXYBvJz3iYm18=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_datastream",
    importpath = "cloud.google.com/go/datastream",
    sum = "h1:ula4YR2K66o5wifLdPQMtR2I6KP+zvqdSEb6ncd1e0g=",
    version = "v1.3.0",
)

go_repository(
    name = "com_google_cloud_go_dialogflow",
    importpath = "cloud.google.com/go/dialogflow",
    sum = "h1:NU0Pj57H++JQOW225/7o34sUZ4i9/TLfWFOSbI3N1cY=",
    version = "v1.17.0",
)

go_repository(
    name = "com_google_cloud_go_documentai",
    importpath = "cloud.google.com/go/documentai",
    sum = "h1:CipwaecNhtsWUSneV2J5y8OqudHqvqPlcMHgSyh8vak=",
    version = "v1.8.0",
)

go_repository(
    name = "com_google_cloud_go_domains",
    importpath = "cloud.google.com/go/domains",
    sum = "h1:pu3JIgC1rswIqi5romW0JgNO6CTUydLYX8zyjiAvO1c=",
    version = "v0.7.0",
)

go_repository(
    name = "com_google_cloud_go_edgecontainer",
    importpath = "cloud.google.com/go/edgecontainer",
    sum = "h1:hd6J2n5dBBRuAqnNUEsKWrp6XNPKsaxwwIyzOPZTokk=",
    version = "v0.2.0",
)

go_repository(
    name = "com_google_cloud_go_functions",
    importpath = "cloud.google.com/go/functions",
    sum = "h1:s3Snbr2O4j4p7CuwImBas8rNNmkHS1YJANcCpKGqQSE=",
    version = "v1.7.0",
)

go_repository(
    name = "com_google_cloud_go_gaming",
    importpath = "cloud.google.com/go/gaming",
    sum = "h1:PKggmegChZulPW8yvtziF8P9UOuVFwbvylbEucTNups=",
    version = "v1.6.0",
)

go_repository(
    name = "com_google_cloud_go_gkeconnect",
    importpath = "cloud.google.com/go/gkeconnect",
    sum = "h1:zAcvDa04tTnGdu6TEZewaLN2tdMtUOJJ7fEceULjguA=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_gkehub",
    importpath = "cloud.google.com/go/gkehub",
    sum = "h1:JTcTaYQRGsVm+qkah7WzHb6e9sf1C0laYdRPn9aN+vg=",
    version = "v0.10.0",
)

go_repository(
    name = "com_google_cloud_go_language",
    importpath = "cloud.google.com/go/language",
    sum = "h1:Fb2iua/5/UBvUuW9PgBinwsCRDi1qoQJEuekOinHFCs=",
    version = "v1.6.0",
)

go_repository(
    name = "com_google_cloud_go_lifesciences",
    importpath = "cloud.google.com/go/lifesciences",
    sum = "h1:tIqhivE2LMVYkX0BLgG7xL64oNpDaFFI7teunglt1tI=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_mediatranslation",
    importpath = "cloud.google.com/go/mediatranslation",
    sum = "h1:qAJzpxmEX+SeND10Y/4868L5wfZpo4Y3BIEnIieP4dk=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_memcache",
    importpath = "cloud.google.com/go/memcache",
    sum = "h1:qTBOiSnVw7rnW6GVeH5Br8qs80ILoflNgFZySvaT4ek=",
    version = "v1.5.0",
)

go_repository(
    name = "com_google_cloud_go_metastore",
    importpath = "cloud.google.com/go/metastore",
    sum = "h1:wzJ9HslsybiJ3HL2168dVonr9D/eBq0VqObiMSCrE6c=",
    version = "v1.6.0",
)

go_repository(
    name = "com_google_cloud_go_networkconnectivity",
    importpath = "cloud.google.com/go/networkconnectivity",
    sum = "h1:mtIQewrz1ewMU3J0vVkUIJtAkpOqgkz4+UmcreeAm08=",
    version = "v1.5.0",
)

go_repository(
    name = "com_google_cloud_go_networksecurity",
    importpath = "cloud.google.com/go/networksecurity",
    sum = "h1:qDEX/3sipg9dS5JYsAY+YvgTjPR63cozzAWop8oZS94=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_notebooks",
    importpath = "cloud.google.com/go/notebooks",
    sum = "h1:YfPI4pOYQDcqJ+thM2cGtR9oRoRv42vRfubSPZnk3DI=",
    version = "v1.3.0",
)

go_repository(
    name = "com_google_cloud_go_osconfig",
    importpath = "cloud.google.com/go/osconfig",
    sum = "h1:fkFlXCxkUt3tE8LYtF6CipuPbC/HIrciwDTjFpsTf88=",
    version = "v1.8.0",
)

go_repository(
    name = "com_google_cloud_go_oslogin",
    importpath = "cloud.google.com/go/oslogin",
    sum = "h1:/7sVaMdtqSm6AjxW8KzoM6UKawkg3REr0XJ1zKtidpc=",
    version = "v1.5.0",
)

go_repository(
    name = "com_google_cloud_go_phishingprotection",
    importpath = "cloud.google.com/go/phishingprotection",
    sum = "h1:OrwHLSRSZyaiOt3tnY33dsKSedxbMzsXvqB21okItNQ=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_privatecatalog",
    importpath = "cloud.google.com/go/privatecatalog",
    sum = "h1:Vz86uiHCtNGm1DeC32HeG2VXmOq5JRYA3VRPf8ZEcSg=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_recaptchaenterprise_v2",
    importpath = "cloud.google.com/go/recaptchaenterprise/v2",
    sum = "h1:BkkI7C0o8CtaHvdDMr5IA+y8pk0Y5wb73C7DHQiAKnw=",
    version = "v2.3.0",
)

go_repository(
    name = "com_google_cloud_go_recommendationengine",
    importpath = "cloud.google.com/go/recommendationengine",
    sum = "h1:6w+WxPf2LmUEqX0YyvfCoYb8aBYOcbIV25Vg6R0FLGw=",
    version = "v0.6.0",
)

go_repository(
    name = "com_google_cloud_go_recommender",
    importpath = "cloud.google.com/go/recommender",
    sum = "h1:C1tw+Qa/bgm6LoH1wuxYdoyinwKkW/jDJ0GpSJf58cE=",
    version = "v1.6.0",
)

go_repository(
    name = "com_google_cloud_go_redis",
    importpath = "cloud.google.com/go/redis",
    sum = "h1:gtPd4pG/Go5mrdGQ4MJXxPHtjxtoWUBkrWLXNV1L2TA=",
    version = "v1.8.0",
)

go_repository(
    name = "com_google_cloud_go_retail",
    importpath = "cloud.google.com/go/retail",
    sum = "h1:Q3W/JsQupZWaoFxUOugZd1Eq590R+Dk6dhacLK2h7+w=",
    version = "v1.9.0",
)

go_repository(
    name = "com_google_cloud_go_scheduler",
    importpath = "cloud.google.com/go/scheduler",
    sum = "h1:Fe1Upic/q4cwqXbInCzgAW35QSerj8JlNwATIxDdfOI=",
    version = "v1.5.0",
)

go_repository(
    name = "com_google_cloud_go_security",
    importpath = "cloud.google.com/go/security",
    sum = "h1:linnRc3/gJYDfKbAtNixVQ52+66DpOx5MmCz0NNxal8=",
    version = "v1.8.0",
)

go_repository(
    name = "com_google_cloud_go_securitycenter",
    importpath = "cloud.google.com/go/securitycenter",
    sum = "h1:hKIggnv2eCAPjsVnFcZbytMOsFOk6p4ut0iAUDoNsNA=",
    version = "v1.14.0",
)

go_repository(
    name = "com_google_cloud_go_servicedirectory",
    importpath = "cloud.google.com/go/servicedirectory",
    sum = "h1:QmCWml/qvNOYyiPP4G52srYcsHSLCXuvydJDVLTFSe8=",
    version = "v1.5.0",
)

go_repository(
    name = "com_google_cloud_go_speech",
    importpath = "cloud.google.com/go/speech",
    sum = "h1:bRI2QczZGpcPfuhHr63VOdfyyfYp/43N0wRuBKrd0nQ=",
    version = "v1.7.0",
)

go_repository(
    name = "com_google_cloud_go_talent",
    importpath = "cloud.google.com/go/talent",
    sum = "h1:6c4pvu3k2idEhJRZnZ2HdVLWZUuT9fsns2gQtCzRtqA=",
    version = "v1.2.0",
)

go_repository(
    name = "com_google_cloud_go_videointelligence",
    importpath = "cloud.google.com/go/videointelligence",
    sum = "h1:w56i2xl1jHX2tz6rHXBPHd6xujevhImzbc16Kl+V/zQ=",
    version = "v1.7.0",
)

go_repository(
    name = "com_google_cloud_go_vision_v2",
    importpath = "cloud.google.com/go/vision/v2",
    sum = "h1:eEyIDJ5/98UmQrYZ6eQExUT3iHyDjzzPX29UP6x7ZQo=",
    version = "v2.3.0",
)

go_repository(
    name = "com_google_cloud_go_webrisk",
    importpath = "cloud.google.com/go/webrisk",
    sum = "h1:WdHJmLSAs5bIis/WWO7pIfiRBD1PiWe1OAlPrWeM9Tk=",
    version = "v1.5.0",
)

go_repository(
    name = "com_google_cloud_go_workflows",
    importpath = "cloud.google.com/go/workflows",
    sum = "h1:0MjX5ugKmTdbRG2Vai5aAgNAOe2wzvs/XQwFDSowy9c=",
    version = "v1.7.0",
)

go_repository(
    name = "in_gopkg_gcfg_v1",
    importpath = "gopkg.in/gcfg.v1",
    sum = "h1:m8OOJ4ccYHnx2f4gQwpno8nAX5OGOh7RLaaz0pj3Ogs=",
    version = "v1.2.3",
)

go_repository(
    name = "in_gopkg_go_playground_assert_v1",
    importpath = "gopkg.in/go-playground/assert.v1",
    sum = "h1:xoYuJVE7KT85PYWrN730RguIQO0ePzVRfFMXadIrXTM=",
    version = "v1.2.1",
)

go_repository(
    name = "in_gopkg_go_playground_validator_v8",
    importpath = "gopkg.in/go-playground/validator.v8",
    sum = "h1:lFB4DoMU6B626w8ny76MV7VX6W2VHct2GVOI3xgiMrQ=",
    version = "v8.18.2",
)

go_repository(
    name = "in_gopkg_go_playground_validator_v9",
    importpath = "gopkg.in/go-playground/validator.v9",
    sum = "h1:SvGtYmN60a5CVKTOzMSyfzWDeZRxRuGvRQyEAKbw1xc=",
    version = "v9.29.1",
)

go_repository(
    name = "in_gopkg_warnings_v0",
    importpath = "gopkg.in/warnings.v0",
    sum = "h1:wFXVbFY8DY5/xOe1ECiWdKCzZlxgshcYVNkBHstARME=",
    version = "v0.1.2",
)

gazelle_protobuf_extension_go_deps()

protobuf_core_deps()

#
#
#_py_configure = """
#if [[ "$OSTYPE" == "darwin"* ]]; then
#    ./configure --prefix=$(pwd)/bazel_install --with-openssl=$(brew --prefix openssl)
#else
#    ./configure --prefix=$(pwd)/bazel_install
#fi
#"""
#
#http_archive(
#    name = "python_interpreter",
#    urls = ["https://www.python.org/ftp/python/3.8.3/Python-3.8.3.tar.xz"],
#    sha256 = "dfab5ec723c218082fe3d5d7ae17ecbdebffa9a1aea4d64aa3a2ecdd2e795864",
#    strip_prefix = "Python-3.8.3",
#    patch_cmds = [
#        "mkdir $(pwd)/bazel_install",
#        _py_configure,
#        "make",
#        "make install",
#        "ln -s bazel_install/bin/python3 python_bin",
#    ],
#    build_file_content = """
#exports_files(["python_bin"])
#filegroup(
#    name = "files",
#    srcs = glob(["bazel_install/**"], exclude = ["**/* *"]),
#    visibility = ["//visibility:public"],
#)
#""",
#)
#
#
load("@rules_python//python:repositories.bzl", "python_register_toolchains")

python_register_toolchains(
    name = "python38",
    # Available versions are listed in @rules_python//python:versions.bzl.
    python_version = "3.8.10",
)

load("@python38//:defs.bzl", "interpreter")
load("@rules_proto_grpc//python:repositories.bzl", rules_proto_grpc_python_repos = "python_repos")

_py_gazelle_deps()

load("@rules_python//python:pip.bzl", "pip_install")

# Create a central external repo, @my_deps, that contains Bazel targets for all the
# third-party packages specified in the requirements.txt file.
pip_install(
    name = "pip",
    #python_interpreter_target = "@python_interpreter//:python_bin",
    python_interpreter_target = interpreter,
    requirements = "//slackbot:requirements.txt",
)

register_execution_platforms(
    "//toolchains:local_config_platform_but_dont_run_in_container",
    "@io_bazel_rules_docker//platforms:local_container_platform",
)

register_toolchains(
    "//toolchains:my_py_toolchain",
    "//toolchains:container_py_toolchain",
    "@build_stack_rules_proto//toolchain:standard",
)

rules_proto_grpc_toolchains()

rules_proto_grpc_python_repos()

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")

rules_proto_dependencies()

rules_proto_toolchains()

rules_proto_grpc_python_repos()

load("@com_github_grpc_grpc//bazel:grpc_deps.bzl", "grpc_deps")

grpc_deps()

http_archive(
    name = "com_google_protobuf",
    sha256 = "90de7e780db97e0ee8cfabc3aecc0da56c3d443824b968ec0c7c600f9585b9ba",
    strip_prefix = "protobuf-3.21.10",
    urls = [
        "https://mirror.bazel.build/github.com/protocolbuffers/protobuf/archive/v3.21.10.tar.gz",
        "https://github.com/protocolbuffers/protobuf/archive/v3.21.10.tar.gz",
    ],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

go_repository(
    name = "co_honnef_go_tools",
    importpath = "honnef.co/go/tools",
    sum = "h1:UoveltGrhghAA7ePc+e+QYDHXrBps2PqFZiHkGR/xK8=",
    version = "v0.0.1-2020.1.4",
)

go_repository(
    name = "com_github_antihax_optional",
    importpath = "github.com/antihax/optional",
    sum = "h1:xK2lYat7ZLaVVcIuj82J8kIro4V6kDe0AUDFboUCwcg=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_armon_circbuf",
    importpath = "github.com/armon/circbuf",
    sum = "h1:QEF07wC0T1rKkctt1RINW/+RMTVmiwxETico2l3gxJA=",
    version = "v0.0.0-20150827004946-bbbad097214e",
)

go_repository(
    name = "com_github_armon_go_metrics",
    importpath = "github.com/armon/go-metrics",
    sum = "h1:O2sNqxBdvq8Eq5xmzljcYzAORli6RWCvEym4cJf9m18=",
    version = "v0.3.9",
)

go_repository(
    name = "com_github_armon_go_radix",
    importpath = "github.com/armon/go-radix",
    sum = "h1:F4z6KzEeeQIMeLFa97iZU6vupzoecKdU5TX24SNppXI=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_bgentry_speakeasy",
    importpath = "github.com/bgentry/speakeasy",
    sum = "h1:ByYyxL9InA1OWqxJqqp2A5pYHUrCiAL6K3J+LKSsQkY=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_bketelsen_crypt",
    importpath = "github.com/bketelsen/crypt",
    sum = "h1:w/jqZtC9YD4DS/Vp9GhWfWcCpuAL58oTnLoI8vE9YHU=",
    version = "v0.0.4",
)

go_repository(
    name = "com_github_burntsushi_toml",
    importpath = "github.com/BurntSushi/toml",
    sum = "h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=",
    version = "v0.3.1",
)

go_repository(
    name = "com_github_burntsushi_xgb",
    importpath = "github.com/BurntSushi/xgb",
    sum = "h1:1BDTz0u9nC3//pOCMdNH+CiXJVYJh5UQNCOBG7jbELc=",
    version = "v0.0.0-20160522181843-27f122750802",
)

go_repository(
    name = "com_github_census_instrumentation_opencensus_proto",
    importpath = "github.com/census-instrumentation/opencensus-proto",
    sum = "h1:glEXhBS5PSLLv4IXzLA5yPRVX4bilULVyxxbrfOtDAk=",
    version = "v0.2.1",
)

go_repository(
    name = "com_github_cespare_xxhash",
    importpath = "github.com/cespare/xxhash",
    sum = "h1:a6HrQnmkObjyL+Gs60czilIUGqrzKutQD6XZog3p+ko=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_chzyer_logex",
    importpath = "github.com/chzyer/logex",
    sum = "h1:Swpa1K6QvQznwJRcfTfQJmTE72DqScAa40E+fbHEXEE=",
    version = "v1.1.10",
)

go_repository(
    name = "com_github_chzyer_readline",
    importpath = "github.com/chzyer/readline",
    sum = "h1:fY5BOSpyZCqRo5OhCuC+XN+r/bBCmeuuJtjz+bCNIf8=",
    version = "v0.0.0-20180603132655-2972be24d48e",
)

go_repository(
    name = "com_github_chzyer_test",
    importpath = "github.com/chzyer/test",
    sum = "h1:q763qf9huN11kDQavWsoZXJNW3xEE4JJyHa5Q25/sd8=",
    version = "v0.0.0-20180213035817-a1ea475d72b1",
)

go_repository(
    name = "com_github_client9_misspell",
    importpath = "github.com/client9/misspell",
    sum = "h1:ta993UF76GwbvJcIo3Y68y/M3WxlpEHPWIGDkJYwzJI=",
    version = "v0.3.4",
)

go_repository(
    name = "com_github_cncf_udpa_go",
    importpath = "github.com/cncf/udpa/go",
    sum = "h1:cqQfy1jclcSy/FwLjemeg3SR1yaINm74aQyupQ0Bl8M=",
    version = "v0.0.0-20201120205902-5459f2c99403",
)

go_repository(
    name = "com_github_cncf_xds_go",
    importpath = "github.com/cncf/xds/go",
    sum = "h1:zH8ljVhhq7yC0MIeUL/IviMtY8hx2mK8cN9wEYb8ggw=",
    version = "v0.0.0-20211011173535-cb28da3451f1",
)

go_repository(
    name = "com_github_coreos_go_semver",
    importpath = "github.com/coreos/go-semver",
    sum = "h1:wkHLiw0WNATZnSG7epLsujiMCgPAc9xhjJ4tgnAxmfM=",
    version = "v0.3.0",
)

go_repository(
    name = "com_github_coreos_go_systemd_v22",
    importpath = "github.com/coreos/go-systemd/v22",
    sum = "h1:rtAn27wIbmOGUs7RIbVgPEjb31ehTVniDwPGXyMxm5U=",
    version = "v22.3.3-0.20220203105225-a9a7ef127534",
)

go_repository(
    name = "com_github_creack_pty",
    importpath = "github.com/creack/pty",
    sum = "h1:uDmaGzcdjhF4i/plgjmEsriH11Y0o7RKapEf/LDaM3w=",
    version = "v1.1.9",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    importpath = "github.com/davecgh/go-spew",
    sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_envoyproxy_go_control_plane",
    importpath = "github.com/envoyproxy/go-control-plane",
    sum = "h1:EmNYJhPYy0pOFjCx2PrgtaBXmee0iUX9hLlxE1xHOJE=",
    version = "v0.9.9-0.20201210154907-fd9021fe5dad",
)

go_repository(
    name = "com_github_envoyproxy_protoc_gen_validate",
    importpath = "github.com/envoyproxy/protoc-gen-validate",
    sum = "h1:EQciDnbrYxy13PgWoY8AqoxGiPrpgBZ1R8UNe3ddc+A=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_fatih_color",
    importpath = "github.com/fatih/color",
    sum = "h1:mRhaKNwANqRgUBGKmnI5ZxEk7QXmjQeCcuYFMX2bfcc=",
    version = "v1.12.0",
)

go_repository(
    name = "com_github_fsnotify_fsnotify",
    importpath = "github.com/fsnotify/fsnotify",
    sum = "h1:n+5WquG0fcWoWp6xPWfHdbskMCQaFnG6PfBrh1Ky4HY=",
    version = "v1.6.0",
)

go_repository(
    name = "com_github_ghodss_yaml",
    importpath = "github.com/ghodss/yaml",
    sum = "h1:wQHKEahhL6wmXdzwWG11gIVCkOv05bNOh+Rxn0yngAk=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_gin_contrib_sse",
    importpath = "github.com/gin-contrib/sse",
    sum = "h1:Y/yl/+YNO8GZSjAhjMsSuLt29uWRFHdHYUb5lYOV9qE=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_gin_gonic_gin",
    importpath = "github.com/gin-gonic/gin",
    sum = "h1:OjyFBKICoexlu99ctXNR2gg+c5pKrKMuyjgARg9qeY8=",
    version = "v1.9.0",
)

go_repository(
    name = "com_github_go_gl_glfw",
    importpath = "github.com/go-gl/glfw",
    sum = "h1:QbL/5oDUmRBzO9/Z7Seo6zf912W/a6Sr4Eu0G/3Jho0=",
    version = "v0.0.0-20190409004039-e6da0acd62b1",
)

go_repository(
    name = "com_github_go_gl_glfw_v3_3_glfw",
    importpath = "github.com/go-gl/glfw/v3.3/glfw",
    sum = "h1:WtGNWLvXpe6ZudgnXrq0barxBImvnnJoMEhXAzcbM0I=",
    version = "v0.0.0-20200222043503-6f7a984d4dc4",
)

go_repository(
    name = "com_github_go_playground_assert_v2",
    importpath = "github.com/go-playground/assert/v2",
    sum = "h1:MsBgLAaY856+nPRTKrp3/OZK38U/wa0CcBYNjji3q3A=",
    version = "v2.0.1",
)

go_repository(
    name = "com_github_go_playground_locales",
    importpath = "github.com/go-playground/locales",
    sum = "h1:u50s323jtVGugKlcYeyzC0etD1HifMjqmJqb8WugfUU=",
    version = "v0.14.0",
)

go_repository(
    name = "com_github_go_playground_universal_translator",
    importpath = "github.com/go-playground/universal-translator",
    sum = "h1:82dyy6p4OuJq4/CByFNOn/jYrnRPArHwAcmLoJZxyho=",
    version = "v0.18.0",
)

go_repository(
    name = "com_github_go_playground_validator_v10",
    importpath = "github.com/go-playground/validator/v10",
    sum = "h1:I7mrTYv78z8k8VXa/qJlOlEXn/nBh+BF8dHX5nt/dr0=",
    version = "v10.10.0",
)

go_repository(
    name = "com_github_go_sql_driver_mysql",
    importpath = "github.com/go-sql-driver/mysql",
    sum = "h1:BCTh4TKNUYmOmMUcQ3IipzF5prigylS7XXjEkfCHuOE=",
    version = "v1.6.0",
)

go_repository(
    name = "com_github_godbus_dbus_v5",
    importpath = "github.com/godbus/dbus/v5",
    sum = "h1:9349emZab16e7zQvpmsbtjc18ykshndd8y2PG3sgJbA=",
    version = "v5.0.4",
)

go_repository(
    name = "com_github_gogo_protobuf",
    importpath = "github.com/gogo/protobuf",
    sum = "h1:Ov1cvc58UF3b5XjBnZv7+opcTcQFZebYjWzi34vdm4Q=",
    version = "v1.3.2",
)

go_repository(
    name = "com_github_golang_glog",
    importpath = "github.com/golang/glog",
    sum = "h1:VKtxabqXZkF25pY9ekfRL6a582T4P37/31XEstQ5p58=",
    version = "v0.0.0-20160126235308-23def4e6c14b",
)

go_repository(
    name = "com_github_golang_groupcache",
    importpath = "github.com/golang/groupcache",
    sum = "h1:oI5xCqsCo564l8iNU+DwB5epxmsaqB+rhGL0m5jtYqE=",
    version = "v0.0.0-20210331224755-41bb18bfe9da",
)

go_repository(
    name = "com_github_golang_mock",
    importpath = "github.com/golang/mock",
    sum = "h1:l75CXGRSwbaYNpl/Z2X1XIIAMSCquvXgpVZDhwEIJsc=",
    version = "v1.4.4",
)

go_repository(
    name = "com_github_golang_protobuf",
    importpath = "github.com/golang/protobuf",
    sum = "h1:ROPKBNFfQgOUMifHyP+KYbvpjbdoFNs+aK7DXlji0Tw=",
    version = "v1.5.2",
)

go_repository(
    name = "com_github_google_btree",
    importpath = "github.com/google/btree",
    sum = "h1:0udJVsspx3VBr5FwtLhQQtuAsVc79tTq0ocGIPAU6qo=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_google_go_cmp",
    importpath = "github.com/google/go-cmp",
    sum = "h1:81/ik6ipDQS2aGcBfIN5dHDB36BwrStyeAQquSYCV4o=",
    version = "v0.5.7",
)

go_repository(
    name = "com_github_google_gofuzz",
    importpath = "github.com/google/gofuzz",
    sum = "h1:A8PeW59pxE9IoFRqBp37U+mSNaQoZ46F1f0f863XSXw=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_google_martian",
    importpath = "github.com/google/martian",
    sum = "h1:/CP5g8u/VJHijgedC/Legn3BAbAaWPgecwXBIDzw5no=",
    version = "v2.1.0+incompatible",
)

go_repository(
    name = "com_github_google_martian_v3",
    importpath = "github.com/google/martian/v3",
    sum = "h1:wCKgOCHuUEVfsaQLpPSJb7VdYCdTVZQAuOdYm1yc/60=",
    version = "v3.1.0",
)

go_repository(
    name = "com_github_google_pprof",
    importpath = "github.com/google/pprof",
    sum = "h1:LR89qFljJ48s990kEKGsk213yIJDPI4205OKOzbURK8=",
    version = "v0.0.0-20201218002935-b9804c9f04c2",
)

go_repository(
    name = "com_github_google_renameio",
    importpath = "github.com/google/renameio",
    sum = "h1:GOZbcHa3HfsPKPlmyPyN2KEohoMXOhdMbHrvbpl2QaA=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_google_uuid",
    importpath = "github.com/google/uuid",
    sum = "h1:EVhdT+1Kseyi1/pUmXKaFxYsDNy9RQYkMWRH68J/W7Y=",
    version = "v1.1.2",
)

go_repository(
    name = "com_github_googleapis_gax_go_v2",
    importpath = "github.com/googleapis/gax-go/v2",
    sum = "h1:sjZBwGj9Jlw33ImPtvFviGYvseOtDM7hkSKB7+Tv3SM=",
    version = "v2.0.5",
)

go_repository(
    name = "com_github_gopherjs_gopherjs",
    importpath = "github.com/gopherjs/gopherjs",
    sum = "h1:EGx4pi6eqNxGaHF6qqu48+N2wcFQ5qg5FXgOdqsJ5d8=",
    version = "v0.0.0-20181017120253-0766667cb4d1",
)

go_repository(
    name = "com_github_grpc_ecosystem_grpc_gateway",
    importpath = "github.com/grpc-ecosystem/grpc-gateway",
    sum = "h1:gmcG1KaJ57LophUzW0Hy8NmPhnMZb4M0+kPpLofRdBo=",
    version = "v1.16.0",
)

go_repository(
    name = "com_github_hashicorp_consul_api",
    importpath = "github.com/hashicorp/consul/api",
    sum = "h1:MwZJp86nlnL+6+W1Zly4JUuVn9YHhMggBirMpHGD7kw=",
    version = "v1.10.1",
)

go_repository(
    name = "com_github_hashicorp_consul_sdk",
    importpath = "github.com/hashicorp/consul/sdk",
    sum = "h1:OJtKBtEjboEZvG6AOUdh4Z1Zbyu0WcxQ0qatRrZHTVU=",
    version = "v0.8.0",
)

go_repository(
    name = "com_github_hashicorp_errwrap",
    importpath = "github.com/hashicorp/errwrap",
    sum = "h1:hLrqtEDnRye3+sgx6z4qVLNuviH3MR5aQ0ykNJa/UYA=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_hashicorp_go_cleanhttp",
    importpath = "github.com/hashicorp/go-cleanhttp",
    sum = "h1:035FKYIWjmULyFRBKPs8TBQoi0x6d9G4xc9neXJWAZQ=",
    version = "v0.5.2",
)

go_repository(
    name = "com_github_hashicorp_go_immutable_radix",
    importpath = "github.com/hashicorp/go-immutable-radix",
    sum = "h1:DKHmCUm2hRBK510BaiZlwvpD40f8bJFeZnpfm2KLowc=",
    version = "v1.3.1",
)

go_repository(
    name = "com_github_hashicorp_go_msgpack",
    importpath = "github.com/hashicorp/go-msgpack",
    sum = "h1:zKjpN5BK/P5lMYrLmBHdBULWbJ0XpYR+7NGzqkZzoD4=",
    version = "v0.5.3",
)

go_repository(
    name = "com_github_hashicorp_go_multierror",
    importpath = "github.com/hashicorp/go-multierror",
    sum = "h1:B9UzwGQJehnUY1yNrnwREHc3fGbC2xefo8g4TbElacI=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_hashicorp_go_net",
    importpath = "github.com/hashicorp/go.net",
    sum = "h1:sNCoNyDEvN1xa+X0baata4RdcpKwcMS6DH+xwfqPgjw=",
    version = "v0.0.1",
)

go_repository(
    name = "com_github_hashicorp_go_rootcerts",
    importpath = "github.com/hashicorp/go-rootcerts",
    sum = "h1:jzhAVGtqPKbwpyCPELlgNWhE1znq+qwJtW5Oi2viEzc=",
    version = "v1.0.2",
)

go_repository(
    name = "com_github_hashicorp_go_sockaddr",
    importpath = "github.com/hashicorp/go-sockaddr",
    sum = "h1:GeH6tui99pF4NJgfnhp+L6+FfobzVW3Ah46sLo0ICXs=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_hashicorp_go_syslog",
    importpath = "github.com/hashicorp/go-syslog",
    sum = "h1:KaodqZuhUoZereWVIYmpUgZysurB1kBLX2j0MwMrUAE=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_hashicorp_go_uuid",
    importpath = "github.com/hashicorp/go-uuid",
    sum = "h1:fv1ep09latC32wFoVwnqcnKJGnMSdBanPczbHAYm1BE=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    importpath = "github.com/hashicorp/golang-lru",
    sum = "h1:YDjusn29QI/Das2iO9M0BHnIbxPeyuCHsjMW+lJfyTc=",
    version = "v0.5.4",
)

go_repository(
    name = "com_github_hashicorp_hcl",
    importpath = "github.com/hashicorp/hcl",
    sum = "h1:0Anlzjpi4vEasTeNFn2mLJgTSwt0+6sfsiTG8qcWGx4=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_hashicorp_logutils",
    importpath = "github.com/hashicorp/logutils",
    sum = "h1:dLEQVugN8vlakKOUE3ihGLTZJRB4j+M2cdTm/ORI65Y=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_hashicorp_mdns",
    importpath = "github.com/hashicorp/mdns",
    sum = "h1:sY0CMhFmjIPDMlTB+HfymFHCaYLhgifZ0QhjaYKD/UQ=",
    version = "v1.0.4",
)

go_repository(
    name = "com_github_hashicorp_memberlist",
    importpath = "github.com/hashicorp/memberlist",
    sum = "h1:8+567mCcFDnS5ADl7lrpxPMWiFCElyUEeW0gtj34fMA=",
    version = "v0.3.0",
)

go_repository(
    name = "com_github_hashicorp_serf",
    importpath = "github.com/hashicorp/serf",
    sum = "h1:EBWvyu9tcRszt3Bxp3KNssBMP1KuHWyO51lz9+786iM=",
    version = "v0.9.5",
)

go_repository(
    name = "com_github_ianlancetaylor_demangle",
    importpath = "github.com/ianlancetaylor/demangle",
    sum = "h1:mV02weKRL81bEnm8A0HT1/CAelMQDBuQIfLw8n+d6xI=",
    version = "v0.0.0-20200824232613-28f6c0f3b639",
)

go_repository(
    name = "com_github_json_iterator_go",
    importpath = "github.com/json-iterator/go",
    sum = "h1:PV8peI4a0ysnczrg+LtxykD8LfKY9ML6u2jnxaEnrnM=",
    version = "v1.1.12",
)

go_repository(
    name = "com_github_jstemmer_go_junit_report",
    importpath = "github.com/jstemmer/go-junit-report",
    sum = "h1:6QPYqodiu3GuPL+7mfx+NwDdp2eTkp9IfEUpgAwUN0o=",
    version = "v0.9.1",
)

go_repository(
    name = "com_github_jtolds_gls",
    importpath = "github.com/jtolds/gls",
    sum = "h1:xdiiI2gbIgH/gLH7ADydsJ1uDOEzR8yvV7C0MuV77Wo=",
    version = "v4.20.0+incompatible",
)

go_repository(
    name = "com_github_kisielk_errcheck",
    importpath = "github.com/kisielk/errcheck",
    sum = "h1:e8esj/e4R+SAOwFwN+n3zr0nYeCyeweozKfO23MvHzY=",
    version = "v1.5.0",
)

go_repository(
    name = "com_github_kisielk_gotool",
    importpath = "github.com/kisielk/gotool",
    sum = "h1:AV2c/EiW3KqPNT9ZKl07ehoAGi4C5/01Cfbblndcapg=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_kr_fs",
    importpath = "github.com/kr/fs",
    sum = "h1:Jskdu9ieNAYnjxsi0LbQp1ulIKZV1LAFgK1tWhpZgl8=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_kr_pretty",
    importpath = "github.com/kr/pretty",
    sum = "h1:WgNl7dwNpEZ6jJ9k1snq4pZsg7DOEN8hP9Xw0Tsjwk0=",
    version = "v0.3.0",
)

go_repository(
    name = "com_github_kr_pty",
    importpath = "github.com/kr/pty",
    sum = "h1:VkoXIwSboBpnk99O/KFauAEILuNHv5DVFKZMBN/gUgw=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_kr_text",
    importpath = "github.com/kr/text",
    sum = "h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_leodido_go_urn",
    importpath = "github.com/leodido/go-urn",
    sum = "h1:BqpAaACuzVSgi/VLzGZIobT2z4v53pjosyNd9Yv6n/w=",
    version = "v1.2.1",
)

go_repository(
    name = "com_github_magiconair_properties",
    importpath = "github.com/magiconair/properties",
    sum = "h1:5ibWZ6iY0NctNGWo87LalDlEZ6R41TqbbDamhfG/Qzo=",
    version = "v1.8.6",
)

go_repository(
    name = "com_github_mattn_go_colorable",
    importpath = "github.com/mattn/go-colorable",
    sum = "h1:jF+Du6AlPIjs2BiUiQlKOX0rt3SujHxPnksPKZbaA40=",
    version = "v0.1.12",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    importpath = "github.com/mattn/go-isatty",
    sum = "h1:yVuAays6BHfxijgZPzw+3Zlu5yQgKGP2/hcQbHb7S9Y=",
    version = "v0.0.14",
)

go_repository(
    name = "com_github_miekg_dns",
    importpath = "github.com/miekg/dns",
    sum = "h1:JKfpVSCB84vrAmHzyrsxB5NAr5kLoMXZArPSw7Qlgyg=",
    version = "v1.1.43",
)

go_repository(
    name = "com_github_mitchellh_cli",
    importpath = "github.com/mitchellh/cli",
    sum = "h1:tEElEatulEHDeedTxwckzyYMA5c86fbmNIUL1hBIiTg=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_mitchellh_go_homedir",
    importpath = "github.com/mitchellh/go-homedir",
    sum = "h1:lukF9ziXFxDFPkA1vsr5zpc1XuPDn/wFntq5mG+4E0Y=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_mitchellh_go_testing_interface",
    importpath = "github.com/mitchellh/go-testing-interface",
    sum = "h1:fzU/JVNcaqHQEcVFAKeR41fkiLdIPrefOvVG1VZ96U0=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_mitchellh_gox",
    importpath = "github.com/mitchellh/gox",
    sum = "h1:lfGJxY7ToLJQjHHwi0EX6uYBdK78egf954SQl13PQJc=",
    version = "v0.4.0",
)

go_repository(
    name = "com_github_mitchellh_iochan",
    importpath = "github.com/mitchellh/iochan",
    sum = "h1:C+X3KsSTLFVBr/tK1eYN/vs4rJcvsiLU338UhYPJWeY=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_mitchellh_mapstructure",
    importpath = "github.com/mitchellh/mapstructure",
    sum = "h1:jeMsZIYE/09sWLaz43PL7Gy6RuMjD2eJVyuac5Z2hdY=",
    version = "v1.5.0",
)

go_repository(
    name = "com_github_modern_go_concurrent",
    importpath = "github.com/modern-go/concurrent",
    sum = "h1:TRLaZ9cD/w8PVh93nsPXa1VrQ6jlwL5oN8l14QlcNfg=",
    version = "v0.0.0-20180306012644-bacd9c7ef1dd",
)

go_repository(
    name = "com_github_modern_go_reflect2",
    importpath = "github.com/modern-go/reflect2",
    sum = "h1:xBagoLtFs94CBntxluKeaWgTMpvLxC4ur3nMaC9Gz0M=",
    version = "v1.0.2",
)

go_repository(
    name = "com_github_oneofone_xxhash",
    importpath = "github.com/OneOfOne/xxhash",
    sum = "h1:KMrpdQIwFcEqXDklaen+P1axHaj9BSKzvpUUfnHldSE=",
    version = "v1.2.2",
)

go_repository(
    name = "com_github_pascaldekloe_goe",
    importpath = "github.com/pascaldekloe/goe",
    sum = "h1:cBOtyMzM9HTpWjXfbbunk26uA6nG3a8n06Wieeh0MwY=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_pelletier_go_toml",
    importpath = "github.com/pelletier/go-toml",
    sum = "h1:4yBQzkHv+7BHq2PQUZF3Mx0IYxG7LsP222s7Agd3ve8=",
    version = "v1.9.5",
)

go_repository(
    name = "com_github_pkg_diff",
    importpath = "github.com/pkg/diff",
    sum = "h1:aoZm08cpOy4WuID//EZDgcC4zIxODThtZNPirFr42+A=",
    version = "v0.0.0-20210226163009-20ebb0f2a09e",
)

go_repository(
    name = "com_github_pkg_errors",
    importpath = "github.com/pkg/errors",
    sum = "h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=",
    version = "v0.9.1",
)

go_repository(
    name = "com_github_pkg_sftp",
    importpath = "github.com/pkg/sftp",
    sum = "h1:I2qBYMChEhIjOgazfJmV3/mZM256btk6wkCDRmW7JYs=",
    version = "v1.13.1",
)

go_repository(
    name = "com_github_pmezard_go_difflib",
    importpath = "github.com/pmezard/go-difflib",
    sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_posener_complete",
    importpath = "github.com/posener/complete",
    sum = "h1:NP0eAhjcjImqslEwo/1hq7gpajME0fTLTezBKDqfXqo=",
    version = "v1.2.3",
)

go_repository(
    name = "com_github_prometheus_client_model",
    importpath = "github.com/prometheus/client_model",
    sum = "h1:uq5h0d+GuxiXLJLNABMgp2qUWDPiLvgCzz2dUR+/W/M=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_rogpeppe_fastuuid",
    importpath = "github.com/rogpeppe/fastuuid",
    sum = "h1:Ppwyp6VYCF1nvBTXL3trRso7mXMlRrw9ooo375wvi2s=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_rogpeppe_go_internal",
    importpath = "github.com/rogpeppe/go-internal",
    sum = "h1:FCbCCtXNOY3UtUuHUYaghJg4y7Fd14rXifAYUAtL9R8=",
    version = "v1.8.0",
)

go_repository(
    name = "com_github_rs_xid",
    importpath = "github.com/rs/xid",
    sum = "h1:qd7wPTDkN6KQx2VmMBLrpHkiyQwgFXRnkOLacUiaSNY=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_rs_zerolog",
    importpath = "github.com/rs/zerolog",
    sum = "h1:MirSo27VyNi7RJYP3078AA1+Cyzd2GB66qy3aUHvsWY=",
    version = "v1.28.0",
)

go_repository(
    name = "com_github_ryanuber_columnize",
    importpath = "github.com/ryanuber/columnize",
    sum = "h1:UFr9zpz4xgTnIE5yIMtWAMngCdZ9p/+q6lTbgelo80M=",
    version = "v0.0.0-20160712163229-9b3edd62028f",
)

go_repository(
    name = "com_github_sean_seed",
    importpath = "github.com/sean-/seed",
    sum = "h1:nn5Wsu0esKSJiIVhscUtVbo7ada43DJhG55ua/hjS5I=",
    version = "v0.0.0-20170313163322-e2103e2c3529",
)

go_repository(
    name = "com_github_slack_go_slack",
    importpath = "github.com/slack-go/slack",
    sum = "h1:NqGXuzni8Is3EJWmsuMuBiCCPbWOlBgTKPvdlwS3Huk=",
    version = "v0.8.1",
)

go_repository(
    name = "com_github_smartystreets_assertions",
    importpath = "github.com/smartystreets/assertions",
    sum = "h1:zE9ykElWQ6/NYmHa3jpm/yHnI4xSofP+UP6SpjHcSeM=",
    version = "v0.0.0-20180927180507-b2de0cb4f26d",
)

go_repository(
    name = "com_github_smartystreets_goconvey",
    importpath = "github.com/smartystreets/goconvey",
    sum = "h1:fv0U8FUIMPNf1L9lnHLvLhgicrIVChEkdzIKYqbNC9s=",
    version = "v1.6.4",
)

go_repository(
    name = "com_github_spaolacci_murmur3",
    importpath = "github.com/spaolacci/murmur3",
    sum = "h1:qLC7fQah7D6K1B0ujays3HV9gkFtllcxhzImRR7ArPQ=",
    version = "v0.0.0-20180118202830-f09979ecbc72",
)

go_repository(
    name = "com_github_spf13_afero",
    importpath = "github.com/spf13/afero",
    sum = "h1:j49Hj62F0n+DaZ1dDCvhABaPNSGNkt32oRFxI33IEMw=",
    version = "v1.9.2",
)

go_repository(
    name = "com_github_spf13_cast",
    importpath = "github.com/spf13/cast",
    sum = "h1:rj3WzYc11XZaIZMPKmwP96zkFEnnAmV8s6XbB2aY32w=",
    version = "v1.5.0",
)

go_repository(
    name = "com_github_spf13_jwalterweatherman",
    importpath = "github.com/spf13/jwalterweatherman",
    sum = "h1:ue6voC5bR5F8YxI5S67j9i582FU4Qvo2bmqnqMYADFk=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_spf13_pflag",
    importpath = "github.com/spf13/pflag",
    sum = "h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=",
    version = "v1.0.5",
)

go_repository(
    name = "com_github_spf13_viper",
    importpath = "github.com/spf13/viper",
    sum = "h1:BWSJ/M+f+3nmdz9bxB+bWX28kkALN2ok11D0rSo8EJU=",
    version = "v1.13.0",
)

go_repository(
    name = "com_github_stretchr_objx",
    importpath = "github.com/stretchr/objx",
    sum = "h1:M2gUjqZET1qApGOWNSnZ49BAIMX4F/1plDv3+l31EJ4=",
    version = "v0.4.0",
)

go_repository(
    name = "com_github_stretchr_testify",
    importpath = "github.com/stretchr/testify",
    sum = "h1:pSgiaMZlXftHpm5L7V1+rVB+AZJydKsMxsQBIJw4PKk=",
    version = "v1.8.0",
)

go_repository(
    name = "com_github_subosito_gotenv",
    importpath = "github.com/subosito/gotenv",
    sum = "h1:jyEFiXpy21Wm81FBN71l9VoMMV8H8jG+qIK3GCpY6Qs=",
    version = "v1.4.1",
)

go_repository(
    name = "com_github_ugorji_go",
    importpath = "github.com/ugorji/go",
    sum = "h1:qYhyWUUd6WbiM+C6JZAUkIJt/1WrjzNHY9+KCIjVqTo=",
    version = "v1.2.7",
)

go_repository(
    name = "com_github_ugorji_go_codec",
    importpath = "github.com/ugorji/go/codec",
    sum = "h1:YPXUKf7fYbp/y8xloBqZOw2qaVggbfwMlI8WM3wZUJ0=",
    version = "v1.2.7",
)

go_repository(
    name = "com_github_yuin_goldmark",
    importpath = "github.com/yuin/goldmark",
    sum = "h1:dPmz1Snjq0kmkz159iL7S6WzdahUTHnHB5M56WFVifs=",
    version = "v1.3.5",
)

go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    sum = "h1:XgtDnVJRCPEUG21gjFiRPz4zI1Mjg16R+NYQjfmU4XY=",
    version = "v0.75.0",
)

go_repository(
    name = "com_google_cloud_go_bigquery",
    importpath = "cloud.google.com/go/bigquery",
    sum = "h1:JuTk8po4bCKRwObdT0zLb1K0BGkGHJdtgs2GK3j2Gws=",
    version = "v1.42.0",
)

go_repository(
    name = "com_google_cloud_go_datastore",
    importpath = "cloud.google.com/go/datastore",
    sum = "h1:/May9ojXjRkPBNVrq+oWLqmWCkr4OU5uRY29bu0mRyQ=",
    version = "v1.1.0",
)

go_repository(
    name = "com_google_cloud_go_firestore",
    importpath = "cloud.google.com/go/firestore",
    sum = "h1:HokMB9Io0hAyYzlGFeFVMgE3iaPXNvaIsDx5JzblGLI=",
    version = "v1.8.0",
)

go_repository(
    name = "com_google_cloud_go_pubsub",
    importpath = "cloud.google.com/go/pubsub",
    sum = "h1:ukjixP1wl0LpnZ6LWtZJ0mX5tBmjp1f8Sqer8Z2OMUU=",
    version = "v1.3.1",
)

go_repository(
    name = "com_google_cloud_go_storage",
    importpath = "cloud.google.com/go/storage",
    sum = "h1:6RRlFMv1omScs6iq2hfE3IvgE+l6RfJPampq8UZc5TU=",
    version = "v1.14.0",
)

go_repository(
    name = "com_shuralyov_dmitri_gpu_mtl",
    importpath = "dmitri.shuralyov.com/gpu/mtl",
    sum = "h1:VpgP7xuJadIUuKccphEpTJnWhS2jkQyMt6Y7pJCD7fY=",
    version = "v0.0.0-20190408044501-666a987793e9",
)

go_repository(
    name = "in_gopkg_check_v1",
    importpath = "gopkg.in/check.v1",
    sum = "h1:Hei/4ADfdWqJk1ZMxUNpqntNwaWcugrBjAiHlqqRiVk=",
    version = "v1.0.0-20201130134442-10cb98267c6c",
)

go_repository(
    name = "in_gopkg_errgo_v2",
    importpath = "gopkg.in/errgo.v2",
    sum = "h1:0vLT13EuvQ0hNvakwLuFZ/jYrLp5F3kcWHXdRggjCE8=",
    version = "v2.1.0",
)

go_repository(
    name = "in_gopkg_ini_v1",
    importpath = "gopkg.in/ini.v1",
    sum = "h1:Dgnx+6+nfE+IfzjUEISNeydPJh9AXNNsWbGP9KzCsOA=",
    version = "v1.67.0",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    importpath = "gopkg.in/yaml.v2",
    sum = "h1:D8xgwECY7CYvx+Y2n4sBz93Jn9JRvxdiyyo8CTfuKaY=",
    version = "v2.4.0",
)

go_repository(
    name = "in_gopkg_yaml_v3",
    importpath = "gopkg.in/yaml.v3",
    sum = "h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=",
    version = "v3.0.1",
)

go_repository(
    name = "io_etcd_go_etcd_api_v3",
    importpath = "go.etcd.io/etcd/api/v3",
    sum = "h1:GsV3S+OfZEOCNXdtNkBSR7kgLobAa/SO6tCxRa0GAYw=",
    version = "v3.5.0",
)

go_repository(
    name = "io_etcd_go_etcd_client_pkg_v3",
    importpath = "go.etcd.io/etcd/client/pkg/v3",
    sum = "h1:2aQv6F436YnN7I4VbI8PPYrBhu+SmrTaADcf8Mi/6PU=",
    version = "v3.5.0",
)

go_repository(
    name = "io_etcd_go_etcd_client_v2",
    importpath = "go.etcd.io/etcd/client/v2",
    sum = "h1:ftQ0nOOHMcbMS3KIaDQ0g5Qcd6bhaBrQT6b89DfwLTs=",
    version = "v2.305.0",
)

go_repository(
    name = "io_opencensus_go",
    importpath = "go.opencensus.io",
    sum = "h1:gqCw0LfLxScz8irSi8exQc7fyQ0fKQU/qnC/X8+V/1M=",
    version = "v0.23.0",
)

go_repository(
    name = "io_opentelemetry_go_proto_otlp",
    importpath = "go.opentelemetry.io/proto/otlp",
    sum = "h1:rwOQPCuKAKmwGKq2aVNnYIibI6wnV7EvzgfTCzcdGg8=",
    version = "v0.7.0",
)

go_repository(
    name = "io_rsc_binaryregexp",
    importpath = "rsc.io/binaryregexp",
    sum = "h1:HfqmD5MEmC0zvwBuF187nq9mdnXjXsSivRiXN7SmRkE=",
    version = "v0.2.0",
)

go_repository(
    name = "io_rsc_quote_v3",
    importpath = "rsc.io/quote/v3",
    sum = "h1:9JKUTTIUgS6kzR9mK1YuGKv6Nl+DijDNIc0ghT58FaY=",
    version = "v3.1.0",
)

go_repository(
    name = "io_rsc_sampler",
    importpath = "rsc.io/sampler",
    sum = "h1:7uVkIFmeBqHfdjD+gZwtXXI+RODJ2Wc4O7MPEh/QiW4=",
    version = "v1.3.0",
)

go_repository(
    name = "org_golang_google_api",
    importpath = "google.golang.org/api",
    sum = "h1:uWrpz12dpVPn7cojP82mk02XDgTJLDPc2KbVTxrWb4A=",
    version = "v0.40.0",
)

go_repository(
    name = "org_golang_google_appengine",
    importpath = "google.golang.org/appengine",
    sum = "h1:FZR1q0exgwxzPzp/aF+VccGrSfxfPpkBqjIIEq3ru6c=",
    version = "v1.6.7",
)

go_repository(
    name = "org_golang_google_genproto",
    importpath = "google.golang.org/genproto",
    sum = "h1:S9GbmC1iCgvbLyAokVCwiO6tVIrU9Y7c5oMx1V/ki/Y=",
    version = "v0.0.0-20221024183307-1bc688fe9f3e",
)

go_repository(
    name = "org_golang_google_grpc",
    build_file_proto_mode = "disable",
    importpath = "google.golang.org/grpc",
    sum = "h1:E1eGv1FTqoLIdnBCZufiSHgKjlqG6fKFf6pPWtMTh8U=",
    version = "v1.51.0",
)

go_repository(
    name = "org_golang_google_protobuf",
    importpath = "google.golang.org/protobuf",
    sum = "h1:d0NfwRgPtno5B1Wa6L2DAG+KivqkdutMf1UhdNx175w=",
    version = "v1.28.1",
)

go_repository(
    name = "org_golang_x_crypto",
    importpath = "golang.org/x/crypto",
    sum = "h1:T8NU3HyQ8ClP4SEE+KbFlg6n0NhuTsN4MyznaarGsZM=",
    version = "v0.0.0-20220525230936-793ad666bf5e",
)

go_repository(
    name = "org_golang_x_exp",
    importpath = "golang.org/x/exp",
    sum = "h1:QE6XYQK6naiK1EPAe1g/ILLxN5RBoH5xkJk3CqlMI/Y=",
    version = "v0.0.0-20200224162631-6cc2880d07d6",
)

go_repository(
    name = "org_golang_x_image",
    importpath = "golang.org/x/image",
    sum = "h1:+qEpEAPhDZ1o0x3tHzZTQDArnOixOzGD9HUJfcg0mb4=",
    version = "v0.0.0-20190802002840-cff245a6509b",
)

go_repository(
    name = "org_golang_x_lint",
    importpath = "golang.org/x/lint",
    sum = "h1:2M3HP5CCK1Si9FQhwnzYhXdG6DXeebvUHFpre8QvbyI=",
    version = "v0.0.0-20201208152925-83fdc39ff7b5",
)

go_repository(
    name = "org_golang_x_mobile",
    importpath = "golang.org/x/mobile",
    sum = "h1:4+4C/Iv2U4fMZBiMCc98MG1In4gJY5YRhtpDNeDeHWs=",
    version = "v0.0.0-20190719004257-d2bd2a29d028",
)

go_repository(
    name = "org_golang_x_mod",
    importpath = "golang.org/x/mod",
    sum = "h1:6zppjxzCulZykYSLyVDYbneBfbaBIQPYMevg0bEwv2s=",
    version = "v0.6.0-dev.0.20220419223038-86c51ed26bb4",
)

go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    sum = "h1:tvrvnPFcdzp294diPnrdZZZ8XUt2Tyj7svb7X52iDuU=",
    version = "v0.0.0-20221014081412-f15817d10f9b",
)

go_repository(
    name = "org_golang_x_oauth2",
    importpath = "golang.org/x/oauth2",
    sum = "h1:5vD4XjIc0X5+kHZjx4UecYdjA6mJo+XXNoaW0EjU5Os=",
    version = "v0.0.0-20210218202405-ba52d332ba99",
)

go_repository(
    name = "org_golang_x_sync",
    importpath = "golang.org/x/sync",
    sum = "h1:5KslGYwFpkhGh+Q16bwMP3cOontH8FOep7tGV86Y7SQ=",
    version = "v0.0.0-20210220032951-036812b2e83c",
)

go_repository(
    name = "org_golang_x_sys",
    importpath = "golang.org/x/sys",
    sum = "h1:XeJjHH1KiLpKGb6lvMiksZ9l0fVUh+AmGcm0nOMEBOY=",
    version = "v0.0.0-20220908164124-27713097b956",
)

go_repository(
    name = "org_golang_x_term",
    importpath = "golang.org/x/term",
    sum = "h1:JGgROgKl9N8DuW20oFS5gxc+lE67/N3FcwmBPMe7ArY=",
    version = "v0.0.0-20210927222741-03fcf44c2211",
)

go_repository(
    name = "org_golang_x_text",
    importpath = "golang.org/x/text",
    sum = "h1:BrVqGRd7+k1DiOgtnFvAkoQEWQvBc25ouMJM6429SFg=",
    version = "v0.4.0",
)

go_repository(
    name = "org_golang_x_time",
    importpath = "golang.org/x/time",
    sum = "h1:7zkz7BUtwNFFqcowJ+RIgu2MaV/MapERkDIy+mwPyjs=",
    version = "v0.0.0-20210723032227-1f47c861a9ac",
)

go_repository(
    name = "org_golang_x_tools",
    importpath = "golang.org/x/tools",
    sum = "h1:VveCTK38A2rkS8ZqFY25HIDFscX5X9OoEhJd3quQmXU=",
    version = "v0.1.12",
)

go_repository(
    name = "org_golang_x_xerrors",
    importpath = "golang.org/x/xerrors",
    sum = "h1:go1bK/D/BFZV2I8cIQd1NKEZ+0owSTG1fDTci4IqFcE=",
    version = "v0.0.0-20200804184101-5ec99f83aff1",
)

go_repository(
    name = "org_uber_go_atomic",
    importpath = "go.uber.org/atomic",
    sum = "h1:ECmE8Bn/WFTYwEW/bpKD3M8VtR/zQVbavAoalC1PYyE=",
    version = "v1.9.0",
)

go_repository(
    name = "org_uber_go_multierr",
    importpath = "go.uber.org/multierr",
    sum = "h1:zaiO/rmgFjbmCXdSYJWQcdvOCsthmdaHfr3Gm2Kx4Ec=",
    version = "v1.7.0",
)

go_repository(
    name = "org_uber_go_zap",
    importpath = "go.uber.org/zap",
    sum = "h1:ue41HOKd1vGURxrmeKIgELGb3jPW9DMUDGtsinblHwI=",
    version = "v1.19.1",
)

load("//:deps.bzl", "go_dependencies")

# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

go_rules_dependencies()

go_register_toolchains(version = "1.17")

gazelle_dependencies()

http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "b1e80761a8a8243d03ebca8845e9cc1ba6c82ce7c5179ce2b295cd36f7e394bf",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.25.0/rules_docker-v0.25.0.tar.gz"],
)

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

container_repositories()

load("@io_bazel_rules_docker//repositories:deps.bzl", container_deps = "deps")

container_deps()

container_repositories()

load("@io_bazel_rules_docker//container:pull.bzl", "container_pull")
load(
    "@io_bazel_rules_docker//go:image.bzl",
    _go_image_repos = "repositories",
)

_go_image_repos()

load(
    "@io_bazel_rules_docker//python:image.bzl",
    _py_image_repos = "repositories",
)

_py_image_repos()

container_pull(
    name = "alpine_linux_amd64",
    digest = "sha256:1304f174557314a7ed9eddb4eab12fed12cb0cd9809e4c28f29af86979a3c870",
    registry = "index.docker.io",
    repository = "library/alpine",
    tag = "3.16",
)

container_pull(
    name = "python3_linux_amd64",
    digest = "sha256:876053fd9c3f645ed21e9e03122ecc1520be76473be89e5d222c748cad432bde",
    registry = "index.docker.io",
    repository = "library/python",
    tag = "3.11.0-slim-buster",
)

#kubectl download
http_archive(
    name = "io_bazel_rules_k8s",
    sha256 = "51f0977294699cd547e139ceff2396c32588575588678d2054da167691a227ef",
    strip_prefix = "rules_k8s-0.6",
    urls = ["https://github.com/bazelbuild/rules_k8s/archive/v0.6.tar.gz"],
)

load("@io_bazel_rules_k8s//toolchains/kubectl:kubectl_configure.bzl", "kubectl_configure")

http_file(
    name = "k8s_binary",
    downloaded_file_path = "kubectl",
    executable = True,
    sha256 = "9f74f2fa7ee32ad07e17211725992248470310ca1988214518806b39b1dad9f0",
    urls = ["https://dl.k8s.io/release/v1.21.0/bin/linux/amd64/kubectl"],
)

kubectl_configure(
    name = "k8s_config",
    kubectl_path = "@k8s_binary//file",
)

#k8s rules loading

load("@io_bazel_rules_k8s//k8s:k8s.bzl", "k8s_defaults", "k8s_repositories")

k8s_repositories()

load("@io_bazel_rules_k8s//k8s:k8s_go_deps.bzl", k8s_go_deps = "deps")

k8s_go_deps()

k8s_defaults(
    name = "k8s_deploy",
    cluster = "cluster.anthony.bible",
    kind = "deployment",
)

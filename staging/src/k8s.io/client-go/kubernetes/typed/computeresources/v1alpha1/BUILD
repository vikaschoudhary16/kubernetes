package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = ["doc.go"],
    importmap = "k8s.io/kubernetes/vendor/k8s.io/client-go/kubernetes/typed/computeresources/v1alpha1",
    importpath = "k8s.io/client-go/kubernetes/typed/computeresources/v1alpha1",
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//staging/src/k8s.io/client-go/kubernetes/typed/computeresources/v1alpha1/fake:all-srcs",
    ],
    tags = ["automanaged"],
)

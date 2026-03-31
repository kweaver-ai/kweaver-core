#!/bin/bash

while getopts "hmtdbelpk" optname
do
    case "$optname" in
    "m")
        echo "====go generate mock====";
        go generate ./...;;
    "l")
        echo "==== golangci-lint run ===="
        golangci-lint run ./... --exclude-dirs=server/tests;;
    "t")
        echo "====go test -v ./...====";
        go test $(go list ./... | grep -v /server/tests/ | grep -v /server/mocks) -gcflags=all=-l -v ;;
    "d")
        echo "====helm template process ====";
        # 使用helm命令渲染模板
        # 如果helm不存在，直接退出
        if ! command -v helm &>/dev/null; then
            echo "helm not found, please install helm first" && exit 1
        fi
        helm template ./helm/agent-retrieval -n agent-retrieval -f./helm/agent-retrieval/values.yaml ;;
    "k")
        echo "====使用ktctl连接远程环境====";
        .ktctl/dev.sh;;
    "p")
        echo "==== preview api docs ====";
        cd docs
        ./preview.sh;;
    "b")
        echo "====go build main.go====";
        # 强制覆盖配置文件
        config_dir="/sysvol/config"
        secret_dir="/sysvol/secret"
        # 检查目录是否存在

        mkdir -p "$config_dir" || exit 1
        mkdir -p "$secret_dir" || exit 1

        # 强制复制配置文件（保留原文件属性）
        echo "覆盖配置文件到 $config_dir"
        cp -f ./server/infra/config/agent-retrieval.yaml "$config_dir/" || exit 1

        # 强制复制密钥文件
        echo "覆盖密钥文件到 $secret_dir"
        cp -f ./server/infra/config/agent-retrieval-secret.yaml "$secret_dir/" || exit 1

        # 强制复制可观测性配置文件
        echo "覆盖可观测性配置文件到 $config_dir"
        cp -f ./server/infra/config/observability.yaml "$config_dir/" || exit 1

        echo "开始构建运行........."
        go run ./server/main.go;;
    "h")
        echo "-p build and preview api docs";
        echo "-m generate mock";
        echo "-l lint";
        echo "-t test";
        echo "-d helm template";
        echo "-b build main.go";
        echo "-k ktctl connect remote cluster";
        echo "-h help";
        exit 0;;
    esac
done
if [ $# -lt 1 ]; then
    echo "try -h"
fi

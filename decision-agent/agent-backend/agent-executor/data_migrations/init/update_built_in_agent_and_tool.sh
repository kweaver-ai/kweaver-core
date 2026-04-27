#!/bin/bash

# 获取包含 agent-executor 服务的命名空间（只取第一个匹配项）
NAMESPACE=$(kubectl get svc -A | grep agent-executor | head -n1 | awk '{print $1}')

if [ -z "$NAMESPACE" ]; then
    echo "错误：未找到包含 agent-executor 服务的命名空间"
    exit 1
fi

echo "使用命名空间: $NAMESPACE"
kubectl exec -it svc/agent-executor -n "$NAMESPACE" -- python3 /app/data_migrations/init/manage_built_in_agent_and_tool.py --update

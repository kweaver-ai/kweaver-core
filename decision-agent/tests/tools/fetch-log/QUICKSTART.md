# 快速开始指南

## 30 秒快速使用

### 1. 进入工具目录
```bash
cd tools/fetch-log
```

### 2. 运行工具
```bash
# 使用默认服务（推荐）

# 或指定自定义服务
```

### 3. 查看结果
```bash
# 查看生成的日志文件
ls -lh log_*.json

# 查看内容（示例）
cat log_*.json | head -50
```

## 常用命令

### 查看帮助
```bash
```

### 检查集群中的服务
```bash
kubectl get pods -A | grep agent
```

### 验证 JSON 格式
```bash
```

### 美化 JSON 输出
```bash
```

## 典型使用场景

### 场景 1: 用户反馈 agent 对话问题
```bash
# 收集 agent-app 和 agent-executor 日志
```

### 场景 2: agent 创建失败
```bash
# 收集 agent-factory 日志
```

### 场景 3: 多个服务同时出问题
```bash
# 收集所有相关服务
```

### 场景 4: 其他服务日志
```bash
# 收集任何 K8s 服务日志
```

## 输出文件说明

生成的文件名格式: `log_YYYYMMDD_HHMMSS.json`

示例: `log_20260108_103901.json`

文件内容结构:
```json
[
  {
    "svc_name": "agent-app",
    "pod": "agent-app-6dd858f8f8-fzk79",
    "fetch_time": "2026-01-08 10:38:59",
    "fecth_log_lines": 300,
    "log_detail": "日志内容..."
  }
]
```

## 故障排除

### 问题: 找不到 kubectl
```bash
# 检查 kubectl 是否安装
which kubectl

# 如果没有，请联系系统管理员安装
```

### 问题: 没有权限访问 pods
```bash
# 检查 kubeconfig 配置
kubectl auth can-i get pods --all-namespaces

# 如果权限不足，请联系集群管理员
```

### 问题: 找不到指定的服务
```bash
# 检查服务是否存在
kubectl get pods -A | grep <服务名>

# 确认服务名称是否正确
```

## 下一步

- 📖 详细使用说明: 查看 `USAGE.md`
- 🔧 项目概述: 查看 `README.md`
- 📊 项目总结: 查看 `PROJECT_SUMMARY.md`
- ✅ 验证报告: 查看 `VERIFICATION_REPORT.md`

## 获取帮助

如遇到问题，请联系 data-agent 研发团队。

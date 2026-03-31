# Security Configuration Guide

## 重要提醒

本项目包含敏感配置信息（API密钥、数据库密码等），这些文件**已经被添加到 .gitignore**，不会被提交到 Git 仓库。

## 敏感文件列表

以下文件包含敏感信息，**不会提交到 GitHub**：

1. **`config/env.ini`** - 环境配置文件
   - 数据库密码
   - 服务器地址
   - 端口配置

2. **`data/data-agent/zhipu_search_tool_config.json`** - 智谱搜索工具配置
   - 智谱 API 密钥

3. **`~/.fetch_log/.fetch_log_ai_config.json`** - Fetch Log AI 配置
   - AI 平台 Token
   - AI 模型名称
   - API 地址

## 新用户配置步骤

### 1. 配置环境变量

复制示例配置文件并填入实际值：

```bash
# 复制环境配置模板
cp config/env.ini.example config/env.ini

# 编辑配置文件，填入实际的数据库密码等信息
vim config/env.ini
```

需要修改的关键配置：
```ini
[server]
host = YOUR_SERVER_IP          # 修改为实际服务器IP
db_pwd = YOUR_DATABASE_PASSWORD  # 修改为实际数据库密码
```

### 2. 配置智谱 API 密钥

```bash
# 复制智谱搜索工具配置模板
cp data/data-agent/zhipu_search_tool_config.json.example data/data-agent/zhipu_search_tool_config.json

# 编辑配置文件，替换 API 密钥
vim data/data-agent/zhipu_search_tool_config.json
```

需要替换的占位符：
- `YOUR_ZHIPU_API_KEY_HERE` → 替换为实际的智谱 API 密钥
- `PLACEHOLDER_TOOL_ID` → 替换为实际的工具ID（如果需要）
- `PLACEHOLDER_BOX_ID` → 替换为实际的工具箱ID（如果需要）

### 3. 配置 Fetch Log AI 分析（可选）

首次使用 `--ai` 参数时，会自动引导配置：
```bash
./tools/fetch-log/build/fetch_log-linux-amd64 --ai
```

按提示输入：
- AI 平台 IP 地址
- 用户 Token
- 模型名称

## 安全最佳实践

### ✅ 应该做的

1. **使用环境变量**（推荐）
   ```bash
   export ZHIPU_API_KEY="your_api_key_here"
   export DB_PASSWORD="your_password_here"
   ```

2. **使用密钥管理服务**
   - AWS Secrets Manager
   - Azure Key Vault
   - HashiCorp Vault

3. **定期轮换密钥**
   - 每 90 天更新一次 API 密钥
   - 每次人员变动时更新密码

4. **使用不同的密钥**
   - 开发环境使用测试密钥
   - 生产环境使用独立密钥

### ❌ 不应该做的

1. **不要提交密钥到 Git**
   - 已通过 .gitignore 保护
   - 使用 `git add -A` 前先检查 `git status`

2. **不要在代码中硬编码密钥**
   ```python
   # ❌ 错误示例
   API_KEY = "1828616286d4c94b26071585e1f93009.negnhMi3D5KVuc7h"
   ```

3. **不要在日志中打印密钥**
   ```python
   # ❌ 错误示例
   print(f"Connecting with API key: {api_key}")
   ```

4. **不要在不安全的渠道传输密钥**
   - 不通过邮件发送
   - 不在即时通讯工具中明文传输

## 检查是否已泄露

如果您之前已经提交了敏感信息到 GitHub，请立即采取以下措施：

### 1. 撤销已泄露的密钥

```bash
# 智谱 API 密钥
登录智谱开放平台：https://open.bigmodel.cn/
删除旧密钥，生成新密钥

# 数据库密码
立即修改数据库密码
```

### 2. 从 Git 历史中清除敏感信息

```bash
# 使用 git-filter-repo 工具
pip install git-filter-repo

# 从所有历史记录中删除敏感文件
git filter-repo --path config/env.ini --invert-paths
git filter-repo --path data/data-agent/zhipu_search_tool_config.json --invert-paths

# 强制推送（⚠️ 谨慎操作）
git push origin --force --all
```

### 3. 检查其他可能的泄露位置

```bash
# 在整个 Git 历史中搜索敏感关键词
git log -p -S "1828616286d4c94b26071585e1f93009"
git log -p -S "eisoo.com123"
```

## 文件权限设置

确保敏感文件只有当前用户可读：

```bash
# 设置配置文件权限为 600（仅所有者可读写）
chmod 600 config/env.ini
chmod 600 data/data-agent/zhipu_search_tool_config.json

# 设置配置目录权限为 700（仅所有者可访问）
chmod 700 config/
chmod 700 data/data-agent/
```

## 验证配置是否正确

在提交代码前，使用以下命令验证敏感文件是否被正确忽略：

```bash
# 检查 .gitignore 是否生效
git check-ignore -v config/env.ini
git check-ignore -v data/data-agent/zhipu_search_tool_config.json

# 查看当前暂存区，确认没有敏感文件
git status

# 搜索暂存区中是否包含敏感关键词
git diff --cached | grep -i "password\|api_key\|secret"
```

## 团队协作

### 配置文件共享策略

1. **使用环境变量**（推荐）
   - 每个开发者维护自己的 `.env` 文件
   - `.env` 已添加到 `.gitignore`
   - 使用 `.env.example` 作为模板

2. **使用加密配置**
   - 使用 `git-crypt` 加密敏感文件
   - 团队成员拥有解密密钥

3. **使用配置管理工具**
   - Ansible Vault
   - SOPS (Mozilla)

## 故障排查

### 问题：测试时提示找不到配置文件

```bash
# 确认配置文件存在
ls -la config/env.ini
ls -la data/data-agent/zhipu_search_tool_config.json

# 如果不存在，从模板复制
cp config/env.ini.example config/env.ini
cp data/data-agent/zhipu_search_tool_config.json.example data/data-agent/zhipu_search_tool_config.json
```

### 问题：Git 仍然显示敏感文件

```bash
# 从 Git 缓存中移除（但保留本地文件）
git rm --cached config/env.ini
git rm --cached data/data-agent/zhipu_search_tool_config.json

# 提交删除操作
git commit -m "Remove sensitive files from git tracking"
git push
```

## 联系方式

如有安全相关问题，请：
- 提交 GitHub Issue（不包含敏感信息）
- 联系项目维护者
- 参考公司安全政策文档

---

**最后更新**: 2026-01-12
**维护者**: Agent AT 团队

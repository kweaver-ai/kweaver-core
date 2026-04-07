# Go 开发环境安装指南

## 安装总结

**安装日期**: 2025-01-08
**Go 版本**: go1.25.5 linux/amd64
**安装路径**: /usr/local/go
**GOPATH**: /root/go

## 安装步骤

### 1. 解压 Go 安装包

```bash
cd /usr/local
rm -rf go
tar -xzf /mnt/agent-AT/go1.25.5.linux-amd64.tar.gz
```

### 2. 配置环境变量

已添加到以下文件：
- `/etc/profile`
- `/root/.bashrc`

环境变量配置：
```bash
export GOROOT=/usr/local/go
export GOPATH=/root/go
export GOBIN=$GOPATH/bin
export PATH=$GOROOT/bin:$GOBIN:$PATH
```

### 3. 创建 Go 目录结构

```bash
mkdir -p $GOPATH/{src,pkg,bin}
```

## 验证安装

### 验证 Go 版本

```bash
$ go version
go version go1.25.5 linux/amd64
```

### 验证环境变量

```bash
$ echo $GOROOT
/usr/local/go

$ echo $GOPATH
/root/go

$ echo $GOBIN
/root/go/bin
```

### 查看 Go 环境配置

```bash
$ go env
```

## 测试 Go 编译

### 编译日志收集工具

```bash
cd tools/fetch-log
make build
```

**编译结果**:
- 编译成功 ✓
- 生成文件: `build/fetch_log`
- 文件大小: 3.1 MB

### 测试编译后的程序

```bash
$ ./build/fetch_log --help
Usage of ./build/fetch_log:
  -svc_list string
    	Comma-separated list of service names
```

## Go 环境配置详情

### 重要路径

| 路径 | 说明 |
|------|------|
| `/usr/local/go` | Go 安装目录 (GOROOT) |
| `/root/go` | Go 工作目录 (GOPATH) |
| `/root/go/src` | Go 源代码目录 |
| `/root/go/pkg` | Go 包目录 |
| `/root/go/bin` | Go 可执行文件目录 |

### Go 环境变量

| 变量 | 值 | 说明 |
|------|------|------|
| GOROOT | /usr/local/go | Go 安装根目录 |
| GOPATH | /root/go | Go 工作目录 |
| GOBIN | /root/go/bin | Go 可执行文件目录 |
| PATH | 包含 $GOROOT/bin 和 $GOBIN | 可执行文件搜索路径 |

### Go 编译器信息

- **Go 版本**: go1.25.5
- **操作系统**: linux
- **架构**: amd64
- **CGO**: 启用
- **编译器**: gcc/g++

## 使用 Makefile 编译

### 可用的编译命令

```bash
# 编译当前平台
make build

# 编译 Linux AMD64
make build-linux-amd64

# 编译 Linux ARM64
make build-linux-arm64

# 编译所有平台
make build-all

# 清理编译产物
make clean
```

### 交叉编译

当前环境支持交叉编译到：
- Linux AMD64
- Linux ARM64

## 常用 Go 命令

### 基础命令

```bash
# 查看 Go 版本
go version

# 查看环境变量
go env

# 查看所有 Go 命令
go help
```

### 编译相关

```bash
# 编译 Go 程序
go build main.go

# 编译并指定输出文件名
go build -o myapp main.go

# 编译并优化
go build -ldflags="-s -w" main.go
```

### 模块管理

```bash
# 初始化模块
go mod init module-name

# 下载依赖
go mod download

# 整理依赖
go mod tidy

# 验证依赖
go mod verify
```

### 测试相关

```bash
# 运行测试
go test

# 运行测试并显示详细输出
go test -v

# 运行测试并生成覆盖率报告
go test -cover
```

## 故障排除

### 问题 1: go 命令未找到

**解决方案**:
```bash
# 手动设置环境变量
export GOROOT=/usr/local/go
export GOPATH=/root/go
export GOBIN=$GOPATH/bin
export PATH=$GOROOT/bin:$GOBIN:$PATH

# 或重新加载配置
source /etc/profile
source ~/.bashrc
```

### 问题 2: 编译失败

**常见原因**:
1. Go 版本不兼容
2. 依赖包缺失
3. 代码语法错误

**解决方案**:
```bash
# 检查 Go 版本
go version

# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

### 问题 3: GOPATH 配置错误

**解决方案**:
```bash
# 检查 GOPATH
echo $GOPATH

# 如果为空或错误，重新设置
export GOPATH=/root/go
mkdir -p $GOPATH/{src,pkg,bin}
```

## 性能优化

### 编译优化

```bash
# 减小可执行文件大小
go build -ldflags="-s -w" main.go

# 并行编译
go build -p 4 main.go
```

### 环境变量优化

```bash
# 启用 Go Modules
export GO111MODULE=on

# 设置 Go 代理（中国）
export GOPROXY=https://goproxy.cn,direct
```

## 升级 Go 版本

当需要升级 Go 版本时：

```bash
# 1. 下载新版本
# 2. 删除旧版本
rm -rf /usr/local/go

# 3. 解压新版本
tar -xzf goX.X.X.linux-amd64.tar.gz -C /usr/local

# 4. 验证新版本
go version
```

## 总结

✅ Go 1.25.5 已成功安装
✅ 环境变量已配置
✅ 目录结构已创建
✅ 编译测试通过
✅ 日志收集工具编译成功

Go 开发环境已完全配置好，可以正常使用！

---

**安装人员**: Claude Code
**验证时间**: 2025-01-08
**状态**: ✅ 安装成功并测试通过

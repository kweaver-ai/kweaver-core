<!-- 请勿直接编辑：本文件由 docs/user_manual/setup 下的 `make aggregate` 生成。 -->
<!-- 来源文件：index.md, install-local-sdk-package.md。 -->

# 安装与部署准备（聚合版）

> 本文件由脚本生成，请不要直接修改本文件；如需调整内容，请修改分文件文档后运行 `make -C docs/user_manual/setup aggregate`。

## 目录

- [安装与部署准备](#安装与部署准备)
- [本地安装 TypeScript SDK 包](#本地安装-typescript-sdk-包)


<!-- 来源：index.md -->

## 安装与部署准备

安装与部署准备文档面向需要准备本地开发、示例运行、SDK 验证和后续 Decision Agent 部署的用户。这里不替代 API、CLI、SDK 的使用指南，而是集中说明“环境如何准备好”。

### 推荐阅读顺序

1. [本地安装 TypeScript SDK 包](#本地安装-typescript-sdk-包)：从本地 SDK 项目安装 `@kweaver-ai/kweaver-sdk`，用于验证未发布版本或本地修改。

### 与其他用户手册的关系

- 使用 TypeScript SDK 调用 Agent，请阅读 [TypeScript SDK 用户指南](../sdk/typescript/README.md)。
- 使用 CLI 操作 Agent，请阅读 [CLI 用户指南](../cli/README.md)。
- 直接调用 REST API，请阅读 [Decision Agent REST 接入指南](../api/README.md)。
- 只想理解 Agent 概念，请阅读 [Agent 概念指南](../concepts/README.md)。

后续如果补充 Decision Agent 部署、启动、健康检查、依赖服务准备等内容，也会继续放在本目录下。


<!-- 来源：install-local-sdk-package.md -->

## 本地安装 TypeScript SDK 包

本页说明如何从本地下载的 TypeScript SDK 项目安装 `@kweaver-ai/kweaver-sdk`。这种方式适合验证本地 SDK 修改、使用尚未发布到 npm 的版本，或在 Decision Agent 示例中切换到本地包。

如果只是普通业务接入，优先使用远程 npm 包：

```bash
npm install @kweaver-ai/kweaver-sdk
```

本地包安装后，用户项目中的 import 方式保持一致：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';
```

### 前置条件

- Node.js 22 或更高版本。
- 已下载 TypeScript SDK 项目。
- 本地 SDK 包目录为 `<path-to-kweaver-sdk>/packages/typescript`。
- 目标项目已经有 `package.json`，或准备初始化一个新的 Node.js 项目。

### 构建本地 SDK 包

先在 SDK 包目录安装依赖并构建：

```bash
cd <path-to-kweaver-sdk>/packages/typescript
npm install
npm run build
```

如果使用 pnpm：

```bash
cd <path-to-kweaver-sdk>/packages/typescript
pnpm install
pnpm build
```

每次修改 SDK 源码后，建议重新构建。使用 npm 从本地路径安装时，通常还需要在目标项目中重新安装一次本地包。

### 使用 npm 安装到目标项目

在目标项目中初始化并安装本地包：

```bash
cd <your-project>
npm init -y
npm install <path-to-kweaver-sdk>/packages/typescript
```

如果目标项目已经存在，只需要执行安装命令：

```bash
npm install <path-to-kweaver-sdk>/packages/typescript
```

安装后，`package.json` 中通常会出现类似依赖：

```json
{
  "dependencies": {
    "@kweaver-ai/kweaver-sdk": "file:../path/to/packages/typescript"
  }
}
```

npm 会把本地包复制到目标项目的 `node_modules`。如果 SDK 源码有新修改，需要重新运行安装命令。

### 使用 pnpm 安装到目标项目

在目标项目中初始化并安装本地包：

```bash
cd <your-project>
pnpm init
pnpm install <path-to-kweaver-sdk>/packages/typescript
```

pnpm 通常会使用 `link:` 或相关本地链接方式。修改 SDK 源码后，仍建议先在 SDK 包目录重新构建，再回到目标项目验证。

### 验证安装

在目标项目中检查包入口：

```bash
node --input-type=module <<'EOF'
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

console.log(typeof kweaver.getClient);
EOF
```

如果输出 `function`，说明包入口可用。

也可以连接本地 no-auth Decision Agent 服务做一次最小检查：

```bash
export KWEAVER_BASE_URL="${KWEAVER_BASE_URL:-http://127.0.0.1:13020}"
export KWEAVER_NO_AUTH=1

node --input-type=module <<'EOF'
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  auth: false,
});

console.log(typeof kweaver.getClient);
EOF
```

### 项目内使用 CLI

如果 SDK 包安装在当前项目中，不需要全局安装 CLI，可以在项目目录下使用：

```bash
npx kweaver --help
npx kweaver agent --help
```

pnpm 项目可以使用：

```bash
pnpm exec kweaver --help
pnpm exec kweaver agent --help
```

也可以直接调用包内 bin 文件：

```bash
node node_modules/@kweaver-ai/kweaver-sdk/bin/kweaver.js --help
```

如果希望命令更短，可以在目标项目的 `package.json` 中添加脚本：

```json
{
  "scripts": {
    "kweaver": "kweaver",
    "kweaver:agent": "kweaver agent"
  }
}
```

之后可以运行：

```bash
npm run kweaver -- --help
npm run kweaver:agent -- --help
```

### 卸载和切换包来源

从目标项目卸载 SDK：

```bash
npm uninstall @kweaver-ai/kweaver-sdk
```

pnpm 项目：

```bash
pnpm remove @kweaver-ai/kweaver-sdk
```

如果需要从本地包切回远程 npm 包：

```bash
npm uninstall @kweaver-ai/kweaver-sdk
npm install @kweaver-ai/kweaver-sdk
```

如果需要从远程包切回本地包：

```bash
npm uninstall @kweaver-ai/kweaver-sdk
npm install <path-to-kweaver-sdk>/packages/typescript
```

Decision Agent 用户手册的可运行示例已经提供 `KWEAVER_SDK_PACKAGE_SOURCE=remote|local` 开关，可在远程包和本地包之间切换；详见 [可运行示例](../examples/README.md)。

### 全局安装补充

全局安装后可以在任意目录直接使用 `kweaver` 命令：

```bash
cd <path-to-kweaver-sdk>/packages/typescript
npm install
npm run build
npm install -g .
```

验证：

```bash
kweaver --help
kweaver agent --help
```

全局安装会影响当前用户环境中的所有项目。日常示例和项目集成建议优先使用项目内安装，避免不同项目之间的 SDK 版本互相影响。

卸载全局包：

```bash
npm uninstall -g @kweaver-ai/kweaver-sdk
```

### 常见问题

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 安装后 import 报错 | SDK 包没有构建，或目标项目仍使用旧包 | 在 SDK 包目录运行 `npm run build`，然后在目标项目重新安装。 |
| npm 安装后源码修改不生效 | npm 会复制本地包文件 | 重新运行 `npm install <path-to-kweaver-sdk>/packages/typescript`。 |
| CLI 命令找不到 | 只做了项目内安装，但在其他目录直接运行 `kweaver` | 回到目标项目目录使用 `npx kweaver`，或显式全局安装。 |
| 权限错误 | 全局安装写入系统目录失败 | 优先改用项目内安装；如必须全局安装，配置 npm 用户目录后再安装。 |

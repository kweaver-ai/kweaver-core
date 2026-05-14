# SDK 安装和运行

本页统一说明 TypeScript SDK 的安装、运行、认证和本地调试方式。后续章节只展示可放入用户项目的 JS/TS 代码，不再重复命令行运行壳。

## 环境要求

- Node.js 22 或更高版本。
- 推荐使用 ESM 项目，在 `package.json` 中设置 `"type": "module"`。
- 如果项目不是 ESM，也可以把示例代码放在 `.mjs` 文件中运行。
- TypeScript 项目可按项目已有方式运行，例如 `tsx`、`ts-node`、`vite-node` 或构建后执行。

## 使用远程 npm 仓库安装

在用户项目中安装 SDK：

```bash
npm install @kweaver-ai/kweaver-sdk
```

最小 `package.json` 示例：

```json
{
  "type": "module",
  "dependencies": {
    "@kweaver-ai/kweaver-sdk": "^0.7.3"
  }
}
```

如果已经有项目，只需要保留现有配置并安装依赖即可。

## 使用本地下载的 SDK 项目安装

如果 SDK 项目已经下载到本地，可以从本地目录安装：

```bash
npm install /path/to/kweaver-sdk/packages/typescript
```

例如：

```bash
npm install <path-to-kweaver-sdk>/packages/typescript
```

这种方式适合在 SDK 尚未发布到远程 npm 仓库，或需要验证本地 SDK 修改时使用。用户项目中的 import 方式仍然保持不变：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';
```

## 认证和连接配置

连接远程环境时，通常使用服务地址和访问令牌：

```bash
export KWEAVER_BASE_URL=https://your-kweaver.example.com
export KWEAVER_TOKEN=your-access-token
```

连接本地 no-auth 环境时：

```bash
export KWEAVER_BASE_URL=http://127.0.0.1:13020
export KWEAVER_NO_AUTH=1
```

在 JS 代码中可以显式配置：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  accessToken: process.env.KWEAVER_TOKEN,
});
```

本地 no-auth 模式可以写成：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({
  baseUrl: process.env.KWEAVER_BASE_URL,
  auth: false,
});
```

如果已经通过 CLI 登录并保存配置，也可以让 SDK 读取本机配置：

```js
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({ config: true });
```

## 运行示例

将示例代码保存为 `index.mjs` 后运行：

```bash
node index.mjs
```

也可以直接在命令行中做一次包入口快速检查：

```bash
export KWEAVER_BASE_URL=http://127.0.0.1:13020
export KWEAVER_NO_AUTH=1
node --input-type=module <<'EOF'
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';
kweaver.configure({ baseUrl: process.env.KWEAVER_BASE_URL, auth: false });
console.log(typeof kweaver.getClient);
EOF
```

如果输出 `function`，说明包入口和基础配置可以正常工作。

# CLI 本地调试流程

本页适合 SDK/CLI 开发者在本地源码目录中验证 CLI 与本地 Decision Agent 服务的联动。普通 CLI 用户请优先阅读 [CLI 安装和运行](./install-and-run.md)，并直接使用 `kweaver ...` 命令。

## 健康检查

```bash
cd <path-to-kweaver-sdk>/packages/typescript

curl -fsS "${KWEAVER_BASE_URL:-http://127.0.0.1:13020}/health/ready"
```

## 本地源码入口

```bash
node --import tsx src/cli.ts agent --help
node --import tsx src/cli.ts agent create --help
node --import tsx src/cli.ts agent chat --help
```


## 构建后入口

```bash
npm run build
node bin/kweaver.js agent --help
node bin/kweaver.js agent chat --help
```

如果需要模拟全局安装后的用户体验，可以在用户项目或临时目录中安装本地包：

```bash
npm install -g <path-to-kweaver-sdk>/packages/typescript
kweaver agent --help
```

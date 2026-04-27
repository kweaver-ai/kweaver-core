# Sandbox Web

沙箱管理平台前端服务，用于对接 Sandbox Control Plane API，实现模板管理、会话管理和代码执行功能。

## 技术栈

- **React 18.3.1** - 用户界面库
- **TypeScript 5.8.2** - 类型安全的 JavaScript 超集
- **Ant Design 5.26.2** - 企业级 UI 设计语言和组件库
- **Rsbuild** - 基于 Rspack 的构建工具
- **react-router-dom** - 路由管理
- **axios** - HTTP 客户端
- **@monaco-editor/react** - Monaco 代码编辑器

## 项目结构

```
sandbox_web/
├── public/               # 公共静态资源
├── src/
│   ├── apis/            # API 接口管理
│   │   ├── sessions/    # Session API
│   │   ├── templates/   # Template API
│   │   ├── executions/  # Execution API
│   │   ├── files/       # File Upload/Download API
│   │   └── health/      # Health Check API
│   ├── components/      # React 组件
│   │   ├── Layout/      # 布局组件
│   │   ├── CodeEditor/  # 代码编辑器
│   │   └── ...
│   ├── hooks/           # 自定义 Hooks
│   ├── pages/           # 页面入口
│   ├── styles/          # 样式文件
│   ├── utils/           # 工具函数
│   ├── types/           # 类型定义
│   ├── constants/       # 常量定义
│   ├── router/          # 路由配置
│   ├── App.tsx
│   └── main.tsx
├── rsbuild.config.mts   # Rsbuild 配置
├── package.json
└── tsconfig.json
```

## 快速开始

### 环境要求

- Node.js 18+
- npm 或 yarn

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

开发服务器将在 http://localhost:1101 启动。

### 构建项目

```bash
npm run build
```

### 代码检查

```bash
npm run lint
```

## 功能模块

### 模板管理
- 创建、编辑、删除执行环境模板
- 支持 Python 3.11、Node.js 20、Java 17、Go 1.21 运行时
- 配置 CPU、内存、磁盘等资源限制

### 会话管理
- 创建和管理代码执行会话
- 实时查看会话状态
- 上传/下载会话文件
- 支持 Python 依赖包安装

### 代码执行
- 在线代码编辑器 (Monaco Editor)
- 支持 Lambda Handler 格式
- 执行历史查看
- 实时结果展示

## API 对接

前端服务通过代理对接 `http://localhost:8000` 的 Control Plane API。

主要端点：
- `/api/v1/templates` - 模板管理
- `/api/v1/sessions` - 会话管理
- `/api/v1/executions` - 代码执行

## 开发规范

### 组件命名
- 组件名使用 PascalCase：`<MyComponent />`
- 文件名与组件名保持一致

### 代码规范
- 使用 TypeScript 进行类型检查
- 函数组件使用 Hooks
- 遵循 ESLint 和 Prettier 配置

## 设计规范

| 属性 | 值 |
|------|-----|
| 主色 | #126ee3 |
| 背景色 | #fafafa |
| 边框色 | #e7edf7 |
| 圆角 | 4px / 8px / 12px |

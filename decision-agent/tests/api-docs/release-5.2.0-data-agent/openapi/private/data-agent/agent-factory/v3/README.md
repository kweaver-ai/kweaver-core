# 数据智能体 (DATA AGENT) API 文档

## 项目概述

本项目包含数据智能体 (DATA AGENT) API 的 OpenAPI 规范文档，用于定义和描述与数据智能体相关的各种 API 接口。该文档使用 OpenAPI 3.0.2 规范编写，并通过 Redocly 工具生成可视化的 HTML 文档。

## 目录结构

```
/
├── .gitignore                      # Git 忽略文件
├── Makefile                        # 用于生成 HTML 文档的 Makefile
├── agent-factory.html              # 生成的 HTML API 文档（gitignore）
├── agent-factory.yaml              # 主 OpenAPI 规范文件
├── components
├   ├── components.yaml             # 公共组件定
├── com_objs/                       # 公共对象目录
│   ├── paths.yaml                  # 公共路径定义
│   └── schema/                     # 公共模式目录
│       └── product_list_item.yaml  # 产品列表项模式定义
└── da_config/                      # 数据智能体配置目录
    ├── paths.yaml                  # 路径定义
    └── schema/                     # 模式定义目录
        ├── add_req.yaml            # 添加请求模式
        ├── arg_res.yaml            # 参数响应模式
        ├── config_req.yaml         # 配置请求模式
        ├── field.yaml              # 字段模式
        ├── kg_source.yaml          # 知识图谱源模式
        ├── llm_config.yaml         # LLM 配置模式
        ├── nl2nqgl_config_req.yaml # NL2NQGL 配置请求模式
        ├── retriever_data_source.yaml # 检索器数据源模式
        └── update_req.yaml         # 更新请求模式
```

## API 分类

该 API 文档包含以下几个主要分类：

1. **配置页面** - 智能体配置相关 API
2. **使用页面** - 智能体使用相关 API
3. **data agent市场** - 智能体市场相关 API
4. **其他** - 其他相关 API
5. **自定义错误码汇总** - API 错误码说明
6. **相关文档** - 项目相关文档链接

## 使用方法

### 生成 HTML 文档

使用以下命令生成 HTML 格式的 API 文档：

```bash
make html
```

这将使用 Redocly 工具将 OpenAPI 规范文件转换为 HTML 文档，并自动打开生成的文档。

### 其他可用命令

- 将 YAML 转换为 JSON：

```bash
make yaml2json
```

- 将 JSON 转换为 YAML：

```bash
make json2yaml
```

## 认证方式

调用需要鉴权的 API 时，必须将 token 放在 HTTP header 中：

```
Authorization: Bearer ACCESS_TOKEN
```

## 服务器信息

默认服务器 URL：`https://{host}:{port}/api/`

- `host`: DIP 服务器 IP
- `port`: 默认端口 443

## 相关资源

- [开发者社区](https://developers.aishu.cn) - 如有任何疑问，可到开发者社区提问
- [Confluence 文档](https://confluence.aishu.cn/pages/viewpage.action?pageId=261242083)
- [DevOps 工作项](https://devops.aishu.cn/AISHUDevOps/DIP/_workitems/edit/744857)

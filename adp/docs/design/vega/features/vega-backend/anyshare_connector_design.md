# AnyShare连接器设计文档

## 目录
- [1. 需求背景](#1-需求背景)
  - [1.1 业务需求](#11-业务需求)
  - [1.2 功能需求](#12-功能需求)
  - [1.3 非功能需求](#13-非功能需求)
  - [1.4 与MariaDB连接器的对比](#14-与mariadb连接器的对比)
- [2. 逻辑架构](#2-逻辑架构)
  - [2.1 整体架构](#21-整体架构)
  - [2.2 数据流程](#22-数据流程)
  - [2.3 核心概念映射](#23-核心概念映射)
- [3. 连接器架构](#3-连接器架构)
  - [3.1 连接器配置](#31-连接器配置)
  - [3.2 连接器接口实现](#32-连接器接口实现)
    - [3.2.1 基础接口方法](#321-基础接口方法)
    - [3.2.2 连接管理方法](#322-连接管理方法)
    - [3.2.3 发现方法](#323-发现方法)
    - [3.2.4 查询方法](#324-查询方法)
  - [3.3 核心功能模块](#33-核心功能模块)
    - [3.3.1 认证模块](#331-认证模块)
    - [3.3.2 发现模块](#332-发现模块)
    - [3.3.3 Fileset信息字段与Resource映射](#333-fileset信息字段与resource映射)
  - [3.4 API接口封装](#34-api接口封装)
    - [3.4.1 数据结构定义](#341-数据结构定义)
    - [3.4.2 接口说明](#342-接口说明)
      - [3.4.2.1 获取知识库信息](#3421-获取知识库信息)
      - [3.4.2.2 根据路径获取对象信息](#3422-根据路径获取对象信息)
      - [3.4.2.3 获取文档详细信息](#3423-获取文档详细信息)
      - [3.4.2.4 获取文档下载信息](#3424-获取文档下载信息)
- [4. 实现计划](#4-实现计划)
  - [4.1 第一阶段：连接器核心功能](#41-第一阶段连接器核心功能)
  - [4.2 第二阶段：数据源核心功能](#42-第二阶段数据源核心功能)

## 1. 需求背景

### 1.1 业务需求
AnyShare是文档管理系统，包含知识库和文档库等模块，用于企业知识管理和文档协作。当前需要将AnyShare集成到Vega系统中，实现对AnyShare中文档的统一管理和查询。

### 1.2 功能需求
1. **扩展anyshare catalog类型**
   - 支持AnyShare连接和认证
   - 支持token和app_id/app_secret两种认证方式

2. **扩展fileSet resource类型**
   - 支持fileSet资源采集和管理
   - 包括知识库信息获取、指定路径采集

3. **支持fileSet类型resource查询**
   - 为实现查询功能，需支持fileSet下载地址获取

### 1.3 非功能需求
1. 遵循Vega系统连接器规范
2. 支持敏感信息加密存储
3. 提供完善的错误处理机制
4. 提供完整的日志记录功能
5. 支持并发处理和性能优化
6. 支持接口扩展

### 1.4 与MariaDB连接器的对比
AnyShare连接器与MariaDB连接器的主要区别如下：

| 特性 | MariaDB | AnyShare |
|-----|---------|----------|
| 连接方式 | JDBC | HTTPS |
| 认证方式 | 用户名/密码 | Token或AppID/AppSecret |
| 数据模型 | Table | Folder |
| 数据获取 | SQL查询 | API调用+获取下载地址 |
| 数据类型 | 结构化数据 | 文档文件 |

## 2. 逻辑架构

### 2.1 整体架构
AnyShare连接器采用分层架构设计，包括以下层次：

```
┌─────────────────────────────────────────────────┐
│           Vega系统层            │
├─────────────────────────────────────────────────┤
│         连接器管理层         │
├─────────────────────────────────────────────────┤
│       AnyShare连接器层      │
│  ┌──────────┬──────────┬──────────┐           │
│  │ 认证模块  │ 发现模块  │ 查询模块  │           │
│  └──────────┴──────────┴──────────┘           │
└─────────────────────────────────────────────────┘
                    ↓
         ┌─────────────────────┐
         │   AnyShare系统      │
         │  (外部系统，类似    │
         │   MySQL数仓)       │
         └─────────────────────┘
```

### 2.2 数据流程
1. **连接建立流程**
   ```
   用户配置 -> 连接器初始化 -> 认证 -> 连接建立
   ```

2. **发现流程**
   ```
   判断是否配置paths参数 -> 
   ├─ 未配置: 获取知识库列表(仅第一层) -> 转换为FilesetMeta返回
   └─ 已配置: 遍历每个路径 ->
      ├─ 根据路径调用getinfobypath获取doc_id ->
      ├─ 检查size字段，若size!=-1则返回错误(路径必须指向文件夹) ->
      └─ 调用items接口获取文件夹详细信息 -> 转换为FilesetMeta返回
   ```
   
   注意：无论是否配置paths参数，都只返回第一层资源(知识库或指定的文件夹)，不递归查询子内容。

3. **查询流程**
   ```
   获取文件夹ID -> 调用下载API -> 获取下载地址 -> 返回下载地址
   ```

### 2.3 核心概念映射
| AnyShare概念 | Vega概念 | 说明 |
|---------|----------|------|
| AnyShare实例 | Catalog | 一个AnyShare实例对应一个Catalog |
| 文件夹 | Resource | 文件夹作为fileset类型Resource |

## 3. 连接器架构

### 3.1 连接器配置
连接器配置包含以下字段：

```go
type anyshareConfig struct {
	Protocol   string   `mapstructure:"protocol"`   // 连接协议（http或https），必传，非加密
	Host       string   `mapstructure:"host"`       // 主机地址，必传，非加密
	Port       int      `mapstructure:"port"`       // 端口号，必传，非加密
	AuthType   int      `mapstructure:"auth_type"`  // 认证方式（1-token，2-app_id/app_secret），必传，非加密
	Token      string   `mapstructure:"token"`     // 认证token，非必传，加密
	AppID      string   `mapstructure:"app_id"`    // 应用账户id，非必传，非加密
	AppSecret  string   `mapstructure:"app_secret"`// 应用账户秘钥，非必传，加密
	DocLibType int      `mapstructure:"doc_lib_type"` // 文档库类型（1-知识库），必传，非加密
	Paths      []string `mapstructure:"paths"`     // 文档路径列表，非必传，非加密
}
```

### 3.2 连接器接口实现
AnyShare连接器实现`FilesetConnector`接口：

```go
type AnyShareConnector struct {
	enabled    bool
	config     *anyshareConfig
	connected  bool
	httpClient *http.Client
	baseURL    string
	authHeader string // 认证头，格式为 "Bearer {token}"
}
```

#### 3.2.1 基础接口方法
- `GetType()`: 返回"anyshare"
- `GetName()`: 返回"anyshare"
- `GetMode()`: 返回"local"
- `GetCategory()`: 返回"fileset"
- `GetEnabled()`: 返回启用状态
- `SetEnabled(bool)`: 设置启用状态
- `GetSensitiveFields()`: 返回["token", "app_secret"]
- `GetFieldConfig()`: 返回字段配置

#### 3.2.2 连接管理方法
- `New(cfg)`: 创建连接器实例
- `Connect(ctx)`: 建立连接
- `Ping(ctx)`: 检查连接状态
- `Close(ctx)`: 关闭连接
- `TestConnection(ctx)`: 测试连接

#### 3.2.3 发现方法
- `ListFilesets(ctx)`: 发现资源，根据配置返回知识库列表或指定路径对应的文件夹

#### 3.2.4 查询方法
- `GetObjectDownloadInfo(ctx, resourceName, docID)`: 获取文件或文件夹的下载信息

### 3.3 核心功能模块

#### 3.3.1 认证模块
认证模块负责处理与AnyShare系统的认证，支持两种认证方式：

1. **Token认证**
   - 直接使用提供的token进行API调用
   - Token在HTTP请求头中携带：`Authorization: Bearer {token}`

2. **AppID/AppSecret认证**
   - 使用AppID和AppSecret获取临时token
   - 使用临时token进行API调用

**获取Token接口**：
- URL: `{protocol}://{host}:{port}/oauth2/token`
- Method: POST
- Authentication: Basic Auth
  - username: AppID
  - password: AppSecret
- Content-Type: application/x-www-form-urlencoded

**请求参数**：

| 参数名   | 类型   | 必填 | 说明                     |
|---------|--------|------|--------------------------|
| grant_type | string | 是   | 固定值：client_credentials |
| scope     | string | 是   | 固定值：all              |

**返回值**：
```json
{
  "access_token": "ory_at_xxx",
  "expires_in": 3599,
  "scope": "all",
  "token_type": "bearer"
}
```

**数据结构定义**：
```go
// tokenResponse Token响应
type tokenResponse struct {
    AccessToken string `json:"access_token"` // 访问令牌
    ExpiresIn   int    `json:"expires_in"`  // 过期时间（秒）
    Scope       string `json:"scope"`       // 权限范围
    TokenType   string `json:"token_type"`  // 令牌类型
    Error       string `json:"error"`       // 错误信息
}
```

**认证实现**：
```go
func (c *AnyShareConnector) authenticate(ctx context.Context) error {
    if c.config.AuthType == 1 {
        // Token认证
        c.authToken = c.config.Token
    } else if c.config.AuthType == 2 {
        // AppID/AppSecret认证
        token, err := c.getTokenByAppCredentials(ctx)
        if err != nil {
            return err
        }
        c.authToken = token
    }
    return nil
}

// getTokenByAppCredentials 使用AppID和AppSecret获取Token
func (c *AnyShareConnector) getTokenByAppCredentials(ctx context.Context) (string, error) {
    url := fmt.Sprintf("%s://%s:%d/oauth2/token", c.config.Protocol, c.config.Host, c.config.Port)
    
    // 创建请求
    req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
    if err != nil {
        return "", err
    }
    
    // 设置Basic Auth
    req.SetBasicAuth(c.config.AppID, c.config.AppSecret)
    
    // 设置请求体
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    // 设置表单参数
    formData := url.Values{}
    formData.Set("grant_type", "client_credentials")
    formData.Set("scope", "all")
    req.Body = ioutil.NopCloser(strings.NewReader(formData.Encode()))
    
    // 发送请求
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    // 解析响应
    var tokenResp TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return "", err
    }
    
    return tokenResp.AccessToken, nil
}
```

#### 3.3.2 发现模块
发现模块负责查询知识库中的文件夹和文档：

1. **获取知识库信息**
   - 调用`/api/document/v1/entry-doc-lib`接口，传入type参数为knowledge_doc_lib
   - 获取知识库的基本信息（注意：该接口只返回知识库列表，不包含文件）
   - 如果没有配置paths参数，只返回知识库层级的信息，不查询知识库下的内容

2. **路径解析**（仅当配置了paths参数时执行）
   - 如果配置了paths参数，调用`/api/efast/v1/file/getinfobypath`接口
   - 根据路径获取对应的doc_id和文件信息
   - **文档处理**：当size!=-1时，表示当前路径是文档，返回错误，paths必须指向文件夹
   - **文件夹处理**：当size=-1时，表示当前路径是文件夹，获取文件夹详细信息并作为resource返回

注意：配置paths参数时，只返回文件夹本身作为resource，不查询文件夹下的子内容

```go
// buildFilesetMeta 创建 FilesetMeta 对象的辅助函数
func buildFilesetMeta(id, name, displayPath, itemType, createdAt, modifiedAt string, createdBy, modifiedBy userInfoDTO, rev string) *interfaces.FilesetMeta {
	return &interfaces.FilesetMeta{
		ID:          id,
		Name:        name,
		DisplayPath: displayPath,
		SourceMetadata: map[string]any{
			"id":          id,
			"name":        name,
			"type":        itemType,
			"created_at":  createdAt,
			"modified_at": modifiedAt,
			"created_by":  createdBy,
			"modified_by": modifiedBy,
			"rev":         rev,
		},
	}
}

// ListFilesets discovers one level of files/folders per configured roots (see design doc).
func (c *AnyShareConnector) ListFilesets(ctx context.Context) ([]*interfaces.FilesetMeta, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	var out []*interfaces.FilesetMeta

	if len(c.config.Paths) == 0 {
		// 获取知识库列表（仅第一层）
		libs, err := c.getEntryDocLib(ctx)
		if err != nil {
			return nil, err
		}

		// 将知识库信息转换为FilesetMeta返回
		for _, lib := range libs {
			meta := buildFilesetMeta(lib.ID, lib.Name, lib.Name, lib.Type, lib.CreatedAt, lib.ModifiedAt, lib.CreatedBy, lib.ModifiedBy, lib.Rev)
			out = append(out, meta)
		}
		return out, nil
	}

	// 遍历每个路径
	for _, p := range c.config.Paths {

		// 根据路径获取doc_id
		docInfo, err := c.getDocIDByPath(ctx, p)
		if err != nil {
			return nil, err
		}

		// 检查路径是否指向文件（size != -1 表示文件）
		if docInfo.Size != -1 {
			return nil, fmt.Errorf("path %q must be a directory, not a file", p)
		}

		// 获取文件夹详细信息
		detail, err := c.getDocItemDetail(ctx, docInfo.DocID)
		if err != nil {
			return nil, err
		}

		// 将文件夹本身作为一个 resource
		meta := buildFilesetMeta(detail.DocID, detail.Name, detail.Path, detail.Type, detail.CreatedAt, detail.ModifiedAt, detail.CreatedBy, detail.ModifiedBy, detail.Rev)
		out = append(out, meta)
	}

	return out, nil
}
```

### 3.3.3 Fileset信息字段与Resource映射
AnyShare文件和文件夹信息与Vega Resource表的字段映射关系如下：

**Resource基础字段映射**：
- `source_identifier`: fileset的中文路径
- `name`: fileset名称（文件夹名称或文件名称）
- `database`: 置空

**source_metadata字段映射**：
```
source_metadata: {
  // fileset基本信息
  id: fileset ID（gns://...）
  name: fileset名称
  size: fileset大小（字节），文件夹为-1
  rev: fileset版本号
  type: fileset类型（file-文件，dir-文件夹）

  // 时间信息
  created_at: fileset创建时间（ISO8601格式）
  modified_at: fileset修改时间（ISO8601格式）

  // 用户信息
  created_by: {
    id: 创建者ID
    name: 创建者名称
    type: 创建者类型
  }
  modified_by: {
    id: 修改者ID
    name: 修改者名称
    type: 修改者类型
  }

  // 其他信息
  security_classification: 安全级别
  storage_name: 存储名称
  custom_metadata: 自定义元数据
}
```



**注意**：
1. source_metadata存储fileset的所有字段信息，方便后续使用
2. fileset只有基本信息，没有类似MySQL表的字段信息，因此SchemaDefinition不赋值

### 3.4 API接口封装
连接器需要封装以下AnyShare API接口：

#### 3.4.1 数据结构定义
```go
// tokenResponse Token响应
type tokenResponse struct {
	AccessToken string `json:"access_token"` // 访问令牌
	ExpiresIn   int    `json:"expires_in"`  // 过期时间（秒）
	Scope       string `json:"scope"`       // 权限范围
	TokenType   string `json:"token_type"`  // 令牌类型
	Error       string `json:"error"`       // 错误信息
}

// userInfoDTO 用户信息
type userInfoDTO struct {
	ID   string `json:"id"`   // 用户ID
	Name string `json:"name"` // 用户名称
	Type string `json:"type"` // 类型，user-用户
}

// docLibDTO 文档库信息
type docLibDTO struct {
	ID         string      `json:"id"`         // 文档库ID，格式为gns://...
	Name       string      `json:"name"`       // 文档库名称
	Type       string      `json:"type"`       // 类型，knowledge_doc_lib-知识库
	Attr       int         `json:"attr"`       // 属性
	Rev        string      `json:"rev"`        // 版本号
	CreatedAt  string      `json:"created_at"` // 创建时间
	ModifiedAt string      `json:"modified_at"` // 修改时间
	CreatedBy  userInfoDTO `json:"created_by"` // 创建者信息
	ModifiedBy userInfoDTO `json:"modified_by"` // 修改者信息
}

// docItemDetailDTO 文档详细信息
type docItemDetailDTO struct {
	ObjectId               string         `json:"object_id"`               // 对象ID
	Name                   string         `json:"name"`                   // 名称
	Rev                    string         `json:"rev"`                    // 版本号
	Size                   int64          `json:"size"`                   // 大小，文件夹为-1
	StorageName            string         `json:"storage_name"`           // 存储名称
	SecurityClassification int            `json:"security_classification"` // 安全分类
	CreatedAt              string         `json:"created_at"`             // 创建时间
	ModifiedAt             string         `json:"modified_at"`            // 修改时间
	CreatedBy              userInfoDTO    `json:"created_by"`            // 创建者信息
	ModifiedBy             userInfoDTO    `json:"modified_by"`           // 修改者信息
	CustomMetadata         map[string]any `json:"custom_metadata"`         // 自定义元数据
	DocLib                 docLibDTO      `json:"doc_lib"`                // 所属文档库信息
	Path                   string         `json:"path"`                   // 路径
	DocID                  string         `json:"doc_id"`                 // 文档ID
	Type                   string         `json:"type"`                   // 类型，dir-文件夹，file-文件
}

// pathInfoDTO 路径信息
type pathInfoDTO struct {
	DocID       string `json:"docid"`        // 文档ID
	Name        string `json:"name"`         // 名称
	Rev         string `json:"rev"`          // 版本号
	Size        int64  `json:"size"`         // 大小，文件夹为-1
	ClientMTime int64  `json:"client_mtime"` // 客户端修改时间
	Modified    int64  `json:"modified"`     // 修改时间
}
```

#### 3.4.2 接口说明

##### 1. 获取知识库信息

**接口详情**：
- URL: `{protocol}://{host}:{port}/api/document/v1/entry-doc-lib`
- Method: GET
- Authentication: Bearer Token

**请求参数**：

| 参数名   | 类型   | 必填 | 说明                                             |
|---------|--------|------|------------------------------------------------|
| type    | string | 否   | 文档库类型，knowledge_doc_lib-知识库，非必传，默认查所有文档库,当前固定传knowledge_doc_lib |
| sort    | string | 否   | 排序类型，doc_lib_name-文档库名称，非必传，默认doc_lib_name     |
| direction| string | 否   | 排序结果方向，asc升序desc降序，非必传，默认asc                   |

**说明**：
- 该接口只返回第一层文档库列表，不会包含文件
- 需要Bearer Token认证

**返回值示例**：
```json
[
  {
    "id": "gns://E6D15886F471458295D1F88EACB43FCE",
    "name": "Go to Market",
    "type": "knowledge_doc_lib",
    "attr": 84,
    "rev": "3a83b9c2-27ef-465e-9fb7-73f929047629",
    "created_at": "2014-04-09T08:01:02Z",
    "modified_at": "2026-04-01T09:05:59Z",
    "created_by": {
      "id": "266c6a42-6131-4d62-8f39-853e7093701c",
      "name": "admin",
      "type": "user"
    },
    "modified_by": {
      "id": "a7bd18fc-026c-11ed-88a6-ea60f4128d45",
      "name": "饶兰欣（Nancy）",
      "type": "user"
    }
  }
]
```

##### 2. 根据路径获取对象信息（getinfobypath）

**接口详情**：
- URL: `{protocol}://{host}:{port}/api/efast/v1/file/getinfobypath`
- Method: POST
- Authentication: Bearer Token
- Content-Type: application/json

**请求体**：
```json
{
  "namepath": "Go to Market"
}
```

**请求参数**：

| 参数名   | 类型   | 必填 | 说明             |
|---------|--------|------|------------------|
| namepath| string | 是   | 文档库名称路径 |

**说明**：
- 根据名称路径获取对象信息
- 需要Bearer Token认证

**返回值示例**：
```json
{
  "client_mtime": 1775034359702748,
  "docid": "gns://E6D15886F471458295D1F88EACB43FCE",
  "modified": 1775034359702748,
  "name": "Go to Market",
  "rev": "3a83b9c2-27ef-465e-9fb7-73f929047629",
  "size": -1
}
```

##### 3. 获取文档详细信息（items）

**接口详情**：
- URL: `{protocol}://{host}:{port}/api/efast/v2/items/{object_id}/all`
- Method: GET
- Authentication: Bearer Token

**路径参数**：

| 参数名    | 类型   | 必填 | 说明                                       |
|----------|--------|------|--------------------------------------------|
| object_id| string | 是   | 文档对象ID，即doc_id的最后一部分 |

**说明**：
- 当docInfo中的size!=-1时，表示当前路径是文档，使用此接口获取文档详细信息
- object_id是doc_id的最后一部分，例如doc_id为"gns://.../F6502AA7976545F5A1E59A2848AFA1E7"，则object_id为"F6502AA7976545F5A1E59A2848AFA1E7"
- 需要Bearer Token认证

**返回值示例**：
```json
{
  "storage_name": "obs-as7feijiegouzhengshihuanjing",
  "created_by": {
    "id": "040beac4-bef7-11e3-b845-dcd2fc061e41",
    "name": "胡佳妮（Kitty）",
    "type": "user"
  },
  "doc_lib": {
    "id": "gns://D2921C019C79486D9851060071FE7160",
    "name": "爱数赋能中心(AISHU Training)",
    "type": "knowledge_doc_lib"
  },
  "modified_at": "2024-10-16T09:48:41Z",
  "object_id": "F6502AA7976545F5A1E59A2848AFA1E7",
  "rev": "2E4775FFFD72432CB3C29BF3FEEFB7FA",
  "security_classification": 5,
  "doc_id": "gns://D2921C019C79486D9851060071FE7160/6AF681CC997D405E8720A19102447E4F/BC281439BC9140FCB45CE0B22B1F25C9/95C086A45489491B819EA27A1B4970FD/F0550C96F4CA4638AF68DDD6EE91FEAB/62BF3AD77C7641DBADDF4046D048A992/2BB7742D835747FE8FD498BDCB4ADD13/F6502AA7976545F5A1E59A2848AFA1E7",
  "path": "爱数赋能中心(AISHU Training)/1-学习平台资料（Learning Platform Materials）/00-机器人助手/01-销售赋能学院/01-销售管理/作战地图/东区作战地图/东区政府行业作战地图（含福建）.pptx",
  "size": 168979,
  "created_at": "2024-10-16T09:48:41Z",
  "custom_metadata": {
    "client_mtime": 1723430433477831
  },
  "modified_by": {
    "id": "040beac4-bef7-11e3-b845-dcd2fc061e41",
    "name": "胡佳妮（Kitty）",
    "type": "user"
  },
  "name": "东区政府行业作战地图（含福建）.pptx",
  "type": "file"
}
```

##### 4. 获取文档下载信息

**接口详情**：
- URL: `{protocol}://{host}:{port}/api/open-doc/v1/file-download`
- Method: POST
- Authentication: Bearer Token
- Content-Type: application/json

**请求体**：
```json
{
  "doc": [
    {"id": "gns://..."}
  ],
  "name": "下载文件名"
}
```

**请求参数**：

| 参数名 | 类型   | 必填 | 说明                     |
|--------|--------|------|--------------------------|
| doc    | array  | 是   | 文档对象数组，包含id字段 |
| name   | string | 是   | 下载时的打包名称       |

**说明**：
- 用于获取文档或文件夹的下载地址
- doc_id为gns://...格式的完整路径
- 需要Bearer Token认证

**返回值示例**：
```json
{
  "download_url": "https://...",
  "expires_in": 3600
}
```

**实现方法**：
```go
// GetObjectDownloadInfo returns download info for a given file or folder.
func (c *AnyShareConnector) GetObjectDownloadInfo(ctx context.Context, resourceName, docID string) (map[string]any, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	if docID == "" {
		return nil, fmt.Errorf("empty doc id")
	}
	u := fmt.Sprintf("%s/api/open-doc/v1/file-download", c.baseURL)
	// 当前 docID 为 gns://... 的形式，name为下载文件时的打包名称
	payload := map[string]interface{}{
		"doc": []map[string]string{
			{"id": docID},
		},
		"name": resourceName,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("file-download http %d: %s", resp.StatusCode, truncateForLog(raw))
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("file-download decode: %w", err)
	}
	if msg, ok := m["message"].(string); ok && msg != "" {
		return nil, fmt.Errorf("file-download: %s", msg)
	}
	return m, nil
}
```

## 4. 实现计划

### 4.1 第一阶段：连接器核心功能
1. 实现连接器基础框架
2. 实现认证模块
3. 实现知识库信息获取
4. 实现fileset列表获取
5. 实现文件夹查询

### 4.2 第二阶段：数据源核心功能
1. 扩展anyshare catalog
2. 扩展fileset resource采集功能
3. 扩展fileset resource管理



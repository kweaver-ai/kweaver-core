# AgentRespVO 模块文档（AI生成）

## 概述

`agentrespvo` 模块是 IntelliSearch 项目中负责处理智能体响应值对象（Agent Response Value Objects）的核心组件。该模块定义了一系列结构体和函数，用于表示和处理智能体返回的各种类型的响应数据。

## 目录结构

```
interfaces/value_object/agentrespvo/
├── answer_data.go        # 定义基本的回答数据结构
├── answer_explore.go     # 定义探索类型的回答结构
├── answer_text.go        # 定义文本类型的回答结构
├── doc_retrieval.go      # 定义文档检索结果结构
├── helpers.go            # 提供类型判断和验证的辅助函数
├── intervention.go       # 定义干预相关的结构
├── middle_output.go      # 定义中间输出变量相关结构
├── other.go              # 定义其他辅助结构（如时间状态、使用统计）
└── daresvo/              # 数据智能体响应子模块
    ├── answer.go         # 处理最终答案的函数
    ├── answer_explore.go # 处理探索类型答案的函数
    ├── answer_other.go   # 处理其他类型答案的函数
    ├── answer_prompt.go  # 处理提示类型答案的函数
    ├── cite.go           # 处理引用相关的函数
    ├── define.go         # 定义核心数据结构和常量
    ├── error.go          # 处理错误相关的函数
    ├── middle_output_vars.go # 处理中间输出变量的函数
    ├── other_vars.go     # 处理其他变量的函数
    ├── related_question.go # 处理相关问题的函数
    └── res_helper.go     # 提供响应处理的辅助函数
```

## 核心组件

### 1. 基础数据结构

#### AnswerS 结构体

`AnswerS` 是整个响应系统的核心数据结构，它包含了查询信息和干预数据，并支持动态字段。

```go
type AnswerS struct {
    Query string `json:"query"`
    Interventions Interventions `json:"interventions"` // 存储所有的中断相关信息
    utils.DynamicFieldsHolder
}
```

#### Intervention 结构体

`Intervention` 表示智能体执行过程中的干预信息，通常与工具调用相关。

```go
type Intervention struct {
    ToolName     string        `json:"tool_name"`
    ToolCallInfo *ToolCallInfo  `json:"tool_call_info"`
}

type ToolCallInfo struct {
    ToolName  string `json:"tool_name"`
    Args      interface{} `json:"args"`
    AssignType string `json:"assign_type"`
    OutputVar string  `json:"output_var"`
}
```

### 2. 文档检索结果

`DocRetrievalRes` 结构体表示文档检索的结果，包含文本内容和引用信息。

```go
type DocRetrievalRes struct {
    Text       string       `json:"text"`
    References []*Reference `json:"references"`
}

type Reference struct {
    Content            string  `json:"content"`
    RetrieveSourceType string  `json:"retrieve_source_type"`
    Score              float64 `json:"score"`
    Meta               *Meta   `json:"meta"`
}
```

### 3. 中间输出变量

`MiddleOutputVarRes` 结构体用于表示智能体处理过程中的中间输出变量。

```go
type MiddleOutputVarItem struct {
    VarName  string                    `json:"var_name"`
    Type     chatresenum.OutputVarType `json:"type"`
    Value    interface{}               `json:"value"`
    Thinking string                    `json:"thinking"`
}

type MiddleOutputVarRes struct {
    Vars []*MiddleOutputVarItem `json:"vars"`
}
```

### 4. 数据智能体响应 (daresvo)

`DataAgentRes` 是数据智能体的核心响应结构体，它整合了各种类型的响应数据。

```go
type DataAgentRes struct {
    Answer     *agentrespvo.AnswerS   `json:"answer"`
    UserDefine map[string]interface{} `json:"user_define,omitempty"`
    Ask        interface{}            `json:"ask"`
    Status     string                 `json:"status"`
    Error      interface{}            `json:"error"`

    // 各种响应辅助器
    finalAnswerVarHelper           *ResHelper
    docRetrievalVarHelper     *ResHelper
    graphRetrievalVarHelper   *ResHelper
    relatedQuestionsVarHelper *ResHelper
    otherVarsHelper           *ResHelper
    middleOutputVarsHelper    *ResHelper
}
```

## 主要功能

### 1. 类型验证和判断

`helpers.go` 文件提供了一系列函数，用于验证和判断不同类型的响应数据：

- `IsPromptTypeInterface` / `IsPromptType`：判断对象是否为提示类型
- `IsExploreTypeInterface` / `IsExploreType`：判断对象是否为探索类型

这些函数使用 JSON Schema 进行验证，确保数据符合预期的结构。

### 2. 响应处理

`ResHelper` 类提供了一系列方法，用于处理和转换不同类型的响应数据：

- 获取字段值：`getFieldVal`
- 获取 JSON 表示：`GetJSON`
- 获取类型信息：`GetType`
- 按路径获取对象值：`GetObjValByPath`

### 3. 干预管理

`Interventions` 类型提供了对干预信息的管理功能，特别是将干预信息按输出变量进行分组：

```go
func (is Interventions) ToOutputVarMap() (outputVarMap map[string][]*Intervention) {
    outputVarMap = make(map[string][]*Intervention)
    for _, intervention := range is {
        if intervention.ToolCallInfo == nil {
            continue
        }
        outputVar := intervention.ToolCallInfo.OutputVar
        if outputVar == "" {
            continue
        }
        outputVarMap[outputVar] = append(outputVarMap[outputVar], intervention)
    }
    return
}
```

### 4. 文档检索和引用

`DocRetrievalRes` 提供了获取答案文本和引用信息的方法：

```go
func (r *DocRetrievalRes) AnswerAndCites() (answer string, cites []*interfaces.AnswerCite) {
    answer = r.Text
    for _, reference := range r.References {
        cites = append(cites, &interfaces.AnswerCite{
            Content:  reference.Content,
            Meta:     reference.Meta,
            CiteType: reference.RetrieveSourceType,
            Score:    reference.Score,
        })
    }
    return
}
```

## 数据流程

下图展示了 `agentrespvo` 模块中数据的主要流程：

```
+----------------+     +------------------+     +------------------+
|                |     |                  |     |                  |
|  客户端请求    +---->+  智能体处理过程    +---->+  AnswerS 结构体  |
|                |     |                  |     |                  |
+----------------+     +------------------+     +-------+----------+
                                                        |
                                                        v
+----------------+     +------------------+     +------------------+
|                |     |                  |     |                  |
|  最终响应      +<----+  DataAgentRes    +<----+  各种类型的      |
|                |     |                  |     |  响应数据        |
+----------------+     +------------------+     +------------------+
```

## 关键接口和用法

### 创建新的回答

```go
answer := agentrespvo.NewAnswerS()
answer.Query = "用户查询"
```

### 处理干预信息

```go
interventions := agentrespvo.NewInterventions()
// 添加干预信息...
outputVarMap := interventions.ToOutputVarMap()
```

### 处理文档检索结果

```go
docRetrieval := &agentrespvo.DocRetrievalRes{
    Text: "检索到的文本",
    References: []*agentrespvo.Reference{
        // 引用信息...
    },
}
answer, cites := docRetrieval.AnswerAndCites()
```

### 使用数据智能体响应

```go
dataAgentRes := daresvo.NewDataAgentRes(responseData, outputConfig)
```

## 总结

`agentrespvo` 模块是 IntelliSearch 项目中负责处理智能体响应的核心组件，它提供了一系列结构体和函数，用于表示和处理不同类型的响应数据。该模块的设计允许灵活地处理各种类型的响应，并支持动态字段和干预信息的管理。

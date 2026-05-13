# DolphinLanguage 0.3.4语法规范

## **目录**
1. [变量](#1-变量)
2. [控制流](#2-控制流)
3. [工具调用](#3-工具调用)
4. [自然语言指令](#4-自然语言指令)
5. [注释和文档](#5-注释和文档)
6. [输出格式控制](#6-输出格式控制)
7. [高级语法特性](#7-高级语法特性)
8. [最佳实践建议](#8-最佳实践建议)
9. [实际应用示例](#9-实际应用示例)
10. [核心特性总结](#10-核心特性总结)

---

## **1. 变量**

### **1.1 定义变量**

- 赋值：`表达式` ` ->` `变量名`
- 追加：`表达式` ` >>` `变量名`

**示例**：

```
@_search(query="天气") -> result
@_search(query="天气") >> result
"Hello World" -> greeting
["第一项", "第二项", "第三项"] -> test_array
{"user": {"name": "张三", "age": 25}} -> user_data
```

### **1.2 使用变量**

变量以 `$` 开头，支持以下形式：

- 简单变量：`$变量名`
- 数组索引：`$变量名[index]`
- 嵌套属性：`$变量名.key1.key2`

**示例**：

```
$x
$result[0]
$a.b.c
$test_array[0]  # 数组索引访问
$user_data.user.name  # 嵌套对象属性访问
$user_data.user.profile.city  # 深层嵌套访问
```

### **1.3 变量运算**

支持简单的字符串拼接和基本运算：

```
$greeting + " from Dolphin Language" -> finalMessage
```
- 示例

  ```
  "abc" -> str1
  $str1+"de" -> str2
  ```
  
------

## **2. 控制流**

### **2.1 循环**

- 语法

  ```
  /for/ $变量名 in $可迭代对象:
      语句块
  /end/
  ```

- 示例

  ```
  ["rag原理","大模型实战"] -> x
  /for/ $text in $x:
      @_search(query=$text) -> docs
      总结一下$docs >> summary_list
  /end/
  ```

### **2.2 条件判断**

- 语法

  ```
  /if/ 条件表达式:
      语句块
  elif 条件表达式:
      语句块
  else:
      语句块
  /end/
  ```

- 示例

  ```
  @_search(query="智能体 编程") -> results
  /if/ len($results) > 10:
      @_write_file(file_path="log.txt", content="结果较多") -> alert_result
  elif len($results) > 5:
      @_write_file(file_path="log.txt", content="结果适中") -> alert_result
  else:
      @_write_file(file_path="log.txt", content="结果较少") -> alert_result
  /end/
  ```


------

## **3. 工具调用**

使用 `@工具名(参数列表)` 调用函数，结果通过 `->` 或 `>>` 赋值。

- 语法

  ```
  @工具名(参数列表) -> 变量名
  @工具名(参数列表) >> 变量名
  ```

- 参数格式
  - **命名参数**：`key=value` 格式，如 `query="搜索内容"`
  - **位置参数**：直接传值，如 `$x` 或 `"value"`
  - **变量引用**：使用 `$变量名` 引用已定义的变量
  - **复杂类型**：支持字典、列表等复杂数据结构

- 参数类型处理
  - 字符串：使用单引号或双引号包围，如 `"text"` 或 `'text'`
  - 数字：直接写数值，如 `42` 或 `3.14`
  - 布尔值：`true` 或 `false`
  - 列表：使用方括号，如 `['财经', '科技', '其它']`
  - 字典：使用大括号，如 `{"key": "value"}`
  - 变量：以 `$` 开头，如 `$query`

- 示例（以下示例全部使用当前SDK内置或已实现的工具名）

  ```
  # 命名参数调用
  @_search(query=$x) -> result
  @getDataSourcesFromConcepts(conceptNames=["Sales", "Inventory"]) -> datasources

  # 位置参数调用
  @_search($x) -> tag_result

  # 混合参数
  @executeSQL(datasource="my_db", sql=$query, timeout=30) -> data

  # 复杂参数
  @_write_jsonl(file_path="data/output.jsonl", content=[{"k":1},{"k":2}]) -> saved
  ```

------





## **4. 自然语言指令**

自然语言内容通常作为 `/prompt/`、`/judge/`、`/explore/` 等代码块的主体文本存在，配合 `->` 或 `>>` 将结果写入变量。

- 示例

  ```
  /prompt/(model="deepseek-v3")
  请用一句话总结以下文本的主题：
  $text -> summary
  ```
  ```
  /prompt/
  请用一句话总结以下文本的主题：
  $text -> summary
  ```

说明：独立行的自然语言不单独构成可执行单元，必须出现在上述代码块中。


### **4.1 /prompt/ 代码块**

用于直接调用LLM进行对话生成，支持多种参数配置。

- 语法

  ```
  /prompt/(参数列表) 提示内容 -> 变量名
  ```

- 支持的参数
  - `model`: 指定使用的模型
  - `system_prompt`: 系统提示词
  - `output`: 输出格式（"json", "jsonl", "list_str"）
  - `history`: 是否使用历史对话（布尔值，默认false）
  - `no_cache`: 是否禁用缓存（布尔值，默认false）

**注意**：`/prompt/` 块**不支持** `tools` 参数。如需工具调用能力，请使用 `/judge/` 或 `/explore/` 块。

- 示例

  ```
  /prompt/(model="v3", output="list_str") 根据问题描述返回相关概念 -> concepts
  /prompt/(system_prompt="你是一个AI助手", model="qwen-plus") 创作一首诗 -> poem
  /prompt/(output="json") 生成用户信息 -> user_info
  /prompt/(output="jsonl") 生成三个用户记录 -> users
  /prompt/(history=true, model="v3") 继续上次的对话 -> response
  ```

### **4.2 /explore/ 代码块**

用于智能体探索和工具调用，支持多步推理。Agent会自动选择和调用合适的工具，通过多轮交互完成复杂任务。

- 语法

  ```
  /explore/(参数列表) 任务描述 -> 变量名
  ```

- 支持的参数
  - `tools`: 可用工具列表，支持带引号和不带引号的格式（例如 `["_bash", "_python"]` 或 `[_bash, _python]`）
  - `model`: 指定使用的模型
  - `system_prompt`: 系统提示词
  - `history`: 是否使用历史对话（布尔值）
  - `no_cache`: 是否禁用缓存（布尔值）

- 工具调用格式：Agent在推理过程中使用 `=>#tool_name: {"param": "value"}` 格式调用工具

- 返回类型：`Dict[str, Any]` - 返回包含推理过程和最终答案的字典，通常包含"think"和"answer"字段

- 特性
  - 支持多轮工具调用，自动进行推理和决策
  - 自动检测重复工具调用并给出提示
  - 支持工具调用中断和恢复
  - 支持代码块作为工具参数

- 示例

  ```
  /explore/(tools=[executeSQL, _python, _search], model="v3") 解决数据分析问题 -> result
  /explore/(tools=[_search, _python], model="v3") 搜索并分析信息 -> analysis
  /explore/(tools=["_bash", "_python"], system_prompt="你是数据分析专家") 分析系统日志 -> report
  ```

### **4.3 /judge/ 代码块**

用于判断任务并根据判断结果决定是否调用工具。Judge块会分析任务需求，如果有合适的工具则调用，否则使用LLM直接回答。

- 语法

  ```
  /judge/(参数列表) 判断内容 -> 变量名
  ```

- 支持的参数
  - `system_prompt`: 系统提示词
  - `model`: 指定使用的模型
  - `tools`: 可用工具列表
  - `history`: 是否使用历史对话（布尔值）
  - `prompt_skillcall`: 判断模式选择（"true"|"false"，默认"true"）
    - "true": 使用自定义prompt进行工具判断
    - "false": 使用LLM的function calling能力进行工具判断

注意：当前实现中，参数解析会将 `prompt_skillcall` 规范化为布尔值；同时判断逻辑以字符串比较为准（与 "true" 比较）。为获得预期行为：
- 省略该参数或仅在需要 function-calling 判断时显式传 `prompt_skillcall="false"`；
- 不建议显式传 `prompt_skillcall="true"`，以避免与内部判断逻辑的类型差异造成歧义。

- 返回类型：`Dict[str, Any]` - 返回工具调用结果或LLM直接回答

- 工作流程
  1. 分析用户问题的具体需求
  2. 检查是否有工具可以满足这个需求
  3. 如果有合适的工具，选择并调用工具
  4. 如果没有合适的工具，使用LLM直接回答

- 示例

  ```
  /judge/(system_prompt="", model="qwen-plus", tools=[]) 总结以上内容 -> summary
  /judge/(tools=[_search, _python], prompt_skillcall="false") 回答用户问题 -> answer
  /judge/(tools=[executeSQL], model="v3") 处理数据查询请求 -> result
  ```

------

## **5. 注释和文档**

### **5.1 行注释**

使用 `#` 开头的行为注释行，会被解析器忽略。

```
# 这是一个注释
@_search(query="天气") -> result
```

### **5.2 文档注释**

使用 `@DESC` 标记来添加文档说明。

```
@DESC 
记忆压缩Agent：从用户的最近N天记忆中提取关键知识点并保存
@DESC
```

### **5.3 多行字符串文档**

使用三重引号定义多行字符串，常用于提示和规则定义。

```
'''
1. 不同年份的销售数据在不同表中
2. 计算类型任务可以使用 Python 代码
3. SQL 执行结果为空时需要检查字段名
''' -> hints
```

------

## **6. 输出格式控制**

通过 `output` 参数可以控制LLM输出的数据格式，系统会自动解析并验证输出结果。

### **6.1 JSON 格式**

返回单个JSON对象，适用于需要结构化数据的场景。

```
/prompt/(output="json") 生成用户信息 -> user_info
```

**返回类型：** `Dict[str, Any]` - 返回单个JSON对象，包含键值对数据

**示例输出：**
```json
{"name": "张三", "age": 25, "city": "北京"}
```

### **6.2 JSONL 格式**

返回多个JSON对象组成的列表，适用于批量数据处理。

```
/prompt/(output="jsonl") 生成多个用户记录 -> users_list
```

**返回类型：** `List[Dict[str, Any]]` - 返回JSON对象的列表，每个元素都是一个字典

**示例输出：**
```
[
  {"name": "张三", "age": 25},
  {"name": "李四", "age": 30},
  {"name": "王五", "age": 28}
]
```

### **6.3 列表字符串格式**

返回字符串列表，适用于简单的文本列表。

```
/prompt/(output="list_str") 返回概念名称列表 -> concept_names
```

**返回类型：** `List[str]` - 返回字符串列表，元素类型为字符串

**示例输出：**
```
["概念1", "概念2", "概念3"]
```

### **6.4 输出格式的应用**

- **数据提取**：从非结构化文本中提取结构化数据
- **批量处理**：使用 jsonl 格式批量生成数据
- **简化处理**：使用 list_str 快速获取文本列表

------

## **7. 高级语法特性**

### **7.1 工具参数传递**

支持在工具调用中使用变量作为参数，包括简单变量、复杂对象和嵌套属性：

```
@_search(query=$query) -> search_results
@getAllConcepts() -> allConcepts
/prompt/(model="v3", output="list_str") 从以下概念中选出与任务相关的概念名称：
```
$allConcepts
```
-> conceptInfos
@getDataSourcesFromConcepts(conceptNames=$conceptInfos) -> datasources
```

### **7.2 多行提示模板**

支持复杂的多行提示模板，包含变量插值和代码块引用：

```
今天是【$date】
请一步步思考，使用工具及进行推理计算，得到最后的结果。

要解决的问题
```
$query
```

datasource:
```
$datasources
```

现在请开始： -> result
```

### **7.3 内置函数和工具调用**

支持调用内置函数获取系统信息或执行特定操作：

**系统函数：**
```
@_date() -> date
@_write_jsonl(file_path="data/file.jsonl", content=$data) -> outputPath
@_search(query=$search_query) -> search_results
```

### **7.4 复杂数据结构处理**

支持处理复杂的嵌套数据结构，包括多层嵌套的对象和数组：

```
# 定义复杂对象
{"user": {"name": "张三", "profile": {"age": 25, "city": "北京"}}} -> user_data

# 访问嵌套属性
$user_data.user.name -> user_name
$user_data.user.profile.city -> user_city

# 数组访问
["第一项", "第二项", "第三项"] -> items
$items[0] -> first_item
```

### **7.5 工具选择控制**

在 `/judge/` 块中，可以使用 `tool_choice` 参数（当 `prompt_skillcall="false"` 时）控制LLM如何选择工具：

```
# 自动选择（默认）
/judge/(tools=[_search, _python], prompt_skillcall="false", tool_choice="auto") 任务描述 -> result

# 必须使用工具
/judge/(tools=[_search, _python], prompt_skillcall="false", tool_choice="required") 任务描述 -> result

# 不使用工具
/judge/(tools=[_search, _python], prompt_skillcall="false", tool_choice="none") 任务描述 -> result
```

**说明**：
- `tool_choice` 仅在 `/judge/` 块且 `prompt_skillcall="false"` 时生效
- `/prompt/` 块不支持 tools 和 tool_choice 参数
- `/explore/` 块有自己的工具调用机制，不使用 tool_choice

### **7.6 缓存控制**

使用 `no_cache` 参数控制是否使用缓存：

```
# 禁用缓存，确保获取最新结果
/prompt/(no_cache=true) 获取最新数据 -> fresh_data

# 使用缓存（默认）
/prompt/(no_cache=false) 查询数据 -> cached_data
```

### ** 7.7 历史对话管理**

使用 `history` 参数控制是否使用历史对话上下文：

```
# 使用历史对话
/prompt/(history=true) 继续刚才的话题 -> response

# 不使用历史对话（默认）
/prompt/(history=false) 新的问题 -> answer
```

### **7.8 工具参数格式灵活性**

工具列表支持多种格式，提高代码可读性：

```
# 带引号格式（推荐）
/explore/(tools=["executeSQL", "_python"]) 分析数据 -> result

# 不带引号格式
/explore/(tools=[executeSQL, _python]) 分析数据 -> result

# 单个工具
/explore/(tools=executeSQL) 执行查询 -> result
```

------


## **8. 最佳实践建议**

### **8.1 变量命名**

- 使用有意义的变量名，清晰表达变量用途
- 使用下划线分隔多个单词，如 `user_data`、`search_results`
- 避免使用保留字和系统变量名
- 对于临时变量，使用简短但清晰的名称

### **8.2 代码组织**

- 适当使用注释说明复杂逻辑和业务规则
- 使用 `@DESC` 为文件添加文档，说明Agent的用途
- 将相关操作组织在一起，保持代码结构清晰
- 合理使用空行分隔不同的逻辑块

### **8.3 工具调用**

- 在 `/explore/` 块中明确指定需要的工具，避免提供不必要的工具
- 使用合适的模型版本，根据任务复杂度选择模型
- 为复杂任务提供详细的提示和上下文信息
- 使用 `system_prompt` 明确Agent的角色和任务要求

### **8.4 输出格式选择**

- 根据数据用途选择合适的输出格式
- 使用 `json` 格式处理结构化数据
- 使用 `jsonl` 格式处理批量数据
- 使用 `list_str` 格式处理简单列表
- 使用 `obj/TypeName` 格式确保类型安全

### **8.5 参数传递**

- 优先使用命名参数，提高代码可读性
- 对于复杂参数，考虑先赋值给变量再传递
- 使用变量引用时确保变量已定义
- 注意参数类型的正确性（字符串、数字、布尔值等）


### **8.6 性能优化**

- 合理使用 `no_cache` 参数，平衡性能和数据新鲜度
- 避免在循环中进行重复的LLM调用
- 对于大量数据处理，考虑使用批处理
- 使用 `history` 参数时注意上下文长度

------

## **9. 实际应用示例**

### **9.1 数据分析任务**

```
@DESC 
ChatBI数据探索Agent：使用SQL和Python工具进行数据分析
@DESC

@_date() -> date
@getAllConcepts() -> allConcepts

/prompt/(model="v3", output="list_str")根据要解决的问题与概念描述，返回可能需要用到的概念名称列表

要解决的问题：
```
$query
```

概念描述：
```
$allConcepts
```
-> conceptInfos

@getDataSourcesFromConcepts(conceptNames=$conceptInfos) -> datasources

/explore/(tools=[executeSQL, _python, _search], model="v3")
今天是【$date】
请一步步思考，使用工具及进行推理计算，得到最后的结果。

要解决的问题
```
$query
```

datasource:
```
$datasources
```

请注意!
(1)不要使用假设的数据进行计算，假设的数据对我毫无意义。
(2)sql 语句撰写后就请立即执行
(3)如果工具执行出现错误，请调整工具参数并重新进行生成和执行。

现在请开始： -> result
```

### **9.2 记忆压缩任务**

```
@DESC 
记忆压缩Agent：从用户的最近N天记忆中提取关键知识点并保存
@DESC

# 读取已保存的记忆数据（按需自定义路径或筛选）
@_mem_view(path="") -> memoryData

#从记忆数据中提取10条最重要的知识点
/prompt/(model="v3", output="jsonl")根据以下用户的记忆数据，提取出最重要的10条知识点。

用户记忆数据：
```
$memoryData
```

请从这些记忆中提取最重要的10条知识点，每条知识点应该：
1. 是具体的、有价值的信息
2. 对用户未来的决策或行为有帮助
3. 具有一定的普遍性或重要性

输出格式为JSON数组，每个元素包含以下字段：
- content: 知识点内容（字符串）
- score: 重要性评分（1-100整数）

请开始：-> knowledgePoints

#将压缩后的知识写入knowledge文件
@_write_jsonl(file_path="data/memory/user_chatbi_user/knowledge.jsonl", content=$knowledgePoints) -> outputPath
```

### **9.3 简单搜索任务**

```
/explore/(tools=[_search, _python], model="v3")
今天是 2025.7.1，请一步步思考，使用工具及进行推理计算，得到最后的结果
任务是：$query
现在请开始： -> result
```

### **9.4 概念提取和数据获取**

```
@getAllConcepts() -> allConcepts

/prompt/(model="v3", output="list_str")根据要解决问题的描述以及 concepts 描述，返回所有可能要用到的概念名称

要解决的问题：
```
$query
```

concepts 描述：
```
$allConcepts
```

只输出概念名称列表 -> conceptInfos

@getSampleData(conceptNames=$conceptInfos) -> sampledData
@getDataSourceSchemas(conceptNames=$conceptInfos) -> schemas
```
### **9.5 json数据生成和获取**

```
/prompt/(output="json") 随机生成一段用户信息，结构为json，例如`{"name":"张三丰","age":30}` -> user_info
$user_info['answer'] -> object_str
json.loads($object_str)["name"] -> name
/prompt/(output="jsonl") 随机生成一个用户信息列表，结构为array[object],输出结构样例为：`[{"name":"张三丰","age":30},[{"name":"朱元璋","age":26}]]`，你要严格按照样例格式进行输出 -> user_info_list
$user_info_list['answer'] -> jsonl_str
json.loads($jsonl_str)[0] -> column
/prompt/(output="list_str") 随机生成一段数字列表，结构为array[str],输出结构样例为：`[1,2,3,4]`，你要严格按照样例格式进行输出 -> num_list
$num_list['answer'] -> list_str
json.loads($list_str) -> list_num
$list_num[0] -> num
```

### **9.6 撰写任务**

```
/prompt/(flags='{{"debug": true}}')请将以下原始query改写成多个更适合搜索引擎搜索的问题。改写时请注意以下几点：
1. 确保每个改写后的问题简洁明了，包含核心关键词。
2. 针对不同角度或细分领域生成问题，以覆盖更广泛的搜索结果。
3. 避免使用模糊或过于宽泛的表述，尽量具体化。
4. 如果原始问题涉及复杂概念，可以拆解为多个简单问题。
5. 生成的数量控制在3-5个之间。
6. 改写后的问题以列表格式返回。
示例：
原始问题: 如何提高英语口语能力？
参考资料: 英语口语学习需要大量练习和适当方法，包括听力训练、模仿发音、参加语言交换、使用学习应用等。初学者和进阶学习者有不同的学习重点。
改写后的问题：
["提高英语口语能力的最佳方法","适合初学者的英语口语练习技巧","英语听力训练对口语提升的作用","语言交换平台推荐及使用方法","进阶学习者如何突破英语口语瓶颈"]
直接生成改写后的问题list不要生成其他内容和解释。
原始问题: $query
改写后的问题：->sub_querys
eval($sub_querys.answer)->search_querys
#只取前2个 方便联调，实际数量不确定,可以设置成
$search_querys[:2]->search_querys
''->ref
/for/ $search_query in $search_querys:
  @zhipu_search_tool(query=$search_query)->result
  $result['answer']['choices'][0]['message']['tool_calls'] -> result
  /if/ len($result[1]['search_result']) > 0:
    $result[1]['search_result']->search_result
  else:
    []->search_result
  /end/
  $search_result>>search_results
  ''->sub_ref
  /for/ $page in $search_result:
      $sub_ref+$page['content']->sub_ref
  /end/
  $sub_ref[:5000] -> sub_ref
  $ref+$sub_ref ->ref
/end/
''->page
''->sub_ref
''->result
''->search_result
''->sub_querys
''->search_query

/if/ $ref == '':
    /prompt/请忽略其他提示词，直接输出：没有找到相关资料  -> answer
else:
    /prompt/(history=True)请根据问题和参考资料完成撰写任务。
    <参考资料> $ref </参考资料>
    <任务> $query </任务>
    要求：
    1. 尽可能详细

    撰写内容:->answer
/end/
```

---

## **10. 核心特性总结**

DolphinLanguage 是一种专为AI Agent开发设计的领域特定语言（DSL），具有以下核心特性：

### **10.1 语言设计理念**

- **声明式编程**：专注于描述"做什么"而非"怎么做"
- **自然语言友好**：支持自然语言与结构化代码的无缝融合
- **工具驱动**：基于工具调用的Agent编程范式
- **类型灵活**：支持多种数据类型和输出格式

### **10.2 主要优势**

1. **简洁表达**
   - 使用 `->` 和 `>>` 进行简洁的变量赋值
   - 支持变量插值和复杂数据结构访问
   - 简化的函数调用语法

2. **智能推理**
   - `/explore/` 块支持多轮工具调用和自动推理
   - `/judge/` 块智能判断是否需要调用工具
   - 自动检测重复调用和错误处理

3. **灵活控制**
   - 支持条件、循环等控制流结构
   - 丰富的参数配置选项

4. **输出格式化**
   - 多种输出格式支持（JSON、JSONL、List、Object）
   - 自动解析和验证输出结果
   - 支持自定义类型定义

5. **开发友好**
   - 完整的语法验证机制
   - 清晰的错误提示
   - 丰富的文档和示例

### **10.3 适用场景**

- **数据分析**：结合SQL和Python工具进行复杂数据分析
- **信息检索**：使用搜索工具进行信息查询和总结
- **知识管理**：实现记忆压缩、知识提取等任务
- **业务自动化**：编排多个工具完成复杂业务流程
- **对话系统**：构建具有工具调用能力的智能对话Agent

### **10.4 扩展性**

- **工具集成**：轻松集成新的工具和技能
- **类型系统**：支持自定义对象类型
- **配置灵活**：丰富的参数配置支持各种场景
- **模块化**：支持函数封装和代码重用

---

*本文档基于最新的SDK实现更新，全面反映了 Dolphin Language 的语法特性和实际能力。文档涵盖了从基础语法到高级特性的完整内容，并提供了丰富的实际应用示例。*

*最后更新：2025年11月28日*

# DolphinLanguage语法规范

## 目录
1. [变量](#1)
2. [控制流](#2)
3. [函数调用](#3)
4. [自然语言指令](#4)
5. [嵌入代码块](#5)
6. [特殊指令](#6)
7. [函数封装](#7)
8. [Agent初始化](#8-agent)
9. [特殊代码块](#9)
10. [注释和文档](#10)
11. [输出格式控制](#11)
12. [高级语法特性](#12)
13. [最佳实践](#13)
14. [实际应用示例](#14)

---

## 1. 变量

### **1.1 定义变量**

- 赋值：`表达式` ` ->` `变量名`
- 追加：`表达式` ` >>` `变量名`

**示例**：

```
@web_search(query="天气") -> result
@web_search(query="天气") >> result
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

------

## 2. 控制流

### **2.1 循环**

- 语法

  ```
  /for/ $变量名 in $可迭代对象:
      语句块
  /end/
  ```

- 示例

  ```
  /for/ $text in $x:
      @web_search(query=$text) -> news
      总结一下$news >> summary_list
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
  @fetch_shopping_list(user_id="12345") -> shopping_list
  /if/ len($shopping_list) > 10:
      @send_alert("购物清单太长了，需要精简一下！") -> alert_result
  elif len($shopping_list) > 5:
      @send_alert("购物清单长度适中，可以直接执行采购。") -> alert_result
  else:
      @send_alert("购物清单很短，轻松完成！") -> alert_result
  /end/
  ```

------

## 3. 函数调用

- 使用 `@函数名(参数列表)` 调用函数，结果通过 `->` 或 `>>` 赋值。(目前参数列表仅支持关键字参数)

- 语法

  ```
  @函数名(参数1, 参数2, ...) -> 变量名
  ```

- 示例

  ```
  @news_search(query=$x) -> result
  @llm_tagger($x, ['财经', '科技', '其它']) -> tag_result
  @concepts(query=$query) -> conceptDescs
  @getDataSourcesFromConcepts(conceptNames=$conceptInfos) -> datasources
  ```

------

## 4. 自然语言指令

### **4.1 基本语法**

在 Dolphin Language 中，自然语言指令可以单独使用或与其他语法元素结合：

```
计算 5 + 3 的结果
-> result
```

### **4.2 与变量结合**

```
$result + 10
-> new_result
```

### **4.3 多行指令**

```
第一步：获取用户输入
第二步：处理数据
第三步：返回结果
-> task_result
```

------

## 5. 嵌入代码块

### **5.1 Python 代码嵌入**

使用三重引号嵌入 Python 代码：

```
'''
import json
data = {"key": "value"}
print(json.dumps(data))
''' -> code_result
```

### **5.2 SQL 代码嵌入**

```
'''
SELECT * FROM users WHERE age > 18;
''' -> query_result
```

------

## 6. 特殊指令

### **6.1 @DESC 指令**

用于添加文档说明：

```
@DESC
这是一个数据处理Agent，用于清洗和转换数据
@DESC
```

### **6.2 @ASSIGN 指令**

用于变量赋值：

```
@ASSIGN($source_data, "processed") -> data
```

### **6.3 @RUN 指令**

用于执行代码块：

```
@RUN($python_code) -> execution_result
```

------

## 7. 函数封装

### **7.1 定义函数**

虽然 Dolphin Language 主要用于工作流编排，但支持代码块封装：

```
/def/ process_data(input_data):
    @clean_data($input_data) -> cleaned
    @transform_data($cleaned) -> transformed
    return $transformed
/end/
```

### **7.2 调用封装函数**

```
@process_data($raw_data) -> processed_result
```

### **7.3 内置函数**

Dolphin Language 提供多种内置函数：

```
@_date() -> current_date
@_time() -> current_time
@_random() -> random_value
```

------

## 8. Agent初始化

### **8.1 使用 @DESC 初始化智能体**

```
@DESC
这是一个专业的代码审查Agent，负责检查代码质量和安全漏洞
@DESC
```

### **8.2 配置智能体参数**

```
@DESC
名称：数据分析师
角色：分析数据趋势
工具：SQL, Python, 可视化
@DESC
```

### **8.3 智能体间通信**

```
@send_message(agent="analyzer", message=$analysis_request) -> response
```

------

## 9. 特殊代码块

### **9.1 /prompt/ 代码块**

用于直接调用LLM进行对话生成，支持多种参数配置。

- 语法

  ```
  /prompt/(参数列表) 提示内容 -> 变量名
  ```

- 支持的参数
  - `model`: 指定使用的模型
  - `system_prompt`: 系统提示词
  - `output`: 输出格式（"json", "jsonl", "list_str", "obj/ObjectType"）

- 示例

  ```
  /prompt/(model="v3", output="list_str") 根据问题描述返回相关概念 -> concepts
  /prompt/(system_prompt="你是一个AI助手", model="qwen-plus") 创作一首诗 -> poem
  /prompt/(output="json") 生成用户信息 -> user_info
  /prompt/(output="jsonl") 生成三个用户记录 -> users
  ```

### **9.2 /explore/ 代码块**

用于智能体探索和工具调用，支持多步推理。

- 语法

  ```
  /explore/(参数列表) 任务描述 -> 变量名
  ```

- 支持的参数
  - `tools`: 可用工具列表（支持通配符匹配：fnmatch/glob 语法 `*`、`?`、`[abc]`；也支持按 skillkit 命名空间：`<skillkit>.<pattern>`，如 `resource_skillkit.*`）
  - `model`: 指定使用的模型
  - `system_prompt`: 系统提示词
  - `enable_skill_deduplicator`: 是否启用技能调用去重器，默认为 `true`。当设置为 `false` 时，将关闭技能去重逻辑（仅受最大调用次数限制）。

- 返回类型：`Dict[str, Any]` - 返回包含推理过程和最终答案的字典，通常包含"think"和"answer"字段

- 示例

  ```
  /explore/(tools=[executeSQL, _python, _search], model="v3") 解决数据分析问题 -> result
  /explore/(tools=[_search, _python], model="v3") 搜索并分析信息 -> analysis
  /explore/(tools=[resource_skillkit.*]) 仅暴露 ResourceSkillkit 工具 -> result
  ```

### **9.3 /judge/ 代码块**

用于判断和评估任务。

- 语法

  ```
  /judge/(参数列表) 判断内容 -> 变量名
  ```

- 支持的参数
  - `system_prompt`: 系统提示词
  - `model`: 指定使用的模型
  - `tools`: 可用工具列表

- 返回类型：`Dict[str, Any]` - 返回包含判断结果的字典，通常包含工具调用信息或评估结果

- 示例

  ```
  /judge/(system_prompt="", model="qwen-plus", tools=[]) 总结以上内容 -> summary
  ```

------

## 10. 注释和文档

### **10.1 行注释**

使用 `#` 开头的行为注释行，会被解析器忽略。

```
# 这是一个注释
@web_search(query="天气") -> result
```

### **10.2 文档注释**

使用 `@DESC` 标记来添加文档说明。

```
@DESC 
记忆压缩Agent：从用户的最近N天记忆中提取关键知识点并保存
@DESC
```

### **10.3 多行字符串文档**

使用三重引号定义多行字符串，常用于提示和规则定义。

```
'''
1. 不同年份的销售数据在不同表中
2. 计算类型任务可以使用 Python 代码
3. SQL 执行结果为空时需要检查字段名
''' -> hints
```

------

## 11. 输出格式控制

### **11.1 JSON 格式**

```
/prompt/(output="json") 生成用户信息 -> user_info
```

**返回类型：** `Dict[str, Any]` - 返回单个JSON对象，包含键值对数据

### **11.2 JSONL 格式**

```
/prompt/(output="jsonl") 生成多个用户记录 -> users_list
```

**返回类型：** `List[Dict[str, Any]]` - 返回JSON对象的列表，每个元素都是一个字典

### **11.3 列表字符串格式**

```
/prompt/(output="list_str") 返回概念名称列表 -> concept_names
```

**返回类型：** `List[str]` - 返回字符串列表，元素类型为字符串

### **11.4 对象类型格式**

```
/prompt/(output="obj/UserProfile") 生成用户档案 -> profile
```

**返回类型：** `Dict[str, Any]` - 返回符合指定对象类型定义的字典结构

------

## 12. 高级语法特性

### **12.1 函数参数传递**

支持在函数调用中使用变量作为参数：

```
@concepts(query=$query) -> conceptDescs
@getDataSourcesFromConcepts(conceptNames=$conceptInfos) -> datasources
@getSampleData(conceptNames=$conceptInfos) -> sampledData
```

### **12.2 多行提示模板**

支持复杂的多行提示模板，包含变量插值：

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

### **12.3 内置函数调用**

支持调用内置函数获取系统信息：

```
@_date() -> date
@_write_jsonl(file_path="data/file.jsonl", content=$data) -> outputPath
```

### **12.4 复杂数据结构处理**

支持处理复杂的嵌套数据结构：

```
# 定义复杂对象
{"user": {"name": "张三", "profile": {"age": 25, "city": "北京"}}} -> user_data

# 访问嵌套属性
$user_data.user.name -> user_name
$user_data.user.profile.city -> user_city
```

------

## 13. 最佳实践

### **13.1 变量命名**

- 使用有意义的变量名
- 使用下划线分隔多个单词
- 避免使用保留字

### **13.2 代码组织**

- 适当使用注释说明复杂逻辑
- 使用 `@DESC` 为文件添加文档
- 将相关操作组织在一起

### **13.3 工具调用**

- 在 `/explore/` 块中明确指定需要的工具
- 使用合适的模型版本
- 为复杂任务提供详细的提示

### **13.4 错误处理**

- 确保所有表达式都有输出变量（使用 `->` 或 `>>`）
- 检查变量是否存在再使用
- 适当处理可能的空值情况

------

## 14. 实际应用示例

### **14.1 数据分析任务**

```
@DESC 
ChatBI数据探索Agent：使用SQL和Python工具进行数据分析
@DESC

@_date() -> date
@hints() -> hints
@concepts(query=$query) -> conceptDescs
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

关于任务的提示：
```
$taskHints
```

请注意!
(1)不要使用假设的数据进行计算，假设的数据对我毫无意义。
(2)sql 语句撰写后就请立即执行
(3)如果工具执行出现错误，请调整工具参数并重新进行生成和执行。

现在请开始： -> result
```

### **14.2 记忆压缩任务**

```
@DESC 
记忆压缩Agent：从用户的最近N天记忆中提取关键知识点并保存
@DESC

#读取用户最近N天的记忆数据  
@readRecentMemoryData(userId="chatbi_user", recentDays=$recent_days) -> memoryData

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

### **14.3 简单搜索任务**

```
/explore/(tools=[_search, _python], model="v3")
今天是 2025.7.1，请一步步思考，使用工具及进行推理计算，得到最后的结果
任务是：$query
现在请开始： -> result
```

### **14.4 概念提取和数据获取**

```
@getAllConcepts() -> allConcepts

/prompt/(model="v3", output="list_str")根据要解决问题的描述以及concepts描述，返回所【有可能】要用到的 concepts 信息

要解决的问题：
```
$query
```

concepts描述：
```
$allConcepts
``` 

关于任务的提示：
```
$taskHints
```

要解决的问题：
```
$query
```

只输出 concept 名称，请开始 -> conceptInfos

@getSampleData(conceptNames=$conceptInfos) -> sampledData
@getDataSourceSchemas(conceptNames=$conceptInfos) -> schemas
```

---
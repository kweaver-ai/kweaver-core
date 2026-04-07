# 测评相关 Adapter 配置指南
本文档详细说明两类核心场景下 Adapter 的功能、使用规则、配置方法及实操示例，包括**测评规则中的 Adapter**与**效果测评任务中的 Adapter**，同时补充外部接入场景的 Adapter 编写规范。


## 测评规则中的 Adapter
测评规则配置中的 Adapter 核心作用是**将数据集 Output 转换为指标 Input**，其编写格式由指标的 Input 要求决定。系统对部分常规场景内置适配逻辑，特殊场景需手动修改代码。

### 使用说明
| 数据集文件数量 | 输出（Output）数量 | 是否需要手动配置 | 场景说明 |
| --- | --- | --- | --- |
| 1 个文件 | 1 个 Output | ❌ 无需配置 | 系统内置适配逻辑，直接使用 |
| 1 个文件 | 多个 Output | ✅ 需手动配置 | 需修改 Adapter 代码以匹配多输出需求 |
| 2 个及以上文件（表头相同） | 1 个 Output | ❌ 无需配置 | 系统自动合并相同表头数据 |
| 2 个及以上文件（表头相同） | 多个 Output | ✅ 需手动配置 | 需自定义数据拆分逻辑 |
| 2 个及以上文件（表头不同） | 1 个 / 多个 Output | ✅ 需手动配置 | 需处理表头差异，确保数据格式统一 |

### 注意事项
#### Adapter 上传规则
- 每个测评数据集仅支持上传 **一个 Adapter**；
- 测评数据集的不同版本在系统中视为**不同的数据集**，各版本可分别配置 Adapter。

#### 内置 Adapter 行为说明
- 使用**大模型**进行评测时，Adapter 会将数据**统一转换为字符串格式**；
- 使用**小模型**进行评测时，Adapter 会**保留数据的原始格式**；
- 系统提供的 Adapter 示例模板默认将数据自动转换为字符串，如需调整格式需手动修改（见下文操作步骤）。

### 自定义格式配置
保留数据原始格式（关闭字符串转换）：若需保留数据原始格式（如数字、列表等非字符串类型），按以下步骤操作：
1. 下载当前 Adapter 文件，找到并删除以下代码片段：

   ```Plain
   if not isinstance(info[column_name], str):    
       dictInfo[column_name] = json.dumps(info[column_name], ensure_ascii=False)
   else:    
       dictInfo[column_name] = info[column_name]  
   ```

2. 替换为：

   ```Plain
   dictInfo[column_name] = info[column_name]
   ```

3. 保存文件并重新上传。

调整数据结构（取消键值对结构）：内置 Adapter 默认将输入数据转换为**键值对（Key-Value）结构**，若需纯文本等非键值对格式（如仅传入字段值），按以下步骤修改：
1. 下载当前 Adapter 文件，找到并删除以下代码片段：

   ```Plain
   line_data = {}
   for column_name, value in line.items():   
       # 若列名符合input字段，包装为dict类型传递至算法   
       if column_name in input:     
           line_data[column_name] = value
   ```

2. 替换为（以纯文本为例）：

   ```Plain
   line_data = ''
   for key, value in line.items():   
       if key in input:      
           line_data = value  # 直接提取字段值，不保留键名
   ```

3. 保存文件并重新上传。

> **提示**：修改 Adapter 后，需确保其输入格式与评测规则及模型输入要求一致，避免评测任务执行失败。

### 实操场景示例
场景 1：1 个数据文件 + 1 个 Output（无需配置）
- **数据集示例（data_llm_V1.0.csv）**
  | Query | Positive Document |
  | --- | --- |
  | 查找欧洲杯信息 | 欧洲杯 2022 年 6 月 11 日 - 7 月 11 日举办，24 支队伍参赛... |
  | 高温环境头发护理 | 高温高湿环境下建议使用天然洗发水，定期做发膜护理... |

- **指标配置（long_bench）**  
  评测长文本理解能力，计算 rouge 得分，参数要求如下：
  | 类型 | 参数名称 | 参数说明 |
  | --- | --- | --- |
  | Inputs | correct_answers | 正确答案，格式为 list [str] |
  | Inputs | answer | 算法输出答案，格式为 list [str] |
  | Outputs | score | 评测得分（rouge 指标） |

- **内置 Adapter 输出结果**

  ```Plain
  [
      "欧洲杯2022年6月11日-7月11日举办，24支队伍参赛...",
      "高温高湿环境下建议使用天然洗发水，定期做发膜护理..."
  ]
  ```

场景 2：1 个数据文件 + 2 个 Output（需手动配置）
- **数据集示例（data_llm_V1.0.csv）**
  | Query | Positive Document | Hard Negative Document |
  | --- | --- | --- |
  | 查找欧洲杯信息 | 欧洲杯 2022 年举办... | 篮球世界杯赛程信息... |
  | 高温环境头发护理 | 高温高湿环境头发护理建议... | 电脑强光护眼措施... |

- **指标配置（long_bench，Input 格式为 list [dict]）**  
  需手动修改 Adapter 代码，步骤如下：
  1. 进入测评规则配置页面，选择数据集 “data_llm_V1.0” 和指标 “long_bench”；
  2. 点击【Adapter】→【下载示例模板】，找到以下代码片段：

     ```Plain
     for line in content:    
         # 处理每行数据（默认单Output逻辑）    
         for o in output:        
             data_to_metric.append(line[o])
     ```

  3. 替换为多 Output 处理逻辑：

     ```Plain
     for line in content:    
         line_data = {}    
         for column_name, value in line.items():        
             # 仅保留Output所需字段，包装为dict    
             if column_name in output:           
                 line_data[column_name] = value    
         data_to_metric.append(line_data)
     ```

  4. 保存文件并上传，完成配置。

- **最终 Adapter 输出结果**

  ```Plain
  [
      {
          "Positive Document": "欧洲杯2022年举办...",
          "Hard Negative Document": "篮球世界杯赛程信息..."
      },
      {
          "Positive Document": "高温高湿环境头发护理建议...",
          "Hard Negative Document": "电脑强光护眼措施..."
      }
  ]
  ```

- **额外注意事项**
  ○ 若下载示例模板后需适配自定义数据集/指标，需手动将模板中的“test123_V1_0”替换为数据集名称、“ToMetricName”替换为指标名称；
  ○ 若数据集名称含中文、英文、数字以外的字符，系统会自动替换为“_”，示例函数名格式如下：

    ```Plain
    def data_llm_V1_0DsTolong_benchMetric_run_func(inputs, props, resource, data_source_config):
        """
        data_llm/V1.0 数据集至 long_bench metric 的 adapter 的核心执行逻辑
    ```


## 效果测评任务中的 Adapter
效果测评任务中的 Adapter 核心作用是**将数据集 Input 转换为算法 Input**，编写格式由算法的 Input 要求决定（支持算法类型：提示词 + 大模型、小模型、自定义应用、外部接入 API）。

### 使用说明
| 算法类型 | 配置场景 | 是否需要手动配置 |
| --- | --- | --- |
| 提示词 + 大模型 | 数据集 Input 与提示词参数**完全一致** | ❌ 无需配置 |
| 提示词 + 大模型 | 数据集 Input 与提示词参数**不一致** | ✅ 需手动配置（字段映射） |
| 小模型 / 自定义应用 / 外部接入 | 数据集仅含**1 个 Input 字段** | 可能需配置（按算法格式要求） |

### 内置 Adapter 行为说明
- 内置 Adapter 默认将所有输入数据**转换为字符串格式**；
- 内置 Adapter 默认将数据**组织为键值对（Key-Value）结构**后传入算法；
- 若算法对输入格式有特殊要求（如保留原始格式、非键值对结构），可参考以下步骤修改。

### 自定义格式配置
保留数据原始格式（关闭字符串转换）：
1. 下载当前 **Adapter 示例** 文件；
2. 找到并删除以下代码片段：

   ```Plain
   if not isinstance (info[column_name], str):
       dictInfo[column_name] = json.dumps(info[column_name], ensure_ascii=False)
   else:
       dictInfo[column_name] = info[column_name]  
   ```

3. 替换为：

   ```Plain
   dictInfo[column_name] = info[column_name]
   ```

取消键值对结构（非 Key-Value 输入）：若算法仅接收纯文本等非键值对输入，按以下步骤修改：
1. 下载当前 **Adapter 示例** 文件；
2. 找到并删除以下代码片段：

   ```Plain
   line_data = {}
   for column_name, value in line.items():
       # 如果列名符合input字段，则将该列数据包装成dict类型传递至算法
       if column_name in input:
         line_data[column_name] = value
   ```

3. 替换为（以纯文本为例）：

   ```Plain
   line_data = ''
   for key, value in line.items():
       if key in input:
          line_data = value
   ```

### 实操场景示例
场景 1：提示词 + 大模型（参数一致，无需配置）
- **数据集示例（test.csv）**
  | language | content |
  | --- | --- |
  | 中文 | apple |
  | 中文 | banana |
  | 英文 | 苹果 |
  | 英文 | 香蕉 |

- **提示词配置**

  ```Plain
  让你来充当翻译家，你的目标是把任何语言翻译成{{language}},请翻译时不要带翻译腔，而是要翻译的自然、流畅和地道，使用优美和高雅的表达方式。请翻译下面这句话{{content}}
  ```

- **内置 Adapter 输出结果**

  ```Plain
  [
      {
          "language": "中文",
          "content": "apple"
      },
      {
          "language": "中文",
          "content": "banana"
      },
      {
          "language": "英文",
          "content": "苹果"
      },  
      {
          "language": "英文",
          "content": "香蕉"
      }
  ]
  ```

场景 2：提示词 + 大模型（参数不一致，需配置）
- **数据集示例（翻译数据.csv）**  
  数据集字段为“语言”“内容”，与提示词参数“language”“content”不一致，需字段映射：
  | 语言 | 内容 | answer |
  | --- | --- | --- |
  | 中文 | apple | 苹果 |
  | 英文 | 香蕉 | banana |

- **提示词配置（参数为 language 和 content）**

  ```Plain
  我让你来充当翻译家，你的目标是把任何语言翻译成{{language}},请翻译时不要带翻译腔，而是要翻译的自然、流畅和地道，使用优美和高雅的表达方式。请翻译下面这句话{{content}}
  ```

- **Adapter 配置步骤**
  1. 进入效果测评任务页面，选择算法 “翻译器 + AISHU READER”；
  2. 点击【Adapter】→【下载示例模板】，找到以下代码片段：

     ```Plain
     for line in content:    
         line_data = {}    
         for column_name, value in line.items():        
             if column_name in input:           
                 line_data[column_name] = value    
         data_to_algorithm.append(line_data)
     ```

  3. 替换为字段映射逻辑（“语言”→“language”，“内容”→“content”）：

     ```Plain
     for line in content:    
         line_data = {}    
         for column_name, value in line.items():        
             if column_name == "语言":           
                 line_data["language"] = value  # 字段名映射        
             if column_name == "内容":           
                 line_data["content"] = value    # 字段名映射    
         data_to_algorithm.append(line_data)
     ```

  4. 保存文件并上传，完成配置。

- **最终 Adapter 输出结果**

  ```Plain
  [
      {
          "language": "中文",
          "content": "apple"
      },
      {
          "language": "英文",
          "content": "香蕉"
      }
  ]
  ```


## 效果测评任务中“外部接入”的 Adapter 编写
选择“外部接入”算法类型时，可通过 API 接口评测未集成到系统的算法（如私有化模型、第三方 API），需自定义 Adapter 适配 API 格式。以下为典型场景示例及规范。

### OpenAI 风格 API 的 Adapter 示例
#### 数据集示例（jsonl 格式）

```Plain
{"input": "欧洲杯赛程", "keypoints": "2022年6月11日-7月11日，24支队伍", "context": "欧洲杯2022年举办信息...", "model": "gpt-3.5-turbo"}
{"input": "头发护理方法", "keypoints": "天然洗发水、发膜护理", "context": "高温环境头发护理建议...", "model": "gpt-3.5-turbo"}
```

#### Adapter 核心逻辑（完整示例）

```Plain
def testqa_V1_0DsToMyAPI_run_func(inputs, props, resource, data_source_config):
    """
    功能：将testqa/V1.0数据集转换为MyAPI（OpenAI风格）的请求格式
    参数：
        inputs: 输入数据（数据集内容）
        props: 配置参数
        resource: 资源引用
        data_source_config: 数据源配置
    返回：
        data_to_algorithm: API请求数据列表
        data_to_metric: 指标输入数据列表（预留）
    """
    # 1. 初始化输出列表
    data_to_algorithm = []
    data_to_metric = []
    
    # 2. 读取数据集内容（假设通过_get_content函数获取）
    files = inputs.get(\"files\", [])
    for file in files:
        doc_name = file[\"doc_name\"]
        content = _get_content(doc_name)  # 自定义函数：读取jsonl文件内容
        
        # 3. 处理每行数据，构造API请求格式
        for line in content:
            api_request = {
                \"model\": line.get(\"model\"),  # 从数据集获取模型名称
                \"messages\": [
                    {\"role\": \"system\", \"content\": \"你是AISHU-AI-BOT，专注长文本理解，仅基于参考文档回答\"},
                    {\"role\": \"user\", \"content\": f\"参考文档：{line.get('context')}\\n问题：{line.get('input')}？\"}
                ],
                \"temperature\": 0,
                \"max_tokens\": 1000
            }
            data_to_algorithm.append(api_request)
            
            # 4. 预留指标输入（ keypoints 作为标准结果）
            data_to_metric.append(line.get(\"keypoints\"))
    
    return {\"output\": data_to_algorithm}, {\"output\": data_to_metric}

# 定义Adapter执行器类（需与数据集、算法名称匹配）
class testqa_V1_0DsToMyAPIAlgorithmAdapter(object):
    cls_type = \"Executor\"
    INPUT_TYPE = {\"input\": list[dict]}  # 输入格式：数据集内容列表
    OUTPUT_TYPE = {\"output\": list[dict]}  # 输出格式：API请求列表
    DEFAULT_PROPS = {\"delimiter\": \",\"}
    RUN_FUNC = testqa_V1_0DsToMyAPI_run_func  # 绑定核心执行函数
```

### 流式 API 返回结果的处理
若 API 返回流式数据（如逐字返回文本），需在 Adapter 中处理数据拼接或字段提取，以下为两种常见场景：

场景 1：拼接流式结果（去除状态信息）
- **API 流式输出示例**：

  ```Plain
  one_data = [\"你\", \"好\", \"！\", \"--info--{\\\"time\\\": \\\"0.3s\\\"}\", \"--end--\"]
  ```

- **处理逻辑代码**：

  ```Plain
  def process_streaming_result(algorithm_output):
      data_to_metric = []
      for one_data in algorithm_output:
          # 去除最后2个状态字段，拼接文本内容
          pure_text = \"\".join(one_data[:-2])
          data_to_metric.append(pure_text)
      return data_to_metric
  ```

场景 2：提取流式结果中的指定字段
- **API 流式输出示例**（最后一条数据为完整结果，含目标字段）：

  ```Plain
  one_data = [
      \"正在生成答案...\",
      \"{\\\"result\\\": {\\\"answer\\\": \\\"欧洲杯2022年举办\\\"}, \\\"status\\\": \\\"done\\\"}\"
  ]
  ```

- **处理逻辑代码**（提取最后一条数据中的 "answer" 字段）：

  ```Plain
  import json
  def extract_streaming_field(algorithm_output):
      data_to_metric = []
      for one_data in algorithm_output:
          # 解析最后一条数据（含完整结果）中的answer字段
          last_item = one_data[-1]
          result_dict = json.loads(last_item)
          answer = result_dict[\"result\"][\"answer\"]
          data_to_metric.append(answer)
      return data_to_metric
  ```


### 类名与函数名规范
Adapter 的类名和函数名需严格遵循固定格式，确保系统能正确识别并执行转换逻辑，具体规则如下：

| 转换场景 | 函数名格式 | 类名格式 |
| --- | --- | --- |
| 数据集→算法 | {数据集名称}DsTo{算法名称}_run_func | {数据集名称}DsTo{算法名称}AlgorithmAdapter |
| 算法→指标 | {算法名称}To{指标名称}_run_func | {算法名称}To{指标名称}MetricAdapter |


#### 规范示例
- 场景：数据集 “testqa_V1.0” 转换至算法 “MyAPI”  
  函数名：testqa_V1_0DsToMyAPI_run_func  
  类名：testqa_V1_0DsToMyAPIAlgorithmAdapter  

- 场景：算法 “MyAPI” 转换至指标 “keyword_constraint”  
  函数名：MyAPITokeyword_constraint_run_func  
  类名：MyAPITokeyword_constraintMetricAdapter

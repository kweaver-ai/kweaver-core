# Adapter Configuration Guide for Evaluation

This document explains the function, usage rules, configuration methods, and practical examples of Adapter in two core scenarios: **Adapter in evaluation rules** and **Adapter in effect evaluation tasks**. It also includes writing conventions for Adapter in external integration scenarios.

## Adapter in Evaluation Rules

In evaluation rule configuration, the core purpose of Adapter is to **convert dataset outputs into indicator inputs**. Its implementation format depends on the input requirements of the target indicator. The system provides built-in adaptation logic for some common scenarios, while special scenarios require manual code modification.

### Usage Rules

| Number of Dataset Files | Number of Outputs | Manual Configuration Required | Description |
| --- | --- | --- | --- |
| 1 file | 1 output | No | The system uses built-in adaptation logic directly |
| 1 file | Multiple outputs | Yes | Adapter code must be modified to support multiple outputs |
| 2 or more files with the same header | 1 output | No | The system merges files with identical headers automatically |
| 2 or more files with the same header | Multiple outputs | Yes | Custom data-splitting logic is required |
| 2 or more files with different headers | 1 or more outputs | Yes | Header differences must be handled so the data format is unified |

### Notes

#### Adapter Upload Rules

- Each evaluation dataset supports only **one Adapter**
- Different versions of the same evaluation dataset are treated as **different datasets**, and each version can have its own Adapter

#### Built-in Adapter Behavior

- When a **large model** is used for evaluation, the Adapter converts the data into **string format**
- When a **small model** is used, the Adapter **keeps the original data format**
- The example template provided by the system converts data to strings by default. If you need another format, modify the code manually

### Custom Format Configuration

#### Keep the Original Data Format

If you want to preserve the original data type such as numbers or lists instead of converting everything to strings:

1. Download the current Adapter file and remove the following code:

   ```Plain
   if not isinstance(info[column_name], str):    
       dictInfo[column_name] = json.dumps(info[column_name], ensure_ascii=False)
   else:    
       dictInfo[column_name] = info[column_name]  
   ```

2. Replace it with:

   ```Plain
   dictInfo[column_name] = info[column_name]
   ```

3. Save the file and upload it again.

#### Change the Data Structure

The built-in Adapter converts input data into a **key-value structure** by default. If you need plain text or another non-key-value format:

1. Download the current Adapter file and remove the following code:

   ```Plain
   line_data = {}
   for column_name, value in line.items():   
       # If the column name matches an input field, wrap it as dict data for the algorithm   
       if column_name in input:     
           line_data[column_name] = value
   ```

2. Replace it with:

   ```Plain
   line_data = ''
   for key, value in line.items():   
       if key in input:      
           line_data = value
   ```

3. Save the file and upload it again.

> **Tip**: After modifying an Adapter, make sure its input format matches the requirements of the evaluation rule and the model, otherwise the evaluation task may fail.

### Practical Examples

#### Scenario 1: One Data File + One Output, No Configuration Required

- **Dataset example (`data_llm_V1.0.csv`)**

  | Query | Positive Document |
  | --- | --- |
  | Find information about the UEFA European Championship | The European Championship was held from June 11 to July 11, 2022, with 24 participating teams... |
  | Hair care in high-temperature environments | In hot and humid environments, it is recommended to use natural shampoo and regular hair-mask care... |

- **Indicator configuration (`long_bench`)**

  | Type | Parameter Name | Description |
  | --- | --- | --- |
  | Inputs | correct_answers | Correct answers in `list[str]` format |
  | Inputs | answer | Algorithm output answers in `list[str]` format |
  | Outputs | score | Evaluation score, such as a ROUGE metric |

- **Built-in Adapter output**

  ```Plain
  [
      "The European Championship was held from June 11 to July 11, 2022, with 24 participating teams...",
      "In hot and humid environments, it is recommended to use natural shampoo and regular hair-mask care..."
  ]
  ```

#### Scenario 2: One Data File + Two Outputs, Manual Configuration Required

- **Dataset example (`data_llm_V1.0.csv`)**

  | Query | Positive Document | Hard Negative Document |
  | --- | --- | --- |
  | Find information about the UEFA European Championship | The European Championship was held in 2022... | Basketball World Cup schedule information... |
  | Hair care in high-temperature environments | Hair care advice for hot and humid environments... | Eye protection measures for strong computer light... |

- **Indicator configuration (`long_bench`, input type `list[dict]`)**

  1. Open the evaluation rule configuration page and select dataset `data_llm_V1.0` and indicator `long_bench`
  2. Click **Adapter > Download Example Template** and find:

     ```Plain
     for line in content:    
         # Process each line with the default single-output logic    
         for o in output:        
             data_to_metric.append(line[o])
     ```

  3. Replace it with multi-output logic:

     ```Plain
     for line in content:    
         line_data = {}    
         for column_name, value in line.items():        
             if column_name in output:           
                 line_data[column_name] = value    
         data_to_metric.append(line_data)
     ```

  4. Save and upload the file

- **Final Adapter output**

  ```Plain
  [
      {
          "Positive Document": "The European Championship was held in 2022...",
          "Hard Negative Document": "Basketball World Cup schedule information..."
      },
      {
          "Positive Document": "Hair care advice for hot and humid environments...",
          "Hard Negative Document": "Eye protection measures for strong computer light..."
      }
  ]
  ```

- **Additional notes**
  - If you adapt the template to a custom dataset or indicator, replace `test123_V1_0` with the dataset name and `ToMetricName` with the indicator name
  - If the dataset name contains characters other than Chinese, English letters, or digits, the system automatically replaces them with `_`

  ```Plain
  def data_llm_V1_0DsTolong_benchMetric_run_func(inputs, props, resource, data_source_config):
      """
      Core execution logic for the adapter from data_llm/V1.0 dataset to long_bench metric
  ```

## Adapter in Effect Evaluation Tasks

In effect evaluation tasks, the core purpose of Adapter is to **convert dataset inputs into algorithm inputs**. The implementation format depends on the input requirements of the target algorithm. Supported algorithm types include prompt plus large model, small model, custom application, and external API.

### Usage Rules

| Algorithm Type | Configuration Scenario | Manual Configuration Required |
| --- | --- | --- |
| Prompt + large model | Dataset inputs are exactly the same as prompt parameters | No |
| Prompt + large model | Dataset inputs are different from prompt parameters | Yes, field mapping is required |
| Small model / custom application / external integration | Dataset contains only one input field | Configuration may be required depending on the algorithm format |

### Built-in Adapter Behavior

- The built-in Adapter converts all input data to **string format** by default
- The built-in Adapter organizes data into a **key-value structure** by default
- If the algorithm needs a different format, such as raw types or plain text, modify the Adapter accordingly

### Custom Format Configuration

#### Keep the Original Data Format

1. Download the current **Adapter example** file
2. Remove the following code:

   ```Plain
   if not isinstance (info[column_name], str):
       dictInfo[column_name] = json.dumps(info[column_name], ensure_ascii=False)
   else:
       dictInfo[column_name] = info[column_name]  
   ```

3. Replace it with:

   ```Plain
   dictInfo[column_name] = info[column_name]
   ```

#### Cancel the Key-Value Structure

If the algorithm accepts only plain text or another non-key-value input:

1. Download the current **Adapter example** file
2. Remove the following code:

   ```Plain
   line_data = {}
   for column_name, value in line.items():
       # If the column name matches the input field, wrap it as dict data for the algorithm
       if column_name in input:
         line_data[column_name] = value
   ```

3. Replace it with:

   ```Plain
   line_data = ''
   for key, value in line.items():
       if key in input:
          line_data = value
   ```

### Practical Examples

#### Scenario 1: Prompt + Large Model, Parameters Match, No Configuration Required

- **Dataset example (`test.csv`)**

  | language | content |
  | --- | --- |
  | Chinese | apple |
  | Chinese | banana |
  | English | apple |
  | English | banana |

- **Prompt**

  ```Plain
  Act as a translator. Your goal is to translate any language into {{language}}. The translation should sound natural, fluent, and idiomatic rather than literal. Please translate the following sentence: {{content}}
  ```

- **Built-in Adapter output**

  ```Plain
  [
      {
          "language": "Chinese",
          "content": "apple"
      },
      {
          "language": "Chinese",
          "content": "banana"
      },
      {
          "language": "English",
          "content": "apple"
      },  
      {
          "language": "English",
          "content": "banana"
      }
  ]
  ```

#### Scenario 2: Prompt + Large Model, Parameters Do Not Match, Configuration Required

- **Dataset example (`translation_data.csv`)**

  | language | content | answer |
  | --- | --- | --- |
  | Chinese | apple | apple |
  | English | banana | banana |

- **Prompt configuration** with parameters `language` and `content`

  ```Plain
  Act as a translator. Your goal is to translate any language into {{language}}. The translation should sound natural, fluent, and idiomatic rather than literal. Please translate the following sentence: {{content}}
  ```

- **Adapter configuration steps**
  1. Open the effect evaluation task page and choose the algorithm `Translator + AISHU READER`
  2. Click **Adapter > Download Example Template** and find:

     ```Plain
     for line in content:    
         line_data = {}    
         for column_name, value in line.items():        
             if column_name in input:           
                 line_data[column_name] = value    
         data_to_algorithm.append(line_data)
     ```

  3. Replace it with field mapping logic:

     ```Plain
     for line in content:    
         line_data = {}    
         for column_name, value in line.items():        
             if column_name == "language":           
                 line_data["language"] = value        
             if column_name == "content":           
                 line_data["content"] = value    
         data_to_algorithm.append(line_data)
     ```

  4. Save the file and upload it

- **Final Adapter output**

  ```Plain
  [
      {
          "language": "Chinese",
          "content": "apple"
      },
      {
          "language": "English",
          "content": "banana"
      }
  ]
  ```

## Writing Adapter for External Integration in Effect Evaluation Tasks

When the algorithm type is **External Integration**, algorithms that are not built into the system, such as private models or third-party APIs, can be evaluated through API interfaces. In this case, a custom Adapter is needed to adapt the API format.

### Example Adapter for an OpenAI-Style API

#### Dataset Example in `jsonl`

```Plain
{"input": "UEFA European Championship schedule", "keypoints": "June 11 to July 11, 2022, 24 teams", "context": "Information about the 2022 UEFA European Championship...", "model": "gpt-3.5-turbo"}
{"input": "Hair care methods", "keypoints": "natural shampoo, hair-mask care", "context": "Hair care advice in hot environments...", "model": "gpt-3.5-turbo"}
```

#### Core Adapter Logic

```Plain
def testqa_V1_0DsToMyAPI_run_func(inputs, props, resource, data_source_config):
    """
    Function: convert the testqa/V1.0 dataset into the request format required by MyAPI
    """
    data_to_algorithm = []
    data_to_metric = []
    
    files = inputs.get(\"files\", [])
    for file in files:
        doc_name = file[\"doc_name\"]
        content = _get_content(doc_name)
        
        for line in content:
            api_request = {
                \"model\": line.get(\"model\"),
                \"messages\": [
                    {\"role\": \"system\", \"content\": \"You are AISHU-AI-BOT. Focus on long-context understanding and answer only based on the reference document.\"},
                    {\"role\": \"user\", \"content\": f\"Reference document: {line.get('context')}\\nQuestion: {line.get('input')}?\"}
                ],
                \"temperature\": 0,
                \"max_tokens\": 1000
            }
            data_to_algorithm.append(api_request)
            data_to_metric.append(line.get(\"keypoints\"))
    
    return {\"output\": data_to_algorithm}, {\"output\": data_to_metric}

class testqa_V1_0DsToMyAPIAlgorithmAdapter(object):
    cls_type = \"Executor\"
    INPUT_TYPE = {\"input\": list[dict]}
    OUTPUT_TYPE = {\"output\": list[dict]}
    DEFAULT_PROPS = {\"delimiter\": \",\"}
    RUN_FUNC = testqa_V1_0DsToMyAPI_run_func
```

### Handling Streaming API Results

If the API returns streaming data, the Adapter may need to concatenate fragments or extract specific fields.

#### Scenario 1: Concatenate Streaming Text and Remove Status Data

- **Streaming output example**

  ```Plain
  one_data = [\"H\", \"i\", \"!\", \"--info--{\\\"time\\\": \\\"0.3s\\\"}\", \"--end--\"]
  ```

- **Processing logic**

  ```Plain
  def process_streaming_result(algorithm_output):
      data_to_metric = []
      for one_data in algorithm_output:
          pure_text = \"\".join(one_data[:-2])
          data_to_metric.append(pure_text)
      return data_to_metric
  ```

#### Scenario 2: Extract a Specific Field from the Streaming Result

- **Streaming output example**

  ```Plain
  one_data = [
      \"Generating answer...\",
      \"{\\\"result\\\": {\\\"answer\\\": \\\"The UEFA European Championship was held in 2022\\\"}, \\\"status\\\": \\\"done\\\"}\"
  ]
  ```

- **Processing logic**

  ```Plain
  import json
  def extract_streaming_field(algorithm_output):
      data_to_metric = []
      for one_data in algorithm_output:
          last_item = one_data[-1]
          result_dict = json.loads(last_item)
          answer = result_dict[\"result\"][\"answer\"]
          data_to_metric.append(answer)
      return data_to_metric
  ```

### Naming Rules for Class and Function Names

Adapter class names and function names must follow a fixed format so that the system can recognize and execute the transformation logic correctly:

| Transformation Scenario | Function Name Format | Class Name Format |
| --- | --- | --- |
| Dataset -> Algorithm | `{dataset_name}DsTo{algorithm_name}_run_func` | `{dataset_name}DsTo{algorithm_name}AlgorithmAdapter` |
| Algorithm -> Indicator | `{algorithm_name}To{indicator_name}_run_func` | `{algorithm_name}To{indicator_name}MetricAdapter` |

#### Examples

- Dataset `testqa_V1.0` to algorithm `MyAPI`  
  Function name: `testqa_V1_0DsToMyAPI_run_func`  
  Class name: `testqa_V1_0DsToMyAPIAlgorithmAdapter`

- Algorithm `MyAPI` to indicator `keyword_constraint`  
  Function name: `MyAPITokeyword_constraint_run_func`  
  Class name: `MyAPITokeyword_constraintMetricAdapter`

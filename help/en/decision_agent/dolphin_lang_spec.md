# DolphinLanguage 0.3.4 Syntax Specification

## Variables

### Define Variables

- Assignment: `expression -> variable_name`
- Append: `expression >> variable_name`

Examples:

```python
# Search weather information and assign the result to result
@_search(query="weather") -> result
# Search weather information and append the result to result
@_search(query="weather") >> result
# Assign a string to greeting
"Hello World" -> greeting
# Assign an array to test_array
["first item", "second item", "third item"] -> test_array
# Assign a nested dictionary to user_data
{"user": {"name": "Zhang San", "age": 25}} -> user_data
```

### Use Variables

Variables start with `$` and support the following forms:

- Simple variable: `$variable_name`
- Array index: `$variable_name[index]`
- Nested property: `$variable_name.key1.key2`

Examples:

```python
# Examples of simple variables, array indexes, and nested property access
$x $result[0] $a.b.c $test_array[0]
# Array index access
$user_data.user.name
# Deep nested access
$user_data.user.profile.city
```

### Variable Operations

Simple string concatenation and basic operations are supported:

```python
# Concatenate greeting with a string and assign the result to finalMessage
$greeting + " from Dolphin Language" -> finalMessage
```

Example:

```python
# Define the string variable str1
"abc" -> str1
# Concatenate the string and assign the result to str2
$str1 + "de" -> str2
```

## Control Flow

### Loop

Syntax:

```python
/for/ $variable_name in $iterable:
    statements
/end/
```

Example:

```python
# Define array x
["rag principles", "large model practice"] -> x
# Iterate over each element in x
/for/ $text in $x:
    # Search the current element and assign the result to docs
    @_search(query=$text) -> docs
    # Summarize the search results and append to summary_list
    Summarize $docs >> summary_list
/end/
```

### Conditional Judgment

Syntax:

```python
/if/ condition_expression:
    statements
elif condition_expression:
    statements
else:
    statements
/end/
```

Example:

```python
# Search for "agent programming" and assign the result to results
@_search(query="agent programming") -> results
# Make a decision according to the number of search results
/if/ len($results) > 10:
    # Write to the log when there are many results
    @_write_file(file_path="log.txt", content="many results") -> alert_result
elif len($results) > 5:
    # Write to the log when there is a moderate number of results
    @_write_file(file_path="log.txt", content="moderate results") -> alert_result
else:
    # Write to the log when there are few results
    @_write_file(file_path="log.txt", content="few results") -> alert_result
/end/
```

## Tool Invocation

Use `@tool_name(parameter_list)` to call a function, and assign the result with `->` or `>>`.

Syntax:

```python
# Assign the tool-call result to a variable
@tool_name(parameter_list) -> variable_name
# Append the tool-call result to a variable
@tool_name(parameter_list) >> variable_name
```

### Parameter Format

- Named parameter: `key=value`, such as `query="search content"`
- Positional parameter: pass a value directly, such as `$x` or `"value"`
- Variable reference: use `$variable_name` to reference a defined variable
- Complex type: dictionaries, lists, and other complex structures are supported

### Parameter Type Handling

| Parameter Type | Format Example |
| --- | --- |
| String | Use single or double quotes, such as `"text"` or `'text'` |
| Number | Write the numeric value directly, such as `42` or `3.14` |
| Boolean | `true` or `false` |
| List | Use square brackets, such as `['finance', 'technology', 'other']` |
| Dictionary | Use curly braces, such as `{"key": "value"}` |
| Variable | Start with `$`, such as `$query` |

Examples:

```python
# Named parameter invocation
# Use variable x as the query condition for search
@_search(query=$x) -> result
# Get data sources related to "Sales" and "Inventory"
@getDataSourcesFromConcepts(conceptNames=["Sales", "Inventory"]) -> datasources

# Positional parameter invocation
# Search using variable x as a positional parameter
@_search($x) -> tag_result

# Mixed parameters
# Execute an SQL query and specify the data source, SQL statement, and timeout
@executeSQL(datasource="my_db", sql=$query, timeout=30) -> data

# Complex parameters
# Write a JSONL file whose content contains two dictionaries
@_write_jsonl(file_path="data/output.jsonl", content=[{"k":1},{"k":2}]) -> saved
```

## Natural Language Instructions

Natural-language content usually appears as the body text of blocks such as `/prompt/`, `/judge/`, and `/explore/`. The result is assigned to a variable with `->` or `>>`.

Example:

```python
# Use the deepseek-v3 model to summarize the theme of the text
/prompt/(model="deepseek-v3")
Summarize the following text in one sentence:
$text -> summary
```

```python
# Summarize the theme of the text with the default model
/prompt/
Summarize the following text in one sentence:
$text -> summary
```

Note: a standalone line of natural language does not form an executable unit by itself. It must appear inside one of the blocks above.

### `/prompt/` Block

Used to call an LLM directly for response generation.

Syntax:

```python
/prompt/(parameter_list) prompt_content -> variable_name
```

Supported parameters:

- `model`: specify the model to use
- `system_prompt`: system prompt
- `output`: output format, such as `"json"`, `"jsonl"`, or `"list_str"`
- `history`: whether to use conversation history, default `false`
- `no_cache`: whether to disable cache, default `false`

Note: the `/prompt/` block does not support the `tools` parameter. If tool-calling capability is required, use `/judge/` or `/explore/`.

Examples:

```python
# Specify model v3 and list_str output to return related concepts
/prompt/(model="v3", output="list_str") Return the concepts related to the problem description -> concepts
# Specify system prompt and model qwen-plus to write a poem
/prompt/(system_prompt="You are an AI assistant", model="qwen-plus") Write a poem -> poem
# Output JSON and generate user information
/prompt/(output="json") Generate user information -> user_info
# Output JSONL and generate three user records
/prompt/(output="jsonl") Generate three user records -> users
# Use conversation history and continue the previous dialog
/prompt/(history=true, model="v3") Continue the previous conversation -> response
```

### `/explore/` Block

Used for agent exploration and tool invocation, with support for multi-step reasoning. The agent automatically selects and calls suitable tools through multi-turn interaction to complete complex tasks.

Syntax:

```python
/explore/(parameter_list) task_description -> variable_name
```

Supported parameters:

- `tools`: available tools. Quoted and unquoted formats are supported, such as `["_bash", "_python"]` or `[_bash, _python]`
- `model`: specify the model to use
- `system_prompt`: system prompt
- `history`: whether to use conversation history
- `no_cache`: whether to disable cache

Tool-call format: during reasoning, the agent uses `=>#tool_name: {"param": "value"}` to invoke tools.

Return type: `Dict[str, Any]`, usually containing fields such as `"think"` and `"answer"`.

Features:

- Supports multi-turn tool invocation with automatic reasoning and decision-making
- Automatically detects repeated tool calls and provides hints
- Supports interruption and resume of tool invocation
- Supports code blocks as tool parameters

Examples:

```python
# Use executeSQL, _python, and _search with model v3 to solve a data analysis problem
/explore/(tools=[executeSQL, _python, _search], model="v3") Solve a data analysis problem -> result
# Use _search and _python with model v3 to search and analyze information
/explore/(tools=[_search, _python], model="v3") Search and analyze information -> analysis
# Use "_bash" and "_python" with a system prompt for system log analysis
/explore/(tools=["_bash", "_python"], system_prompt="You are a data analysis expert") Analyze system logs -> report
```

### `/judge/` Block

Used to judge a task and decide whether to invoke tools. The judge block analyzes the requirement. If a suitable tool exists, it invokes the tool. Otherwise, it answers directly with an LLM.

Syntax:

```python
/judge/(parameter_list) judgment_content -> variable_name
```

Supported parameters:

- `system_prompt`: system prompt
- `model`: specify the model
- `tools`: available tool list
- `history`: whether to use conversation history
- `prompt_skillcall`: choose the judgment mode, `"true"` or `"false"`, default `"true"`
  - `"true"`: use a custom prompt for tool judgment
  - `"false"`: use the model's function-calling capability for tool judgment

Notes:

- In the current implementation, parameter parsing normalizes `prompt_skillcall` to a boolean, while the judgment logic still compares it as a string. To avoid ambiguity:
  1. Omit the parameter, or explicitly pass `prompt_skillcall="false"` only when function-calling judgment is required
  2. It is not recommended to pass `prompt_skillcall="true"` explicitly
- Return type: `Dict[str, Any]`, containing either tool-call results or a direct LLM answer

Workflow:

1. Analyze the specific needs of the user question
2. Check whether any tool can satisfy the need
3. If a suitable tool exists, choose and invoke it
4. If no suitable tool exists, answer directly with the LLM

Examples:

```python
# No tools are available. Summarize the content with qwen-plus
/judge/(system_prompt="", model="qwen-plus", tools=[]) Summarize the content above -> summary
# Use _search and _python, and let function calling determine whether to use tools
/judge/(tools=[_search, _python], prompt_skillcall="false") Answer the user question -> answer
# Use executeSQL with model v3 to handle a data query request
/judge/(tools=[executeSQL], model="v3") Handle a data query request -> result
```

## Comments and Documentation

### Line Comments

Lines starting with `#` are treated as comments and ignored by the parser.

```python
# This is a comment
# Search weather information and assign the result to result
@_search(query="weather") -> result
```

### Documentation Comments

Use the `@DESC` marker to add documentation.

```python
@DESC
Memory Compression Agent: extract and save key knowledge points from the user's memory over the last N days
@DESC
```

### Multi-line String Documentation

Use triple quotes to define multi-line strings, often for prompts and rule definitions.

```python
'''
1. Sales data from different years is stored in different tables
2. Python code can be used for calculation tasks
3. If the SQL result is empty, check the field names
''' -> hints
```

## Output Format Control

Use the `output` parameter to control the output data format of the LLM. The system automatically parses and validates the output.

### JSON Format

Returns a single JSON object and is suitable for structured data scenarios.

```python
# Output JSON and generate user information
/prompt/(output="json") Generate user information -> user_info
```

Return type: `Dict[str, Any]`

Example output:

```json
{"name": "Zhang San", "age": 25, "city": "Beijing"}
```

### JSONL Format

Returns a list of JSON objects and is suitable for batch data processing.

```python
# Output JSONL and generate multiple user records
/prompt/(output="jsonl") Generate multiple user records -> users_list
```

Return type: `List[Dict[str, Any]]`

Example output:

```python
{"name": "Zhang San", "age": 25},
{"name": "Li Si", "age": 30},
{"name": "Wang Wu", "age": 28}
```

### List of Strings Format

Returns a list of strings and is suitable for simple text lists.

```python
# Output list_str and return a list of concept names
/prompt/(output="list_str") Return a list of concept names -> concept_names
```

Return type: `List[str]`

Example output:

```python
["Concept 1", "Concept 2", "Concept 3"]
```

### Applications of Output Formats

- Data extraction: extract structured data from unstructured text
- Batch processing: use `jsonl` for bulk generation
- Simplified processing: use `list_str` to quickly get text lists

## Advanced Syntax Features

### Tool Parameter Passing

Variables can be used as tool parameters, including simple variables, complex objects, and nested properties:

```python
# Use variable query as the search parameter
@_search(query=$query) -> search_results
# Get all concepts
@getAllConcepts() -> allConcepts

# Use model v3 and list_str output to select concepts related to the task
/prompt/(model="v3", output="list_str") Select the concept names related to the task from the following concepts:
$allConcepts
-> conceptInfos

# Get data sources based on the selected concepts
@getDataSourcesFromConcepts(conceptNames=$conceptInfos) -> datasources
```

### Multi-line Prompt Template

Complex multi-line prompt templates are supported, including variable interpolation and code block references:

```python
Today is [$date]
Please think step by step, use tools, and perform reasoning and calculation to get the final result.

Problem to solve
$query

datasource:
$datasources

Now begin: -> result
```

### Built-in Functions and Tool Invocation

Built-in functions can be called to get system information or perform specific operations.

System functions:

```python
# Get the current date
@_date() -> date
# Write data into a JSONL file
@_write_jsonl(file_path="data/file.jsonl", content=$data) -> outputPath
# Search with variable search_query
@_search(query=$search_query) -> search_results
```

### Complex Data Structure Handling

Complex nested structures are supported, including deeply nested objects and arrays:

```python
# Define a complex object user_data
{"user": {"name": "Zhang San", "profile": {"age": 25, "city": "Beijing"}}} -> user_data

# Access nested properties
$user_data.user.name -> user_name
$user_data.user.profile.city -> user_city

# Array access
["first item", "second item", "third item"] -> items
$items[0] -> first_item
```

### Tool Selection Control

Inside a `/judge/` block, when `prompt_skillcall="false"`, the `tool_choice` parameter can be used to control how the LLM selects tools:

```python
# Automatically choose a tool
/judge/(tools=[_search, _python], prompt_skillcall="false",
tool_choice="auto") Task description -> result

# Must use a tool
/judge/(tools=[_search, _python], prompt_skillcall="false",
tool_choice="required") Task description -> result

# Do not use tools
/judge/(tools=[_search, _python], prompt_skillcall="false",
tool_choice="none") Task description -> result
```

Notes:

- `tool_choice` works only inside `/judge/` and only when `prompt_skillcall="false"`
- `/prompt/` does not support `tools` or `tool_choice`
- `/explore/` has its own tool-calling mechanism and does not use `tool_choice`

### Cache Control

Use the `no_cache` parameter to control whether cache is used:

```python
# Disable cache to ensure the latest result is returned
/prompt/(no_cache=true) Get the latest data -> fresh_data

# Use cache
/prompt/(no_cache=false) Query data -> cached_data
```

### Conversation History Management

Use the `history` parameter to control whether conversation history is used:

```python
# Use conversation history to continue the previous topic
/prompt/(history=true) Continue the previous topic -> response

# Do not use history and handle a new question
/prompt/(history=false) A new question -> answer
```

### Flexible Tool List Formats

Multiple formats are supported for the tool list to improve readability:

```python
# Quoted format
/explore/(tools=["executeSQL", "_python"]) Analyze data -> result

# Unquoted format
/explore/(tools=[executeSQL, _python]) Analyze data -> result

# Single tool
/explore/(tools=executeSQL) Execute a query -> result
```

## Best Practices

### Variable Naming

- Use meaningful variable names that clearly reflect their purpose
- Use underscores to separate multiple words, such as `user_data` and `search_results`
- Avoid reserved words and system variable names
- For temporary variables, use short but clear names

### Code Organization

- Use comments appropriately to explain complex logic and business rules
- Use `@DESC` to document the purpose of the Agent or file
- Organize related operations together to keep the code structure clear
- Use blank lines to separate logical blocks

### Tool Invocation

- In `/explore/`, specify only the tools that are actually needed
- Choose an appropriate model version according to task complexity
- Provide detailed prompt text and context information for complex tasks
- Use `system_prompt` to clarify the agent role and task requirements

### Output Format Selection

- Choose an output format according to the usage of the data
- Use `json` for structured data
- Use `jsonl` for batch data
- Use `list_str` for simple lists
- Use `obj/TypeName` when type safety is needed

### Parameter Passing

- Prefer named parameters to improve readability
- For complex parameters, consider assigning them to a variable first
- Ensure referenced variables are already defined
- Pay attention to parameter types such as strings, numbers, and booleans

### Performance Optimization

- Use `no_cache` reasonably to balance performance and freshness
- Avoid repeated LLM calls inside loops
- Use batch processing when working with large datasets
- When enabling `history`, pay attention to context length

## Real-world Examples

### Data Analysis Task

```python
@DESC
ChatBI Data Exploration Agent: use SQL and Python tools for data analysis
@DESC

# Get the current date
@_date() -> date
# Get all concepts
@getAllConcepts() -> allConcepts

# Use model v3 and list_str output to select concepts related to the problem
/prompt/(model="v3", output="list_str") Based on the problem to solve and the concept descriptions, return the list of concept names that may be needed
Problem to solve:
$query

Concept descriptions:
$allConcepts
-> conceptInfos

# Get data sources based on the selected concepts
@getDataSourcesFromConcepts(conceptNames=$conceptInfos) -> datasources

# Use executeSQL, _python, and _search with model v3 for data analysis
/explore/(tools=[executeSQL, _python, _search], model="v3")
Today is [$date]
Please think step by step, use tools, and perform reasoning and calculation to get the final result.

Problem to solve
$query

datasource:
$datasources

Please note!
(1) Do not use assumed data for calculation. Assumed data is meaningless here.
(2) Execute the SQL immediately after writing it.
(3) If a tool call fails, adjust the parameters and try again.
Now begin: -> result
```

### Memory Compression Task

```python
@DESC
Memory Compression Agent: extract and save key knowledge points from the user's memory in the last N days
@DESC

# Read the stored memory data
@_mem_view(path="") -> memoryData

# Use model v3 and jsonl output to extract the 10 most important knowledge points
/prompt/(model="v3", output="jsonl") Based on the following user memory data, extract the 10 most important knowledge points.
User memory data:
$memoryData

Please extract the 10 most important knowledge points from these memories. Each one should:
1. Be concrete and valuable
2. Help future user decisions or behavior
3. Have general importance

Output a JSON array in which each element contains:
- content: the knowledge point content
- score: importance score, integer from 1 to 100

Please begin:-> knowledgePoints

# Write the compressed knowledge into the knowledge file
@_write_jsonl(file_path="data/memory/user_chatbi_user/knowledge.jsonl",
content=$knowledgePoints) -> outputPath
```

### Simple Search Task

```python
# Use _search and _python with model v3 to perform a search task
/explore/(tools=[_search, _python], model="v3")
Today is 2025.7.1. Please think step by step, use tools, and perform reasoning and calculation to get the final result.
Task: $query
Now begin: -> result
```

### Concept Extraction and Data Retrieval

```python
# Get all concepts
@getAllConcepts() -> allConcepts

# Use model v3 and list_str output to return all concept names that may be needed
/prompt/(model="v3", output="list_str") Based on the problem description and the concepts description, return the names of all concepts that may be needed
Problem to solve:
$query

concepts description:
$allConcepts

Output only the concept name list -> conceptInfos

# Get sample data from the selected concepts
@getSampleData(conceptNames=$conceptInfos) -> sampledData
# Get data source schemas from the selected concepts
@getDataSourceSchemas(conceptNames=$conceptInfos) -> schemas
```

### JSON Data Generation and Retrieval

```python
# Output JSON and generate a random user information object
/prompt/(output="json") Generate random user information in JSON, for example `{"name":"Zhang Sanfeng","age":30}` -> user_info
# Extract the answer field
$user_info['answer'] -> object_str
# Parse the JSON string and get the name field
json.loads($object_str)["name"] -> name

# Output JSONL and generate a random list of user information
/prompt/(output="jsonl") Generate a random user information list in array[object] format. Example:
`[{"name":"Zhang Sanfeng","age":30},{"name":"Zhu Yuanzhang","age":26}]`. You must follow the sample format strictly -> user_info_list
# Extract the answer field
$user_info_list['answer'] -> jsonl_str
# Parse the JSON string and get the first element
json.loads($jsonl_str)[0] -> column

# Output list_str and generate a random number list
/prompt/(output="list_str") Generate a random list of numbers in array[str] format. Example:
`[1,2,3,4]`. You must follow the sample format strictly -> num_list
# Extract the answer field
$num_list['answer'] -> list_str
# Parse the JSON string
json.loads($list_str) -> list_num
# Get the first element
$list_num[0] -> num
```

### Writing Task

```python
# Enable debug and rewrite the original query into multiple search-friendly questions
/prompt/(flags='{"debug": true}') Rewrite the following raw query into multiple questions that are more suitable for search engines.
Please follow these rules:
1. Keep each rewritten question concise and focused on key terms.
2. Generate questions from different angles or subdomains to cover broader search results.
3. Avoid vague expressions and make them as specific as possible.
4. If the original question contains complex concepts, split it into simpler questions.
5. Generate 3 to 5 questions.
6. Return the rewritten questions in list format.

Example:
Original question: How can I improve spoken English?
Reference material: Improving spoken English requires lots of practice and the right methods, including listening training, pronunciation imitation, language exchange, and learning apps. Beginners and advanced learners focus on different things.
Rewritten questions:
["Best methods to improve spoken English","Spoken English practice tips for beginners","How listening training improves speaking","Recommended language exchange platforms and how to use them","How advanced learners can break through spoken English bottlenecks"]

Only generate the rewritten question list. Do not generate any extra explanation.
Original question: $query
Rewritten questions:->sub_querys

# Parse the rewritten question list
eval($sub_querys.answer)->search_querys
# Use only the first two questions
$search_querys[:2]->search_querys
# Initialize ref as an empty string
''->ref

# Iterate over each search query
/for/ $search_query in $search_querys:
    # Search the current query with zhipu_search_tool
    @zhipu_search_tool(query=$search_query)->result
    # Extract the tool call section from the search result
    $result['answer']['choices'][0]['message']['tool_calls'] -> result
    
    # Judge the number of search results
    /if/ len($result[1]['search_result']) > 0:
        # Extract the search results when not empty
        $result[1]['search_result']->search_result
    else:
        # Assign an empty list when there is no result
        []->search_result
    /end/
    
    # Append current search results to the total result list
    $search_result>>search_results
    # Initialize sub_ref as an empty string
    ''->sub_ref
    
    # Traverse each page in the current search result
    /for/ $page in $search_result:
        # Concatenate page content to sub_ref
        $sub_ref+$page['content']->sub_ref
    /end/
    
    # Keep only the first 5000 characters of sub_ref
    $sub_ref[:5000] -> sub_ref
    # Append sub_ref to ref
    $ref+$sub_ref ->ref
/end/

# Initialize related variables to empty strings
''->page
''->sub_ref
''->result
''->search_result
''->sub_querys
''->search_query

# Judge whether ref is empty
/if/ $ref == '':
    # Tell the user no related material was found
    /prompt/Please ignore all other prompts and output directly: no related material was found -> answer
else:
    # Use history and complete the writing task based on the reference material
    /prompt/(history=True) Please complete the writing task according to the problem and the reference material.
    <reference material> $ref </reference material>
    <task> $query </task>
    Requirements:
    1. Be as detailed as possible

    Written content:->answer
/end/
```

## Summary of Core Features

DolphinLanguage is a domain-specific language designed for AI Agent development. It has the following core features:

### Language Design Philosophy

- Declarative programming: focus on **what to do** instead of **how to do it**
- Natural-language friendly: seamless integration of natural language and structured code
- Tool-driven: an Agent programming paradigm centered on tool invocation
- Flexible types: supports multiple data types and output formats

### Main Advantages

1. **Concise expression**
   - Uses `->` and `>>` for compact variable assignment
   - Supports variable interpolation and complex data-structure access
   - Simplified function-invocation syntax
2. **Intelligent reasoning**
   - `/explore/` supports multi-turn tool invocation and automatic reasoning
   - `/judge/` can decide intelligently whether a tool call is needed
   - Automatic repeated-call detection and error handling
3. **Flexible control**
   - Supports control-flow structures such as conditionals and loops
   - Provides rich parameter configuration options
4. **Output formatting**
   - Supports multiple output formats such as JSON, JSONL, List, and Object
   - Automatically parses and validates output results
   - Supports custom type definitions
5. **Developer-friendly**
   - Complete syntax validation
   - Clear error messages
   - Rich documentation and examples

### Typical Use Cases

- Data analysis: combine SQL and Python tools for complex analysis
- Information retrieval: use search tools for querying and summarizing information
- Knowledge management: implement memory compression, knowledge extraction, and similar tasks
- Business automation: orchestrate multiple tools to complete complex business processes
- Conversational systems: build intelligent conversational agents with tool-calling capability

### Extensibility

- Tool integration: easily integrate new tools and skills
- Type system: supports custom object types
- Flexible configuration: rich parameters support many scenarios
- Modularization: supports function encapsulation and code reuse

This document is based on the latest SDK implementation and reflects the syntax features and actual capabilities of DolphinLanguage. It covers content from basic syntax to advanced features and includes rich real-world examples.

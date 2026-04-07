package drivenadapters

// import (
// 	"context"
// 	"fmt"
// 	"testing"

// 	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
// 	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
// 	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
// 	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
// 	. "github.com/smartystreets/goconvey/convey"
// )

// func TestChatCompletion(t *testing.T) {
// 	um := &mfModelAPIClient{
// 		baseURL:    "http://mf-model-api.anyshare:9898/api/private/mf-model-api",
// 		logger:     logger.DefaultLogger(),
// 		httpClient: rest.NewHTTPClient(),
// 	}
// 	Convey("TestChatCompletion:同步请求", t, func() {
// 		resp, err := um.ChatCompletion(context.Background(), &interfaces.ChatCompletionReq{
// 			Model: "",
// 			Messages: []interfaces.ChatCompletionMessage{
// 				{
// 					Role:    "system",
// 					Content: "# 函数生成 Prompt 模板\r\n\r\n## 角色\r\n你是一个专门用于生成事件驱动型Python工具代码的智能助手\r\n\r\n## 目标\r\n你的任务是根据用户的自然语言描述、可选的元数据(inputs,outputs)以及可选的已安装依赖库列表,编写一个符合严格格式规范的Python脚本\r\n\r\n## 代码模板规范\r\n所有生成的脚本必须严格遵循以下结构,不得更改:\r\n\r\n1. **导入模块**:\r\n    - 必须导入: `from typing import Dict, Any`\r\n    - **如果用户提供了已安装依赖库列表**:\r\n        - 必须使用用户提供的库列表来选择实现方案\r\n        - 优先使用用户环境已安装的库\r\n        - 确保导入的库都在用户的依赖库列表中\r\n        - 示例: `import requests`, `import json`, `from datetime import datetime`\r\n    - **如果未提供依赖库列表**:\r\n        - 优先使用 Python 标准库实现功能\r\n        - 仅在必要时使用通用的流行第三方库（如 `requests`）\r\n\r\n2. **处理函数 (Handler)**:\r\n    ```python\r\n    def handler(event: Dict[str, Any]):\r\n        \"\"\"\r\n        [工具功能的简要描述]\r\n\r\n        Parameters:\r\n        event: dict\r\n            [描述 event 中预期的输入参数]\r\n\r\n        Return:\r\n        [描述返回的数据对象]\r\n        \"\"\"\r\n\r\n        # 业务逻辑实现\r\n        return result\r\n    ```\r\n\r\n3. **本地测试块 (Test Block)**:\r\n    ```python\r\n    if __name__ == '__main__':\r\n        # 本地测试代码\r\n        print(\"--- Start Local Test ---\")\r\n        test_event = { ... } # 构造一个符合 inputs 要求的测试事件\r\n        print(\"Input:\", test_event)\r\n        print(\"Result:\", handler(test_event))\r\n        print(\"--- End Local Test ---\")\r\n    ```\r\n\r\n## 逻辑实现规则\r\n\r\n### 1. 依赖库处理\r\n- **如果用户提供了已安装依赖库列表**:\r\n    - 必须在导入部分只使用列表中的库\r\n    - 选择列表中最适合的库来实现功能逻辑\r\n    - 如果功能可以用标准库实现，优先使用标准库\r\n    - 如果必须使用第三方库，从提供的列表中选择\r\n    - 确保导入的库都被实际使用\r\n- **如果未提供依赖库列表**:\r\n    - 优先使用 Python 标准库实现功能\r\n    - 如果必须使用第三方库，选择通用、流行的库（如 `requests`）\r\n    - 仅在代码逻辑确实需要第三方库时才导入和使用\r\n\r\n### 2. 输入处理 (Inputs)\r\n- 所有输入参数必须通过 `event` 字典传递\r\n- **如果提供了 `inputs` 元数据**:\r\n    - 遍历每一个定义的输入项\r\n    - **提取**: 使用 `event.get(\"name\", default_value)` 获取参数。优先使用元数据中定义的 `default`，如果没有则使用合理的默认值或 `None`\r\n    - **校验**: 如果 `required` 为 `true`，必须检查参数是否存在。如果缺失，应抛出 `ValueError` 或返回包含错误信息的字典\r\n    - **类型转换**: 必须强制执行 `type` 定义的类型转换（例如：将字符串转为 `int`, `float`, 或者解析 JSON 数组/对象）\r\n    - **复杂嵌套结构处理**:\r\n        - 如果参数类型为 `object` 且包含 `sub_parameters`，需要递归处理嵌套对象\r\n        - 如果参数类型为 `array` 且包含 `sub_parameters`，需要处理数组元素类型\r\n- **如果未提供 `inputs` 元数据**:\r\n    - 根据用户描述推断必要的输入参数\r\n    - 使用防御性编程（例如 `event.get()`）来处理潜在的缺失键\r\n\r\n### 3. 输出处理 (Outputs)\r\n- 函数必须返回一个可序列化的对象（通常是 `dict`）\r\n- **如果提供了 `outputs` 元数据**:\r\n    - 确保返回字典的键值结构与 `outputs` 定义完全匹配\r\n- **如果没有提供**:\r\n    - 返回一个结构清晰的字典，例如 `{\"result\": ...}` 或 `{\"message\": ...}`\r\n\r\n### 4. 代码质量保证\r\n- 确保代码中使用的所有输入参数都能从 event 中获取\r\n- 返回值结构清晰，易于理解\r\n\r\n### 5. 通用规则\r\n- 核心逻辑周围必须包含适当的错误处理（`try/except` 块）\r\n- 如果提供了依赖库列表，代码逻辑必须与列表中的库匹配\r\n- 确保所有导入的库都被实际使用，避免无效导入\r\n- 代码必须是自包含的\r\n\r\n## 输出格式\r\n请严格按照以下结构输出最终结果。不要包含额外的 Markdown 标记（如 ```python）。\r\n\r\nfrom typing import Dict, Any\r\ndef handler(event):\r\n    # 代码内容...\r\n    pass\r\n\r\n接下来我会输入简短的代码内容或者需求描述,请直接给出生成的代码结果,不要输出任何其他内容\r\n请严格按照正确格式输出纯Python代码,不要使用代码块标记\r\n如果输入内容意义不明确或者输入为空白,你需要给出较为泛用的代码\r\n",
// 				},
// 				{
// 					Role:    "user",
// 					Content: "写一个工具，计算两个数字 a 和 b 的和",
// 				},
// 			},
// 			Temperature:      0.1,
// 			TopP:             0.1,
// 			TopK:             20,
// 			Stream:           false,
// 			FrequencyPenalty: 0.1,
// 			PresencePenalty:  0.1,
// 			MaxTokens:        2048,
// 		})
// 		fmt.Println(err)
// 		So(err, ShouldBeNil)
// 		So(resp, ShouldNotBeNil)
// 		fmt.Println(utils.ObjectToJSON(resp))
// 		// So(len(resp.Choices), ShouldEqual, 1)
// 		// So(resp.Choices[0].Message.Role, ShouldEqual, "assistant")
// 		// So(resp.Choices[0].Message.Content, ShouldNotBeEmpty)
// 		// fmt.Println(resp.Choices[0].Message.Content)
// 	})
// 	Convey("TestChatCompletion:流式请求", t, func() {
// 		ctx, cancel := context.WithCancel(context.Background())
// 		defer cancel() // 确保在测试结束时清理资源
// 		messageCh, errCh, err := um.StreamChatCompletion(ctx, &interfaces.ChatCompletionReq{
// 			Model: "",
// 			Messages: []interfaces.ChatCompletionMessage{
// 				{
// 					Role:    "system",
// 					Content: "# 函数生成 Prompt 模板\r\n\r\n## 角色\r\n你是一个专门用于生成事件驱动型Python工具代码的智能助手\r\n\r\n## 目标\r\n你的任务是根据用户的自然语言描述、可选的元数据(inputs,outputs)以及可选的已安装依赖库列表,编写一个符合严格格式规范的Python脚本\r\n\r\n## 代码模板规范\r\n所有生成的脚本必须严格遵循以下结构,不得更改:\r\n\r\n1. **导入模块**:\r\n    - 必须导入: `from typing import Dict, Any`\r\n    - **如果用户提供了已安装依赖库列表**:\r\n        - 必须使用用户提供的库列表来选择实现方案\r\n        - 优先使用用户环境已安装的库\r\n        - 确保导入的库都在用户的依赖库列表中\r\n        - 示例: `import requests`, `import json`, `from datetime import datetime`\r\n    - **如果未提供依赖库列表**:\r\n        - 优先使用 Python 标准库实现功能\r\n        - 仅在必要时使用通用的流行第三方库（如 `requests`）\r\n\r\n2. **处理函数 (Handler)**:\r\n    ```python\r\n    def handler(event: Dict[str, Any]):\r\n        \"\"\"\r\n        [工具功能的简要描述]\r\n\r\n        Parameters:\r\n        event: dict\r\n            [描述 event 中预期的输入参数]\r\n\r\n        Return:\r\n        [描述返回的数据对象]\r\n        \"\"\"\r\n\r\n        # 业务逻辑实现\r\n        return result\r\n    ```\r\n\r\n3. **本地测试块 (Test Block)**:\r\n    ```python\r\n    if __name__ == '__main__':\r\n        # 本地测试代码\r\n        print(\"--- Start Local Test ---\")\r\n        test_event = { ... } # 构造一个符合 inputs 要求的测试事件\r\n        print(\"Input:\", test_event)\r\n        print(\"Result:\", handler(test_event))\r\n        print(\"--- End Local Test ---\")\r\n    ```\r\n\r\n## 逻辑实现规则\r\n\r\n### 1. 依赖库处理\r\n- **如果用户提供了已安装依赖库列表**:\r\n    - 必须在导入部分只使用列表中的库\r\n    - 选择列表中最适合的库来实现功能逻辑\r\n    - 如果功能可以用标准库实现，优先使用标准库\r\n    - 如果必须使用第三方库，从提供的列表中选择\r\n    - 确保导入的库都被实际使用\r\n- **如果未提供依赖库列表**:\r\n    - 优先使用 Python 标准库实现功能\r\n    - 如果必须使用第三方库，选择通用、流行的库（如 `requests`）\r\n    - 仅在代码逻辑确实需要第三方库时才导入和使用\r\n\r\n### 2. 输入处理 (Inputs)\r\n- 所有输入参数必须通过 `event` 字典传递\r\n- **如果提供了 `inputs` 元数据**:\r\n    - 遍历每一个定义的输入项\r\n    - **提取**: 使用 `event.get(\"name\", default_value)` 获取参数。优先使用元数据中定义的 `default`，如果没有则使用合理的默认值或 `None`\r\n    - **校验**: 如果 `required` 为 `true`，必须检查参数是否存在。如果缺失，应抛出 `ValueError` 或返回包含错误信息的字典\r\n    - **类型转换**: 必须强制执行 `type` 定义的类型转换（例如：将字符串转为 `int`, `float`, 或者解析 JSON 数组/对象）\r\n    - **复杂嵌套结构处理**:\r\n        - 如果参数类型为 `object` 且包含 `sub_parameters`，需要递归处理嵌套对象\r\n        - 如果参数类型为 `array` 且包含 `sub_parameters`，需要处理数组元素类型\r\n- **如果未提供 `inputs` 元数据**:\r\n    - 根据用户描述推断必要的输入参数\r\n    - 使用防御性编程（例如 `event.get()`）来处理潜在的缺失键\r\n\r\n### 3. 输出处理 (Outputs)\r\n- 函数必须返回一个可序列化的对象（通常是 `dict`）\r\n- **如果提供了 `outputs` 元数据**:\r\n    - 确保返回字典的键值结构与 `outputs` 定义完全匹配\r\n- **如果没有提供**:\r\n    - 返回一个结构清晰的字典，例如 `{\"result\": ...}` 或 `{\"message\": ...}`\r\n\r\n### 4. 代码质量保证\r\n- 确保代码中使用的所有输入参数都能从 event 中获取\r\n- 返回值结构清晰，易于理解\r\n\r\n### 5. 通用规则\r\n- 核心逻辑周围必须包含适当的错误处理（`try/except` 块）\r\n- 如果提供了依赖库列表，代码逻辑必须与列表中的库匹配\r\n- 确保所有导入的库都被实际使用，避免无效导入\r\n- 代码必须是自包含的\r\n\r\n## 输出格式\r\n请严格按照以下结构输出最终结果。不要包含额外的 Markdown 标记（如 ```python）。\r\n\r\nfrom typing import Dict, Any\r\ndef handler(event):\r\n    # 代码内容...\r\n    pass\r\n\r\n接下来我会输入简短的代码内容或者需求描述,请直接给出生成的代码结果,不要输出任何其他内容\r\n请严格按照正确格式输出纯Python代码,不要使用代码块标记\r\n如果输入内容意义不明确或者输入为空白,你需要给出较为泛用的代码\r\n",
// 				},
// 				{
// 					Role:    "user",
// 					Content: "写一个工具，计算两个数字 a 和 b 的和",
// 				},
// 			},
// 			Temperature:      0.1,
// 			TopP:             0.1,
// 			TopK:             20,
// 			Stream:           true,
// 			FrequencyPenalty: 0.1,
// 			PresencePenalty:  0.1,
// 			MaxTokens:        2048,
// 		})
// 		fmt.Println(err)
// 		So(err, ShouldBeNil)
// 		So(messageCh, ShouldNotBeNil)
// 		So(errCh, ShouldNotBeNil)
// 	loop:
// 		for {
// 			select {
// 			case msg := <-messageCh:
// 				fmt.Println(msg)
// 				if msg == "data: [DONE]" {
// 					break loop
// 				}
// 			case err := <-errCh:
// 				fmt.Println(err)
// 				break loop
// 			case <-ctx.Done():
// 				fmt.Println("context done")
// 				break loop
// 			}
// 		}
// 	})
// }

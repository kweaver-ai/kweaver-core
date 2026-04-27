package v3agentconfigsvc

import (
	"fmt"
	"strings"
)

const (
	enUS = "en-US"
	zhCN = "zh-CN"
	zhTW = "zh-TW"
)

func openingRemarksSystemPrompt(language string) (prompt string) {
	var userInput, outputExample, promptTemplate string

	switch language {
	case enUS:
		// 英文版本
		userInput = `Name: Document Reading Assistant
Description: I am an AI assistant specialized in helping users read and understand various types of documents.
Skills List: Document summary generation, key information extraction, content explanation, question answering, format conversion
Knowledge Sources:
- Document Library: A
- Directory: B
- Knowledge Repository: C
Please generate an opening statement based on the information above.
Requirements:
1. Introduce yourself, where you get your knowledge from, what skills you have, and what your role is.
2. Welcome users at the end.
3. Output using Markdown syntax or plain text.`

		outputExample = `# 👋 Hello, I'm your Document Reading Assistant!

I can help you read and understand various types of documents. My knowledge sources include:
- Document Library: A
- Directory: B
- Knowledge Repository: C

These resources enable me to provide you with professional document analysis services.

## My skills include:
- 📝 Generating document summaries to help you quickly grasp core content
- 🔍 Extracting key information and highlights from documents
- 💡 Explaining complex document content to make it easier to understand
- ❓ Answering your questions about document content
- 🔄 Converting document content into different formats

Whether you need to understand technical documentation, analyze academic papers, or simplify complex user manuals, I can help!

Feel free to ask me questions anytime, and let me help you solve various challenges in document reading!`

		promptTemplate = `Generate an opening statement based on the user's input.
Requirements: Refer to the example below and directly output the content requested by the user. Do not output anything else, such as "OK". Do not wrap the output in """.
Language requirements for the opening statement: Use the same language as the user's input. If the user's input is in English, the generated opening statement should also be in English.
Reference example (content wrapped in """): 
User input content: """%s""", generated opening statement: """%s""".`

	case zhTW:
		// 繁体中文版本
		userInput = `名稱：文檔閱讀小助手
簡介：我是一個專門幫助用戶閱讀和理解各類文檔的AI助手。
技能列表：文檔摘要生成、關鍵信息提取、內容解釋、問題解答、格式轉換
知識來源：
- 文檔庫：A
- 目錄：B
- 知識倉庫：C
請根據上面提供的信息生成一個開場白。
要求：
1. 內容為介紹自己能從什麼地方獲取知識，並擁有什麼樣的技能，你的角色是什麼等。
2. 最後歡迎用戶來使用。
3. 使用Markdown語法或普通文本來輸出。`

		outputExample = `# 👋 你好，我是文檔閱讀小助手！

我可以幫助你閱讀和理解各種類型的文檔。我的知識來源包括：
- 文檔庫: A
- 目錄: B
- 知識倉庫: C

這些資源使我能夠為你提供專業的文檔分析服務。

## 我的技能包括：
- 📝 生成文檔摘要，幫你快速把握核心內容
- 🔍 提取文檔中的關鍵信息和重點
- 💡 解釋複雜的文檔內容，使其更容易理解
- ❓ 回答你關於文檔內容的問題
- 🔄 將文檔內容轉換為不同的格式

無論你是需要理解技術文檔、分析學術論文，還是簡化複雜的用戶手冊，我都能為你提供幫助！

歡迎隨時向我提問，讓我幫你解決文檔閱讀中的各種難題！`

		promptTemplate = `根據用戶輸入的內容生成開場白。
要求：參考下面的示例直接輸出用戶要求的內容，不要輸出其他任何內容，如"好的"等。輸出內容不要用"""包裹。
生成開場白的語言要求：跟用戶輸入的語言保持一致，假如用戶輸入的是英語，則生成的開場白也是英文的。
參考示例（"""包裹的為相關內容）：
用戶輸入內容為："""%s""", 生成的開場白為："""%s"""。`

	default: // 默认使用简体中文
		userInput = `名称：文档阅读小助手
简介：我是一个专门帮助用户阅读和理解各类文档的AI助手。
技能列表：文档摘要生成、关键信息提取、内容解释、问题解答、格式转换
知识来源：
- 文档库：A
- 目录：B
- 知识仓库：C
请根据上面提供的信息生成一个开场白。
要求：
1. 内容为介绍自己能从什么地方获取知识，并拥有什么样的技能，你的角色是什么等。
2. 最后欢迎用户来使用。
3. 使用Markdown语法或普通文本来输出。`

		outputExample = `# 👋 你好，我是文档阅读小助手！

我可以帮助你阅读和理解各种类型的文档。我的知识来源包括：
- 文档库: A
- 目录: B
- 知识仓库: C

这些资源使我能够为你提供专业的文档分析服务。

## 我的技能包括：
- 📝 生成文档摘要，帮你快速把握核心内容
- 🔍 提取文档中的关键信息和重点
- 💡 解释复杂的文档内容，使其更容易理解
- ❓ 回答你关于文档内容的问题
- 🔄 将文档内容转换为不同的格式

无论你是需要理解技术文档、分析学术论文，还是简化复杂的用户手册，我都能为你提供帮助！

欢迎随时向我提问，让我帮你解决文档阅读中的各种难题！`

		promptTemplate = `根据用户输入的内容生成开场白。
要求：参考下面的示例直接输出用户要求的内容，不要输出其他任何内容，如"好的"等。输出内容不要用"""包裹。
生成开场白的语言要求：跟用户输入的语言保持一致，假如用户输入的是英语，则生成的开场白也是英文的。
参考示例（"""包裹的为相关内容）：
用户输入内容为："""%s""", 生成的开场白为："""%s"""。`
	}

	prompt = fmt.Sprintf(promptTemplate, userInput, outputExample)

	return prompt
}

func systemPrompt(language string) (prompt string) {
	var userInput, outputExample, promptTemplate string

	switch language {
	case enUS:
		// 英文版本
		userInput = `Name: Document Reading Assistant
Description: I am an AI assistant specialized in helping users read and understand various types of documents.
Skills List: Document summary generation, key information extraction, content explanation, question answering, format conversion
Knowledge Sources:
- Document Library: A
- Directory: B
- Knowledge Repository: C
Please generate a "Personality and Response Logic" based on the information above.
Requirements:
1. Describe the assistant's personality traits, language style, and communication approach
2. Explain how the assistant uses its knowledge and skills to answer user questions
3. Specify the assistant's strategies for handling different types of questions
4. Ensure the personality aligns with the assistant's professional domain and functions
5. Use concise language, not exceeding 300 words`

		outputExample = `### Personality and Style
- Professional and friendly, using clear and concise language
- Patient and detailed, skilled at explaining complex concepts
- Structured response style, utilizing headings, lists, and emphasis formatting
- Calm and professional tone, avoiding being too casual or too formal

### Knowledge Application
- Prioritize retrieving information from Document Library A, Directory B, and Knowledge Repository C
- Utilize appropriate skills based on question type: summary generation, information extraction, content explanation, etc.
- Reference specific sources in responses to enhance credibility

### Question Handling Strategies
- Document summary requests: Provide structured summaries highlighting key points
- Information extraction questions: Precisely locate and extract key content, ranked by importance
- Explanation requests: Transform complex content into easy-to-understand language, using analogies where appropriate
- Format conversion needs: Preserve the essence of original content while optimizing presentation

When beyond knowledge scope, clearly explain limitations and provide alternative suggestions.`

		promptTemplate = `Generate a "Personality and Response Logic" based on the user's input.
Requirements: Refer to the example below and directly output the content requested by the user. Do not output anything else, such as "OK". Do not wrap the output in """.
Language requirements for the personality and response logic: Use the same language as the user's input. If the user's input is in English, the generated personality and response logic should also be in English.
Reference example (content wrapped in """): 
User input content: """%s""", generated "Personality and Response Logic": """%s""".`

	case zhTW:
		// 繁体中文版本
		userInput = `名稱：文檔閱讀小助手
簡介：我是一個專門幫助用戶閱讀和理解各類文檔的AI助手。
技能列表：文檔摘要生成、關鍵信息提取、內容解釋、問題解答、格式轉換
知識來源：
- 文檔庫：A
- 目錄：B
- 知識倉庫：C
請根據上面提供的信息生成一個"人設及回覆邏輯"。
要求：
1. 描述助手的性格特點、語言風格和溝通方式
2. 說明助手如何利用自己的知識和技能回答用戶問題
3. 明確助手在面對不同類型問題時的處理策略
4. 確保人設與助手的專業領域和功能相符
5. 使用簡潔明了的語言，不超過300字`

		outputExample = `### 性格與風格
- 專業且友好，使用清晰簡潔的語言
- 耐心細緻，善於解釋複雜概念
- 回覆風格結構化，善用標題、列表和強調格式
- 語氣平和專業，避免過於隨意或過於嚴肅

### 知識應用
- 優先從文檔庫A、目錄B和知識倉庫C中獲取信息
- 根據問題類型調用相應技能：摘要生成、信息提取、內容解釋等
- 在回答中引用具體來源，增強可信度

### 問題處理策略
- 文檔摘要請求：提供結構化摘要，突出關鍵點
- 信息提取問題：精確定位並提取關鍵內容，按重要性排序
- 解釋請求：將複雜內容轉化為易懂語言，適當使用類比
- 格式轉換需求：保留原始內容精髓，優化呈現方式

當超出知識範圍時，清晰說明限制並提供替代建議。`

		promptTemplate = `根據用戶輸入的內容生成"人設及回覆邏輯"。
要求：參考下面的示例直接輸出用戶要求的內容，不要輸出其他任何內容，如"好的"等。輸出內容不要用"""包裹。
生成人設及回覆邏輯的語言要求：跟用戶輸入的語言保持一致，假如用戶輸入的是英語，則生成的人設及回覆邏輯也是英文的。
參考示例（"""包裹的為相關內容）：
用戶輸入內容為："""%s""", 生成的"人設及回覆邏輯"為："""%s"""。`

	default: // 默认使用简体中文
		userInput = `
		你是一个提示词工程师，你可以根据用户的需求描述和简短提示词内容，给出优化后的结构化的提示词，使其遵循特定的模式和规则，从而方便有效理解信息。
		格式和规则如下：
		-----------------------------------------------------------
		## Role : [请填写你想定义的角色名称]
		
		## Background : [请描述角色的背景信息，例如其历史、来源或特定的知识背景]
		
		## Preferences : [请描述角色的偏好或特定风格，例如对某种设计或文化的偏好]
		
		## Profile :
		
		- language: 中文
		- description: [请简短描述该角色的主要功能，50 字以内]
		
		## Goals :
		[请列出该角色的主要目标 1]
		[请列出该角色的主要目标 2]
		...
		
		## Constrains :
		[请列出该角色在互动中必须遵循的限制条件 1]
		[请列出该角色在互动中必须遵循的限制条件 2]
		...
		
		## Skills :
		
		[为了在限制条件下实现目标，该角色需要拥有的技能 1]
		[为了在限制条件下实现目标，该角色需要拥有的技能 2]
		...
		
		## Examples :
		
		[提供一个输出示例 1，展示角色的可能回答或行为]
		[提供一个输出示例 2]
		...
		
		## OutputFormat :
		
		[请描述该角色的工作流程的第一步]
		[请描述该角色的工作流程的第二步]
		...
		
		## Initialization : 作为 [角色名称], 拥有 [列举技能], 严格遵守 [列举限制条件], 使用默认 [选择语言] 与用户对话，友好的欢迎用户。然后介绍自己，并提示用户输入.
		-----------------------------------------------------------
		输出结果示例1：
		# Role:知识探索专家
		
		## Profile:
		- language: 中文
		- description: 我是一个专门用于提问并解答有关特定知识点的 AI 角色。
		
		## Goals:
		提出并尝试解答有关用户指定知识点的三个关键问题：其来源、其本质、其发展。
		
		## Constrains:
		1. 对于不在你知识库中的信息, 明确告知用户你不知道
		2. 你不擅长客套, 不会进行没有意义的夸奖和客气对话
		3. 解释完概念即结束对话, 不会询问是否有其它问题
		
		## Skills:
		1. 具有强大的知识获取和整合能力
		2. 拥有广泛的知识库, 掌握提问和回答的技巧
		3. 拥有排版审美, 会利用序号, 缩进, 分隔线和换行符等等来美化信息排版
		4. 擅长使用比喻的方式来让用户理解知识
		5. 惜字如金, 不说废话
		
		## Workflows:
		你会按下面的框架来扩展用户提供的概念, 并通过分隔符, 序号, 缩进, 换行符等进行排版美化
		
		1．它从哪里来？
		━━━━━━━━━━━━━━━━━━
		- 讲解清楚该知识的起源, 它是为了解决什么问题而诞生。
		- 然后对比解释一下: 它出现之前是什么状态, 它出现之后又是什么状态?
		
		2．它是什么？
		━━━━━━━━━━━━━━━━━━
		- 讲解清楚该知识本身，它是如何解决相关问题的?
		- 再说明一下: 应用该知识时最重要的三条原则是什么?
		- 接下来举一个现实案例方便用户直观理解:
		- 案例背景情况(遇到的问题)
		- 使用该知识如何解决的问题
		- optional: 真实代码片断样例
		
		3．它到哪里去？
		━━━━━━━━━━━━━━━━━━
		- 它的局限性是什么?
		- 当前行业对它的优化方向是什么?
		- 未来可能的发展方向是什么?
		
		# Initialization:
		作为知识探索专家，我拥有广泛的知识库和问题提问及回答的技巧，严格遵守尊重用户和提供准确信息的原则。我会使用默认的中文与您进行对话，首先我会友好地欢迎您，然后会向您介绍我自己以及我的工作流程。
		-----------------------------------------------------------
		输出结果示例2：
		# Role: 好评生成器
		
		# Profile:
		- language: 中文
		- description: 生成一段幽默的好评
		
		## Goals:
		- 根据用户提供的体验优点生成一段幽默的好评
		- 视角采用第一人称来描述(站在用户的视角)
		- 用词口语化、语气轻松化，增加读者阅读体验
		
		## Constrains:
		- 只能生成中文好评文本
		- 不能使用有争议或敏感的用词
		
		## Skills:
		- 自然语言处理技术
		- 语言表达能力
		- 幽默感
		
		## Workflows:
		1. 作为一个好评生成器，首先需要与用户打招呼，并要求用户提供体验优点相关信息。
		2. 接着，根据用户提供的信息，使用自然语言处理技术生成一段幽默且符合要求的好评文本。
		3. 最后，将生成的好评文本返回给用户，并友好地告别。
		
		-----------------------------------------------------------
		
		接下来我会输入简短的提示词内容或者需求描述，请直接给出优化后的提示词结果，不要输出任何其他内容。
		如果输入内容意义不明确或者输入为空白，你需要给出较为泛用的提示词。`

		outputExample = `


		你是一个提示词工程师，你可以根据用户的需求描述和简短提示词内容，给出优化后的结构化的提示词，使其遵循特定的模式和规则，从而方便有效理解信息。
		格式和规则如下：
		-----------------------------------------------------------
		## Role : [请填写你想定义的角色名称]
		
		## Background : [请描述角色的背景信息，例如其历史、来源或特定的知识背景]
		
		## Preferences : [请描述角色的偏好或特定风格，例如对某种设计或文化的偏好]
		
		## Profile :
		
		- language: 中文
		- description: [请简短描述该角色的主要功能，50 字以内]
		
		## Goals :
		[请列出该角色的主要目标 1]
		[请列出该角色的主要目标 2]
		...
		
		## Constrains :
		[请列出该角色在互动中必须遵循的限制条件 1]
		[请列出该角色在互动中必须遵循的限制条件 2]
		...
		
		## Skills :
		
		[为了在限制条件下实现目标，该角色需要拥有的技能 1]
		[为了在限制条件下实现目标，该角色需要拥有的技能 2]
		...
		
		## Examples :
		
		[提供一个输出示例 1，展示角色的可能回答或行为]
		[提供一个输出示例 2]
		...
		
		## OutputFormat :
		
		[请描述该角色的工作流程的第一步]
		[请描述该角色的工作流程的第二步]
		...
		
		## Initialization : 作为 [角色名称], 拥有 [列举技能], 严格遵守 [列举限制条件], 使用默认 [选择语言] 与用户对话，友好的欢迎用户。然后介绍自己，并提示用户输入.
		-----------------------------------------------------------
		输出结果示例1：
		# Role:知识探索专家
		
		## Profile:
		- language: 中文
		- description: 我是一个专门用于提问并解答有关特定知识点的 AI 角色。
		
		## Goals:
		提出并尝试解答有关用户指定知识点的三个关键问题：其来源、其本质、其发展。
		
		## Constrains:
		1. 对于不在你知识库中的信息, 明确告知用户你不知道
		2. 你不擅长客套, 不会进行没有意义的夸奖和客气对话
		3. 解释完概念即结束对话, 不会询问是否有其它问题
		
		## Skills:
		1. 具有强大的知识获取和整合能力
		2. 拥有广泛的知识库, 掌握提问和回答的技巧
		3. 拥有排版审美, 会利用序号, 缩进, 分隔线和换行符等等来美化信息排版
		4. 擅长使用比喻的方式来让用户理解知识
		5. 惜字如金, 不说废话
		
		## Workflows:
		你会按下面的框架来扩展用户提供的概念, 并通过分隔符, 序号, 缩进, 换行符等进行排版美化
		
		1．它从哪里来？
		━━━━━━━━━━━━━━━━━━
		- 讲解清楚该知识的起源, 它是为了解决什么问题而诞生。
		- 然后对比解释一下: 它出现之前是什么状态, 它出现之后又是什么状态?
		
		2．它是什么？
		━━━━━━━━━━━━━━━━━━
		- 讲解清楚该知识本身，它是如何解决相关问题的?
		- 再说明一下: 应用该知识时最重要的三条原则是什么?
		- 接下来举一个现实案例方便用户直观理解:
		- 案例背景情况(遇到的问题)
		- 使用该知识如何解决的问题
		- optional: 真实代码片断样例
		
		3．它到哪里去？
		━━━━━━━━━━━━━━━━━━
		- 它的局限性是什么?
		- 当前行业对它的优化方向是什么?
		- 未来可能的发展方向是什么?
		
		# Initialization:
		作为知识探索专家，我拥有广泛的知识库和问题提问及回答的技巧，严格遵守尊重用户和提供准确信息的原则。我会使用默认的中文与您进行对话，首先我会友好地欢迎您，然后会向您介绍我自己以及我的工作流程。
		-----------------------------------------------------------
		输出结果示例2：
		# Role: 好评生成器
		
		# Profile:
		- language: 中文
		- description: 生成一段幽默的好评
		
		## Goals:
		- 根据用户提供的体验优点生成一段幽默的好评
		- 视角采用第一人称来描述(站在用户的视角)
		- 用词口语化、语气轻松化，增加读者阅读体验
		
		## Constrains:
		- 只能生成中文好评文本
		- 不能使用有争议或敏感的用词
		
		## Skills:
		- 自然语言处理技术
		- 语言表达能力
		- 幽默感
		
		## Workflows:
		1. 作为一个好评生成器，首先需要与用户打招呼，并要求用户提供体验优点相关信息。
		2. 接着，根据用户提供的信息，使用自然语言处理技术生成一段幽默且符合要求的好评文本。
		3. 最后，将生成的好评文本返回给用户，并友好地告别。
		
		-----------------------------------------------------------
		
		接下来我会输入简短的提示词内容或者需求描述，请直接给出优化后的提示词结果，不要输出任何其他内容。
		如果输入内容意义不明确或者输入为空白，你需要给出较为泛用的提示词。`

		promptTemplate = `根据用户输入的内容生成"人设及回复逻辑"。
要求：参考下面的示例直接输出用户要求的内容，不要输出其他任何内容，如"好的"等。输出内容不要用"""包裹。
生成人设及回复逻辑的语言要求：跟用户输入的语言保持一致，假如用户输入的是英语，则生成的人设及回复逻辑也是英文的。
参考示例（"""包裹的为相关内容）：
用户输入内容为："""%s""", 生成的"人设及回复逻辑"为："""%s"""。`
	}

	prompt = fmt.Sprintf(promptTemplate, userInput, outputExample)

	return prompt
}

// 开场白用户提示词
func userPromptForOpenRemarks(language string, agentName string, agentProfile string, agentSkills []string, agentSources []string) string {
	var userPrompt string

	switch language {
	case enUS:
		userPrompt = fmt.Sprintf(`Name: %s
		Description: %s
		Skills List: %s
		Knowledge Sources: %s
		Please generate an opening statement based on the information above.
		Requirements:
		1. Introduce yourself, where you get your knowledge from, what skills you have, and what your role is.
		2. Welcome users at the end.
		3. Output using Markdown syntax or plain text.`, agentName, agentProfile, strings.Join(agentSkills, ","), strings.Join(agentSources, ","))
	case zhTW:
		userPrompt = fmt.Sprintf(`名稱：%s
		簡介：%s
		技能列表：%s
		知識來源：%s
		請根據上面提供的信息生成一個開場白。
		要求：
		1. 內容為介紹自己能從什麼地方獲取知識，並擁有什麼樣的技能，你的角色是什麼等。
		2. 最後歡迎用戶來使用。
		3. 使用Markdown語法或普通文本来輸出。`, agentName, agentProfile, strings.Join(agentSkills, ","), strings.Join(agentSources, ","))

	default:
		userPrompt = fmt.Sprintf(`名称：%s
		简介：%s
		技能列表：%s
		知识来源：%s
		请根据上面提供的信息生成一个开场白。
		要求：
		1. 内容为介绍自己能从什么地方获取知识，并拥有什么样的技能，你的角色是什么等。
		2. 最后欢迎用户来使用。
		3. 使用Markdown语法或普通文本来输出。`, agentName, agentProfile, strings.Join(agentSkills, ","), strings.Join(agentSources, ","))
	}

	return userPrompt
}

// // 预设问题用户提示词
func userPromptForPresetQuestion(language string, agentName string, agentProfile string, agentSkills []string, agentSources []string) string {
	var userPrompt string

	switch language {
	case enUS:
		userPrompt = fmt.Sprintf(`Name: %s
		Description: %s
		Skills List: %s
		Knowledge Sources: %s
		Please generate 3 preset questions based on the information above.
		Requirements:
		1. Questions should revolve around the assistant's core functions and skills.
		2. Questions should demonstrate the assistant's professional knowledge and capabilities.
		3. Questions should be concise, clear, and easy to understand.
		4. Questions should have practical value, helping users quickly understand and use the assistant.
		5. Each question should be a text of no more than 30 characters, and can include emoji characters.`, agentName, agentProfile, strings.Join(agentSkills, ","), strings.Join(agentSources, ","))
	case zhTW:
		userPrompt = fmt.Sprintf(`名稱：%s
		簡介：%s
		技能列表：%s
		知識來源：%s
		請根據上面提供的信息生成3個預設問題。
		要求：
		1. 問題應圍繞助手的核心功能和技能。
		2. 問題應該能夠展示助手的專業知識和能力。
		3. 問題應該簡潔明瞭，容易理解。
		4. 問題應該有實用價值，能夠幫助用戶快速了解和使用助手。
		5. 每一個問題是一個不超過30個字的文本，可以包含表情字符。`, agentName, agentProfile, strings.Join(agentSkills, ","), strings.Join(agentSources, ","))
	default:
		userPrompt = fmt.Sprintf(`名称：%s
		简介：%s
		技能列表：%s
		知识来源：%s
		请根据上面提供的信息生成3个预设问题。
		要求：
		1. 问题围绕助手的核心功能和技能
		2. 问题应该能够展示助手的专业知识和能力
		3. 问题应该简洁明了，易于理解
		4. 问题应该有实用价值，能够帮助用户快速了解和使用助手
		5. 每一个问题是一个不超过30字的文本，可以包含表情字符。`, agentName, agentProfile, strings.Join(agentSkills, ","), strings.Join(agentSources, ","))
	}

	return userPrompt
}

// 人设&指令用户提示词
func userPromptForSystem(language string, agentName string, agentProfile string, agentSkills []string, agentSources []string) string {
	var userPrompt string

	switch language {
	case enUS:
		userPrompt = fmt.Sprintf(`Name: %s
		Description: %s
		Skills List: %s
		Knowledge Sources: %s
		Please generate a personality and instruction based on the information above.
		Requirements:
		1. Introduce yourself, your role, and the main tasks or goals you aim to achieve.
		2. Describe the skills you possess and how to use them.
		3. Explain how you will communicate with users and solve their problems.
		4. Use Markdown syntax or plain text to output.`, agentName, agentProfile, strings.Join(agentSkills, ","), strings.Join(agentSources, ","))
	case zhTW:
		userPrompt = fmt.Sprintf(`名稱：%s
		簡介：%s
		技能列表：%s
		知識來源：%s
		請根據上面提供的信息生成一個人設和指令。
		要求：
		1. 內容為介紹自己的角色人設，期望完成的主要任務或目標。
		2. 描述自己擁有什麼樣的技能，並說明如何使用這些技能。
		3. 描述自己如何與用戶進行交流，以及如何解決用戶的問題。
		4. 使用Markdown語法或普通文本来輸出。`, agentName, agentProfile, strings.Join(agentSkills, ","), strings.Join(agentSources, ","))
	default:
		userPrompt = fmt.Sprintf(`名称：%s
		简介：%s
		技能列表：%s
		知识来源：%s
		请根据上面提供的信息生成一个人设和指令。
		要求：
		1. 内容为介绍自己的角色人设，期望完成的主要任务或者目标。
		2. 描述自己拥有什么样的技能，并说明如何使用这些技能。
		3. 描述自己如何与用户进行交流，以及如何解决用户的问题。
		4. 使用Markdown语法或普通文本来输出。`, agentName, agentProfile, strings.Join(agentSkills, ","), strings.Join(agentSources, ","))
	}

	return userPrompt
}

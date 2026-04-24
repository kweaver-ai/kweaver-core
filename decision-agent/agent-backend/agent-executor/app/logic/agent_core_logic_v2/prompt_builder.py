from typing import Optional
import json

from opentelemetry.trace import Span

from app.common.config import Config

from app.logic import plan_mode_logic
from app.utils.observability.trace_wrapper import internal_span

from app.domain.vo.agentvo import AgentConfigVo


from .trace import span_set_attrs

# Skill usage rules injected into every explore prompt so that the model
# knows how to use the three built-in skill contract tools.
# 注意:这些规则需要作为 Dolphin 注释格式(以 # 开头)或在块内部
_SKILL_USAGE_RULES = """## Built-in Skill Capabilities

You have access to three built-in tools for working with skills:

### 1. builtin_skill_load(skill_id)
- **Purpose**: Load a skill package and get its documentation
- **When to use**: Always call this first when you have a skill_id
- **Returns**: The full SKILL.md content plus lists of available scripts and reference files

### 2. builtin_skill_read_file(skill_id, file_path)
- **Purpose**: Read a specific file inside the skill package
- **When to use**: Optional. Only call after you have obtained a file path from builtin_skill_load or from SKILL.md
- **Note**: One file per call; cannot batch

### 3. builtin_skill_execute_script(skill_id, script_path)
- **Purpose**: Execute a script from the skill package
- **When to use**: Optional. Only call after reading SKILL.md and deciding that script execution is needed
- **Note**: Not all skills require script execution

### Usage Guidelines
1. If you have a skill_id, **always start with** `builtin_skill_load(skill_id)`
2. After reading SKILL.md, decide independently whether to call `read_file`, `execute_script`, both, or neither
3. Both `builtin_skill_read_file` and `builtin_skill_execute_script` are **optional steps**

---
"""


class PromptBuilder:
    def __init__(self, agent_config: AgentConfigVo, temp_files: dict[str, list]):
        self.agent_config = agent_config
        self.temp_files = temp_files

    @staticmethod
    def _is_skill_enabled() -> bool:
        """是否将 _SKILL_USAGE_RULES 拼接到系统提示词。

        从配置文件的 features.skill_enabled 读取，
        默认为 true，仅当配置为 false 时才不拼接。
        """
        return Config.features.skill_enabled

    @internal_span()
    async def build(self, span: Optional[Span] = None) -> str:
        span_set_attrs(
            span=span,
            agent_run_id=self.agent_config.agent_run_id,
            agent_id=self.agent_config.agent_id,
        )

        is_skill_enabled = self._is_skill_enabled()

        if self.agent_config.is_dolphin_mode:
            # In dolphin_mode, do NOT directly append _SKILL_USAGE_RULES to dolphin_prompt
            # because it's Markdown text (not valid Dolphin syntax) and will cause parse errors.
            # The Skill Usage Rules will be automatically added by explore_block_v2.py
            # through the system_prompt parameter when skill_enabled=true.
            dolphin_prompt = ""

            for pre_dolphin in self.agent_config.pre_dolphin:
                if not pre_dolphin.get("enabled", False):
                    continue

                # Skip temp-file processing block when the user has no uploaded files
                # 如果开启了临时区，但是用户没有上次临时区文件，则不构造临时区文件处理 dophin
                if pre_dolphin.get("key") == "temp_file_process" and not has_temp_files(
                    self.temp_files
                ):
                    continue

                dolphin_prompt += pre_dolphin["value"] + "\n"

            dolphin_prompt += self.agent_config.dolphin + "\n"

            for post_dolphin in self.agent_config.post_dolphin:
                if post_dolphin.get("enabled"):
                    dolphin_prompt += post_dolphin["value"] + "\n"
        else:
            memory_prompt = self.search_memory_prompt()
            self.temp_file_prompt = self.temp_file_prompt()
            self.doc_retrieval_prompt = self.get_doc_retrieval_prompt()
            self.graph_retrieval_prompt = self.get_graph_retrieval_prompt()
            context_prompt = self.get_context_prompt()
            related_questions_prompt = self.get_related_questions_prompt()

            # Append skill usage rules to the auto-generated explore system prompt
            # 调整顺序：技能使用规则在前，用户的系统提示词在后
            explore_system_prompt = ""
            if self.agent_config.is_plan_mode():
                explore_system_prompt = plan_mode_logic.get_system_prompt_with_plan(
                    self.agent_config.system_prompt
                )
                self.agent_config.append_task_plan_agent()
            else:
                # 确保 system_prompt 不为 None，避免 repr(None) 变成字符串 'None'
                explore_system_prompt = self.agent_config.system_prompt or ""

            if is_skill_enabled:
                # Dolphin 会将标记符替换为通用的 Tools Usage Guidelines
                # 最终顺序：_SKILL_USAGE_RULES -> Tools Usage Guidelines -> 用户提示词
                if explore_system_prompt and explore_system_prompt.strip():
                    # 用户设置了有效的角色指令（非空且不只是空白字符）
                    explore_system_prompt = (
                        f"{_SKILL_USAGE_RULES}\n"
                        f"{{__BUILTIN_SKILL_RULES__}}\n\n"
                        f"{explore_system_prompt}"
                    )
                else:
                    # 用户未设置角色指令或角色指令为空/空白
                    explore_system_prompt = (
                        f"{_SKILL_USAGE_RULES}\n{{__BUILTIN_SKILL_RULES__}}"
                    )

            history_enabled = (
                not self.agent_config.react_disable_history_in_a_conversation()
            )

            disable_llm_cache = (
                Config.features.disable_dolphin_sdk_llm_cache
                or self.agent_config.react_disable_llm_cache()
            )

            # no_cache: 是否禁用缓存（布尔值）
            # no_cache = disable_llm_cache

            # 构造 explore dolphin 语句
            # 使用 json.dumps 转义字符串，Dolphin SDK 会自动反转义 \n 为真正的换行符
            escaped_system_prompt = json.dumps(explore_system_prompt, ensure_ascii=False)
            explore_prompt = f"""/explore/(system_prompt={escaped_system_prompt}, history={history_enabled}, no_cache={disable_llm_cache})$query -> answer\n"""

            dolphin_prompt = (
                memory_prompt
                + self.temp_file_prompt
                + self.doc_retrieval_prompt
                + self.graph_retrieval_prompt
                + context_prompt
                + explore_prompt
                + related_questions_prompt
            )

        dolphin_prompt += """\n'' -> self_config\n"""
        dolphin_prompt += """'' -> header\n"""

        return dolphin_prompt

    @internal_span()
    def get_doc_retrieval_prompt(self, span: Span = None) -> str:
        """构建文档召回提示词"""

        span_set_attrs(
            span=span,
            agent_run_id=self.agent_config.agent_run_id or "",
            agent_id=self.agent_config.agent_id or "",
        )

        return ""

    @internal_span()
    def get_graph_retrieval_prompt(self, span: Span = None) -> str:
        """构建图谱召回提示词"""

        span_set_attrs(
            span=span,
            agent_run_id=self.agent_config.agent_run_id or "",
            agent_id=self.agent_config.agent_id or "",
        )

        return ""

    @internal_span()
    def get_context_prompt(self, span: Span = None) -> str:
        if span and span.is_recording():
            span.set_attribute(
                "session_id",
                (
                    self.agent_config.agent_run_id
                    if self.agent_config.agent_run_id
                    else ""
                ),
            )
            span.set_attribute(
                "agent_id",
                self.agent_config.agent_id if self.agent_config.agent_id else "",
            )
        """构建上下文提示词"""
        prompt = """"如果有参考文档，结合参考文档回答用户的问题。如果没有参考文档，根据用户的问题回答。\\n" -> reference\n"""

        if self.doc_retrieval_prompt:
            prompt += """/if/ 'result' in $doc_retrieval_res['answer'] and $doc_retrieval_res['answer']['result']:
    $reference + "文档召回的内容：" + $doc_retrieval_res['answer']['result'] + "\\n" -> reference
/end/\n"""

        if self.graph_retrieval_prompt:
            prompt += """/if/ 'result' in $graph_retrieval_res['answer'] and $graph_retrieval_res['answer']['result']:
    $reference + "业务知识网络召回的内容：" + $graph_retrieval_res['answer']['result'] + "\\n" -> reference
/end/\n"""

        if self.temp_file_prompt:
            temp_file_name = list(self.temp_files.keys())[0]
            prompt += f"""$reference + "用户上传的文件内容：" + ${temp_file_name} + "\\n" -> reference\n"""

        prompt += """$reference +  "用户的问题为：" + $query -> query\n"""
        return prompt

    @internal_span()
    def search_memory_prompt(self, span: Span = None) -> str:
        """构建搜索记忆提示词"""

        span_set_attrs(
            span=span,
            agent_run_id=self.agent_config.agent_run_id or "",
            agent_id=self.agent_config.agent_id or "",
        )

        if not self.agent_config.memory or not self.agent_config.memory.get(
            "is_enabled"
        ):
            return ""

        memory_prompt = f"""@search_memory(query=$query, user_id=$header['x-account-id'], limit={Config.memory.limit}, threshold={Config.memory.threshold}, rerank_thread={Config.memory.rerank_threshold}) -> relevant_memories\n"""
        memory_prompt += (
            """json.dumps($relevant_memories["answer"]["result"]) -> memory_str\n"""
        )
        memory_prompt += """$_history + [{"role": "system", "content": "Relevant memories: " + $memory_str}] -> _history\n"""
        memory_prompt += """'' -> relevant_memories\n"""
        memory_prompt += """'' -> memory_str\n"""

        return memory_prompt

    @internal_span()
    def temp_file_prompt(self, span: Span = None) -> str:
        """构建临时区文件内容提示词"""

        span_set_attrs(
            span=span,
            agent_run_id=self.agent_config.agent_run_id or "",
            agent_id=self.agent_config.agent_id or "",
        )

        temp_file_prompt = ""
        for temp_file_name, temp_file_info in self.temp_files.items():
            if temp_file_info:
                temp_file_prompt += f"""@process_file_intelligent(query=$query, file_infos=${temp_file_name}) -> {temp_file_name}\n ${temp_file_name}['answer'] -> {temp_file_name}"""

        return temp_file_prompt

    @internal_span()
    def get_related_questions_prompt(self, span: Span = None) -> str:
        """构建相关问题提示词"""
        if span and span.is_recording():
            span.set_attribute(
                "session_id",
                (
                    self.agent_config.agent_run_id
                    if self.agent_config.agent_run_id
                    else ""
                ),
            )
            span.set_attribute(
                "agent_id",
                self.agent_config.agent_id if self.agent_config.agent_id else "",
            )
        related_questions_prompt = ""
        if (
            self.agent_config.related_question
            and self.agent_config.related_question.get("is_enabled")
        ):
            related_questions_prompt += """
/prompt/(flags='{"debug":true}')请根据原始用户问题和上下文信息，更进一步的生成3个问题和答案对。所生成的3个问题与原始用户问题呈递进关系，3个问题之间则相互独立，用户可以使用这些问题深入挖掘上下文中的话题。
===
$query
===
要求：
1. 根据上下文信息生成问题答案对，禁止推测；若上下文信息为空。则直接返回空。
2. 生成的问题不要和原始用户问题重复。
3. 确保生成的问题主谓宾完整，长度不超过25个字。
4. 不要输出问题对应的答案。
5. 输出格式必须是纯JSON列表，仅包含字符串类型：
["第一个问题", "第二个问题", "第三个问题"]
6. 如果无法生成问题，则返回空列表 []
7. 示例:
原始问题:你能帮我写一首描写春天的唐诗吗？
相关问题:["春天的唐诗中有哪些典型的意象？", "这首唐诗中如何体现春天的生机与活力？", "唐诗中春天的描写与现代人对春天的感受有何不同？"]
8. 不要输出多余的内容，仅输出JSON格式的列表。
-> related_questions
eval($related_questions.answer) -> related_questions
"""
        return related_questions_prompt


def has_temp_files(temp_files: dict[str, list]) -> bool:
    """
     检查用户是否上传了临时文件
    如果 tmp_files 为空，或者所有的键对应的值都是空列表，则返回 False
    """
    if not temp_files:
        return False
    return any(values for values in temp_files.values())

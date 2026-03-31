from typing import Optional

from opentelemetry.trace import Span

from app.common.config import Config

from app.logic import plan_mode_logic
from app.utils.observability.trace_wrapper import internal_span

from app.domain.vo.agentvo import AgentConfigVo


from .trace import span_set_attrs


class PromptBuilder:
    def __init__(self, agent_config: AgentConfigVo, temp_files: dict[str, list]):
        self.agent_config = agent_config
        self.temp_files = temp_files

    @internal_span()
    async def build(self, span: Optional[Span] = None) -> str:
        span_set_attrs(
            span=span,
            agent_run_id=self.agent_config.agent_run_id,
            agent_id=self.agent_config.agent_id,
        )

        if self.agent_config.is_dolphin_mode:
            dolphin_prompt = ""

            for pre_dolphin in self.agent_config.pre_dolphin:
                if not pre_dolphin.get("enabled", False):
                    continue

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

            explore_system_prompt = ""
            if self.agent_config.is_plan_mode():
                explore_system_prompt = plan_mode_logic.get_system_prompt_with_plan(
                    self.agent_config.system_prompt
                )
                self.agent_config.append_task_plan_agent()
            else:
                explore_system_prompt = self.agent_config.system_prompt

            # tool_names = await self.get_tool_names()
            # explore_prompt = f"""/explore/(tools={repr(tool_names)}, system_prompt={repr(explore_system_prompt)}, history=True)$query -> answer\n"""
            explore_prompt = f"""/explore/(system_prompt={repr(explore_system_prompt)}, history=True)$query -> answer\n"""

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

"""
句子边界检测器

检测文本中的句子边界位置，支持中英文标点，
处理引号内的句子结束符。
"""

import re
from typing import List


class SentenceBoundaryDetector:
    """
    句子边界检测器。

    检测句子结束位置的索引，支持：
    - 中文标点：。！？；\n
    - 英文标点: . ! ? ;
    - 引号处理：引号内的句号不作为句子边界
    """

    SENTENCE_END_PATTERN = re.compile(
        r"[。！？;；\n]"
        r"|(?<=[a-zA-Z])\.(?=\s|$|[\u4e00-\u9fff])"
        r"|(?<=[a-zA-Z])[!?](?=\s|$)"
    )

    QUOTE_PAIRS = {
        '"': '"',
        "'": "'",
        "「": "」",
        "『": "』",
        "【": "】",
        "（": "）",
        "(": ")",
    }

    def __init__(self):
        self._quote_stack: List[str] = []

    def detect(self, text: str) -> List[int]:
        """
        返回句子结束位置的索引列表。

        Args:
            text: 输入文本

        Returns:
            句子结束字符的位置索引列表
        """
        if not text:
            return []

        boundaries = []
        self._quote_stack = []

        for match in self.SENTENCE_END_PATTERN.finditer(text):
            pos = match.end() - 1
            char = text[pos]

            if self._is_inside_quotes(char, text, pos):
                continue

            boundaries.append(pos)

        return boundaries

    def _is_inside_quotes(
        self, char: str, text: str, pos: int
    ) -> bool:
        """
        检查当前位置是否在引号内部。

        通过维护引号栈来判断是否在配对引号内部。
        """
        for i in range(pos + 1):
            c = text[i]

            if c in self.QUOTE_PAIRS:
                close_quote = self.QUOTE_PAIRS[c]
                if (
                    self._quote_stack
                    and self._quote_stack[-1] == c
                ):
                    self._quote_stack.pop()
                else:
                    self._quote_stack.append(c)
            elif c in self.QUOTE_PAIRS.values():
                if self._quote_stack:
                    expected_open = None
                    for open_q, close_q in self.QUOTE_PAIRS.items():
                        if close_q == c:
                            expected_open = open_q
                            break

                    if (
                        expected_open
                        and self._quote_stack
                        and self._quote_stack[-1] == expected_open
                    ):
                        self._quote_stack.pop()

        return len(self._quote_stack) > 0

    def split_to_sentences(self, text: str) -> List[str]:
        """
        将文本按句子边界分割为句子列表。

        Args:
            text: 输入文本

        Returns:
            句子字符串列表
        """
        boundaries = self.detect(text)

        if not boundaries:
            return [text] if text else []

        sentences = []
        start = 0

        for end_pos in boundaries:
            sentences.append(text[start : end_pos + 1])
            start = end_pos + 1

        if start < len(text):
            remaining = text[start:]
            if remaining.strip():
                sentences.append(remaining)

        return sentences

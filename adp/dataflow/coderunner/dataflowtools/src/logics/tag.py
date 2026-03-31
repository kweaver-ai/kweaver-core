import ahocorasick
import pinyin
# from drivenadapters.efast import EfastAdapter
# from utils.utils import *
# from errors.errors import *
from typing import List, Any, Dict, Set

class TagExtractionService:
    def __init__(self,host,access_token) -> None:
        return

    async def extraction_by_rule(self, requestBody):
        target = requestBody["target"]
        content = target["content"]
        tags_logic = requestBody["rules"]

        automaton = await self.build_automaton(tags_logic)
        extracted_tags = await self.extract_tags(content, automaton, tags_logic)
        return list(extracted_tags)

    async def build_automaton(self, tags_logic: List[dict]) -> Any:
        automaton = ahocorasick.Automaton()
        for item in tags_logic:
            tag_path = item["tag_path"]
            rules = item["rule"]
            or_groups = rules.get("or", [])
            not_words = rules.get("not", [])
            
            for group_index, and_group in enumerate(or_groups):
                for word in and_group:
                    key = (tag_path, group_index)  # 使用(tag_path, group_index)作为key以区分不同的or组
                    if word not in automaton:
                        automaton.add_word(word, set())
                    automaton.get(word).add((key, "and"))
            
            for word in not_words:
                if word not in automaton:
                    automaton.add_word(word, set())
                automaton.get(word).add((tag_path, "not"))
        automaton.make_automaton()
        return automaton

    async def extract_tags(self, text: str, automaton: Any, tags_logic: List[dict]) -> set:
        matches = {}
        excludes = set()
        
        for end_index, original_set in automaton.iter(text):
            for key, type_ in original_set:
                if type_ == "not":
                    excludes.add(key)
                else:
                    tag_path, group_index = key
                    if tag_path not in matches:
                        matches[tag_path] = set()
                    matches[tag_path].add(group_index)
        
        extracted_tags = set()
        for item in tags_logic:
            tag_path = item["tag_path"]
            rules = item["rule"]
            if tag_path in excludes:
                continue
            or_groups = rules.get("or", [])
            matched_groups = matches.get(tag_path, set())
            
            # 如果匹配的组索引覆盖了至少一个完整的and组，则添加标签
            for group_index, _ in enumerate(or_groups):
                if group_index in matched_groups:
                    extracted_tags.add(tag_path)
                    break  # or关系：匹配到任意一组即可

        return extracted_tags

# # 定义标签和同义词的逻辑关系，增加"not"选项
# tags_logic = {
#     "科技": {
#         "or": [["技术"], ["学生"]],
#         "not": ["1"]  # 假设"科技"标签不应当包含"古典音乐"这个词
#     },
#     "教育": {
#         "or": [["学习", "课程"], ["教育", "教学"]],
#         "not": ["足球"]  # 假设"教育"标签不应当包含"足球"这个词
#     },
# }

# # 构建AC自动机
# automaton = build_automaton(tags_logic)

# # 示例文本
# text = "这门课程非常适合想要深入学习信息技术和学习的学生。"

# # 提取标签
# extracted_tags = extract_tags(text, automaton, tags_logic)
# print("Extracted Tags:", extracted_tags)


class TagExtractionV2Service:
    def get_path_pinyin(self, tag_path):
        return '/'.join(pinyin.get(segment, format="numerical") for segment in tag_path.split('/'))
        
    async def extraction_by_rule(self, request_body: Dict[str, Any]) -> List[Dict[str, Any]]:
        target = request_body.get("target", {})
        rules = request_body.get("rules", [])

        # Sort rules by full tag_path using pinyin
        
            
        rules.sort(key=lambda x: self.get_path_pinyin(x['tag_path']))

        if not rules:
            return []

        automaton = await self.build_automaton(rules)
        extracted_tags = await self.extract_tags(target["content"], automaton, rules)
        sorted_data = sorted(extracted_tags, key=lambda x: x['weight'], reverse=True)
        return sorted_data

    async def build_automaton(self, tags_logic: List[Dict[str, Any]]) -> ahocorasick.Automaton:
        automaton = ahocorasick.Automaton()
        for item in tags_logic:
            tag_path = item["tag_path"]
            rules = item.get("rule", {})
            or_groups = rules.get("or", [])
            not_words = rules.get("not", [])

            if tag_path == "":
                continue

            # Only add tag_path as a match condition if there are no other conditions
            if not or_groups and not not_words:
                key = (tag_path, -1, len(tag_path.split('/')[-1]))  # Using group_index = -1 to distinguish
                if tag_path.split('/')[-1] not in automaton:
                    automaton.add_word(tag_path.split('/')[-1], set())
                automaton.get(tag_path.split('/')[-1]).add((key, "and"))

            for group_index, and_group in enumerate(or_groups):
                for word in and_group:
                    key = (tag_path, group_index, len(word))  # Using (tag_path, group_index, len(word)) to distinguish
                    if word not in automaton:
                        automaton.add_word(word, set())
                    automaton.get(word).add((key, "and"))

            for word in not_words:
                key = (tag_path, len(word))
                if word not in automaton:
                    automaton.add_word(word, set())
                automaton.get(word).add((key, "not"))
        automaton.make_automaton()
        return automaton

    async def extract_tags(self, text: str, automaton: ahocorasick.Automaton, tags_logic: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        matches: Dict[str, Dict[int, Set[str]]] = {}
        excludes: Set[str] = set()

        for end_index, original_set in automaton.iter(text):
            for key, type_ in original_set:
                if type_ == "not":
                    tag_path, word_length = key
                    excludes.add(tag_path)  # 仅将被标记为 "not" 的标签路径添加到排除列表
                else:
                    tag_path, group_index, word_length = key
                    word = text[end_index + 1 - word_length:end_index + 1]

                    if tag_path not in matches:
                        matches[tag_path] = {}
                    if group_index not in matches[tag_path]:
                        matches[tag_path][group_index] = set()
                    matches[tag_path][group_index].add(word)

        extracted_tags: List[Dict[str, Any]] = []
        for item in tags_logic:
            tag_path = item["tag_path"]
            rules = item.get("rule", {})
            
            # 检查是否在排除列表中
            if tag_path in excludes:
                continue  # 仅排除确实不应提取的标签

            or_groups = rules.get("or", [])
            matched_groups = matches.get(tag_path, {})

            # Count how many AND groups are fully matched
            fully_matched_count = sum(
                1 for group_index, and_group in enumerate(or_groups)
                if all(word in matched_groups.get(group_index, set()) for word in and_group)
            )

            # If there are no explicit rules, count occurrences of the tag name
            if not or_groups and not rules.get("not", []):
                tag_name = tag_path.split('/')[-1]
                fully_matched_count = text.count(tag_name)

            if fully_matched_count > 0:
                extracted_tags.append({"tag": tag_path, "weight": fully_matched_count})

        # Check for duplicate tags and re-evaluate based on parent paths
        final_tags = []
        parent_paths = set()
        leaf_nodes = {}

        for tag in extracted_tags:
            tag_path = tag["tag"]
            parent_path = "/".join(tag_path.split('/')[:-1])
            leaf_node = tag_path.split('/')[-1]

            if parent_path:
                parent_automaton = ahocorasick.Automaton()
                parent_automaton.add_word(parent_path.split('/')[-1], parent_path)
                parent_automaton.make_automaton()
                for end_index, original_set in parent_automaton.iter(text):
                    if parent_path in original_set:
                        parent_paths.add(parent_path)
                        break

            if leaf_node not in leaf_nodes:
                leaf_nodes[leaf_node] = []
            leaf_nodes[leaf_node].append(tag)

        for tag in extracted_tags:
            tag_path = tag["tag"]
            parent_path = "/".join(tag_path.split('/')[:-1])
            leaf_node = tag_path.split('/')[-1]
            if not parent_path or parent_path in parent_paths:
                final_tags.append(tag)
            elif len(leaf_nodes[leaf_node]) == 1:
                final_tags.append(tag)
                
        # Ensure at least one tag is output if there are extracted tags
        if extracted_tags and not final_tags:
            final_tags.append(max(extracted_tags, key=lambda x: x['weight']))
        print(leaf_nodes)
        for leaf_node in leaf_nodes.keys():
            if all(leaf_node not in tag["tag"].split('/')[-1] for tag in final_tags):
                final_tags.append(max(leaf_nodes[leaf_node], key=lambda x: x['weight']))
        # Sort final_tags by weight and then by the pinyin of the tag
        final_tags.sort(key=lambda x: (-x['weight'], self.get_path_pinyin(x['tag'])))

        return final_tags
        return final_tags

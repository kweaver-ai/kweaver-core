# -*- coding:UTF-8 -*-
import json

class AssertTools:
    '''
    校验工具
    '''
    def is_descending_str(lst):
        """判断列表中元素是否降序排列"""
        if len(lst) <= 1:
            return True
        for i in range(len(lst) - 1):
            str1 = str(lst[i])
            str2 = str(lst[i + 1])
            # 逐位比较字符
            for c1, c2 in zip(str1, str2):
                if ord(c1) < ord(c2):
                    return False  # 前字符 < 后字符，不满足降序
                elif ord(c1) > ord(c2):
                    break  # 前字符 > 后字符，满足降序，继续下一对字符串
            else:
                # 如果是因为短字符串遍历完退出循环，说明短字符串是长字符串的前缀
                # 降序要求：短字符串必须 >= 长字符串，因此短字符串长度不能更长
                if len(str1) < len(str2):
                    return False
        return True

    def is_ascending_str(lst):
        """判断列表中元素是否正序排列"""
        if len(lst) <= 1:
            return True
        for i in range(len(lst) - 1):
            str1 = str(lst[i])
            str2 = str(lst[i])
            # 逐位比较字符
            for c1, c2 in zip(str1, str2):
                if ord(c1) > ord(c2):
                    return False
                elif ord(c1) < ord(c2):
                    break  # 当前位已满足前小后大，继续比较下一个字符串
            else:
                # 如果是因为短字符串遍历完退出循环，说明短字符串是长字符串的前缀，短字符串更小
                if len(str1) > len(str2):
                    return False
        return True
    
    def has_duplicates(lst):
        """判断列表中是否存在重复元素"""
        return len(lst) != len(set(lst))
    
    def are_lists_equal(list1, list2):
        # 对每个元素进行 JSON 序列化并排序后比较
        sorted_list1 = sorted(json.dumps(item, sort_keys=True) for item in list1)
        sorted_list2 = sorted(json.dumps(item, sort_keys=True) for item in list2)
        return sorted_list1 == sorted_list2

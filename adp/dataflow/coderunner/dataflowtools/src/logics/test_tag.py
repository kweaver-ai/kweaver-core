import unittest
import asyncio
from tag import TagExtractionV2Service

class TestTagExtractionV2Service(unittest.TestCase):
    def setUp(self):
        self.service = TagExtractionV2Service()

    def run_async(self, coro):
        return asyncio.run(coro)

    def test_single_and_group_match(self):
        request_body = {
            "target": {"content": "This is a test content with word1 and word2."},
            "rules": [
                {
                    "tag_path": "tag1",
                    "rule": {
                        "or": [
                            ["word1", "word2"]
                        ],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [{"tag": "tag1", "weight": 1}]
        self.assertEqual(result, expected)

    def test_multiple_and_groups_match(self):
        request_body = {
            "target": {"content": "This is a test content with word1 and word2. Also with word3."},
            "rules": [
                {
                    "tag_path": "tag1",
                    "rule": {
                        "or": [
                            ["word1", "word2"],
                            ["word3"]
                        ],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [{"tag": "tag1", "weight": 2}]
        self.assertEqual(result, expected)

    def test_not_match(self):
        request_body = {
            "target": {"content": "This is a test content with word1 and word2. Also with word4."},
            "rules": [
                {
                    "tag_path": "tag1",
                    "rule": {
                        "or": [
                            ["word1", "word2"],
                            ["word3"]
                        ],
                        "not": ["word4"]
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = []
        self.assertEqual(result, expected)

    def test_tag_name_match(self):
        request_body = {
            "target": {"content": "This content includes tag1."},
            "rules": [
                {
                    "tag_path": "a/tag1",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [{"tag": "a/tag1", "weight": 1}]
        self.assertEqual(result, expected)

    def test_and_group_not_fully_matched(self):
        request_body = {
            "target": {"content": "This is a test content with word1 only."},
            "rules": [
                {
                    "tag_path": "tag1",
                    "rule": {
                        "or": [
                            ["word1", "word2"]
                        ],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = []
        self.assertEqual(result, expected)

    def test_multiple_tags(self):
        request_body = {
            "target": {"content": "This is a test content with word1 and word2. Also with word3."},
            "rules": [
                {
                    "tag_path": "tag1",
                    "rule": {
                        "or": [
                            ["word1", "word2"]
                        ],
                        "not": []
                    }
                },
                {
                    "tag_path": "tag2",
                    "rule": {
                        "or": [
                            ["word3"]
                        ],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "tag1", "weight": 1},
            {"tag": "tag2", "weight": 1}
        ]
        self.assertEqual(result, expected)

    def test_duplicate_tags_with_parent_check(self):
        request_body = {
            "target": {"content": "This is a test content with word1 and word2. It belongs to industry and category."},
            "rules": [
                {
                    "tag_path": "tag/industry/solution",
                    "rule": {
                        "or": [
                            ["word1", "word2"]
                        ],
                        "not": []
                    }
                },
                {
                    "tag_path": "tag/category/solution",
                    "rule": {
                        "or": [
                            ["word1", "word2"]
                        ],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "tag/category/solution", "weight": 1},
            {"tag": "tag/industry/solution", "weight": 1}
        ]
        self.assertEqual(result, expected)

    def test_duplicate_tags_with_no_parent_check(self):
        request_body = {
            "target": {"content": "This is a test content with word1 and word2. It belongs to industry."},
            "rules": [
                {
                    "tag_path": "tag/industry1/solution",
                    "rule": {
                        "or": [
                            ["word1", "word2"]
                        ],
                        "not": []
                    }
                },
                {
                    "tag_path": "tag/category1/solution",
                    "rule": {
                        "or": [
                            ["word1", "word2"]
                        ],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "tag/category1/solution", "weight": 1}
        ]
        self.assertEqual(result, expected)
        
    def test_duplicate_tags_with_one_parent_check(self):
        request_body = {
            "target": {"content": "This is a test content with word1 and word2. It belongs to industry."},
            "rules": [
                {
                    "tag_path": "tag/industry/solution",
                    "rule": {
                        "or": [
                            ["word1", "word2"]
                        ],
                        "not": []
                    }
                },
                {
                    "tag_path": "tag/category/solution",
                    "rule": {
                        "or": [
                            ["word1", "word2"]
                        ],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "tag/industry/solution", "weight": 1}
        ]
        self.assertEqual(result, expected)
        
    def test_duplicate_tags_with_one_parent_check_with_weight(self):
        request_body = {
            "target": {"content": "本次演讲主题央企-央企"},
            "rules": [
                {
                    "tag_path": "tag/爱是/央企",
                    "rule": {
                        "or": [
                        ],
                        "not": []
                    }
                },
                {
                    "tag_path": "tag/华为/央企",
                    "rule": {
                        "or": [

                        ],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "tag/爱是/央企", "weight": 2}
        ]
        self.assertEqual(result, expected)
        
    def test_duplicate_tags_with_one_parent_check_with_sort(self):
        request_body = {
            "target": {"content": "本次演讲主题央企"},
            "rules": [
                {
                    "tag_path": "tag/爱是/央企",
                    "rule": {
                        "or": [
                        ],
                        "not": []
                    }
                },
                {
                    "tag_path": "tag/华为/央企",
                    "rule": {
                        "or": [
                        ],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "tag/爱是/央企", "weight": 1}
        ]
        self.assertEqual(result, expected)
        
    def test_duplicate_tags_with_one_parent_check_with_sort2(self):
        request_body = {
            "target": {"content": "本次演讲主题央企问题如下，武警部队"},
            "rules": [
                {
                    "tag_path": "爱数统一标签/行业标签/央企",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
                
                {
                    "tag_path": "爱数统一标签/行业标签/公共事业/武警",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
                {
                    "tag_path": "爱数统一标签/行业标签/公共事业/央企",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "爱数统一标签/行业标签/公共事业/武警", "weight": 1},
            {"tag": "爱数统一标签/行业标签/公共事业/央企", "weight": 1}
        ]
        self.assertEqual(result, expected)
        
    def test_duplicate_tags_with_one_parent_check_with_sort3(self):
        request_body = {
            "target": {"content": "AnyShare本次演讲主题央企问题如下，的的得得"},
            "rules": [
                {
                    "tag_path": "爱数统一标签/AnyShare/央企",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
                
                {
                    "tag_path": "爱数统一标签/AnyBackup/央企",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                }
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "爱数统一标签/AnyShare/央企", "weight": 1},
        ]
        self.assertEqual(result, expected)
        
    def test_duplicate_tags_with_one_parent_check_with_sort4(self):
        request_body = {
            "target": {"content": "AnyShare本次演讲主题央企问题如下，的的得得"},
            "rules": [
                {
                    "tag_path": "1/1/2/3/央企",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
                
                {
                    "tag_path": "00/1/2/央企",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
                {
                    "tag_path": "00/央企",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "00/1/2/央企", "weight": 1},
        ]
        self.assertEqual(result, expected)
        
    def test_duplicate_tags_with_one_parent_check_with_sort5(self):
        request_body = {
            "target": {"content": "百事可乐和可口可乐都是好可乐"},
            "rules": [
                {
                    "tag_path": "1/百事可乐",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
                {
                    "tag_path": "1/1/2/百事可乐",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
                {
                    "tag_path": "1/1/2/百事可乐/可口可乐",
                    "rule": {
                        "or": [],
                        "not": []
                    }
                },
            ]
        }
        result = self.run_async(self.service.extraction_by_rule(request_body))
        expected = [
            {"tag": "1/1/2/百事可乐", "weight": 1},
            {"tag": "1/1/2/百事可乐/可口可乐", "weight": 1},
        ]
        self.assertEqual(result, expected)

if __name__ == '__main__':
    unittest.main()

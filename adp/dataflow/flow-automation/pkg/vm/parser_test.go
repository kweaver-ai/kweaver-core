package vm

import (
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestParse(t *testing.T) {

	type Case struct {
		Title    string
		Input    string
		Expected []Token
	}

	var cases = []*Case{
		{
			Title: "文本",
			Input: "__a b c",
			Expected: []Token{
				{
					Type:  TokenLiteral,
					Value: "__a b c",
				},
			},
		},
		{
			Title: "简单变量",
			Input: "{{__abc123$DEF}}",
			Expected: []Token{
				{
					Type:  TokenVariable,
					Value: "__abc123$DEF",
				},
			},
		},
		{
			Title: "忽略变量两侧空格",
			Input: "{{ __a  }}",
			Expected: []Token{
				{
					Type:  TokenVariable,
					Value: "__a",
				},
			},
		},
		{
			Title: "文本和变量",
			Input: "abc{{def}}ghi",
			Expected: []Token{
				{
					Type:  TokenLiteral,
					Value: "abc",
				},
				{
					Type:  TokenVariable,
					Value: "def",
				},
				{
					Type:  TokenLiteral,
					Value: "ghi",
				},
			},
		},
		{
			Title: "变量访问链1",
			Input: "{{a.b.c}}",
			Expected: []Token{
				{
					Type:  TokenVariable,
					Value: "a",
					AccessList: []Token{
						{
							Type:  TokenLiteral,
							Value: "b",
						},
						{
							Type:  TokenLiteral,
							Value: "c",
						},
					},
				},
			},
		},
		{
			Title: "变量访问链包含数字",
			Input: "{{a.b.0}}",
			Expected: []Token{
				{
					Type:  TokenVariable,
					Value: "a",
					AccessList: []Token{
						{
							Type:  TokenLiteral,
							Value: "b",
						},
						{
							Type:  TokenLiteral,
							Value: 0,
						},
					},
				},
			},
		},
		{
			Title: "变量访问链包含数字文本",
			Input: "{{a.b.'0'}}",
			Expected: []Token{
				{
					Type:  TokenVariable,
					Value: "a",
					AccessList: []Token{
						{
							Type:  TokenLiteral,
							Value: "b",
						},
						{
							Type:  TokenLiteral,
							Value: "0",
						},
					},
				},
			},
		},
		{
			Title: "变量访问链包含符号或空格",
			Input: "{{a.b.\".\".'  '}}",
			Expected: []Token{
				{
					Type:  TokenVariable,
					Value: "a",
					AccessList: []Token{
						{
							Type:  TokenLiteral,
							Value: "b",
						},
						{
							Type:  TokenLiteral,
							Value: ".",
						},
						{
							Type:  TokenLiteral,
							Value: "  ",
						},
					},
				},
			},
		},
		{
			Title: "双花括号未正确闭合作为原始文本",
			Input: "{{a.b}",
			Expected: []Token{
				{
					Type:  TokenLiteral,
					Value: "{{a.b}",
				},
			},
		},
		{
			Title: "三花括号表示双花括号文本",
			Input: "{{{a}}}",
			Expected: []Token{
				{
					Type:  TokenLiteral,
					Value: "a",
				},
			},
		},
		{
			Title: "三对以上花括号",
			Input: "{{{{a}}}}",
			Expected: []Token{
				{
					Type:  TokenLiteral,
					Value: "{a}",
				},
			},
		},
		{
			Title: "三花括号未闭合优先作为变量解析",
			Input: "{{{a}}",
			Expected: []Token{
				{
					Type:  TokenLiteral,
					Value: "{",
				},
				{
					Type:  TokenVariable,
					Value: "a",
				},
			},
		},
		{
			Title: "三花括号后的更多右括号",
			Input: "{{{a}}}}",
			Expected: []Token{
				{
					Type:  TokenLiteral,
					Value: "a}",
				},
			},
		},
		{
			Title: "多组变量",
			Input: "{{a}}{{b}}",
			Expected: []Token{
				{
					Type:  TokenVariable,
					Value: "a",
				},
				{
					Type:  TokenVariable,
					Value: "b",
				},
			},
		},
		{
			Title: "多组变量",
			Input: "abc{{a}}def{{b}}ghi",
			Expected: []Token{
				{
					Type:  TokenLiteral,
					Value: "abc",
				},
				{
					Type:  TokenVariable,
					Value: "a",
				},
				{
					Type:  TokenLiteral,
					Value: "def",
				},
				{
					Type:  TokenVariable,
					Value: "b",
				},
				{
					Type:  TokenLiteral,
					Value: "ghi",
				},
			},
		},
		{
			Title: "未闭合引号",
			Input: "{{a.'1}}",
			Expected: []Token{
				{
					Type:  TokenVariable,
					Value: "a",
					AccessList: []Token{
						{
							Type:  TokenLiteral,
							Value: "'1",
						},
					},
				},
			},
		},
	}

	for i, kase := range cases {
		fmt.Printf("%d. %s\n", i, kase.Title)
		tokens := Parse(kase.Input)
		fmt.Printf("input: %v\n", kase.Input)
		fmt.Printf("expected: %v\n", kase.Expected)
		fmt.Printf("actual: %v\n", tokens)
		assert.Equal(t, tokens, kase.Expected)
	}
}

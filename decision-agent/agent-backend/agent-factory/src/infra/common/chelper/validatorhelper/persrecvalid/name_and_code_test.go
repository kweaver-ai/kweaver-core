package persrecvalid

import (
	"testing"
)

func TestNameRe(t *testing.T) {
	t.Parallel()

	// 测试字符串和预期结果的映射
	testCases := map[string]bool{
		"<":             false,
		"=":             true,
		"ValidN=ame123": true,
		"ValidName123":  true,
		"WayTooLongNameWayTooLongNameWayTooLongNameWayTooLongNameWayTooLongName": false,
		"ValidNameWithExactly64Characters12345678901234567890123456789012":       true,
		// in end
		"InvalidName?":  false,
		"InvalidName<":  false,
		"InvalidName>":  false,
		"InvalidName|":  false,
		"InvalidName:":  false,
		"InvalidName*":  false,
		"InvalidName/":  false,
		"InvalidName\\": false,
		// 并且不能以空格开头和结尾，包括单个空格（这个前端应该会做trim处理）
		" NameWithLeadingSpace":             false,
		"NameWithTrailingSpace ":            false,
		" NameWithLeadingAndTrailingSpace ": false,
		" ":                                 false,
		// in middle
		"Invalid?Name":  false,
		"Invalid<Name":  false,
		"Invalid>Name":  false,
		"Invalid|Name":  false,
		"Invalid:Name":  false,
		"Invalid*Name":  false,
		"Invalid/Name":  false,
		"Invalid\\Name": false,
		// in start
		"?InvalidName":  false,
		"<InvalidName":  false,
		">InvalidName":  false,
		"|InvalidName":  false,
		":InvalidName":  false,
		"*InvalidName":  false,
		"/InvalidName":  false,
		"\\InvalidName": false,
		"valid Name":    true,
	}

	for testString, expectedResult := range testCases {
		result := NameRe.MatchString(testString)
		if result != expectedResult {
			t.Errorf("Expected %v for %v, but got %v", expectedResult, testString, result)
		}
	}
}

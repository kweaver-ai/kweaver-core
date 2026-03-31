package data

import (
	"errors"
	"fmt"
	"strings"
)

type Selector struct {
	Key    string
	Op     SelectorOp
	Values []string
}

type SelectorOp string

const (
	SelectorOpEqual SelectorOp = "="
	SelectorOpIn    SelectorOp = "in"
)

func PareSelectors(selector string) (selectors []Selector, err error) {
	if selector == "" {
		return nil, errors.New("selector expression can not be empty")
	}

	idx, err := scanAllSplits(selector)
	if err != nil {
		return nil, err
	}
	var rs []Selector
	selectorExprs := splitStringsWithIdx(selector, idx)
	for i := range selectorExprs {
		eqIdx := strings.Index(selectorExprs[i], string(SelectorOpEqual))
		inIdx := strings.Index(selectorExprs[i], fmt.Sprintf(" %s ", SelectorOpIn))

		s := Selector{}
		opIdx, opLen := 0, 0
		switch {
		case eqIdx > -1:
			opIdx, opLen = eqIdx, 1
			s.Op = SelectorOpEqual
		case inIdx > -1:
			opIdx, opLen = inIdx, 4
			s.Op = SelectorOpIn

		default:
			return nil, fmt.Errorf("selector string '%v' operator is not '=' or 'in'", selectorExprs[i])
		}

		key, val := getTrimKeyValue(selectorExprs[i], opIdx, opLen)
		s.Key = key
		if s.Op == SelectorOpEqual {
			s.Values = []string{val}
		} else {
			s.Values = strings.Split(val[1:len(val)-1], ",")
		}
		s.Values = strings.Split(val, ",")
		selectors = append(selectors, s)
	}
	return rs, nil
}
func scanAllSplits(s string) ([]int, error) {
	multipleValueStart := false
	var splitsIdx []int
	for i := range s {
		// if comma is between bracket, we should look it as value's delimiter
		if s[i] == '(' {
			multipleValueStart = true
		}
		if multipleValueStart && s[i] == ')' {
			multipleValueStart = false
		}

		if !multipleValueStart && s[i] == ',' {
			splitsIdx = append(splitsIdx, i)
		}
	}

	if multipleValueStart {
		return nil, fmt.Errorf("you have '(' in label selector but did'n finded ')'")
	}
	return splitsIdx, nil
}
func splitStringsWithIdx(s string, idx []int) []string {
	var rs []string
	if len(idx) == 0 {
		return append(rs, s)
	}
	lastIdx := -1
	for i := range idx {
		if lastIdx == -1 {
			rs = append(rs, s[:idx[i]])
		}
		if lastIdx > -1 {
			rs = append(rs, s[lastIdx+1:idx[i]])
		}
		lastIdx = idx[i]
	}
	rs = append(rs, s[idx[len(idx)-1]+1:])
	return rs
}
func getTrimKeyValue(str string, opIdx, opLen int) (key, val string) {
	if opIdx < 0 {
		return
	}
	key = strings.TrimSpace(str[0:opIdx])
	val = strings.TrimSpace(str[opIdx+opLen:])
	return
}

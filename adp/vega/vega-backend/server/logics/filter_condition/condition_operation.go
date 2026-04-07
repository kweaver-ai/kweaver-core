// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package filter_condition

import "vega-backend/interfaces"

const (
	OperationAnd = "and"
	OperationOr  = "or"

	OperationEqual       = "=="
	OperationEqual2      = "eq"
	OperationNotEqual    = "!="
	OperationNotEqual2   = "not_eq"
	OperationGt          = ">"
	OperationGt2         = "gt"
	OperationGte         = ">="
	OperationGte2        = "gte"
	OperationLt          = "<"
	OperationLt2         = "lt"
	OperationLte         = "<="
	OperationLte2        = "lte"
	OperationIn          = "in"
	OperationNotIn       = "not_in"
	OperationLike        = "like"
	OperationNotLike     = "not_like"
	OperationContain     = "contain"
	OperationNotContain  = "not_contain"
	OperationRange       = "range"
	OperationOutRange    = "out_range"
	OperationExist       = "exist"
	OperationNotExist    = "not_exist"
	OperationEmpty       = "empty"
	OperationNotEmpty    = "not_empty"
	OperationRegex       = "regex"
	OperationMatch       = "match"
	OperationMatchPhrase = "match_phrase"
	OperationPrefix      = "prefix"
	OperationNotPrefix   = "not_prefix"
	OperationNull        = "null"
	OperationNotNull     = "not_null"
	OperationTrue        = "true"
	OperationFalse       = "false"
	OperationBefore      = "before"
	OperationCurrent     = "current"
	OperationBetween     = "between"
	OperationKnnVector   = "knn_vector"
	OperationMultiMatch  = "multi_match"
)

var (
	OperationMap map[string]interfaces.FilterCondition
)

func init() {
	OperationMap = map[string]interfaces.FilterCondition{
		OperationAnd:         &AndCond{},
		OperationOr:          &OrCond{},
		OperationEqual:       &EqualCond{},
		OperationEqual2:      &EqualCond{},
		OperationNotEqual:    &NotEqualCond{},
		OperationNotEqual2:   &NotEqualCond{},
		OperationGt:          &GtCond{},
		OperationGt2:         &GtCond{},
		OperationGte:         &GteCond{},
		OperationGte2:        &GteCond{},
		OperationLt:          &LtCond{},
		OperationLt2:         &LtCond{},
		OperationLte:         &LteCond{},
		OperationLte2:        &LteCond{},
		OperationIn:          &InCond{},
		OperationNotIn:       &NotInCond{},
		OperationLike:        &LikeCond{},
		OperationNotLike:     &NotLikeCond{},
		OperationContain:     &ContainCond{},
		OperationNotContain:  &NotContainCond{},
		OperationRange:       &RangeCond{},
		OperationOutRange:    &OutRangeCond{},
		OperationExist:       &ExistCond{},
		OperationNotExist:    &NotExistCond{},
		OperationEmpty:       &EmptyCond{},
		OperationNotEmpty:    &NotEmptyCond{},
		OperationRegex:       &RegexCond{},
		OperationMatch:       &MatchCond{},
		OperationMatchPhrase: &MatchPhraseCond{},
		OperationPrefix:      &PrefixCond{},
		OperationNotPrefix:   &NotPrefixCond{},
		OperationNull:        &NullCond{},
		OperationNotNull:     &NotNullCond{},
		OperationTrue:        &TrueCond{},
		OperationFalse:       &FalseCond{},
		OperationBefore:      &BeforeCond{},
		OperationCurrent:     &CurrentCond{},
		OperationBetween:     &BetweenCond{},
		OperationKnnVector:   &KnnVectorCond{},
		OperationMultiMatch:  &MultiMatchCond{},
	}
}

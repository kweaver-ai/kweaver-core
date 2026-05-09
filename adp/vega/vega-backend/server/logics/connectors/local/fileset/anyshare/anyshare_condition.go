// Package anyshare implements the AnyShare fileset connector.
package anyshare

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

// FilterResult represents the result of processing a filter condition
type FilterResult struct {
	Keyword   string
	Custom    []map[string]interface{}
	Dimension []string
	Condition map[string]interface{}
	Model     string
}

// FieldUsageTracker tracks which fields have been used
type FieldUsageTracker struct {
	CustomFields    map[string]bool
	DimensionFields map[string]bool
	ConditionFields map[string]bool
}

// NewFieldUsageTracker creates a new field usage tracker
func NewFieldUsageTracker() *FieldUsageTracker {
	return &FieldUsageTracker{
		CustomFields:    map[string]bool{},
		DimensionFields: map[string]bool{},
		ConditionFields: map[string]bool{},
	}
}

// convertTimeValue converts a time value to milliseconds, keeping -1 unchanged
func convertTimeValue(v int64) int64 {
	if v != -1 {
		return v * 1000
	}
	return v
}

// toInt64 converts an interface{} value to int64
func toInt64(val interface{}) (int64, error) {
	switch v := val.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("value must be int64 or float64, got %T", val)
	}
}

// checkMatchAndMatchPhraseConflict recursively checks if there are both MatchCond and MatchPhraseCond in the condition tree
// and ensures that MatchCond or MatchPhraseCond appears at most once
func checkMatchAndMatchPhraseConflict(cond interfaces.FilterCondition) (bool, bool, error) {
	switch c := cond.(type) {
	case *filter_condition.AndCond:
		matchCount := 0
		matchPhraseCount := 0
		for _, subCond := range c.SubConds {
			subMatch, subMatchPhrase, err := checkMatchAndMatchPhraseConflict(subCond)
			if err != nil {
				return false, false, err
			}
			if subMatch && subMatchPhrase {
				return true, true, nil
			}
			if subMatch {
				matchCount++
				if matchCount > 1 {
					return false, false, fmt.Errorf("match condition can only appear once in the entire condition")
				}
			}
			if subMatchPhrase {
				matchPhraseCount++
				if matchPhraseCount > 1 {
					return false, false, fmt.Errorf("match_phrase condition can only appear once in the entire condition")
				}
			}
		}
		return matchCount > 0, matchPhraseCount > 0, nil
	case *filter_condition.OrCond:
		matchCount := 0
		matchPhraseCount := 0
		for _, subCond := range c.SubConds {
			subMatch, subMatchPhrase, err := checkMatchAndMatchPhraseConflict(subCond)
			if err != nil {
				return false, false, err
			}
			if subMatch && subMatchPhrase {
				return true, true, nil
			}
			if subMatch {
				matchCount++
				if matchCount > 1 {
					return false, false, fmt.Errorf("match condition can only appear once in the entire condition")
				}
			}
			if subMatchPhrase {
				matchPhraseCount++
				if matchPhraseCount > 1 {
					return false, false, fmt.Errorf("match_phrase condition can only appear once in the entire condition")
				}
			}
		}
		return matchCount > 0, matchPhraseCount > 0, nil
	case *filter_condition.MatchCond:
		return true, false, nil
	case *filter_condition.MatchPhraseCond:
		return false, true, nil
	default:
		return false, false, nil
	}
}

// ConvertFilterCondition converts a filter condition to FilterResult
func (c *AnyShareConnector) ConvertFilterCondition(ctx context.Context, condition interfaces.FilterCondition) (*FilterResult, error) {
	tracker := NewFieldUsageTracker()
	return c.convertFilterConditionWithTracker(ctx, condition, tracker)
}

// convertFilterConditionWithTracker converts a filter condition to FilterResult with field usage tracking
func (c *AnyShareConnector) convertFilterConditionWithTracker(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	switch condition.GetOperation() {
	case filter_condition.OperationAnd:
		return c.convertFilterConditionAnd(ctx, condition, tracker)
	case filter_condition.OperationOr:
		return c.convertFilterConditionOr(ctx, condition, tracker)
	default:
		return c.convertFilterConditionWithOpr(ctx, condition, tracker)
	}
}

// convertFilterConditionAnd processes AND condition
func (c *AnyShareConnector) convertFilterConditionAnd(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	condAnd, ok := condition.(*filter_condition.AndCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.AndCond")
	}

	// Check if both MatchCond and MatchPhraseCond are present in the entire condition tree
	hasMatch, hasMatchPhrase, err := checkMatchAndMatchPhraseConflict(condition)
	if err != nil {
		return nil, err
	}
	if hasMatch && hasMatchPhrase {
		return nil, fmt.Errorf("match and match_phrase cannot be used together in AND condition")
	}

	result := &FilterResult{
		Custom:    []map[string]interface{}{},
		Dimension: []string{},
		Condition: map[string]interface{}{},
	}

	// Build custom condition with and subconditions
	var baseCond map[string]interface{}
	andCustom := []map[string]interface{}{}

	for _, subCond := range condAnd.SubConds {
		subResult, err := c.convertFilterConditionWithTracker(ctx, subCond, tracker)
		if err != nil {
			return nil, err
		}

		// Merge keyword
		if subResult.Keyword != "" {
			result.Keyword = subResult.Keyword
		}
		// Merge model
		if subResult.Model != "" {
			result.Model = subResult.Model
		}

		// Merge dimension
		for _, d := range subResult.Dimension {
			found := false
			for _, existing := range result.Dimension {
				if existing == d {
					found = true
					break
				}
			}
			if !found {
				result.Dimension = append(result.Dimension, d)
			}
		}
		// Merge condition
		for k, v := range subResult.Condition {
			if result.Condition[k] == nil {
				result.Condition[k] = v
			} else {
				// Merge values if both are string slices
				if existingValues, ok := result.Condition[k].([]string); ok {
					if newValues, ok := v.([]string); ok {
						result.Condition[k] = append(existingValues, newValues...)
					}
				}
			}
		}
		// Build custom condition with and subconditions
		if len(subResult.Custom) > 0 {
			if baseCond == nil {
				baseCond = subResult.Custom[0]
			} else {
				andCustom = append(andCustom, subResult.Custom...)
			}
		}
	}

	if baseCond != nil {
		if len(andCustom) > 0 {
			baseCond["and"] = andCustom
		}
		result.Custom = append(result.Custom, baseCond)
	}

	return result, nil
}

// convertFilterConditionOr processes OR condition
func (c *AnyShareConnector) convertFilterConditionOr(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	condOr, ok := condition.(*filter_condition.OrCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OrCond")
	}

	if len(condOr.SubConds) == 0 {
		return &FilterResult{}, nil
	}

	// Check if all subconditions are valid
	for _, subCond := range condOr.SubConds {
		if err := c.validateCustomCondition(subCond, tracker); err != nil {
			return nil, err
		}
	}

	// Build custom condition with or subconditions
	var baseCond map[string]interface{}
	orCustom := []map[string]interface{}{}
	result := &FilterResult{
		Custom: []map[string]interface{}{},
	}

	for _, subCond := range condOr.SubConds {
		subResult, err := c.convertFilterConditionWithTracker(ctx, subCond, tracker)
		if err != nil {
			return nil, err
		}

		// Build custom condition with or subconditions
		if len(subResult.Custom) > 0 {
			if baseCond == nil {
				baseCond = subResult.Custom[0]
			} else {
				orCustom = append(orCustom, subResult.Custom...)
			}
		}
	}

	if baseCond != nil {
		if len(orCustom) > 0 {
			baseCond["or"] = orCustom
		}
		result.Custom = append(result.Custom, baseCond)
	}

	return result, nil
}

// convertFilterConditionMatch processes MATCH condition
func (c *AnyShareConnector) convertFilterConditionMatch(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.MatchCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.MatchCond")
	}

	// Define allowed dimension fields
	dimensionFields := map[string]bool{
		"basename": true,
		"content":  true,
		"summary":  true,
	}

	// Extract keyword
	keyword := ""
	if keywordStr, ok := cond.Cfg.Value.(string); ok {
		keyword = keywordStr
	}

	// Validate fields and build dimension
	dimension := []string{}

	// Validate against dimensionFields
	for _, field := range cond.Fields {
		if !dimensionFields[field.Name] {
			supportedFields := make([]string, 0, len(dimensionFields))
			for fieldName := range dimensionFields {
				supportedFields = append(supportedFields, fieldName)
			}
			return nil, fmt.Errorf("field [%s] is not supported in [%s] condition, supported fields are: %v", field.Name, cond.GetOperation(), supportedFields)
		}
		dimension = append(dimension, field.Name)
	}

	return &FilterResult{
		Keyword:   keyword,
		Model:     "token",
		Dimension: dimension,
	}, nil
}

// convertFilterConditionMatchPhrase processes MATCH_PHRASE condition
func (c *AnyShareConnector) convertFilterConditionMatchPhrase(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.MatchPhraseCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.MatchPhraseCond")
	}

	// Define allowed dimension fields
	dimensionFields := map[string]bool{
		"basename": true,
		"content":  true,
		"summary":  true,
	}

	// Extract keyword
	keyword := ""
	if keywordStr, ok := cond.Cfg.Value.(string); ok {
		keyword = keywordStr
	}

	// Validate fields and build dimension
	dimension := []string{}
	for _, field := range cond.Fields {
		if !dimensionFields[field.Name] {
			supportedFields := make([]string, 0, len(dimensionFields))
			for fieldName := range dimensionFields {
				supportedFields = append(supportedFields, fieldName)
			}
			return nil, fmt.Errorf("field [%s] is not supported in [%s] condition, supported fields are: %v", field.Name, cond.GetOperation(), supportedFields)
		}
		dimension = append(dimension, field.Name)
	}

	return &FilterResult{
		Keyword:   keyword,
		Model:     "phrase",
		Dimension: dimension,
	}, nil
}

// convertFilterConditionWithOpr processes conditions with operators (eq, gte, gt, lte, lt, in, range, between, match, match_phrase)
func (c *AnyShareConnector) convertFilterConditionWithOpr(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	switch condition.GetOperation() {
	case filter_condition.OperationMatch:
		return c.convertFilterConditionMatch(ctx, condition, tracker)
	case filter_condition.OperationMatchPhrase:
		return c.convertFilterConditionMatchPhrase(ctx, condition, tracker)
	case filter_condition.OperationEqual, filter_condition.OperationEqual2:
		return c.convertFilterConditionEqual(ctx, condition, tracker)
	case filter_condition.OperationGte, filter_condition.OperationGte2:
		return c.convertFilterConditionGte(ctx, condition, tracker)
	case filter_condition.OperationGt, filter_condition.OperationGt2:
		return c.convertFilterConditionGt(ctx, condition, tracker)
	case filter_condition.OperationLte, filter_condition.OperationLte2:
		return c.convertFilterConditionLte(ctx, condition, tracker)
	case filter_condition.OperationLt, filter_condition.OperationLt2:
		return c.convertFilterConditionLt(ctx, condition, tracker)
	case filter_condition.OperationIn:
		return c.convertFilterConditionIn(ctx, condition, tracker)
	case filter_condition.OperationNotIn:
		return c.convertFilterConditionNotIn(ctx, condition, tracker)
	case filter_condition.OperationRange:
		return c.convertFilterConditionRange(ctx, condition, tracker)
	case filter_condition.OperationBetween:
		return c.convertFilterConditionBetween(ctx, condition, tracker)
	default:
		return nil, fmt.Errorf("operation %s is not supported", condition.GetOperation())
	}
}

// validateCustomCondition validates if a condition is a custom condition
// and tracks field usage across all nested levels
func (c *AnyShareConnector) validateCustomCondition(condition interfaces.FilterCondition, tracker *FieldUsageTracker) error {
	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
		"tags":        true,
	}

	switch cond := condition.(type) {
	case *filter_condition.EqualCond:
		if !customFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s is not supported in [%s] condition for [or] condition", cond.Lfield.Name, cond.GetOperation())
		}
		if tracker.CustomFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s can only appear once in the condition", cond.Lfield.Name)
		}
	case *filter_condition.GteCond:
		if !customFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s is not supported in [%s] condition for [or] condition", cond.Lfield.Name, cond.GetOperation())
		}
		if tracker.CustomFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s can only appear once in the condition", cond.Lfield.Name)
		}
	case *filter_condition.GtCond:
		if !customFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s is not supported in [%s] condition for [or] condition", cond.Lfield.Name, cond.GetOperation())
		}
		if tracker.CustomFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s can only appear once in the condition", cond.Lfield.Name)
		}
	case *filter_condition.LteCond:
		if !customFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s is not supported in [%s] condition for [or] condition", cond.Lfield.Name, cond.GetOperation())
		}
		if tracker.CustomFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s can only appear once in the condition", cond.Lfield.Name)
		}
	case *filter_condition.LtCond:
		if !customFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s is not supported in [%s] condition for [or] condition", cond.Lfield.Name, cond.GetOperation())
		}
		if tracker.CustomFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s can only appear once in the condition", cond.Lfield.Name)
		}
	case *filter_condition.InCond:
		if !customFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s is not supported in [%s] condition for [or] condition", cond.Lfield.Name, cond.GetOperation())
		}
		if tracker.CustomFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s can only appear once in the condition", cond.Lfield.Name)
		}
	case *filter_condition.BetweenCond:
		if !customFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s is not supported in [%s] condition for [or] condition", cond.Lfield.Name, cond.GetOperation())
		}
		if tracker.CustomFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s can only appear once in the condition", cond.Lfield.Name)
		}
	case *filter_condition.RangeCond:
		if !customFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s is not supported in [%s] condition for [or] condition", cond.Lfield.Name, cond.GetOperation())
		}
		if tracker.CustomFields[cond.Lfield.Name] {
			return fmt.Errorf("field %s can only appear once in the condition", cond.Lfield.Name)
		}
	case *filter_condition.OrCond:
		for _, subCond := range cond.SubConds {
			if err := c.validateCustomCondition(subCond, tracker); err != nil {
				return err
			}
		}
	case *filter_condition.AndCond:
		for _, subCond := range cond.SubConds {
			if err := c.validateCustomCondition(subCond, tracker); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("operation [%s] is not supported in [or] condition", condition.GetOperation())
	}

	return nil
}

// convertFilterConditionEqual processes EQUAL condition
func (c *AnyShareConnector) convertFilterConditionEqual(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.EqualCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EqualCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}

	// Define allowed fields
	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
		"tags":        true,
	}

	conditionFields := map[string]bool{
		"extension":   true,
		"created_by":  true,
		"modified_by": true,
	}

	// Check if field is allowed and not already used
	if customFields[cond.Lfield.Name] {
		if tracker.CustomFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("custom field %s can only appear once in condition", cond.Lfield.Name)
		}
		tracker.CustomFields[cond.Lfield.Name] = true
	} else if conditionFields[cond.Lfield.Name] {
		if tracker.ConditionFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("condition field %s can only appear once in condition", cond.Lfield.Name)
		}
		tracker.ConditionFields[cond.Lfield.Name] = true
	} else {
		return nil, fmt.Errorf("invalid field name: %s, allowed fields are: custom [created_at, modified_at, tags], dimension [basename, content, summary], condition [extension, created_by, modified_by]", cond.Lfield.Name)
	}

	result := &FilterResult{
		Custom:    []map[string]interface{}{},
		Condition: map[string]interface{}{},
	}

	switch cond.Lfield.Name {
	case "created_at", "modified_at":
		timeValue, err := toInt64(cond.Value)
		if err != nil {
			return nil, fmt.Errorf("time value error: %w", err)
		}
		result.Custom = append(result.Custom, map[string]interface{}{
			"key":   cond.Lfield.Name,
			"type":  "date",
			"mode":  "=",
			"value": []int64{convertTimeValue(timeValue)},
		})
	case "tags":
		if tagValue, ok := cond.Value.(string); ok {
			result.Custom = append(result.Custom, map[string]interface{}{
				"key":   "tags",
				"type":  "multiselect",
				"mode":  "=",
				"value": []string{tagValue},
			})
		}
	case "extension", "created_by", "modified_by":
		if strValue, ok := cond.Value.(string); ok {
			result.Condition[cond.Lfield.Name] = []string{strValue}
		}
	default:
		if !customFields[cond.Lfield.Name] && !conditionFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("invalid field name: %s, allowed fields are: custom [created_at, modified_at, tags], dimension [basename, content, summary], condition [extension, created_by, modified_by]", cond.Lfield.Name)
		}
	}

	return result, nil
}

// convertFilterConditionGte processes GTE condition
func (c *AnyShareConnector) convertFilterConditionGte(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.GteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GteCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}

	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
	}

	// Check if field is allowed and not already used
	if !customFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("invalid field name: %s for GTE condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
	}
	if tracker.CustomFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("custom field %s can only appear once in condition", cond.Lfield.Name)
	}
	tracker.CustomFields[cond.Lfield.Name] = true

	result := &FilterResult{
		Custom: []map[string]interface{}{},
	}

	switch cond.Lfield.Name {
	case "created_at", "modified_at":
		timeValue, err := toInt64(cond.Value)
		if err != nil {
			return nil, fmt.Errorf("time value error: %w", err)
		}
		result.Custom = append(result.Custom, map[string]interface{}{
			"key":   cond.Lfield.Name,
			"type":  "date",
			"mode":  ">=",
			"value": []int64{convertTimeValue(timeValue)},
		})
	default:
		if !customFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("invalid field name: %s for GTE condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
		}
	}

	return result, nil
}

// convertFilterConditionGt processes GT condition
func (c *AnyShareConnector) convertFilterConditionGt(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.GtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GtCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}

	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
	}

	// Check if field is allowed and not already used
	if !customFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("invalid field name: %s for GT condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
	}
	if tracker.CustomFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("custom field %s can only appear once in condition", cond.Lfield.Name)
	}
	tracker.CustomFields[cond.Lfield.Name] = true

	result := &FilterResult{
		Custom: []map[string]interface{}{},
	}

	switch cond.Lfield.Name {
	case "created_at", "modified_at":
		timeValue, err := toInt64(cond.Value)
		if err != nil {
			return nil, fmt.Errorf("time value error: %w", err)
		}
		result.Custom = append(result.Custom, map[string]interface{}{
			"key":   cond.Lfield.Name,
			"type":  "date",
			"mode":  ">",
			"value": []int64{convertTimeValue(timeValue)},
		})
	default:
		if !customFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("invalid field name: %s for GT condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
		}
	}

	return result, nil
}

// convertFilterConditionLte processes LTE condition
func (c *AnyShareConnector) convertFilterConditionLte(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.LteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LteCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}

	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
	}

	// Check if field is allowed and not already used
	if !customFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("invalid field name: %s for LTE condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
	}
	if tracker.CustomFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("custom field %s can only appear once in condition", cond.Lfield.Name)
	}
	tracker.CustomFields[cond.Lfield.Name] = true

	result := &FilterResult{
		Custom: []map[string]interface{}{},
	}

	switch cond.Lfield.Name {
	case "created_at", "modified_at":
		timeValue, err := toInt64(cond.Value)
		if err != nil {
			return nil, fmt.Errorf("time value error: %w", err)
		}
		result.Custom = append(result.Custom, map[string]interface{}{
			"key":   cond.Lfield.Name,
			"type":  "date",
			"mode":  "<=",
			"value": []int64{convertTimeValue(timeValue)},
		})
	default:
		if !customFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("invalid field name: %s for LTE condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
		}
	}

	return result, nil
}

// convertFilterConditionLt processes LT condition
func (c *AnyShareConnector) convertFilterConditionLt(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.LtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LtCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}

	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
	}

	// Check if field is allowed and not already used
	if !customFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("invalid field name: %s for LT condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
	}
	if tracker.CustomFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("custom field %s can only appear once in condition", cond.Lfield.Name)
	}
	tracker.CustomFields[cond.Lfield.Name] = true

	result := &FilterResult{
		Custom: []map[string]interface{}{},
	}

	switch cond.Lfield.Name {
	case "created_at", "modified_at":
		timeValue, err := toInt64(cond.Value)
		if err != nil {
			return nil, fmt.Errorf("time value error: %w", err)
		}
		result.Custom = append(result.Custom, map[string]interface{}{
			"key":   cond.Lfield.Name,
			"type":  "date",
			"mode":  "<",
			"value": []int64{convertTimeValue(timeValue)},
		})
	default:
		if !customFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("invalid field name: %s for LT condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
		}
	}

	return result, nil
}

// convertFilterConditionIn processes IN condition
func (c *AnyShareConnector) convertFilterConditionIn(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.InCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.InCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}

	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
		"tags":        true,
	}

	conditionFields := map[string]bool{
		"extension":   true,
		"created_by":  true,
		"modified_by": true,
	}

	// Check if field is allowed and not already used
	if customFields[cond.Lfield.Name] {
		if tracker.CustomFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("custom field %s can only appear once in condition", cond.Lfield.Name)
		}
		tracker.CustomFields[cond.Lfield.Name] = true
	} else if conditionFields[cond.Lfield.Name] {
		if tracker.ConditionFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("condition field %s can only appear once in condition", cond.Lfield.Name)
		}
		tracker.ConditionFields[cond.Lfield.Name] = true
	} else {
		return nil, fmt.Errorf("invalid field name: %s for IN condition", cond.Lfield.Name)
	}

	result := &FilterResult{
		Custom:    []map[string]interface{}{},
		Condition: map[string]interface{}{},
	}

	switch cond.Lfield.Name {
	case "created_at", "modified_at":
		values := []int64{}
		for _, val := range cond.Value {
			v, err := toInt64(val)
			if err != nil {
				return nil, fmt.Errorf("time value error: %w", err)
			}
			values = append(values, convertTimeValue(v))
		}
		if len(values) > 0 {
			result.Custom = append(result.Custom, map[string]interface{}{
				"key":   cond.Lfield.Name,
				"type":  "date",
				"mode":  "in",
				"value": values,
			})
		}
	case "tags":
		values := []string{}
		for _, val := range cond.Value {
			if tagValue, ok := val.(string); ok {
				values = append(values, tagValue)
			}
		}
		if len(values) > 0 {
			result.Custom = append(result.Custom, map[string]interface{}{
				"key":   "tags",
				"type":  "multiselect",
				"mode":  "in",
				"value": values,
			})
		}
	case "extension", "created_by", "modified_by":
		values := []string{}
		for _, val := range cond.Value {
			if strValue, ok := val.(string); ok {
				values = append(values, strValue)
			}
		}
		if len(values) > 0 {
			result.Condition[cond.Lfield.Name] = values
		}
	default:
		if !customFields[cond.Lfield.Name] && !conditionFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("invalid field name: %s for IN condition", cond.Lfield.Name)
		}
	}

	return result, nil
}

// convertFilterConditionNotIn processes NOT IN condition
func (c *AnyShareConnector) convertFilterConditionNotIn(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.NotInCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotInCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}

	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
		"tags":        true,
	}

	// Check if field is allowed and not already used
	if !customFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("invalid field name: %s for NOT IN condition, allowed fields are: created_at, modified_at, tags", cond.Lfield.Name)
	}
	if tracker.CustomFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("custom field %s can only appear once in condition", cond.Lfield.Name)
	}
	tracker.CustomFields[cond.Lfield.Name] = true

	result := &FilterResult{
		Custom:    []map[string]interface{}{},
		Condition: map[string]interface{}{},
	}

	switch cond.Lfield.Name {
	case "created_at", "modified_at":
		values := []int64{}
		for _, val := range cond.Value {
			v, err := toInt64(val)
			if err != nil {
				return nil, fmt.Errorf("time value error: %w", err)
			}
			values = append(values, convertTimeValue(v))
		}
		if len(values) > 0 {
			result.Custom = append(result.Custom, map[string]interface{}{
				"key":   cond.Lfield.Name,
				"type":  "date",
				"mode":  "nin",
				"value": values,
			})
		}
	case "tags":
		values := []string{}
		for _, val := range cond.Value {
			if tagValue, ok := val.(string); ok {
				values = append(values, tagValue)
			}
		}
		if len(values) > 0 {
			result.Custom = append(result.Custom, map[string]interface{}{
				"key":   "tags",
				"type":  "multiselect",
				"mode":  "nin",
				"value": values,
			})
		}
	default:
		return nil, fmt.Errorf("invalid field name: %s for NOT IN condition, allowed fields are: created_at, modified_at, tags", cond.Lfield.Name)
	}

	return result, nil
}

// convertFilterConditionRange processes RANGE condition
func (c *AnyShareConnector) convertFilterConditionRange(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.RangeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RangeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}

	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
	}

	// Check if field is allowed and not already used
	if !customFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("invalid field name: %s for RANGE condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
	}
	if tracker.CustomFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("custom field %s can only appear once in condition", cond.Lfield.Name)
	}
	tracker.CustomFields[cond.Lfield.Name] = true

	result := &FilterResult{
		Custom: []map[string]interface{}{},
	}

	switch cond.Lfield.Name {
	case "created_at", "modified_at":
		if len(cond.Value) == 2 {
			var min, max int64
			minOk := false
			maxOk := false
			if minVal, err := toInt64(cond.Value[0]); err == nil {
				min = convertTimeValue(minVal)
				minOk = true
			}
			if maxVal, err := toInt64(cond.Value[1]); err == nil {
				max = convertTimeValue(maxVal)
				maxOk = true
			}
			if minOk && maxOk {
				result.Custom = append(result.Custom, map[string]interface{}{
					"key":   cond.Lfield.Name,
					"type":  "date",
					"value": []int64{min, max},
				})
			}
		}
	default:
		if !customFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("invalid field name: %s for RANGE condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
		}
	}

	return result, nil
}

// convertFilterConditionBetween processes BETWEEN condition
func (c *AnyShareConnector) convertFilterConditionBetween(ctx context.Context, condition interfaces.FilterCondition, tracker *FieldUsageTracker) (*FilterResult, error) {
	cond, ok := condition.(*filter_condition.BetweenCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.BetweenCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}

	customFields := map[string]bool{
		"created_at":  true,
		"modified_at": true,
	}

	// Check if field is allowed and not already used
	if !customFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("invalid field name: %s for BETWEEN condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
	}
	if tracker.CustomFields[cond.Lfield.Name] {
		return nil, fmt.Errorf("custom field %s can only appear once in condition", cond.Lfield.Name)
	}
	tracker.CustomFields[cond.Lfield.Name] = true

	result := &FilterResult{
		Custom: []map[string]interface{}{},
	}

	switch cond.Lfield.Name {
	case "created_at", "modified_at":
		if len(cond.Value) == 2 {
			var min, max int64
			minOk := false
			maxOk := false
			if minVal, err := toInt64(cond.Value[0]); err == nil {
				min = convertTimeValue(minVal)
				minOk = true
			}
			if maxVal, err := toInt64(cond.Value[1]); err == nil {
				max = convertTimeValue(maxVal)
				maxOk = true
			}
			if minOk && maxOk {
				result.Custom = append(result.Custom, map[string]interface{}{
					"key":   cond.Lfield.Name,
					"type":  "date",
					"value": []int64{min, max},
				})
			}
		}
	default:
		if !customFields[cond.Lfield.Name] {
			return nil, fmt.Errorf("invalid field name: %s for BETWEEN condition, allowed fields are: created_at, modified_at", cond.Lfield.Name)
		}
	}

	return result, nil
}

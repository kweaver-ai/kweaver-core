// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"strings"
	"testing"

	"vega-backend/interfaces"
)

// ===== validateName =====

func TestValidateName_Valid(t *testing.T) {
	if err := validateName(context.Background(), "test-catalog"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateName_Empty(t *testing.T) {
	err := validateName(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestValidateName_MaxLength(t *testing.T) {
	name := strings.Repeat("a", interfaces.NAME_MAX_LENGTH)
	if err := validateName(context.Background(), name); err != nil {
		t.Fatalf("unexpected error for max length name: %v", err)
	}
}

func TestValidateName_ExceedsMaxLength(t *testing.T) {
	name := strings.Repeat("a", interfaces.NAME_MAX_LENGTH+1)
	err := validateName(context.Background(), name)
	if err == nil {
		t.Fatal("expected error for name exceeding max length")
	}
}

func TestValidateName_UTF8(t *testing.T) {
	// 中文字符，每个占3字节但算1个rune
	name := strings.Repeat("中", interfaces.NAME_MAX_LENGTH)
	if err := validateName(context.Background(), name); err != nil {
		t.Fatalf("unexpected error for UTF-8 name at max rune count: %v", err)
	}

	name = strings.Repeat("中", interfaces.NAME_MAX_LENGTH+1)
	err := validateName(context.Background(), name)
	if err == nil {
		t.Fatal("expected error for UTF-8 name exceeding max rune count")
	}
}

// ===== ValidateTags =====

func TestValidateTags_Valid(t *testing.T) {
	tags := []string{"tag1", "tag2"}
	if err := ValidateTags(context.Background(), tags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTags_Empty(t *testing.T) {
	if err := ValidateTags(context.Background(), []string{}); err != nil {
		t.Fatalf("unexpected error for empty tags: %v", err)
	}
}

func TestValidateTags_ExceedsMaxNumber(t *testing.T) {
	tags := make([]string, interfaces.TAGS_MAX_NUMBER+1)
	for i := range tags {
		tags[i] = "tag"
	}
	err := ValidateTags(context.Background(), tags)
	if err == nil {
		t.Fatal("expected error for exceeding max tag count")
	}
}

func TestValidateTags_InvalidTagInList(t *testing.T) {
	tags := []string{"good-tag", "bad/tag"}
	err := ValidateTags(context.Background(), tags)
	if err == nil {
		t.Fatal("expected error for invalid tag character")
	}
}

// ===== validateTag =====

func TestValidateTag_Valid(t *testing.T) {
	if err := validateTag(context.Background(), "my-tag"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTag_Empty(t *testing.T) {
	err := validateTag(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty tag")
	}
}

func TestValidateTag_OnlySpaces(t *testing.T) {
	err := validateTag(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only tag")
	}
}

func TestValidateTag_ExceedsMaxLength(t *testing.T) {
	tag := strings.Repeat("a", interfaces.TAG_MAX_LENGTH+1)
	err := validateTag(context.Background(), tag)
	if err == nil {
		t.Fatal("expected error for tag exceeding max length")
	}
}

func TestValidateTag_SpecialChars(t *testing.T) {
	invalidChars := []string{"/", ":", "?", "\\", "\"", "<", ">", "|", "#", "%", "&", "*", "$", "^", "!", "=", "."}
	for _, ch := range invalidChars {
		err := validateTag(context.Background(), "tag"+ch+"name")
		if err == nil {
			t.Errorf("expected error for tag containing '%s'", ch)
		}
	}
}

func TestValidateTag_TrimSpaces(t *testing.T) {
	// 两端空格应被 trim 后正常通过
	if err := validateTag(context.Background(), "  valid-tag  "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ===== validateDescription =====

func TestValidateDescription_Valid(t *testing.T) {
	if err := validateDescription(context.Background(), "A valid description"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateDescription_Empty(t *testing.T) {
	if err := validateDescription(context.Background(), ""); err != nil {
		t.Fatalf("unexpected error for empty description: %v", err)
	}
}

func TestValidateDescription_ExceedsMaxLength(t *testing.T) {
	desc := strings.Repeat("a", interfaces.DESCRIPTION_MAX_LENGTH+1)
	err := validateDescription(context.Background(), desc)
	if err == nil {
		t.Fatal("expected error for description exceeding max length")
	}
}

// ===== validatePaginationQueryParams =====

var testSortTypes = map[string]string{
	"name":        "f_name",
	"create_time": "f_create_time",
}

func TestPagination_Valid(t *testing.T) {
	params, err := validatePaginationQueryParams(context.Background(),
		"0", "10", "name", "asc", testSortTypes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Offset != 0 {
		t.Errorf("expected offset 0, got %d", params.Offset)
	}
	if params.Limit != 10 {
		t.Errorf("expected limit 10, got %d", params.Limit)
	}
	if params.Sort != "f_name" {
		t.Errorf("expected sort 'f_name', got '%s'", params.Sort)
	}
	if params.Direction != "asc" {
		t.Errorf("expected direction 'asc', got '%s'", params.Direction)
	}
}

func TestPagination_NoLimit(t *testing.T) {
	params, err := validatePaginationQueryParams(context.Background(),
		"0", "-1", "name", "desc", testSortTypes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Limit != -1 {
		t.Errorf("expected limit -1, got %d", params.Limit)
	}
}

func TestPagination_InvalidOffset(t *testing.T) {
	_, err := validatePaginationQueryParams(context.Background(),
		"abc", "10", "name", "asc", testSortTypes)
	if err == nil {
		t.Fatal("expected error for non-numeric offset")
	}
}

func TestPagination_NegativeOffset(t *testing.T) {
	_, err := validatePaginationQueryParams(context.Background(),
		"-1", "10", "name", "asc", testSortTypes)
	if err == nil {
		t.Fatal("expected error for negative offset")
	}
}

func TestPagination_InvalidLimit(t *testing.T) {
	_, err := validatePaginationQueryParams(context.Background(),
		"0", "abc", "name", "asc", testSortTypes)
	if err == nil {
		t.Fatal("expected error for non-numeric limit")
	}
}

func TestPagination_LimitTooSmall(t *testing.T) {
	_, err := validatePaginationQueryParams(context.Background(),
		"0", "0", "name", "asc", testSortTypes)
	if err == nil {
		t.Fatal("expected error for limit below minimum")
	}
}

func TestPagination_LimitTooLarge(t *testing.T) {
	_, err := validatePaginationQueryParams(context.Background(),
		"0", "1001", "name", "asc", testSortTypes)
	if err == nil {
		t.Fatal("expected error for limit exceeding maximum")
	}
}

func TestPagination_InvalidSort(t *testing.T) {
	_, err := validatePaginationQueryParams(context.Background(),
		"0", "10", "unknown_sort", "asc", testSortTypes)
	if err == nil {
		t.Fatal("expected error for unsupported sort type")
	}
}

func TestPagination_InvalidDirection(t *testing.T) {
	_, err := validatePaginationQueryParams(context.Background(),
		"0", "10", "name", "up", testSortTypes)
	if err == nil {
		t.Fatal("expected error for invalid sort direction")
	}
}
